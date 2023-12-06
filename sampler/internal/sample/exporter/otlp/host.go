package otlp

import (
	"go.opentelemetry.io/collector/component"
)

type Host struct{}

// ReportFatalError does not have to do anything because otlpexporter.LogsExporter doesn't report errors
// if it ever reports fatal errors, it should notify the Exporter.
func (h *Host) ReportFatalError(_ error) {}

func (h *Host) GetFactory(_ component.Kind, _ component.Type) component.Factory {
	return nil
}
func (h *Host) GetExtensions() map[component.ID]component.Component {
	return make(map[component.ID]component.Component)
}
func (h *Host) GetExporters() map[component.DataType]map[component.ID]component.Component {
	return make(map[component.DataType]map[component.ID]component.Component)
}
