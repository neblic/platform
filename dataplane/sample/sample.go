package sample

import (
	"time"

	"github.com/neblic/platform/controlplane/control"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"

	conventions "go.opentelemetry.io/collector/semconv/v1.9.0"
)

const (
	// resource attributes
	rlSamplerNameKey = "sampler_name"
	// sample attributes
	lrSampleStreamsUIDsKey = "stream_uids"
	lrSampleKey            = "sample_key"
	lrSampleTypeKey        = "sample_type"
	lrSampleEncodingKey    = "sample_encoding"
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
			logRecord.Attributes().Range(func(k string, v pcommon.Value) bool {
				switch k {
				case lrSampleStreamsUIDsKey:
					var streamUIDs []control.SamplerStreamUID
					for k := 0; k < v.Slice().Len(); k++ {
						lrStreamUID := v.Slice().At(k)
						streamUIDs = append(streamUIDs, control.SamplerStreamUID(lrStreamUID.Str()))
					}
					sample.Streams = streamUIDs
				case lrSampleKey:
					sample.Key = v.Str()
				case lrSampleTypeKey:
					sample.Type = control.ParseSampleType(v.Str())
				case lrSampleEncodingKey:
					sample.Encoding = ParseSampleEncoding(v.Str())
				default:
					// any other attribute is considered metadata
					if sample.Metadata == nil {
						sample.Metadata = make(map[MetadataKey]string)
					}
					sample.Metadata[MetadataKey(k)] = v.Str()
				}

				return true
			})

			if sample.Encoding == JSONEncoding {
				sample.Data = []byte(logRecord.Body().Str())
			} else {
				sample.Data = logRecord.Body().Bytes().AsRaw()
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

			// set known attributes
			logRecord.Attributes().PutStr(lrSampleTypeKey, smpl.Type.String())
			logRecord.Attributes().PutStr(lrSampleEncodingKey, smpl.Encoding.String())
			logRecord.Attributes().PutStr(lrSampleKey, smpl.Key)
			lrStreamUIDs := logRecord.Attributes().PutEmptySlice(lrSampleStreamsUIDsKey)
			lrStreamUIDs.EnsureCapacity(len(smpl.Streams))
			for _, stream := range smpl.Streams {
				lrStreamUIDs.AppendEmpty().SetStr(string(stream))
			}

			// set body
			if smpl.Encoding == JSONEncoding {
				logRecord.Body().SetStr(string(smpl.Data))
			} else {
				logRecord.Body().SetEmptyBytes().FromRaw(smpl.Data)
			}

			// set metadata as other attributes
			for k, v := range smpl.Metadata {
				logRecord.Attributes().PutStr(string(k), v)
			}
		}
	}

	return logs
}

type Encoding uint8

const (
	UnknownEncoding Encoding = iota
	JSONEncoding
)

func (s Encoding) String() string {
	switch s {
	case UnknownEncoding:
		return "unknown"
	case JSONEncoding:
		return "json"
	default:
		return "unknown"
	}
}

func ParseSampleEncoding(enc string) Encoding {
	switch enc {
	case "json":
		return JSONEncoding
	default:
		return UnknownEncoding
	}
}

type MetadataKey string

const (
	EventUID  MetadataKey = "event_uid"
	EventRule MetadataKey = "event_rule"
	DigestUID MetadataKey = "digest_uid"
)

// Sample defines a sample to be exported
type Sample struct {
	Ts       time.Time
	Type     control.SampleType
	Streams  []control.SamplerStreamUID
	Encoding Encoding
	Key      string
	Data     []byte
	Metadata map[MetadataKey]string
}

// SamplerSamples contains a group of samples that originate from the same resource e.g. operator, service...
type SamplerSamples struct {
	ResourceName string
	SamplerName  string
	Samples      []Sample
}
