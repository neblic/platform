package dataplane

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/neblic/platform/dataplane/metric"
	"github.com/neblic/platform/dataplane/protos"
	"github.com/neblic/platform/dataplane/sample"
	"google.golang.org/protobuf/encoding/protojson"
)

func (p *Processor) processEvent(samplerMetrics metric.SamplerMetrics, attributes metric.DatapointAttributes, event sample.EventOTLPLog) error {
	eventUUIDString := string(event.UID())
	eventUUID, err := uuid.Parse(eventUUIDString)
	if err != nil {
		return fmt.Errorf("cannot parse event UUID: %v", err)
	}

	streams := event.StreamUIDs()
	if len(streams) == 0 {
		return fmt.Errorf("no streams in sample")
	}
	if len(streams) > 1 {
		return fmt.Errorf("multiple streams in sample")
	}
	streamUUID := string(streams[0])

	attributes = attributes.WithEventUID(eventUUID).WithStreamUID(streamUUID)

	samplerMetrics.AppendSum(eventUUID, "", "", true, metric.AggregationTemporalityDelta).AppendInt64Datapoint(attributes).SetValue(1)

	return nil
}

func generateValueDigestMetrics(samplerMetrics metric.SamplerMetrics, uid uuid.UUID, path metric.Path, attributes metric.DatapointAttributes, field string, value *protos.ValueValue) {
	anyPath := path.AddPart(field, metric.AnyType)
	samplerMetrics.AppendSum(uid, anyPath, "total_count", true, metric.AggregationTemporalityDelta).AppendInt64Datapoint(attributes).SetValue(int64(value.TotalCount))
	samplerMetrics.AppendSum(uid, anyPath, "null_count", true, metric.AggregationTemporalityDelta).AppendInt64Datapoint(attributes).SetValue(int64(value.NullCount))

	if value.Number != nil {
		numberPath := path.AddPart(field, metric.NumberType)
		samplerMetrics.AppendSum(uid, numberPath, "total_count", true, metric.AggregationTemporalityDelta).AppendInt64Datapoint(attributes).SetValue(int64(value.Number.TotalCount))
		samplerMetrics.AppendSum(uid, numberPath, "default_count", true, metric.AggregationTemporalityDelta).AppendInt64Datapoint(attributes).SetValue(int64(value.Number.DefaultCount))
		samplerMetrics.AppendSum(uid, numberPath, "null_count", true, metric.AggregationTemporalityDelta).AppendInt64Datapoint(attributes).SetValue(int64(value.Number.NullCount))
		samplerMetrics.AppendGauge(uid, numberPath, "min").AppendFloat64Datapoint(attributes).SetValue(value.Number.Min.Value)
		samplerMetrics.AppendGauge(uid, numberPath, "avg").AppendFloat64Datapoint(attributes).SetValue(value.Number.Avg.Sum / float64(value.Number.Avg.Count))
		samplerMetrics.AppendGauge(uid, numberPath, "max").AppendFloat64Datapoint(attributes).SetValue(value.Number.Max.Value)
		samplerMetrics.AppendGauge(uid, numberPath, "cardinality").AppendInt64Datapoint(attributes).SetValue(int64(value.Number.HyperLogLog.Cardinality))
	}
	if value.String_ != nil {
		stringPath := path.AddPart(field, metric.StringType)
		samplerMetrics.AppendSum(uid, stringPath, "total_count", true, metric.AggregationTemporalityDelta).AppendInt64Datapoint(attributes).SetValue(int64(value.String_.TotalCount))
		samplerMetrics.AppendSum(uid, stringPath, "default_count", true, metric.AggregationTemporalityDelta).AppendInt64Datapoint(attributes).SetValue(int64(value.String_.DefaultCount))
		samplerMetrics.AppendSum(uid, stringPath, "null_count", true, metric.AggregationTemporalityDelta).AppendInt64Datapoint(attributes).SetValue(int64(value.String_.NullCount))
		samplerMetrics.AppendGauge(uid, stringPath, "cardinality").AppendInt64Datapoint(attributes).SetValue(int64(value.String_.HyperLogLog.Cardinality))
		samplerMetrics.AppendGauge(uid, stringPath, "length_min").AppendFloat64Datapoint(attributes).SetValue(value.String_.Length.Min.Value)
		samplerMetrics.AppendGauge(uid, stringPath, "length_avg").AppendFloat64Datapoint(attributes).SetValue(value.String_.Length.Avg.Sum / float64(value.String_.Length.Avg.Count))
		samplerMetrics.AppendGauge(uid, stringPath, "length_max").AppendFloat64Datapoint(attributes).SetValue(value.String_.Length.Max.Value)
		samplerMetrics.AppendGauge(uid, stringPath, "length_cardinality").AppendInt64Datapoint(attributes).SetValue(int64(value.String_.Length.HyperLogLog.Cardinality))
	}
	if value.Boolean != nil {
		booleanPath := path.AddPart(field, metric.BooleanType)
		samplerMetrics.AppendSum(uid, booleanPath, "total_count", true, metric.AggregationTemporalityDelta).AppendInt64Datapoint(attributes).SetValue(int64(value.Boolean.TotalCount))
		samplerMetrics.AppendSum(uid, booleanPath, "default_count", true, metric.AggregationTemporalityDelta).AppendInt64Datapoint(attributes).SetValue(int64(value.Boolean.DefaultCount))
		samplerMetrics.AppendSum(uid, booleanPath, "null_count", true, metric.AggregationTemporalityDelta).AppendInt64Datapoint(attributes).SetValue(int64(value.Boolean.NullCount))
		samplerMetrics.AppendSum(uid, booleanPath, "false_count", true, metric.AggregationTemporalityDelta).AppendInt64Datapoint(attributes).SetValue(int64(value.Boolean.FalseCount))
		samplerMetrics.AppendSum(uid, booleanPath, "true_count", true, metric.AggregationTemporalityDelta).AppendInt64Datapoint(attributes).SetValue(int64(value.Boolean.TrueCount))
	}
	if value.Array != nil {
		arrayPath := path.AddPart(field, metric.ArrayType)
		samplerMetrics.AppendSum(uid, arrayPath, "total_count", true, metric.AggregationTemporalityDelta).AppendInt64Datapoint(attributes).SetValue(int64(value.Array.TotalCount))
		samplerMetrics.AppendSum(uid, arrayPath, "default_count", true, metric.AggregationTemporalityDelta).AppendInt64Datapoint(attributes).SetValue(int64(value.Array.DefaultCount))
		samplerMetrics.AppendSum(uid, arrayPath, "null_count", true, metric.AggregationTemporalityDelta).AppendInt64Datapoint(attributes).SetValue(int64(value.Array.NullCount))
		generateValueDigestMetrics(samplerMetrics, uid, arrayPath, attributes, "*", value.Array.Values)
	}
	if value.Obj != nil {
		objPath := path.AddPart(field, metric.ObjectType)
		samplerMetrics.AppendSum(uid, objPath, "total_count", true, metric.AggregationTemporalityDelta).AppendInt64Datapoint(attributes).SetValue(int64(value.Obj.TotalCount))
		samplerMetrics.AppendSum(uid, objPath, "default_count", true, metric.AggregationTemporalityDelta).AppendInt64Datapoint(attributes).SetValue(int64(value.Obj.DefaultCount))
		samplerMetrics.AppendSum(uid, objPath, "null_count", true, metric.AggregationTemporalityDelta).AppendInt64Datapoint(attributes).SetValue(int64(value.Obj.NullCount))
		for field, value := range value.Obj.Fields {
			generateValueDigestMetrics(samplerMetrics, uid, objPath, attributes, field, value)
		}
	}
}

