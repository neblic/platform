package otlp

import (
	"github.com/neblic/platform/sampler/internal/sample/exporter"
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

// TODO: in general, the produced log could be more compact but this way, we keep it human-readable
// eventually, we will want to binary encode as much as possible and create a human-readable
// representation when necessary
func fromSamplerSamples(resourceSmpls []exporter.SamplerSamples) plog.Logs {
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
			logRecord.SetTimestamp(pcommon.Timestamp(smpl.Ts.UnixNano()))
			if smpl.Encoding == exporter.JSONSampleEncoding {
				logRecord.Body().SetStr(string(smpl.Data))
			} else {
				logRecord.Body().SetEmptyBytes().FromRaw(smpl.Data)
			}

			// build attributes values
			lrSamplingRuleUIDs := logRecord.Attributes().PutEmptySlice(lrSampleStreamsUIDsKey)
			lrSamplingRuleUIDs.EnsureCapacity(len(smpl.Streams))
			for _, stream := range smpl.Streams {
				lrSamplingRuleUIDs.AppendEmpty().SetStr(string(stream))
			}

			logRecord.Attributes().PutStr(lrSampleTypeKey, smpl.Type.String())
			logRecord.Attributes().PutStr(lrSampleEncodingKey, smpl.Encoding.String())
		}
	}

	return logs
}
