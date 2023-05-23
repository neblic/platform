package digest

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/neblic/platform/controlplane/data"
	"github.com/neblic/platform/sampler/internal/sample"
	"github.com/neblic/platform/sampler/internal/sample/exporter"
)

// // Move to data package

const defaultDigestStMaxProcessedFields = 100
const defaultDigestFlushPeriod = time.Minute
const defaultDigestBufferSize = 1000

/////

type Settings struct {
	ResourceName string
	SamplerName  string

	NotifyErr func(error)
	Exporter  exporter.Exporter
}

type Digester struct {
	ctx          context.Context
	resourceName string
	samplerName  string

	notifyErr func(error)
	exporter  exporter.Exporter

	digestsConfig []data.Digest
	workers       map[data.SamplerDigestUID]*worker
}

func NewDigester(ctx context.Context, settings Settings) *Digester {
	return &Digester{
		ctx:          ctx,
		resourceName: settings.ResourceName,
		samplerName:  settings.SamplerName,

		notifyErr: settings.NotifyErr,
		exporter:  settings.Exporter,

		workers: make(map[data.SamplerDigestUID]*worker),
	}
}

func (d *Digester) buildWorkerSettings(digestCfg data.Digest) (workerSettings, error) {
	var digest Digest
	switch digestCfg.Type {
	case data.DigestTypeSt:
		digest = NewStDigest(digestCfg.St.MaxProcessedFields, d.notifyErr)
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

func (d *Digester) SetDigestsConfig(digestsCfg []data.Digest) {
	for _, digestCfg := range digestsCfg {
		if existingWorker, ok := d.workers[digestCfg.UID]; ok {
			existingWorker.stop()
			delete(d.workers, digestCfg.UID)
		}

		newWorkerSettings, err := d.buildWorkerSettings(digestCfg)
		if err != nil {
			d.notifyErr(errors.New("unknown digest type"))
			continue
		}

		w := newWorker(d.ctx, newWorkerSettings)
		d.workers[digestCfg.UID] = w

		go w.run()
	}

	digestCfgsMap := make(map[data.SamplerDigestUID]data.Digest)
	for _, digestCfg := range digestsCfg {
		digestCfgsMap[digestCfg.UID] = digestCfg
	}

	for uid, existingWorker := range d.workers {
		if _, ok := digestCfgsMap[uid]; !ok {
			existingWorker.stop()
			delete(d.workers, uid)
		}
	}

	d.digestsConfig = digestsCfg
}

func (d *Digester) ProcessSample(streams []data.SamplerStreamUID, sampleData *sample.Data) {
	for _, stream := range streams {
		for _, worker := range d.workers {
			if worker.streamUID == stream {
				worker.processSample(sampleData)
			}
		}
	}
}

type workerSettings struct {
	streamUID    data.SamplerStreamUID
	resourceName string
	samplerName  string

	samplesToFlush int
	inChBufferSize int
	flushPeriod    time.Duration
	digest         Digest
	exporter       exporter.Exporter

	notifyErr func(error)
}

type worker struct {
	ctx             context.Context
	processSampleCh chan *sample.Data

	workerSettings
}

func newWorker(ctx context.Context, settings workerSettings) *worker {
	return &worker{
		ctx:             ctx,
		workerSettings:  settings,
		processSampleCh: make(chan *sample.Data, settings.inChBufferSize),
	}
}

func (w *worker) String() string {
	return fmt.Sprintf("worker(StreamUID: %s, Digest: %s)", w.streamUID, w.digest)
}

func (w *worker) processSample(sampleData *sample.Data) {
	select {
	case w.processSampleCh <- sampleData:
		w.samplesToFlush += 1
	default:
		w.notifyErr(fmt.Errorf("%s buffer is full", w))
	}
}

func (w *worker) run() {
	ticker := time.NewTicker(w.flushPeriod)
loop:
	for {
		select {
		case <-w.ctx.Done():
			break loop
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

func (w *worker) buildDigestSample(digestData []byte) exporter.SamplerSamples {
	return exporter.SamplerSamples{
		ResourceName: w.resourceName,
		SamplerName:  w.samplerName,
		Samples: []exporter.Sample{{
			Ts:       time.Now(),
			Type:     exporter.StructDigestSampleType,
			Streams:  []data.SamplerStreamUID{w.streamUID},
			Encoding: exporter.JSONSampleEncoding,
			Data:     digestData,
		}},
	}
}

func (w *worker) exportDigest() {
	if w.samplesToFlush <= 0 {
		return
	}

	digestData, err := w.digest.JSON()
	if err != nil {
		w.notifyErr(err)
	}

	smpl := w.buildDigestSample(digestData)
	err = w.exporter.Export(w.ctx, []exporter.SamplerSamples{smpl})
	if err != nil {
		w.notifyErr(err)
	}

	w.digest.Reset()
	w.samplesToFlush = 0
}

func (w *worker) stop() {
	close(w.processSampleCh)
}
