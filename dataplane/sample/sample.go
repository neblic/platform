package sample

import (
	"time"

	"github.com/neblic/platform/controlplane/data"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"

	conventions "go.opentelemetry.io/collector/semconv/v1.9.0"
)

const (
	rlSamplerNameKey       = "sampler_name"
	lrSampleTypeKey        = "sample_type"
	lrSampleEncodingKey    = "sample_encoding"
	lrSampleStreamsUIDsKey = "stream_uids"
)

func OTLPLogsToSamples(logs plog.Logs) []SamplerSamples {
	var samplersSamples []SamplerSamples

	for i := 0; i < logs.ResourceLogs().Len(); i++ {
		samplerSamples := SamplerSamples{}
		rLog := logs.ResourceLogs().At(i)

		if samplerName, ok := rLog.Resource().Attributes().Get(rlSamplerNameKey); ok {
			samplerSamples.SamplerName = samplerName.Str()
		}

		if resourceName, ok := rLog.Resource().Attributes().Get(conventions.AttributeServiceName); ok {
			samplerSamples.ResourceName = resourceName.Str()
		}

		if rLog.ScopeLogs().Len() != 1 {
			continue
		}

		slogs := rLog.ScopeLogs().At(0)
		for j := 0; j < slogs.LogRecords().Len(); j++ {
			sample := Sample{}
			logRecord := slogs.LogRecords().At(j)

			sample.Ts = logRecord.Timestamp().AsTime()
			if sampleType, ok := logRecord.Attributes().Get(lrSampleTypeKey); ok {
				sample.Type = ParseSampleType(sampleType.Str())
			}
			if sampleEncoding, ok := logRecord.Attributes().Get(lrSampleEncodingKey); ok {
				sample.Encoding = ParseSampleEncoding(sampleEncoding.Str())
			}

			if sample.Encoding == JSONSampleEncoding {
				sample.Data = []byte(logRecord.Body().Str())
			} else {
				sample.Data = logRecord.Body().Bytes().AsRaw()
			}

			lrStreamUIDs, ok := logRecord.Attributes().Get(lrSampleStreamsUIDsKey)
			if ok {
				var streamUIDs []data.SamplerStreamUID
				for k := 0; k < lrStreamUIDs.Slice().Len(); k++ {
					lrStreamUID := lrStreamUIDs.Slice().At(k)
					streamUIDs = append(streamUIDs, data.SamplerStreamUID(lrStreamUID.Str()))
				}
				sample.Streams = streamUIDs
			}

			samplerSamples.Samples = append(samplerSamples.Samples, sample)
		}

		samplersSamples = append(samplersSamples, samplerSamples)
	}

	return samplersSamples
}

// TODO: in general, the produced log could be more compact but this way, we keep it human-readable
// eventually, we will want to binary encode as much as possible and create a human-readable
// representation when necessary
func SamplesToOTLPLogs(resourceSmpls []SamplerSamples) plog.Logs {
	logs := plog.NewLogs()

	rlogs := logs.ResourceLogs()
	rlogs.EnsureCapacity(len(resourceSmpls))
	for i := 0; i < len(resourceSmpls); i++ {
		resourceSmpl := resourceSmpls[i]

		rlog := rlogs.AppendEmpty()
		rlog.Resource().Attributes().PutStr(rlSamplerNameKey, resourceSmpl.SamplerName)
		rlog.Resource().Attributes().PutStr(conventions.AttributeServiceName, resourceSmpl.ResourceName)

		// we consider each sampler name to be an scope
		// each samples batch belongs to one sampler, so we only have one scope
		slogs := rlog.ScopeLogs()
		slog := slogs.AppendEmpty()

		// TODO: it could also make sense to add the sampler name as a scope attribute
		// slog.Scope().Attributes().PutStr(rlSamplerNameKey, resourceSmpl.SamplerName)

		logRecords := slog.LogRecords()
		logRecords.EnsureCapacity(len(resourceSmpl.Samples))
		for j := 0; j < len(resourceSmpl.Samples); j++ {
			smpl := resourceSmpl.Samples[j]

			// we consider each sample to be a LogRecord
			logRecord := logRecords.AppendEmpty()
			logRecord.SetTimestamp(pcommon.Timestamp(smpl.Ts.UTC().UnixNano()))
			logRecord.Attributes().PutStr(lrSampleTypeKey, smpl.Type.String())
			logRecord.Attributes().PutStr(lrSampleEncodingKey, smpl.Encoding.String())

			if smpl.Encoding == JSONSampleEncoding {
				logRecord.Body().SetStr(string(smpl.Data))
			} else {
				logRecord.Body().SetEmptyBytes().FromRaw(smpl.Data)
			}

			// build attributes values
			lrStreamUIDs := logRecord.Attributes().PutEmptySlice(lrSampleStreamsUIDsKey)
			lrStreamUIDs.EnsureCapacity(len(smpl.Streams))
			for _, stream := range smpl.Streams {
				lrStreamUIDs.AppendEmpty().SetStr(string(stream))
			}
		}
	}

	return logs
}

type SampleType uint8

const (
	UnknownSampleType SampleType = iota
	RawSampleType
	StructDigestSampleType
)

type SampleEncoding uint8

func (s SampleType) String() string {
	switch s {
	case UnknownSampleType:
		return "unknown"
	case RawSampleType:
		return "raw"
	case StructDigestSampleType:
		return "struct-digest"
	default:
		return "unknown"
	}
}

func ParseSampleType(t string) SampleType {
	switch t {
	case "raw":
		return RawSampleType
	case "struct-digest":
		return StructDigestSampleType
	default:
		return UnknownSampleType
	}
}

const (
	UnknownSampleEncoding SampleEncoding = iota
	JSONSampleEncoding
)

func (s SampleEncoding) String() string {
	switch s {
	case UnknownSampleEncoding:
		return "unknown"
	case JSONSampleEncoding:
		return "json"
	default:
		return "unknown"
	}
}

func ParseSampleEncoding(enc string) SampleEncoding {
	switch enc {
	case "json":
		return JSONSampleEncoding
	default:
		return UnknownSampleEncoding
	}
}

// Sample defines a sample to be exported
type Sample struct {
	Ts       time.Time
	Type     SampleType
	Streams  []data.SamplerStreamUID
	Encoding SampleEncoding
	Data     []byte
}

// SamplerSamples contains a group of samples that originate from the same resource e.g. operator, service...
type SamplerSamples struct {
	ResourceName string
	SamplerName  string
	Samples      []Sample
}
