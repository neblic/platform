package sample

import (
	"reflect"
	"testing"
	"time"

	"github.com/neblic/platform/controlplane/control"
)

func TestRangeWithType(t *testing.T) {
	logs := NewOTLPLogs()

	samplerLogs := logs.AppendSamplerOTLPLogs("resource1", "sampler1")

	rawSample1 := samplerLogs.AppendRawSampleOTLPLog()
	rawSample1.SetTimestamp(time.Now())
	rawSample1.SetStreamUIDs([]control.SamplerStreamUID{"stream1"})
	rawSample1.SetSampleKey("key1")
	rawSample1.SetSampleRawData(JSONEncoding, []byte(`{"foo":"bar1"}`))

	rawSample2 := samplerLogs.AppendRawSampleOTLPLog()
	rawSample2.SetTimestamp(time.Now())
	rawSample2.SetStreamUIDs([]control.SamplerStreamUID{"stream2"})
	rawSample2.SetSampleKey("key2")
	rawSample2.SetSampleRawData(JSONEncoding, []byte(`{"foo":"bar2"}`))

	eventSample1 := samplerLogs.AppendEventOTLPLog()
	eventSample1.SetTimestamp(time.Now())
	eventSample1.SetStreamUIDs([]control.SamplerStreamUID{"stream3"})
	eventSample1.SetSampleKey("key3")
	eventSample1.SetSampleRawData(JSONEncoding, []byte(`{"foo":"bar3"}`))

	gotRawSamples := []RawSampleOTLPLog{}
	wantRawSamples := []RawSampleOTLPLog{rawSample1, rawSample2}
	RangeWithType[RawSampleOTLPLog](logs, func(resource, sample string, log RawSampleOTLPLog) {
		gotRawSamples = append(gotRawSamples, log)
	})
	if !reflect.DeepEqual(gotRawSamples, wantRawSamples) {
		t.Errorf("RangeWithType[RawSampleOTLPLog]() = %v, want %v", gotRawSamples, wantRawSamples)
	}

	gotEvents := []EventOTLPLog{}
	wantEvents := []EventOTLPLog{eventSample1}
	RangeWithType[EventOTLPLog](logs, func(resource, sample string, log EventOTLPLog) {
		gotEvents = append(gotEvents, log)
	})
	if !reflect.DeepEqual(gotEvents, wantEvents) {
		t.Errorf("RangeWithType[EventOTLPLog]() = %v, want %v", gotEvents, wantEvents)
	}

	gotValueDigests := []ValueDigestOTLPLog{}
	wantValueDigests := []ValueDigestOTLPLog{}
	RangeWithType[ValueDigestOTLPLog](logs, func(resource, sample string, log ValueDigestOTLPLog) {
		gotValueDigests = append(gotValueDigests, log)
	})
	if !reflect.DeepEqual(gotValueDigests, wantValueDigests) {
		t.Errorf("RangeWithType() = %v, want %v", gotValueDigests, wantValueDigests)
	}
}

func TestRange(t *testing.T) {
	logs := NewOTLPLogs()

	samplerLogs := logs.AppendSamplerOTLPLogs("resource1", "sampler1")

	rawSample1 := samplerLogs.AppendRawSampleOTLPLog()
	rawSample1.SetTimestamp(time.Now())
	rawSample1.SetStreamUIDs([]control.SamplerStreamUID{"stream1"})
	rawSample1.SetSampleKey("key1")
	rawSample1.SetSampleRawData(JSONEncoding, []byte(`{"foo":"bar1"}`))

	rawSample2 := samplerLogs.AppendRawSampleOTLPLog()
	rawSample2.SetTimestamp(time.Now())
	rawSample2.SetStreamUIDs([]control.SamplerStreamUID{"stream2"})
	rawSample2.SetSampleKey("key2")
	rawSample2.SetSampleRawData(JSONEncoding, []byte(`{"foo":"bar2"}`))

	eventSample1 := samplerLogs.AppendEventOTLPLog()
	eventSample1.SetTimestamp(time.Now())
	eventSample1.SetStreamNames([]string{"stream3"})
	eventSample1.SetSampleKey("key3")
	eventSample1.SetSampleRawData(JSONEncoding, []byte(`{"foo":"bar3"}`))

	gotLogs := []OTLPLog{}
	wantLogs := []OTLPLog{rawSample1, rawSample2, eventSample1}
	Range(logs, func(resource, sample string, log OTLPLog) {
		gotLogs = append(gotLogs, log)
	})
	if !reflect.DeepEqual(gotLogs, wantLogs) {
		t.Errorf("Range() = %v, want %v", gotLogs, wantLogs)
	}
}
