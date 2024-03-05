package digest

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/neblic/platform/controlplane/control"
	dpsample "github.com/neblic/platform/dataplane/sample"
	"github.com/neblic/platform/internal/pkg/data"
	"github.com/neblic/platform/internal/pkg/exporter"
	"github.com/neblic/platform/logging"
)

// // Move to data package

const defaultDigestStMaxProcessedFields = 100
const defaultDigestFlushPeriod = time.Minute
const defaultDigestBufferSize = 1000

/////

type Settings struct {
	ResourceName        string
	SamplerName         string
	ComputationLocation control.ComputationLocation

	NotifyErr func(error)
	Exporter  exporter.LogsExporter
	Logger    logging.Logger
}

type Digester struct {
	resourceName        string
	samplerName         string
	computationLocation control.ComputationLocation

	notifyErr func(error)
	exporter  exporter.LogsExporter
	logger    logging.Logger

	digestsConfig map[control.SamplerDigestUID]control.Digest
	workers       map[control.SamplerDigestUID]*worker
	sync          bool
}

func NewDigester(settings Settings) *Digester {
	return &Digester{
		resourceName:        settings.ResourceName,
		samplerName:         settings.SamplerName,
		computationLocation: settings.ComputationLocation,
		notifyErr:           settings.NotifyErr,
		exporter:            settings.Exporter,
		logger:              settings.Logger,

		workers: make(map[control.SamplerDigestUID]*worker),
	}
}

func (d *Digester) buildWorkerSettings(digestCfg control.Digest) (workerSettings, error) {
	var digest Digest
	switch digestCfg.Type {
	case control.DigestTypeSt:
		digest = NewStDigest(digestCfg.St.MaxProcessedFields, d.notifyErr)
	case control.DigestTypeValue:
		digest = NewValue(digestCfg.Value.MaxProcessedFields)
	default:
		return workerSettings{}, errors.New("unknown digest type")
	}

	// set default values if unset or incorrect
	flushPeriod := digestCfg.FlushPeriod
	if digestCfg.FlushPeriod <= 0 {
		flushPeriod = defaultDigestFlushPeriod
	}

	bufferSize := digestCfg.BufferSize
	if digestCfg.BufferSize <= 0 {
		bufferSize = defaultDigestBufferSize
	}

	return workerSettings{
		digestUID:    digestCfg.UID,
		streamUID:    digestCfg.StreamUID,
		resourceName: d.resourceName,
		samplerName:  d.samplerName,

		digest:         digest,
		flushPeriod:    flushPeriod,
		inChBufferSize: bufferSize,
		exporter:       d.exporter,

		notifyErr: d.notifyErr,
	}, nil
}

func (d *Digester) SetDigestsConfig(digestCfgs map[control.SamplerDigestUID]control.Digest) {
	for _, digestCfg := range digestCfgs {
		if digestCfg.ComputationLocation != d.computationLocation {
			d.logger.Debug("Skipping digest worker", "config", digestCfg, "reason", "computation location mismatch")
			continue
		}

		if existingWorker, ok := d.workers[digestCfg.UID]; ok {
			existingWorker.stop()
			delete(d.workers, digestCfg.UID)
		}

		newWorkerSettings, err := d.buildWorkerSettings(digestCfg)
		if err != nil {
			d.notifyErr(errors.New("unknown digest type"))
			continue
		}

		w := newWorker(newWorkerSettings)
		d.workers[digestCfg.UID] = w

		d.logger.Debug("Starting digest worker", "config", digestCfg)
		go w.run()
	}

	for uid, existingWorker := range d.workers {
		if _, ok := digestCfgs[uid]; !ok {
			existingWorker.stop()
			delete(d.workers, uid)
		}
	}

	d.digestsConfig = digestCfgs
}

func (d *Digester) SetSync(sync bool) {
	d.sync = sync
}

