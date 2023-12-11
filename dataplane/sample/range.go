package sample

import (
	conventions "go.opentelemetry.io/collector/semconv/v1.9.0"
)

// RangeSamplers iterates over the samplers and runs the callback for each one. Samplers added to
// the OTLPLogs after this function is called will not be visited
func RangeSamplers(otlpLogs OTLPLogs, fn func(resource, sampler string, samplerLogs SamplerOTLPLogs)) {
	resourceLogsLen := otlpLogs.logs.ResourceLogs().Len()
	for i := 0; i < resourceLogsLen; i++ {
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

// RangeSamplerLogs iterates over the logs and runs the callback for each log. Logs added to
// the SamplerOTLPLogs after this function is called will not be visited
func RangeSamplerLogs(samplerOtlpLogs SamplerOTLPLogs, fn func(otlpLog interface{})) {
	scopeLogsLen := samplerOtlpLogs.scopeLogs.LogRecords().Len()
	for j := 0; j < scopeLogsLen; j++ {
		logRecord := samplerOtlpLogs.scopeLogs.LogRecords().At(j)
		fn(OTLPLogFrom(logRecord))
	}
}

func RangeSamplerLogsWithType[T OTLPLog](samplerOtlpLogs SamplerOTLPLogs, fn func(otlpLog T)) {
	targetSampleType := OTLPLogToSampleType[T]()

	scopeLogsLen := samplerOtlpLogs.scopeLogs.LogRecords().Len()
	for j := 0; j < scopeLogsLen; j++ {

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

// Range iterates over the logs and runs the callback for each log record that matches the type provided.
// Logs added to the SamplerOTLPLogs after this function is called will not be visited.
func RangeWithType[T OTLPLog](otlpLogs OTLPLogs, fn func(resource, sampler string, otlpLog T)) {
	RangeSamplers(otlpLogs, func(resource, sample string, samplerLogs SamplerOTLPLogs) {
		RangeSamplerLogsWithType[T](samplerLogs, func(otlpLog T) {
			fn(resource, sample, otlpLog)
		})
	})
}

// Range iterates over the logs and runs the callback for each log.
// Logs added to the SamplerOTLPLogs after this function is called will not be visited.
func Range(otlpLogs OTLPLogs, fn func(resource, sampler string, otlpLog interface{})) {
	RangeSamplers(otlpLogs, func(resource, sample string, samplerLogs SamplerOTLPLogs) {
		RangeSamplerLogs(samplerLogs, func(otlpLog any) {
			fn(resource, sample, otlpLog)
		})
	})
}
