package mock

import (
	context "context"

	"github.com/neblic/platform/dataplane/metric"
)

// MetricsExporter is a mock of MetricsExporter interface.
type MetricsExporter struct {
	Metrics []metric.Metrics
}

// NewMetricsExporter creates a new mock instance.
func NewMetricsExporter() *MetricsExporter {
	return &MetricsExporter{
		Metrics: []metric.Metrics{},
	}
}

// Consume mocks base method.
func (m *MetricsExporter) Export(_ context.Context, _ metric.Metrics) error {
	return nil
}

// Close mocks base method.
func (m *MetricsExporter) Close(_ context.Context) error {
	return nil
}