func (p *Processor) handleValueDigest(samplerMetrics metric.SamplerMetrics, attributes metric.DatapointAttributes, valueDigest sample.ValueDigestOTLPLog) error {
	digestUUIDString := string(valueDigest.UID())
	digestUUID, err := uuid.Parse(digestUUIDString)
	if err != nil {
		return err
	}

	// Decode the value digest
	objDigest := new(protos.ObjValue)
	switch valueDigest.SampleEncoding() {
	case sample.JSONEncoding:
		err = protojson.Unmarshal(valueDigest.SampleRawData(), objDigest)
	default:
		err = fmt.Errorf("unsupported encoding %s", valueDigest.SampleEncoding())
	}
	if err != nil {
		return err
	}

	streams := valueDigest.StreamUIDs()
	if len(streams) == 0 {
		return fmt.Errorf("no streams in sample")
	}
	if len(streams) > 1 {
		return fmt.Errorf("multiple streams in sample")
	}
	streamUUID := string(streams[0])

	attributes = attributes.WithDigestUID(digestUUID).WithStreamUID(streamUUID)

	path := metric.NewPath().AddPart("$", metric.ObjectType)

	samplerMetrics.AppendSum(digestUUID, path, "total_count", true, metric.AggregationTemporalityDelta).AppendInt64Datapoint(attributes).SetValue(int64(objDigest.TotalCount))
	samplerMetrics.AppendSum(digestUUID, path, "default_count", true, metric.AggregationTemporalityDelta).AppendInt64Datapoint(attributes).SetValue(int64(objDigest.DefaultCount))
	samplerMetrics.AppendSum(digestUUID, path, "null_count", true, metric.AggregationTemporalityDelta).AppendInt64Datapoint(attributes).SetValue(int64(objDigest.NullCount))

	for field, value := range objDigest.Fields {
		generateValueDigestMetrics(samplerMetrics, digestUUID, path, attributes, field, value)
	}

	return nil
}

func (p *Processor) ComputeMetrics(otlpLogs sample.OTLPLogs) (metric.Metrics, error) {
	var errs error

	metrics := metric.NewMetrics()

	attributes := metric.NewDatapointAttributes()
	sample.RangeSamplers(otlpLogs, func(resource string, sampler string, samplerLogs sample.SamplerOTLPLogs) {
		samplerMetrics := metrics.AppendSamplerMetrics(resource, sampler)
		sample.RangeSamplerLogs(samplerLogs, func(otlpLog sample.OTLPLog) {
			var err error

			switch v := otlpLog.(type) {
			case sample.EventOTLPLog:
				attributes = attributes.WithTs(v.Timestamp())
				attributes = attributes.WithSampleType(v.SampleType())
				err = p.processEvent(samplerMetrics, attributes, v)
			case sample.ValueDigestOTLPLog:
				attributes = attributes.WithTs(v.Timestamp())
				attributes = attributes.WithSampleType(v.SampleType())
				err = p.handleValueDigest(samplerMetrics, attributes, v)
			default:
			}

			if err != nil {
				errs = errors.Join(errs, err)
			}

		})

	})

	return metrics, errs
}
