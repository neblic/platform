package sample

import (
	"reflect"
	"testing"
	"time"

	"github.com/neblic/platform/controlplane/control"
	"go.opentelemetry.io/collector/pdata/plog"
)

func TestOTLPLogFrom(t *testing.T) {
	otlpLogs := NewOTLPLogs()
	otlpSamplerLogs := otlpLogs.AppendSamplerOTLPLogs("resource1", "sampler1")
	otlpRawSampleLog := otlpSamplerLogs.AppendRawSampleOTLPLog()
	otlpRawSampleLog.SetTimestamp(time.Now())

	if got := OTLPLogFrom(otlpRawSampleLog.Record()); !reflect.DeepEqual(got, RawSampleOTLPLogFrom(otlpRawSampleLog.Record())) {
		t.Errorf("NewOTLPLogFrom() = %v, want %v", got, RawSampleOTLPLogFrom(otlpRawSampleLog.Record()))
	}
}

func TestOTLPLogToSampleType(t *testing.T) {
	if got := OTLPLogToSampleType[RawSampleOTLPLog](); got != control.RawSampleType {
		t.Errorf("OTLPLogToSampleType() = %v, want %v", got, control.RawSampleType.String())
	}
}

func TestSetAndGetSampleRawData(t *testing.T) {
	// create a new baseOTLPLog instance
	b := baseOTLPLog{
		logRecord: plog.NewLogRecord(),
	}

	// set sample raw data using SetSampleRawData function
	rawData := []byte("sample raw data")
	b.SetSampleRawData(JSONEncoding, rawData)

	// get sample raw data using SampleRawData function
	gotEncoding := b.SampleEncoding()
	gotRawData := b.SampleRawData()

	// check if the returned values are equal to the expected values
	if !reflect.DeepEqual(gotRawData, rawData) {
		t.Errorf("SampleRawData() raw data = %v, want %v", gotRawData, rawData)
	}
	if gotEncoding != JSONEncoding {
		t.Errorf("SampleRawData() encoding = %v, want %v", gotEncoding, JSONEncoding)
	}

	// set sample raw data using SetSampleRawData function with a different encoding
	rawData2 := []byte("sample raw data 2")
	b.SetSampleRawData(UnknownEncoding, rawData2)

	// get sample raw data using SampleRawData function
	gotEncoding2 := b.SampleEncoding()
	gotRawData2 := b.SampleRawData()

	// check if the returned values are equal to the expected values
	if !reflect.DeepEqual(gotRawData2, rawData2) {
		t.Errorf("SampleRawData() raw data = %v, want %v", gotRawData2, rawData2)
	}
	if gotEncoding2 != UnknownEncoding {
		t.Errorf("SampleRawData() encoding = %v, want %v", gotEncoding2, UnknownEncoding)
	}
}
func TestStreams(t *testing.T) {
	// create a new baseOTLPLog instance
	b := baseOTLPLog{
		logRecord: plog.NewLogRecord(),
	}

	// set streams using setStreams function
	streams := []control.SamplerStreamUID{"stream1", "stream2"}
	b.SetStreamUIDs(streams)

	// get streams using Streams function
	gotStreams := b.StreamUIDs()

	// check if the returned streams are equal to the expected streams
	if !reflect.DeepEqual(gotStreams, streams) {
		t.Errorf("Streams() = %v, want %v", gotStreams, streams)
	}
}

func TestSetAndGetSampleKey(t *testing.T) {
	// create a new baseOTLPLog instance
	b := baseOTLPLog{
		logRecord: plog.NewLogRecord(),
	}

	// set sample key using SetSampleKey function
	key := "sample key"
	b.SetSampleKey(key)

	// get sample key using SampleKey function
	gotKey := b.SampleKey()

	// check if the returned value is equal to the expected value
	if gotKey != key {
		t.Errorf("SampleKey() = %v, want %v", gotKey, key)
	}
}

func TestBaseOTLPLog_SetAndGetTimestamp(t *testing.T) {
	// create a new baseOTLPLog instance
	b := baseOTLPLog{
		logRecord: plog.NewLogRecord(),
	}

	// set timestamp using SetTimestamp function
	ts := time.Now()
	b.SetTimestamp(ts)

	// get timestamp using Timestamp function
	gotTs := b.Timestamp()

	// check if the returned value is equal to the expected value
	if !reflect.DeepEqual(gotTs, ts.UTC()) {
		t.Errorf("Timestamp() = %v, want %v", gotTs, ts.UTC())
	}
}

func TestEventOTLPLog_RuleExpression(t *testing.T) {
	// create a new EventOTLPLog instance
	e := EventOTLPLogFrom(plog.NewLogRecord())

	// set rule expression using SetRuleExpression function
	ruleExpr := "sample rule expression"
	e.SetRuleExpression(ruleExpr)

	// get rule expression using RuleExpression function
	gotRuleExpr := e.RuleExpression()

	// check if the returned value is equal to the expected value
	if gotRuleExpr != ruleExpr {
		t.Errorf("RuleExpression() = %v, want %v", gotRuleExpr, ruleExpr)
	}
}

func TestStructDigestOTLPLog_SetUID(t *testing.T) {
	// create a new StructDigestOTLPLog instance
	s := StructDigestOTLPLogFrom(plog.NewLogRecord())

	// set UID using SetUID function
	uid := control.SamplerDigestUID("sample uid")
	s.SetUID(uid)

	// get UID using UID function
	gotUID := s.UID()

	// check if the returned value is equal to the expected value
	if gotUID != uid {
		t.Errorf("UID() = %v, want %v", gotUID, uid)
	}
}

func TestValueDigestOTLPLog_SetUID(t *testing.T) {
	// create a new ValueDigestOTLPLog instance
	s := ValueDigestOTLPLogFrom(plog.NewLogRecord())

	// set UID using SetUID function
	uid := control.SamplerDigestUID("sample uid")
	s.SetUID(uid)

	// get UID using UID function
	gotUID := s.UID()

	// check if the returned value is equal to the expected value
	if gotUID != uid {
		t.Errorf("UID() = %v, want %v", gotUID, uid)
	}
}

func TestEventOTLPLog_SetUID(t *testing.T) {
	// create a new EventOTLPLog instance
	s := EventOTLPLogFrom(plog.NewLogRecord())

	// set UID using SetUID function
	uid := control.SamplerEventUID("sample uid")
	s.SetUID(uid)

	// get UID using UID function
	gotUID := s.UID()

	// check if the returned value is equal to the expected value
	if gotUID != uid {
		t.Errorf("UID() = %v, want %v", gotUID, uid)
	}
}
