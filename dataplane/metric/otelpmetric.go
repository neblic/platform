package metric

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	conventions "go.opentelemetry.io/collector/semconv/v1.9.0"
)

const (
	// sample attributes
	OTLPSampleStreamUIDKey   = "com.neblic.sample.stream.uid"
	OTLPSampleStreamNamesKey = "com.neblic.sample.stream.name"
	OTLPSampleSampleTypeKey  = "com.neblic.sample.type"
	OTLPSampleFieldPathKey   = "com.neblic.sample.field.path"
	OTLPSampleNameKey        = "com.neblic.sample.name"
	OTLPSampleEventUIDKey    = "com.neblic.sample.event.uid"
	OTLPSampleDigestUIDKey   = "com.neblic.sample.digest.uid"
)

type AggregationTemporality pmetric.AggregationTemporality

var (
	AggregationTemporalityDelta      = AggregationTemporality(pmetric.AggregationTemporalityDelta)
	AggregationTemporalityCumulative = AggregationTemporality(pmetric.AggregationTemporalityCumulative)
)

type Number interface {
	float64 | int64
}

type Datapoint[T Number] struct {
	datapoint pmetric.NumberDataPoint
}

func newDatapoint[T Number](metricName string, metricPath Path, datapoint pmetric.NumberDataPoint, datapointAttributes DatapointAttributes) Datapoint[T] {
	attributes := datapointAttributes.attributes
	if !metricPath.IsEmpty() {
		attributes.PutStr(OTLPSampleFieldPathKey, metricPath.String())
	}
	if metricName != "" {
		attributes.PutStr(OTLPSampleNameKey, metricName)
	}
	attributes.CopyTo(datapoint.Attributes())

	datapoint.SetTimestamp(pcommon.Timestamp(datapointAttributes.tsUnixNano))

	return Datapoint[T]{
		datapoint: datapoint,
	}
}

func (d Datapoint[T]) SetValue(value T) {
	switch v := any(value).(type) {
	case float64:
		d.datapoint.SetDoubleValue(v)
	case int64:
		d.datapoint.SetIntValue(v)
	default:
		panic(fmt.Sprintf("unsupported type %T", v))
	}
}

func getMetricName(uid uuid.UUID, path Path, name string) string {
	// Metric name has to be unique across all resources and samplers. In order to achieve that, we generate a name.
	// Metric name has to follow [a-zA-Z_:][a-zA-Z0-9_:]*) regex defined by the prometheus naming conventions. That means
	// it can only contain alphanumeric characters, colons (which cannot be used by users) and underscores, and it has to
	// start with a letter or underscore
	metricNameUUID := uid
	if !path.IsEmpty() {
		metricNameUUID = uuid.NewSHA1(metricNameUUID, []byte(path))
	}
	if name != "" {
		metricNameUUID = uuid.NewSHA1(metricNameUUID, []byte(name))
	}

	// Normalize metric name UUID
	// - Replace dashes with underscores to follow prometheus naming conventions
	// - Name starting with a number are not valid. Always prefix the name with an underscore
	metricNameUUIDString := "_" + strings.ReplaceAll(metricNameUUID.String(), "-", "_")

	return metricNameUUIDString
}

// Base implementation with common functionality
type Metric struct {
	path       Path
	name       string
	metric     pmetric.Metric
	datapoints pmetric.NumberDataPointSlice
}

func newMetric(path Path, name string, metric pmetric.Metric, datapoints pmetric.NumberDataPointSlice) Metric {
	return Metric{
		path:       path,
		name:       name,
		metric:     metric,
		datapoints: datapoints,
	}
}

func (m Metric) AppendInt64Datapoint(attributes DatapointAttributes) Datapoint[int64] {
	return newDatapoint[int64](m.name, m.path, m.datapoints.AppendEmpty(), attributes)

}

func (m Metric) AppendFloat64Datapoint(attributes DatapointAttributes) Datapoint[float64] {
	return newDatapoint[float64](m.name, m.path, m.datapoints.AppendEmpty(), attributes)
}

func (m Metric) Record() pmetric.Metric {
	return m.metric
}

type SamplerMetrics struct {
	resourceMetrics pmetric.ResourceMetrics
	scopeMetrics    pmetric.ScopeMetrics
}

func (s SamplerMetrics) AppendSum(uid uuid.UUID, path Path, name string, isMonotonic bool, temporality AggregationTemporality) Metric {
	metric := s.scopeMetrics.Metrics().AppendEmpty()
	metric.SetName(getMetricName(uid, path, name))

	sum := metric.SetEmptySum()
	sum.SetIsMonotonic(isMonotonic)
	sum.SetAggregationTemporality(pmetric.AggregationTemporality(temporality))

	return newMetric(path, name, metric, sum.DataPoints())
}

func (s SamplerMetrics) AppendGauge(uid uuid.UUID, path Path, name string) Metric {
	metric := s.scopeMetrics.Metrics().AppendEmpty()
	metric.SetName(getMetricName(uid, path, name))

	gauge := metric.SetEmptyGauge()

	return newMetric(path, name, metric, gauge.DataPoints())
}

type Metrics struct {
	metrics pmetric.Metrics
}

func NewMetrics() Metrics {
	return Metrics{
		metrics: pmetric.NewMetrics(),
	}
}

func MetricsFrom(pmetrics pmetric.Metrics) Metrics {
	return Metrics{
		metrics: pmetrics,
	}
}

func (m *Metrics) AppendSamplerMetrics(resource string, sampler string) SamplerMetrics {
	resourceMetrics := m.metrics.ResourceMetrics().AppendEmpty()
	resourceMetrics.Resource().Attributes().PutStr(conventions.AttributeServiceName, resource)

	scopeMetrics := resourceMetrics.ScopeMetrics().AppendEmpty()
	scopeMetrics.Scope().SetName(sampler)

	return SamplerMetrics{
		resourceMetrics: resourceMetrics,
		scopeMetrics:    scopeMetrics,
	}
}

func (m Metrics) Metrics() pmetric.Metrics {
	return m.metrics
}

func (m Metrics) Len() int {
	return m.metrics.ResourceMetrics().Len()
}
