package event

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/neblic/platform/controlplane/control"
	"github.com/neblic/platform/dataplane/sample"
)

func TestEventor_ProcessSample(t *testing.T) {
	// Create a new Eventor
	eventor, err := NewEventor(Settings{ResourceName: "resource1", SamplerName: "sampler1"})
	if err != nil {
		t.Fatalf("NewEventor() returned an error: %v", err)
	}

	// Set the configuration for the Eventor
	event1UUID := control.SamplerEventUID(uuid.NewString())
	event2UUID := control.SamplerEventUID(uuid.NewString())
	events := control.Events{
		event1UUID: {
			UID:       event1UUID,
			Name:      "event1",
			StreamUID: "stream1",
			Rule: control.Rule{
				Lang:       control.SrlCel,
				Expression: `sample.foo == "bar"`,
			},
			Limiter: control.LimiterConfig{
				Limit: 10,
			},
		},
		event2UUID: {
			UID:       event2UUID,
			Name:      "event2",
			StreamUID: "stream1",
			Rule: control.Rule{
				Lang:       control.SrlCel,
				Expression: `sample.foo == "baz"`,
			},
			Limiter: control.LimiterConfig{
				// Limit set to 0 means no event will be generated
				Limit: 0,
			},
		},
	}
	streams := control.Streams{
		"stream1": {
			UID:  "stream1",
			Name: "stream1",
		},
	}
	err = eventor.SetEventsConfig(events, streams)
	if err != nil {
		t.Fatalf("SetEventsConfig() returned an error: %v", err)
	}

	// Create some sample data
	logs := sample.NewOTLPLogs()
	samplerLogs := logs.AppendSamplerOTLPLogs("resource1", "sampler1")

	rawSample1 := samplerLogs.AppendRawSampleOTLPLog()
	rawSample1.SetTimestamp(time.Now())
	rawSample1.SetStreamUIDs([]control.SamplerStreamUID{"stream1"})
	rawSample1.SetSampleKey("key1")
	rawSample1.SetSampleRawData(sample.JSONEncoding, []byte(`{"foo":"bar"}`))

	rawSample2 := samplerLogs.AppendRawSampleOTLPLog()
	rawSample2.SetTimestamp(time.Now())
	rawSample2.SetStreamUIDs([]control.SamplerStreamUID{"stream1"})
	rawSample2.SetSampleKey("key2")
	rawSample2.SetSampleRawData(sample.JSONEncoding, []byte(`{"foo":"baz"}`))

	// Process the sample data
	err = eventor.ProcessSample(samplerLogs)
	if err != nil {
		t.Fatalf("ProcessSample() returned an error: %v", err)
	}

	// Check that the expected events were created in the sample data
	expectedEventSamples := []struct {
		UID            control.SamplerEventUID
		Streams        []control.SamplerStreamUID
		SampleKey      string
		SampleRawData  []byte
		RuleExpression string
	}{
		{
			UID:            event1UUID,
			Streams:        []control.SamplerStreamUID{"stream1"},
			SampleKey:      "key1",
			SampleRawData:  []byte(`{"foo":"bar"}`),
			RuleExpression: `sample.foo == "bar"`,
		},
	}
	i := 0
	sample.RangeSamplerLogsWithType[sample.EventOTLPLog](samplerLogs, func(otlpLog sample.EventOTLPLog) {
		if i >= len(expectedEventSamples) {
			t.Errorf("More events than the expected were created")
			return
		}

		expectedEventSample := expectedEventSamples[i]

		if otlpLog.UID() != expectedEventSample.UID {
			t.Errorf("Event %s has incorrect UID: got %s, want %s", otlpLog.UID(), otlpLog.UID(), expectedEventSample.UID)
		}
		if !reflect.DeepEqual(otlpLog.StreamUIDs(), expectedEventSample.Streams) {
			t.Errorf("Event %s has incorrect streams: got %v, want %v", otlpLog.UID(), otlpLog.StreamUIDs(), expectedEventSample.Streams)
		}
		if otlpLog.SampleKey() != expectedEventSample.SampleKey {
			t.Errorf("Event %s has incorrect sample key: got %s, want %s", otlpLog.UID(), otlpLog.SampleKey(), expectedEventSample.SampleKey)
		}
		if !reflect.DeepEqual(otlpLog.SampleRawData(), expectedEventSample.SampleRawData) {
			t.Errorf("Event %s has incorrect sample data: got %v, want %v", otlpLog.UID(), otlpLog.SampleRawData(), expectedEventSample.SampleRawData)
		}
		if otlpLog.RuleExpression() != expectedEventSample.RuleExpression {
			t.Errorf("Event %s has incorrect rule expression: got %s, want %s", otlpLog.UID(), otlpLog.RuleExpression(), expectedEventSample.RuleExpression)
		}

		i++

	})

	if i != len(expectedEventSamples) {
		t.Errorf("Not enough events were created: got %d, want %d", i, len(expectedEventSamples))
	}
}
