package controlplane

import (
	"fmt"
	"io"

	"github.com/neblic/platform/controlplane/control"
	"github.com/olekukonko/tablewriter"
	"golang.org/x/exp/slices"
)

func writeTable(header []string, rows [][]string, mergeColumnsByIndex []int, writer io.Writer) {
	writer.Write([]byte("\n"))

	table := tablewriter.NewWriter(writer)
	table.SetHeader(header)
	if mergeColumnsByIndex != nil {
		table.SetAutoMergeCellsByColumnIndex(mergeColumnsByIndex)
	}
	table.SetRowLine(true)
	table.SetCenterSeparator("|")
	table.AppendBulk(rows)
	table.Render()
}

// ListResourcesView shows a table with all the resources
type ListResourcesView struct {
	header           []string
	deduplicatedRows map[string]struct{}
	rows             [][]string
}

func NewListResourcesView() *ListResourcesView {
	return &ListResourcesView{
		header:           []string{"Resource"},
		deduplicatedRows: map[string]struct{}{},
		rows:             [][]string{},
	}
}

func (lrv *ListResourcesView) AddSampler(sampler *control.Sampler) {
	if _, ok := lrv.deduplicatedRows[sampler.Resource]; ok {
		return
	}

	lrv.deduplicatedRows[sampler.Resource] = struct{}{}
	lrv.rows = append(lrv.rows, []string{sampler.Resource})
}

func (lrv *ListResourcesView) Render(writer io.Writer) {
	// Sort rows by resource
	slices.SortStableFunc(lrv.rows, func(a []string, b []string) int {
		// Order rows by resource (first entry)
		return cmpStrings(a[0], b[0])
	})

	writeTable(lrv.header, lrv.rows, nil, writer)
}

// ListSamplersView shows a table with all samplers and their stats.
type ListSamplersView struct {
	header []string
	rows   [][]string
}

func NewListSamplersView() *ListSamplersView {
	return &ListSamplersView{
		header: []string{"Resource", "Sampler", "Stats"},
		rows:   [][]string{},
	}
}

func (lsv *ListSamplersView) AddSampler(sampler *control.Sampler) {
	stats := sampler.SamplingStats

	lsv.rows = append(lsv.rows,
		[]string{
			sampler.Resource,
			sampler.Name,
			fmt.Sprintf("Evaluated: %d, Exported: %d, Digested: %d", stats.SamplesEvaluated, stats.SamplesExported, stats.SamplesDigested)})
}

func (lsv *ListSamplersView) Render(writer io.Writer) {
	// Sort rows by resource, rows with the same resource must be ordered by sampler.
	slices.SortStableFunc(lsv.rows, func(a []string, b []string) int {
		if a[0] != b[0] {
			// The resource is not the same in the two rows. Order by resource (first entry)
			return cmpStrings(a[0], b[0])
		} else {
			// The resource is the same in the two rows. Order by sampler (second entry)
			return cmpStrings(a[1], b[1])
		}
	})

	writeTable(lsv.header, lsv.rows, []int{0}, writer)
}

// ListSamplersConfigView shows a table with all samplers and their config.
type ListSamplersConfigView struct {
	header []string
	rows   [][]string
}

func NewListSamplersConfigView() *ListSamplersConfigView {
	return &ListSamplersConfigView{
		header: []string{"Resource", "Sampler", "Limiter In", "Sampling In", "Limiter Out"},
		rows:   [][]string{},
	}
}

func (lscv *ListSamplersConfigView) AddSampler(sampler *control.Sampler) {
	limiterIn := "none"
	if sampler.Config.LimiterIn != nil {
		limiterIn = fmt.Sprintf("%d", sampler.Config.LimiterIn.Limit)
	}

	samplingIn := "none"
	if sampler.Config.SamplingIn != nil {
		switch sampler.Config.SamplingIn.SamplingType {
		case control.DeterministicSamplingType:
			samplingIn = fmt.Sprintf("Type: Deterministic, SampleRate: %d, SampleEmtpyDeterminant: %t",
				sampler.Config.SamplingIn.DeterministicSampling.SampleRate,
				sampler.Config.SamplingIn.DeterministicSampling.SampleEmptyDeterminant,
			)
		default:
			samplingIn = "Type: Unknown"
		}
	}

	limiterOut := "none"
	if sampler.Config.LimiterOut != nil {
		limiterOut = fmt.Sprintf("%d", sampler.Config.LimiterOut.Limit)
	}

	lscv.rows = append(lscv.rows,
		[]string{
			sampler.Resource,
			sampler.Name,
			limiterIn,
			samplingIn,
			limiterOut,
		},
	)
}

func (lscv *ListSamplersConfigView) Render(writer io.Writer) {
	// Sort rows by resource, rows with the same resource must be ordered by sampler.
	slices.SortStableFunc(lscv.rows, func(a []string, b []string) int {
		if a[0] != b[0] {
			// The resource is not the same in the two rows. Order by resource (first entry)
			return cmpStrings(a[0], b[0])
		} else {
			// The resource is the same in the two rows. Order by sampler (second entry)
			return cmpStrings(a[1], b[1])
		}
	})

	writeTable(lscv.header, lscv.rows, []int{0}, writer)
}

