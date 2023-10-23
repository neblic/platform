package sample

import (
	conventions "go.opentelemetry.io/collector/semconv/v1.9.0"
)

func RangeSamplers(otlpLogs OTLPLogs, fn func(resource, sample string, samplerLogs SamplerOTLPLogs)) {
	for i := 0; i < otlpLogs.logs.ResourceLogs().Len(); i++ {
		rLog := otlpLogs.logs.ResourceLogs().At(i)

		var sampler string
		if samplerValue, ok := rLog.Resource().Attributes().Get(rlSamplerNameKey); ok {
			sampler = samplerValue.Str()
		}

		var resource string
		if resourceValue, ok := rLog.Resource().Attributes().Get(conventions.AttributeServiceName); ok {
			resource = resourceValue.Str()
		}

		if rLog.ScopeLogs().Len() != 1 {
			panic("expected only one scope log")
		}

		slogs := rLog.ScopeLogs().At(0)

		samplerLogs := SamplerOTLPLogs{
			resourceLogs: rLog,
			scopeLogs:    slogs,
		}

		fn(resource, sampler, samplerLogs)
	}
}

func RangeSamplerLogs(samplerOtlpLogs SamplerOTLPLogs, fn func(otlpLog interface{})) {
	for j := 0; j < samplerOtlpLogs.scopeLogs.LogRecords().Len(); j++ {
		logRecord := samplerOtlpLogs.scopeLogs.LogRecords().At(j)
		fn(OTLPLogFrom(logRecord))
	}
}

func RangeSamplerLogsWithType[T OTLPLog](samplerOtlpLogs SamplerOTLPLogs, fn func(otlpLog T)) {
	targetSampleType := OTLPLogToSampleType[T]()

	for j := 0; j < samplerOtlpLogs.scopeLogs.LogRecords().Len(); j++ {

		logRecord := samplerOtlpLogs.scopeLogs.LogRecords().At(j)

		sampleType := getSampleType(logRecord)

		// Check if the log record sample type is the same as type as the one provided
		// in the range function. If that's the case, run the callback
		if sampleType == targetSampleType {
			otlpLog := OTLPLogFrom(logRecord).(T)
			fn(otlpLog)
		}
	}
}

// Range iterates over the logs and runs the callback for each log record that matches the type provided
func RangeWithType[T OTLPLog](otlpLogs OTLPLogs, fn func(resource, sample string, otlpLog T)) {
	RangeSamplers(otlpLogs, func(resource, sample string, samplerLogs SamplerOTLPLogs) {
		RangeSamplerLogsWithType[T](samplerLogs, func(otlpLog T) {
			fn(resource, sample, otlpLog)
		})
	})
}

// Range iterates over the logs and runs the callback for each log
func Range(otlpLogs OTLPLogs, fn func(resource, sample string, otlpLog interface{})) {
	RangeSamplers(otlpLogs, func(resource, sample string, samplerLogs SamplerOTLPLogs) {
		RangeSamplerLogs(samplerLogs, func(otlpLog any) {
			fn(resource, sample, otlpLog)
		})
	})
}
