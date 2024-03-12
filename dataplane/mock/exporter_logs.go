package mock

import (
	context "context"

	sample "github.com/neblic/platform/dataplane/sample"
)

// LogsExporter is a mock of LogsExporter interface.
type LogsExporter struct {
	Configs       []sample.ConfigOTLPLog
	StructDigests []sample.StructDigestOTLPLog
	ValueDigests  []sample.ValueDigestOTLPLog
}

// NewLogsConsumer creates a new mock instance.
func NewLogsExporter() *LogsExporter {
	return &LogsExporter{
		Configs:       []sample.ConfigOTLPLog{},
		StructDigests: []sample.StructDigestOTLPLog{},
		ValueDigests:  []sample.ValueDigestOTLPLog{},
	}
}

// Consume mocks base method.
func (m *LogsExporter) Export(_ context.Context, logs sample.OTLPLogs) error {
	sample.Range(logs, func(resourceName string, sampleName string, log sample.OTLPLog) {
		switch v := log.(type) {
		case sample.ConfigOTLPLog:
			m.Configs = append(m.Configs, v)
		case sample.StructDigestOTLPLog:
			m.StructDigests = append(m.StructDigests, v)
		case sample.ValueDigestOTLPLog:
			m.ValueDigests = append(m.ValueDigests, v)
		}
	})

	return nil
}

// Close mocks base method.
func (m *LogsExporter) Close(_ context.Context) error {
	return nil
}