func (d *Digester) ProcessSample(streams []control.SamplerStreamUID, sampleData *data.Data) bool {
	processed := false
	for _, stream := range streams {
		for _, worker := range d.workers {
			if worker.streamUID == stream {
				if d.sync {
					worker.processSampleSync(sampleData)
				} else {
					worker.processSample(sampleData)
				}
				processed = true
			}
		}
	}

	return processed
}

func (d *Digester) Close() error {
	for _, worker := range d.workers {
		worker.stop()
	}

	return nil
}

type workerSettings struct {
	digestUID    control.SamplerDigestUID
	streamUID    control.SamplerStreamUID
	resourceName string
	samplerName  string

	samplesToFlush int
	inChBufferSize int
	flushPeriod    time.Duration
	digest         Digest
	exporter       exporter.LogsExporter

	notifyErr func(error)
}

type worker struct {
	processSampleCh chan *data.Data

	workerSettings
}

func newWorker(settings workerSettings) *worker {
	return &worker{
		workerSettings:  settings,
		processSampleCh: make(chan *data.Data, settings.inChBufferSize),
	}
}

func (w *worker) String() string {
	return fmt.Sprintf("worker(StreamUID: %s, Digest: %s)", w.streamUID, w.digest)
}

func (w *worker) processSample(sampleData *data.Data) {
	select {
	case w.processSampleCh <- sampleData:
		w.samplesToFlush++
	default:
		w.notifyErr(fmt.Errorf("%s buffer is full", w))
	}
}

func (w *worker) processSampleSync(sampleData *data.Data) {
	if err := w.digest.AddSampleData(sampleData); err != nil {
		w.notifyErr(err)
	}
}

func (w *worker) run() {
	ticker := time.NewTicker(w.flushPeriod)
loop:
	for {
		select {
		case sampleData, more := <-w.processSampleCh:
			if !more {
				break loop
			}

			if err := w.digest.AddSampleData(sampleData); err != nil {
				w.notifyErr(err)
			}
		case <-ticker.C:
			w.exportDigest()
		}
	}

	w.exportDigest()
	ticker.Stop()
}

func (w *worker) buildDigestSample(digestData []byte) dpsample.OTLPLogs {
	otlpLogs := dpsample.NewOTLPLogs()
	samplerOtlpLogs := otlpLogs.AppendSamplerOTLPLogs(w.resourceName, w.samplerName)

	switch w.digest.SampleType() {
	case control.StructDigestSampleType:
		digestOtlpLog := samplerOtlpLogs.AppendStructDigestOTLPLog()
		digestOtlpLog.SetUID(w.digestUID)
		digestOtlpLog.SetTimestamp(time.Now())
		digestOtlpLog.SetStreamUIDs([]control.SamplerStreamUID{w.streamUID})
		digestOtlpLog.SetSampleRawData(dpsample.JSONEncoding, digestData)
	case control.ValueDigestSampleType:
		digestOtlpLog := samplerOtlpLogs.AppendValueDigestOTLPLog()
		digestOtlpLog.SetUID(w.digestUID)
		digestOtlpLog.SetTimestamp(time.Now())
		digestOtlpLog.SetStreamUIDs([]control.SamplerStreamUID{w.streamUID})
		digestOtlpLog.SetSampleRawData(dpsample.JSONEncoding, digestData)
	default:
		panic(fmt.Errorf("unknown digest sample type %s", w.digest.SampleType()))
	}

	return otlpLogs
}

func (w *worker) exportDigest() {
	if w.samplesToFlush <= 0 {
		return
	}

	digestData, err := w.digest.JSON()
	if err != nil {
		w.notifyErr(err)
	}

	otlpLogs := w.buildDigestSample(digestData)
	err = w.exporter.Export(context.Background(), otlpLogs)
	if err != nil {
		w.notifyErr(err)
	}

	w.digest.Reset()
	w.samplesToFlush = 0
}

func (w *worker) stop() {
	close(w.processSampleCh)

	//TODO: block until last digest have been sent
}
