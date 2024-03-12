package dataplane

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/neblic/platform/controlplane/control"
	"github.com/neblic/platform/dataplane/metric"
	"github.com/neblic/platform/dataplane/sample"
	"github.com/neblic/platform/logging"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/pmetrictest"
)

func TestProcessor_ComputeMetrics(t *testing.T) {
	type fields struct {
		logger logging.Logger
	}
	type args struct {
		otlpLogs sample.OTLPLogs
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    metric.Metrics
		wantErr bool
	}{
		{
			name: "process event",
			args: args{
				otlpLogs: func() sample.OTLPLogs {
					logs := sample.NewOTLPLogs()
					samplerLogs := logs.AppendSamplerOTLPLogs("resource1", "sampler1")
					event := samplerLogs.AppendEventOTLPLog()
					event.SetUID("550e8400-e29b-41d4-a716-446655440000")
					event.SetTimestamp(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC))
					event.SetStreamUIDs([]control.SamplerStreamUID{"660e8400-e29b-41d4-a716-446655440000"})
					event.SetSampleKey("key1")
					event.SetSampleRawData(sample.JSONEncoding, []byte("{\"id\": 1}"))
					event.SetRuleExpression("sample.id==1")
					return logs
				}(),
			},
			want: func() metric.Metrics {
				attributes := metric.NewDatapointAttributes().
					WithEventUID(uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")).
					WithStreamUID("660e8400-e29b-41d4-a716-446655440000").
					WithTs(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)).
					WithSampleType(control.EventSampleType)
				metrics := metric.NewMetrics()
				samplerMetrics := metrics.AppendSamplerMetrics("resource1", "sampler1")
				samplerMetrics.AppendSum(uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), "", "", true, metric.AggregationTemporalityDelta).AppendInt64Datapoint(attributes).SetValue(1)
				return metrics
			}(),
		},
		{
			name: "process empty digest",
			args: args{
				otlpLogs: func() sample.OTLPLogs {
					logs := sample.NewOTLPLogs()
					samplerLogs := logs.AppendSamplerOTLPLogs("resource1", "sampler1")
					digest := samplerLogs.AppendValueDigestOTLPLog()
					digest.SetUID("550e8400-e29b-41d4-a716-446655440000")
					digest.SetTimestamp(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC))
					digest.SetStreamUIDs([]control.SamplerStreamUID{"660e8400-e29b-41d4-a716-446655440000"})
					digest.SetSampleKey("key1")
					digest.SetSampleRawData(sample.JSONEncoding, []byte(`{"totalCount": 1, "defaultCount": 0, "nullCount": 1}`))
					return logs
				}(),
			},
			want: func() metric.Metrics {
				path := metric.NewPath().AddPart("$", metric.ObjectType)
				attributes := metric.NewDatapointAttributes().
					WithDigestUID(uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")).
					WithStreamUID("660e8400-e29b-41d4-a716-446655440000").
					WithTs(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)).
					WithSampleType(control.ValueDigestSampleType)
				metrics := metric.NewMetrics()
				samplerMetrics := metrics.AppendSamplerMetrics("resource1", "sampler1")
				samplerMetrics.AppendSum(uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), path, "total_count", true, metric.AggregationTemporalityDelta).AppendInt64Datapoint(attributes).SetValue(int64(1))
				samplerMetrics.AppendSum(uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), path, "default_count", true, metric.AggregationTemporalityDelta).AppendInt64Datapoint(attributes).SetValue(int64(0))
				samplerMetrics.AppendSum(uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), path, "null_count", true, metric.AggregationTemporalityDelta).AppendInt64Datapoint(attributes).SetValue(int64(1))
				return metrics
			}(),
		},
		{
			name: "process number digest",
			args: args{
				otlpLogs: func() sample.OTLPLogs {
					logs := sample.NewOTLPLogs()
					samplerLogs := logs.AppendSamplerOTLPLogs("resource1", "sampler1")
					digest := samplerLogs.AppendValueDigestOTLPLog()
					digest.SetUID("550e8400-e29b-41d4-a716-446655440000")
					digest.SetTimestamp(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC))
					digest.SetStreamUIDs([]control.SamplerStreamUID{"660e8400-e29b-41d4-a716-446655440000"})
					digest.SetSampleKey("key1")
					digest.SetSampleRawData(sample.JSONEncoding, []byte(`
						{
							"totalCount": 60,
							"defaultCount": 0,
							"nullCount": 0,
							"fields": {
								"field1": {
									"totalCount": "60",
									"nullCount": "0",
									"number": {
										"totalCount": "60",
										"defaultCount": "0",
										"nullCount": "0",
										"min": {
											"value": 1
										},
										"avg": {
											"sum": 95,
											"count": "60"
										},
										"max": {
											"value": 2
										},
										"hyperLogLog": {
											"data": "AQ4AAQAAAAIBabCuApwXqAAAAAAAAAAAAAAAAA==",
											"cardinality": "2"
										}
									}
								}
							}
						}`))
					return logs
				}(),
			},
			want: func() metric.Metrics {
				path := metric.NewPath().AddPart("$", metric.ObjectType)
				anyPath := path.AddPart("field1", metric.AnyType)
				numberPath := path.AddPart("field1", metric.NumberType)
				attributes := metric.NewDatapointAttributes().
					WithDigestUID(uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")).
					WithStreamUID("660e8400-e29b-41d4-a716-446655440000").
					WithTs(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)).
					WithSampleType(control.ValueDigestSampleType)
				metrics := metric.NewMetrics()
				samplerMetrics := metrics.AppendSamplerMetrics("resource1", "sampler1")
				samplerMetrics.AppendSum(uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), path, "total_count", true, metric.AggregationTemporalityDelta).AppendInt64Datapoint(attributes).SetValue(int64(60))
				samplerMetrics.AppendSum(uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), path, "default_count", true, metric.AggregationTemporalityDelta).AppendInt64Datapoint(attributes).SetValue(int64(0))
				samplerMetrics.AppendSum(uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), path, "null_count", true, metric.AggregationTemporalityDelta).AppendInt64Datapoint(attributes).SetValue(int64(0))
				samplerMetrics.AppendSum(uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), anyPath, "total_count", true, metric.AggregationTemporalityDelta).AppendInt64Datapoint(attributes).SetValue(int64(60))
				samplerMetrics.AppendSum(uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), anyPath, "null_count", true, metric.AggregationTemporalityDelta).AppendInt64Datapoint(attributes).SetValue(int64(0))
				samplerMetrics.AppendSum(uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), numberPath, "total_count", true, metric.AggregationTemporalityDelta).AppendInt64Datapoint(attributes).SetValue(60)
				samplerMetrics.AppendSum(uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), numberPath, "default_count", true, metric.AggregationTemporalityDelta).AppendInt64Datapoint(attributes).SetValue(0)
				samplerMetrics.AppendSum(uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), numberPath, "null_count", true, metric.AggregationTemporalityDelta).AppendInt64Datapoint(attributes).SetValue(0)
				samplerMetrics.AppendGauge(uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), numberPath, "min").AppendFloat64Datapoint(attributes).SetValue(1)
				samplerMetrics.AppendGauge(uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), numberPath, "avg").AppendFloat64Datapoint(attributes).SetValue(1.5833333333333333)
				samplerMetrics.AppendGauge(uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), numberPath, "max").AppendFloat64Datapoint(attributes).SetValue(2)
				samplerMetrics.AppendGauge(uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), numberPath, "cardinality").AppendInt64Datapoint(attributes).SetValue(int64(2))

				return metrics
			}(),
		},
		{
			name: "process string digest",
			args: args{
				otlpLogs: func() sample.OTLPLogs {
					logs := sample.NewOTLPLogs()
					samplerLogs := logs.AppendSamplerOTLPLogs("resource1", "sampler1")
					digest := samplerLogs.AppendValueDigestOTLPLog()
					digest.SetUID("550e8400-e29b-41d4-a716-446655440000")
					digest.SetTimestamp(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC))
					digest.SetStreamUIDs([]control.SamplerStreamUID{"660e8400-e29b-41d4-a716-446655440000"})
					digest.SetSampleKey("key1")
					digest.SetSampleRawData(sample.JSONEncoding, []byte(`
						{
							"totalCount": 60,
							"defaultCount": 0,
							"nullCount": 0,
							"fields": {
								"field1": {
									"totalCount": "60",
									"nullCount": "0",
									"number": {
										"totalCount": "60",
										"defaultCount": "0",
										"nullCount": "0",
										"min": {
											"value": 1
										},
										"avg": {
											"sum": 95,
											"count": "60"
										},
										"max": {
											"value": 2
										},
										"hyperLogLog": {
											"data": "AQ4AAQAAAAIBabCuApwXqAAAAAAAAAAAAAAAAA==",
											"cardinality": "2"
										}
									}
								}
							}
						}`))
					return logs
				}(),
			},
			want: func() metric.Metrics {
				path := metric.NewPath().AddPart("$", metric.ObjectType)
				anyPath := path.AddPart("field1", metric.AnyType)
				numberPath := path.AddPart("field1", metric.NumberType)
				attributes := metric.NewDatapointAttributes().
					WithDigestUID(uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")).
					WithStreamUID("660e8400-e29b-41d4-a716-446655440000").
					WithTs(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)).
					WithSampleType(control.ValueDigestSampleType)
				metrics := metric.NewMetrics()
				samplerMetrics := metrics.AppendSamplerMetrics("resource1", "sampler1")
				samplerMetrics.AppendSum(uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), path, "total_count", true, metric.AggregationTemporalityDelta).AppendInt64Datapoint(attributes).SetValue(int64(60))
				samplerMetrics.AppendSum(uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), path, "default_count", true, metric.AggregationTemporalityDelta).AppendInt64Datapoint(attributes).SetValue(int64(0))
				samplerMetrics.AppendSum(uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), path, "null_count", true, metric.AggregationTemporalityDelta).AppendInt64Datapoint(attributes).SetValue(int64(0))
				samplerMetrics.AppendSum(uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), anyPath, "total_count", true, metric.AggregationTemporalityDelta).AppendInt64Datapoint(attributes).SetValue(int64(60))
				samplerMetrics.AppendSum(uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), anyPath, "null_count", true, metric.AggregationTemporalityDelta).AppendInt64Datapoint(attributes).SetValue(int64(0))
				samplerMetrics.AppendSum(uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), numberPath, "total_count", true, metric.AggregationTemporalityDelta).AppendInt64Datapoint(attributes).SetValue(60)
				samplerMetrics.AppendSum(uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), numberPath, "default_count", true, metric.AggregationTemporalityDelta).AppendInt64Datapoint(attributes).SetValue(0)
				samplerMetrics.AppendSum(uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), numberPath, "null_count", true, metric.AggregationTemporalityDelta).AppendInt64Datapoint(attributes).SetValue(0)
				samplerMetrics.AppendGauge(uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), numberPath, "min").AppendFloat64Datapoint(attributes).SetValue(1)
				samplerMetrics.AppendGauge(uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), numberPath, "avg").AppendFloat64Datapoint(attributes).SetValue(1.5833333333333333)
				samplerMetrics.AppendGauge(uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), numberPath, "max").AppendFloat64Datapoint(attributes).SetValue(2)
				samplerMetrics.AppendGauge(uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), numberPath, "cardinality").AppendInt64Datapoint(attributes).SetValue(int64(2))

				return metrics
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ph := &Processor{
				logger: tt.fields.logger,
			}
			got, err := ph.ComputeMetrics(tt.args.otlpLogs)
			if (err != nil) != tt.wantErr {
				t.Errorf("Processor.Process() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err := pmetrictest.CompareMetrics(got.Metrics(), tt.want.Metrics()); err != nil {
				t.Errorf("Processor.Process() = %v", err)
			}
		})
	}
}
