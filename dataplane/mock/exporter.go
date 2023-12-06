package mock

import (
	context "context"

	sample "github.com/neblic/platform/dataplane/sample"
)

// Exporter is a mock of Exporter interface.
type Exporter struct {
	Configs       []sample.ConfigOTLPLog
	StructDigests []sample.StructDigestOTLPLog
	ValueDigests  []sample.ValueDigestOTLPLog
}

// NewExporter creates a new mock instance.
func NewExporter() *Exporter {
	return &Exporter{
		Configs:       []sample.ConfigOTLPLog{},
		StructDigests: []sample.StructDigestOTLPLog{},
		ValueDigests:  []sample.ValueDigestOTLPLog{},
	}
}

// Export mocks base method.
func (m *Exporter) Export(_ context.Context, logs sample.OTLPLogs) error {
	sample.Range(logs, func(resourceName string, sampleName string, log any) {
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
func (m *Exporter) Close(_ context.Context) error {
	return nil
}