// ListStreamsView shows a table with all the configured streams for each resource and sampler. Data
// is ordered by resource and sampler.
type ListStreamsView struct {
	header []string
	rows   [][]string
}

func NewListStreamsView() *ListStreamsView {
	return &ListStreamsView{
		header: []string{"Resource", "Sampler", "Stream"},
		rows:   [][]string{},
	}
}

func (lsv *ListStreamsView) AddSampler(sampler *control.Sampler) {
	for _, stream := range sampler.Config.Streams {
		streamStr := fmt.Sprintf("Name: %s, Rule: %s, ExportRawSamples: %t",
			stream.Name,
			stream.StreamRule,
			stream.ExportRawSamples,
		)
		if stream.Keyed.Enabled {
			streamStr += fmt.Sprintf(", Keyed: {TTL: %s, MaxKeys: %d}", stream.Keyed.TTL.String(), stream.Keyed.MaxKeys)
		}
		lsv.rows = append(lsv.rows, []string{sampler.Resource, sampler.Name, streamStr})
	}
}

func (lsv *ListStreamsView) Render(writer io.Writer) {
	// Sort rows by resource, rows with the same resource must be ordered by sampler.
	slices.SortStableFunc(lsv.rows, func(a []string, b []string) int {
		if a[0] != b[0] {
			// The resource is not the same in the two rows. Order by resource (first entry)
			return cmpStrings(a[0], b[0])
		} else {
			// The resource is the same in the two rows. Order by sampler (second entry)
			return cmpStrings(a[1], b[1])
		}
	})

	writeTable(lsv.header, lsv.rows, []int{0, 1}, writer)
}

// ListDigestsView shows a table with all the configured digests for each resource and sampler. Data
// is ordered by resource and sampler.
type ListDigestsView struct {
	header []string
	rows   [][]string
}

func NewListDigestsView() *ListDigestsView {
	return &ListDigestsView{
		header: []string{"Resource", "Sampler", "Digest"},
		rows:   [][]string{},
	}
}

func (ldv *ListDigestsView) AddSampler(sampler *control.Sampler) {
	for _, stream := range sampler.Config.Streams {
		for _, digest := range sampler.Config.Digests {
			if stream.UID == digest.StreamUID {

				var typeInfo string
				switch digest.Type {
				case control.DigestTypeSt:
					typeInfo = fmt.Sprintf("Type: Structure, MaxProcessedFields: %d", digest.St.MaxProcessedFields)
				case control.DigestTypeValue:
					typeInfo = fmt.Sprintf("Type: Value, MaxProcessedFields: %d", digest.Value.MaxProcessedFields)
				}

				digest := fmt.Sprintf("Name: %s, Stream: %s, FlushPeriod: %s, %s",
					digest.Name,
					stream.Name,
					digest.FlushPeriod,
					typeInfo,
				)

				ldv.rows = append(ldv.rows, []string{
					sampler.Resource,
					sampler.Name,
					digest,
				})
			}
		}
	}
}

func (ldv *ListDigestsView) Render(writer io.Writer) {
	// Sort rows by resource, rows with the same resource must be ordered by sampler.
	slices.SortStableFunc(ldv.rows, func(a []string, b []string) int {
		if a[0] != b[0] {
			// The resource is not the same in the two rows. Order by resource (first entry)
			return cmpStrings(a[0], b[0])
		} else {
			// The resource is the same in the two rows. Order by sampler (second entry)
			return cmpStrings(a[1], b[1])
		}
	})

	writeTable(ldv.header, ldv.rows, []int{0, 1}, writer)
}

// ListEventsView shows a table with all the events for each resource and sampler. Data
// is ordered by resource and sampler.
type ListEventsView struct {
	header []string
	rows   [][]string
}

func NewListEventsView() *ListEventsView {
	return &ListEventsView{
		header: []string{"Resource", "Sampler", "Events"},
		rows:   [][]string{},
	}
}

func (lev *ListEventsView) AddSampler(sampler *control.Sampler) {
	for _, stream := range sampler.Config.Streams {
		for _, event := range sampler.Config.Events {
			if stream.UID == event.StreamUID {
				eventInfo := fmt.Sprintf("Name: %s, Stream: %s, DataType: %s, Rule: %s",
					event.Name,
					stream.Name,
					event.SampleType,
					event.Rule,
				)
				lev.rows = append(lev.rows, []string{
					sampler.Resource,
					sampler.Name,
					eventInfo,
				})
			}
		}
	}
}

func (lev *ListEventsView) Render(writer io.Writer) {
	// Sort rows by resource, rows with the same resource must be ordered by sampler.
	slices.SortStableFunc(lev.rows, func(a []string, b []string) int {
		if a[0] != b[0] {
			// The resource is not the same in the two rows. Order by resource (first entry)
			return cmpStrings(a[0], b[0])
		} else {
			// The resource is the same in the two rows. Order by sampler (second entry)
			return cmpStrings(a[1], b[1])
		}
	})

	writeTable(lev.header, lev.rows, []int{0, 1}, writer)
}
