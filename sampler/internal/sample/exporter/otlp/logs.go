package otlp

import (
	"github.com/neblic/platform/sampler/internal/sample"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	conventions "go.opentelemetry.io/collector/semconv/v1.9.0"
)

const (
	rlSamplerNameKey      = "sampler_name"
	lrClientUIDsAttrKey   = "client_uids"
	lrSamplingRuleUIDsKey = "sampling_rule_uids"
)

func fromResourceSamples(resourceSmpls []sample.ResourceSamples) plog.Logs {
	logs := plog.NewLogs()

	rlogs := logs.ResourceLogs()
	rlogs.EnsureCapacity(len(resourceSmpls))
	for i := 0; i < len(resourceSmpls); i++ {
		resourceSmpl := resourceSmpls[i]

		rlog := rlogs.AppendEmpty()
		rlog.Resource().Attributes().PutStr(conventions.AttributeServiceName, resourceSmpl.ResourceName)
		rlog.Resource().Attributes().PutStr(rlSamplerNameKey, resourceSmpl.SamplerName)

		slogs := rlog.ScopeLogs()
		slogs.EnsureCapacity(len(resourceSmpl.SamplersSamples))
		for j := 0; j < len(resourceSmpl.SamplersSamples); j++ {
			samplerSamples := resourceSmpl.SamplersSamples[j]

			slog := slogs.AppendEmpty()
			logRecords := slog.LogRecords()
			logRecords.EnsureCapacity(len(samplerSamples.Samples))
			for k := 0; k < len(samplerSamples.Samples); k++ {
				sample := samplerSamples.Samples[k]

				// we consider each sample to be a LogRecord
				logRecord := logRecords.AppendEmpty()

				logRecord.SetTimestamp(pcommon.Timestamp(sample.Ts.UnixNano()))
				logRecord.Body().SetEmptyMap().FromRaw(sample.Data)
				lrClientUIDs := logRecord.Attributes().PutEmptySlice(lrClientUIDsAttrKey)
				lrSamplingRuleUIDs := logRecord.Attributes().PutEmptySlice(lrSamplingRuleUIDsKey)

				// build attributes values
				lrClientUIDs.EnsureCapacity(len(sample.Matches))
				lrSamplingRuleUIDs.EnsureCapacity(len(sample.Matches))
				for _, match := range sample.Matches {
					lrSamplingRuleUIDs.AppendEmpty().SetStr(string(match.StreamUID))
				}

			}
		}
	}

	return logs
}
