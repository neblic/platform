package sample

import (
	"fmt"
	"time"

	"github.com/neblic/platform/controlplane/control"
	"github.com/neblic/platform/internal/pkg/data"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	conventions "go.opentelemetry.io/collector/semconv/v1.9.0"
)

type OTLPLog interface {
	RawSampleOTLPLog |
		StructDigestOTLPLog |
		ValueDigestOTLPLog |
		EventOTLPLog |
		ConfigOTLPLog
}

func OTLPLogFrom(logRecord plog.LogRecord) any {
	var otlpLog any
	switch getSampleType(logRecord) {
	case control.RawSampleType:
		otlpLog = RawSampleOTLPLogFrom(logRecord)
	case control.StructDigestSampleType:
		otlpLog = StructDigestOTLPLogFrom(logRecord)
	case control.ValueDigestSampleType:
		otlpLog = ValueDigestOTLPLogFrom(logRecord)
	case control.EventSampleType:
		otlpLog = EventOTLPLogFrom(logRecord)
	case control.ConfigSampleType:
		otlpLog = ConfigOTLPLogFrom(logRecord)
	default:
		panic(fmt.Errorf("unknown OTLPLog type %T", otlpLog))
	}

	return otlpLog
}

func OTLPLogToSampleType[T OTLPLog]() control.SampleType {
	var targetSampleType control.SampleType
	var otlpLog T

	switch any(otlpLog).(type) {
	case RawSampleOTLPLog:
		targetSampleType = control.RawSampleType
	case StructDigestOTLPLog:
		targetSampleType = control.StructDigestSampleType
	case ValueDigestOTLPLog:
		targetSampleType = control.ValueDigestSampleType
	case EventOTLPLog:
		targetSampleType = control.EventSampleType
	case ConfigOTLPLog:
		targetSampleType = control.ConfigSampleType
	default:
		panic(fmt.Errorf("unknown OTLPLog type %T", otlpLog))
	}

	return targetSampleType
}

func getSampleType(logRecord plog.LogRecord) control.SampleType {
	value, ok := logRecord.Attributes().Get(lrSampleTypeKey)
	if !ok {
		return control.UnknownSampleType
	}

	return control.ParseSampleType(value.Str())
}

// Base implementation with common functionality
type baseOTLPLog struct {
	logRecord plog.LogRecord
}

func newBaseOTLPLog(logRecord plog.LogRecord, sampleType control.SampleType) baseOTLPLog {
	logRecord.Attributes().PutStr(lrSampleTypeKey, sampleType.String())

	return baseOTLPLog{
		logRecord: logRecord,
	}
}

func (b baseOTLPLog) Timestamp() time.Time {
	return b.logRecord.Timestamp().AsTime()
}

func (b baseOTLPLog) SetTimestamp(ts time.Time) {
	b.logRecord.SetTimestamp(pcommon.Timestamp(ts.UTC().UnixNano()))
}

func (b baseOTLPLog) StreamsStr() []string {
	value, ok := b.logRecord.Attributes().Get(lrSampleStreamsUIDsKey)
	if !ok {
		return []string{}
	}

	streams := make([]string, value.Slice().Len())
	for i := 0; i < value.Slice().Len(); i++ {
		streams[i] = value.Slice().At(i).Str()
	}

	return streams
}

func (b baseOTLPLog) Streams() []control.SamplerStreamUID {
	value, ok := b.logRecord.Attributes().Get(lrSampleStreamsUIDsKey)
	if !ok {
		return []control.SamplerStreamUID{}
	}

	streams := make([]control.SamplerStreamUID, value.Slice().Len())
	for i := 0; i < value.Slice().Len(); i++ {
		streams[i] = control.SamplerStreamUID(value.Slice().At(i).Str())
	}

	return streams
}

func (b baseOTLPLog) SetStreams(streams []control.SamplerStreamUID) {
	lrStreamUIDs := b.logRecord.Attributes().PutEmptySlice(lrSampleStreamsUIDsKey)
	lrStreamUIDs.EnsureCapacity(len(streams))
	for _, stream := range streams {
		lrStreamUIDs.AppendEmpty().SetStr(string(stream))
	}
}

func (b baseOTLPLog) SampleType() control.SampleType {
	return getSampleType(b.logRecord)
}

func (b baseOTLPLog) SampleKey() string {
	value, ok := b.logRecord.Attributes().Get(lrSampleKey)
	if !ok {
		return ""
	}

	return value.Str()
}

func (b baseOTLPLog) SetSampleKey(key string) {
	b.logRecord.Attributes().PutStr(lrSampleKey, key)
}

func (b baseOTLPLog) SampleEncoding() Encoding {
	value, ok := b.logRecord.Attributes().Get(lrSampleEncodingKey)
	if !ok {
		return UnknownEncoding
	}

	return ParseSampleEncoding(value.Str())
}

func (b baseOTLPLog) SampleRawData() []byte {
	if b.SampleEncoding() == JSONEncoding {
		return []byte(b.logRecord.Body().Str())
	}

	return b.logRecord.Body().Bytes().AsRaw()
}

func (b baseOTLPLog) SetSampleRawData(encoding Encoding, data []byte) {
	b.logRecord.Attributes().PutStr(lrSampleEncodingKey, encoding.String())

	if encoding == JSONEncoding {
		b.logRecord.Body().SetStr(string(data))
	} else {
		b.logRecord.Body().SetEmptyBytes().FromRaw(data)
	}
}

func (b baseOTLPLog) SampleData() (*data.Data, error) {
	var record *data.Data
	var err error

	switch b.SampleEncoding() {
	case JSONEncoding:
		record = data.NewSampleDataFromJSON(b.logRecord.Body().Str())
	default:
		err = fmt.Errorf("unknown encoding %s", b.SampleEncoding().String())
	}

	return record, err
}

func (b baseOTLPLog) Record() plog.LogRecord {
	return b.logRecord
}

// RawSampleOTLPLog implementation
type RawSampleOTLPLog struct {
	baseOTLPLog
}

func RawSampleOTLPLogFrom(logRecord plog.LogRecord) RawSampleOTLPLog {
	return RawSampleOTLPLog{
		baseOTLPLog: baseOTLPLog{
			logRecord: logRecord,
		},
	}
}

// StructDigestOTLPLog implementation
type StructDigestOTLPLog struct {
	baseOTLPLog
}

func StructDigestOTLPLogFrom(logRecord plog.LogRecord) StructDigestOTLPLog {
	return StructDigestOTLPLog{
		baseOTLPLog: baseOTLPLog{
			logRecord: logRecord,
		},
	}
}

func (e StructDigestOTLPLog) UID() control.SamplerDigestUID {
	value, ok := e.logRecord.Attributes().Get(string(DigestUID))
	if !ok {
		return ""
	}

	return control.SamplerDigestUID(value.Str())
}

func (e StructDigestOTLPLog) SetUID(uid control.SamplerDigestUID) {
	e.logRecord.Attributes().PutStr(string(DigestUID), string(uid))
}

// ValueDigestOTLPLog implementation
type ValueDigestOTLPLog struct {
	baseOTLPLog
}

func ValueDigestOTLPLogFrom(logRecord plog.LogRecord) ValueDigestOTLPLog {
	return ValueDigestOTLPLog{
		baseOTLPLog: baseOTLPLog{
			logRecord: logRecord,
		},
	}
}

func (e ValueDigestOTLPLog) UID() control.SamplerDigestUID {
	value, ok := e.logRecord.Attributes().Get(string(DigestUID))
	if !ok {
		return ""
	}

	return control.SamplerDigestUID(value.Str())
}

func (e ValueDigestOTLPLog) SetUID(uid control.SamplerDigestUID) {
	e.logRecord.Attributes().PutStr(string(DigestUID), string(uid))
}

// EventOTLPLog implementation
type EventOTLPLog struct {
	baseOTLPLog
}

func EventOTLPLogFrom(logRecord plog.LogRecord) EventOTLPLog {
	return EventOTLPLog{
		baseOTLPLog: baseOTLPLog{
			logRecord: logRecord,
		},
	}
}

func (e EventOTLPLog) UID() control.SamplerEventUID {
	value, ok := e.logRecord.Attributes().Get(string(EventUID))
	if !ok {
		return ""
	}

	return control.SamplerEventUID(value.Str())
}

func (e EventOTLPLog) SetUID(uid control.SamplerEventUID) {
	e.logRecord.Attributes().PutStr(string(EventUID), string(uid))
}

func (e EventOTLPLog) RuleExpression() string {
	value, ok := e.logRecord.Attributes().Get(string(EventRule))
	if !ok {
		return ""
	}

	return value.Str()
}

func (e EventOTLPLog) SetRuleExpression(ruleExpression string) {
	e.logRecord.Attributes().PutStr(string(EventRule), ruleExpression)
}

// ConfigOTLPLog implementation
type ConfigOTLPLog struct {
	baseOTLPLog
}

func ConfigOTLPLogFrom(logRecord plog.LogRecord) ConfigOTLPLog {
	return ConfigOTLPLog{
		baseOTLPLog: baseOTLPLog{
			logRecord: logRecord,
		},
	}
}

type SamplerOTLPLogs struct {
	resourceLogs plog.ResourceLogs
	scopeLogs    plog.ScopeLogs
}

func (s SamplerOTLPLogs) AppendRawSampleOTLPLog() RawSampleOTLPLog {
	logRecord := s.scopeLogs.LogRecords().AppendEmpty()
	return RawSampleOTLPLog{
		baseOTLPLog: newBaseOTLPLog(logRecord, control.RawSampleType),
	}
}

func (s *SamplerOTLPLogs) AppendStructDigestOTLPLog() StructDigestOTLPLog {
	logRecord := s.scopeLogs.LogRecords().AppendEmpty()
	return StructDigestOTLPLog{
		baseOTLPLog: newBaseOTLPLog(logRecord, control.StructDigestSampleType),
	}
}

func (s SamplerOTLPLogs) AppendValueDigestOTLPLog() ValueDigestOTLPLog {
	logRecord := s.scopeLogs.LogRecords().AppendEmpty()
	return ValueDigestOTLPLog{
		baseOTLPLog: newBaseOTLPLog(logRecord, control.ValueDigestSampleType),
	}
}

func (s SamplerOTLPLogs) AppendEventOTLPLog() EventOTLPLog {
	logRecord := s.scopeLogs.LogRecords().AppendEmpty()
	return EventOTLPLog{
		baseOTLPLog: newBaseOTLPLog(logRecord, control.EventSampleType),
	}
}

func (s SamplerOTLPLogs) AppendConfigOTLPLog() ConfigOTLPLog {
	logRecord := s.scopeLogs.LogRecords().AppendEmpty()
	return ConfigOTLPLog{
		baseOTLPLog: newBaseOTLPLog(logRecord, control.ConfigSampleType),
	}
}

type OTLPLogs struct {
	logs plog.Logs
}

func NewOTLPLogs() OTLPLogs {
	return OTLPLogs{
		logs: plog.NewLogs(),
	}
}

func OTLPLogsFrom(plogs plog.Logs) OTLPLogs {
	return OTLPLogs{
		logs: plogs,
	}
}

func (l *OTLPLogs) AppendSamplerOTLPLogs(resource string, sampler string) SamplerOTLPLogs {
	resourceLogs := l.logs.ResourceLogs().AppendEmpty()
	resourceLogs.Resource().Attributes().PutStr(conventions.AttributeServiceName, resource)
	resourceLogs.Resource().Attributes().PutStr(rlSamplerNameKey, sampler)

	scopeLogs := resourceLogs.ScopeLogs().AppendEmpty()

	return SamplerOTLPLogs{
		resourceLogs: resourceLogs,
		scopeLogs:    scopeLogs,
	}
}

func (l OTLPLogs) MoveAndAppendTo(dest OTLPLogs) {
	l.logs.ResourceLogs().MoveAndAppendTo(dest.logs.ResourceLogs())
}

// RemoveOTLPLogIf calls f sequentially for each element present in the otlp logs.
// f receives an interface containing a RawSampleOTLPLog, EventOTLPLog, etc.
// If f returns true, the element is removed from otlp logs.
func (l OTLPLogs) RemoveOTLPLogIf(f func(otlpLog any) bool) {
	resourceLogs := l.logs.ResourceLogs()
	for i := 0; i < resourceLogs.Len(); i++ {
		scopeLogs := resourceLogs.At(i).ScopeLogs()
		for j := 0; j < scopeLogs.Len(); j++ {
			logRecords := scopeLogs.At(j).LogRecords()
			logRecords.RemoveIf(func(logRecord plog.LogRecord) bool {
				return f(OTLPLogFrom(logRecord))
			})
		}
	}
}

func (l OTLPLogs) Logs() plog.Logs {
	return l.logs
}

func (l OTLPLogs) Len() int {
	return l.logs.ResourceLogs().Len()
}
