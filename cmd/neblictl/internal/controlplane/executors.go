package controlplane

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/neblic/platform/cmd/neblictl/internal"
	"github.com/neblic/platform/cmd/neblictl/internal/interpoler"
	"github.com/neblic/platform/controlplane/control"
	"github.com/olekukonko/tablewriter"
	"golang.org/x/exp/slices"
)

type named interface {
	GetName() string
}

func getEntryByName[U comparable, T named](entries map[U]T, name string) (T, bool) {
	var entry T
	var ok bool
	for _, entry = range entries {
		if entry.GetName() == name {
			ok = true
			break
		}
	}
	return entry, ok
}

func writeTable(header []string, rows [][]string, mergeColumnsByIndex []int, writer *internal.Writer) {
	writer.WriteString("\n")

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

func cmpStrings(a, b string) int {
	if a < b {
		return -1
	} else if a > b {
		return 1
	} else {
		return 0
	}
}

type Executors struct {
	controlPlaneClient *Client
}

func NewExecutors(controlPlaneClient *Client) *Executors {
	return &Executors{
		controlPlaneClient: controlPlaneClient,
	}
}

func (e *Executors) ResourcesList(ctx context.Context, parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
	// Get all samplers
	samplers, err := e.controlPlaneClient.getAllSamplers(ctx, false)

	// Build deduplicated list of rows
	header := []string{"Resource"}
	rows := [][]string{}
	resources := map[string]bool{}
	for sampler := range samplers {
		if _, ok := resources[sampler.resource]; !ok {
			rows = append(rows, []string{sampler.resource})
			resources[sampler.resource] = true
		}
	}

	// Sort rows by resource
	slices.SortStableFunc(rows, func(a []string, b []string) int {
		// Order rows by resource (first entry)
		return cmpStrings(a[0], b[0])
	})

	// Write table
	writeTable(header, rows, nil, writer)

	if err != nil && errors.Is(err, context.Canceled) {
		writer.WriteStringf("\n\nWarn: internal state was not updated because %s, results could be outdated\n", err)
	}

	return nil
}

func (e *Executors) SamplersList(ctx context.Context, parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
	samplerParameter, _ := parameters.Get("sampler-name")
	resourceParameter, _ := parameters.Get("resource-name")

	resourceAndSamplers, err := e.controlPlaneClient.getSamplers(ctx, resourceParameter.Value, samplerParameter.Value, "*", false)
	if err != nil {
		return err
	}

	header := []string{"Resource", "Sampler", "Stats"}
	rows := [][]string{}
	for resourceAndSamplerEntry, samplerData := range resourceAndSamplers {
		statsInfo := fmt.Sprintf("Evaluated: %d, Exported: %d, Digested: %d",
			samplerData.SamplingStats.SamplesEvaluated,
			samplerData.SamplingStats.SamplesExported,
			samplerData.SamplingStats.SamplesDigested)
		rows = append(rows, []string{resourceAndSamplerEntry.resource,
			resourceAndSamplerEntry.sampler,
			statsInfo,
		})
	}

	// Sort rows by resource, rows with the same resource must be ordered by sampler.
	slices.SortStableFunc(rows, func(a []string, b []string) int {
		if a[0] != b[0] {
			// The resource is not the same in the two rows. Order by resource (first entry)
			return cmpStrings(a[0], b[0])
		} else {
			// The resource is the same in the two rows. Order by sampler (second entry)
			return cmpStrings(a[1], b[1])
		}
	})

	writeTable(header, rows, []int{0}, writer)

	if err != nil && errors.Is(err, context.Canceled) {
		writer.WriteStringf("\n\nWarn: internal state was not updated because %s, results could be outdated\n", err)
	}

	return nil
}

func (e *Executors) SamplersListConfig(ctx context.Context, parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
	samplerParameter, _ := parameters.Get("sampler-name")
	resourceParameter, _ := parameters.Get("resource-name")

	resourceAndSamplers, err := e.controlPlaneClient.getSamplers(ctx, resourceParameter.Value, samplerParameter.Value, "*", false)
	if err != nil {
		return err
	}

	header := []string{"Resource", "Sampler", "Limiter In", "Sampling In", "Limiter Out"}
	rows := [][]string{}
	for resourceAndSamplerEntry, samplerData := range resourceAndSamplers {
		limiterIn := "default"
		if samplerData.Config.LimiterIn != nil {
			limiterIn = fmt.Sprintf("%d", samplerData.Config.LimiterIn.Limit)
		}

		samplingIn := "default"
		if samplerData.Config.SamplingIn != nil {
			switch samplerData.Config.SamplingIn.SamplingType {
			case control.DeterministicSamplingType:
				samplingIn = fmt.Sprintf("Type: Deterministic, SampleRate: %d, SampleEmtpyDeterminant: %t",
					samplerData.Config.SamplingIn.DeterministicSampling.SampleRate,
					samplerData.Config.SamplingIn.DeterministicSampling.SampleEmptyDeterminant,
				)
			default:
				samplingIn = "Type: Unknown"
			}
		}

		limiterOut := "default"
		if samplerData.Config.LimiterOut != nil {
			limiterOut = fmt.Sprintf("%d", samplerData.Config.LimiterOut.Limit)
		}

		rows = append(rows, []string{resourceAndSamplerEntry.resource,
			resourceAndSamplerEntry.sampler,
			limiterIn,
			samplingIn,
			limiterOut,
		})
	}

	// Sort rows by resource, rows with the same resource must be ordered by sampler.
	slices.SortStableFunc(rows, func(a []string, b []string) int {
		if a[0] != b[0] {
			// The resource is not the same in the two rows. Order by resource (first entry)
			return cmpStrings(a[0], b[0])
		} else {
			// The resource is the same in the two rows. Order by sampler (second entry)
			return cmpStrings(a[1], b[1])
		}
	})

	writeTable(header, rows, []int{0}, writer)

	if err != nil && errors.Is(err, context.Canceled) {
		writer.WriteStringf("\n\nWarn: internal state was not updated because %s, results could be outdated\n", err)
	}

	return nil
}

func (e *Executors) StreamsList(ctx context.Context, parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
	// Get options
	samplerParameter, _ := parameters.Get("sampler-name")
	resourceParameter, _ := parameters.Get("resource-name")

	// Compute list of targeted resources and samplers
	resourceAndSamplers, err := e.controlPlaneClient.getSamplers(ctx, resourceParameter.Value, samplerParameter.Value, "*", false)

	// Build table rows
	header := []string{"Resource", "Sampler", "Stream"}
	rows := [][]string{}
	for _, samplerData := range resourceAndSamplers {
		for _, stream := range samplerData.Config.Streams {
			streamInfo := fmt.Sprintf("Name: %s, Rule: %s, ExportRawSamples: %t",
				stream.Name,
				stream.StreamRule,
				stream.ExportRawSamples,
			)
			rows = append(rows, []string{samplerData.Resource, samplerData.Name, streamInfo})
		}
	}

	// Sort rows by resource, rows with the same resource must be ordered by sampler.
	slices.SortStableFunc(rows, func(a []string, b []string) int {
		if a[0] != b[0] {
			// The resource is not the same in the two rows. Order by resource (first entry)
			return cmpStrings(a[0], b[0])
		} else {
			// The resource is the same in the two rows. Order by sampler (second entry)
			return cmpStrings(a[1], b[1])
		}
	})

	// Write table
	writeTable(header, rows, []int{0, 1}, writer)

	if err != nil && errors.Is(err, context.Canceled) {
		writer.WriteStringf("\n\nWarn: internal state was not updated because %s, results could be outdated\n", err)
	}

	return nil
}

func (e *Executors) StreamsCreate(ctx context.Context, parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
	resourceParameter, _ := parameters.Get("resource-name")
	samplerParameter, _ := parameters.Get("sampler-name")
	streamRuleParameter, _ := parameters.Get("rule")
	streamNameParameter, _ := parameters.Get("stream-name")

	exportRawParameter, _ := parameters.Get("export-raw")
	exportRawBool, err := strconv.ParseBool(exportRawParameter.Value)
	if err != nil {
		return fmt.Errorf("export-raw must be a boolean")
	}

	// Compute list of targeted resources and samplers
	resourceAndSamplers, err := e.controlPlaneClient.getSamplers(ctx, resourceParameter.Value, samplerParameter.Value, "*", false)
	if err != nil {
		return err
	}

	// Create rules one by one
	for resourceAndSamplerEntry, samplerControl := range resourceAndSamplers {
		// Check that the stream does not exist
		_, ok := getEntryByName[control.SamplerStreamUID](samplerControl.Config.Streams, streamNameParameter.Value)
		if ok {
			writer.WriteStringf("%s.%s: Stream already exists\n", resourceAndSamplerEntry.resource, resourceAndSamplerEntry.sampler)
			continue
		}

		update := &control.SamplerConfigUpdate{
			StreamUpdates: []control.StreamUpdate{
				{
					Op: control.StreamUpsert,
					Stream: control.Stream{
						UID:  control.SamplerStreamUID(uuid.New().String()),
						Name: streamNameParameter.Value,
						StreamRule: control.Rule{
							Lang:       control.SrlCel,
							Expression: streamRuleParameter.Value,
						},
						ExportRawSamples: exportRawBool,
					},
				},
			},
		}

		// Propagate new configuration
		err := e.controlPlaneClient.setSamplerConfig(ctx, resourceAndSamplerEntry.sampler, resourceAndSamplerEntry.resource, update)
		if err != nil {
			writer.WriteStringf("%s.%s: Could not create the stream because %v\n", resourceAndSamplerEntry.resource, resourceAndSamplerEntry.sampler, err)
			continue
		}

		writer.WriteStringf("%s.%s: Stream successfully created\n", resourceAndSamplerEntry.resource, resourceAndSamplerEntry.sampler)

	}

	return nil
}

func (e *Executors) StreamsUpdate(ctx context.Context, parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
	resourceParameter, _ := parameters.Get("resource-name")
	samplerParameter, _ := parameters.Get("sampler-name")
	streamNameParameter, _ := parameters.Get("stream-name")
	updatedRuleParameter, _ := parameters.Get("updated-rule")

	exportRawParameter, _ := parameters.Get("export-raw")
	exportRawBool, err := strconv.ParseBool(exportRawParameter.Value)
	if err != nil {
		return fmt.Errorf("export-raw must be a boolean")
	}

	// Compute list of targeted resources and samplers
	resourceAndSamplers, err := e.controlPlaneClient.getSamplers(ctx, resourceParameter.Value, samplerParameter.Value, streamNameParameter.Value, false)
	if err != nil {
		return err
	}

	// Update streams one by one
	for resourceAndSamplerEntry, samplerControl := range resourceAndSamplers {

		// Find stream UID
		stream, ok := getEntryByName(samplerControl.Config.Streams, streamNameParameter.Value)
		if !ok {
			writer.WriteStringf("%s.%s: Stream does not exist\n", resourceAndSamplerEntry.resource, resourceAndSamplerEntry.sampler)
			continue
		}

		// Modify sampling rule to existing config
		update := &control.SamplerConfigUpdate{
			StreamUpdates: []control.StreamUpdate{
				{
					Op: control.StreamUpsert,
					Stream: control.Stream{
						UID:  stream.UID,
						Name: streamNameParameter.Value,
						StreamRule: control.Rule{
							Lang:       control.SrlCel,
							Expression: updatedRuleParameter.Value,
						},
						ExportRawSamples: exportRawBool,
					},
				},
			},
		}

		// Propagate new configuration
		err := e.controlPlaneClient.setSamplerConfig(ctx, resourceAndSamplerEntry.sampler, resourceAndSamplerEntry.resource, update)
		if err != nil {
			writer.WriteStringf("%s.%s: Could not update the stream because %v\n", resourceAndSamplerEntry.resource, resourceAndSamplerEntry.sampler, err)
			continue
		}

		// Write output
		writer.WriteStringf("%s.%s: Stream successfully updated\n", resourceAndSamplerEntry.resource, resourceAndSamplerEntry.sampler)
	}

	return nil
}

func (e *Executors) StreamsDelete(ctx context.Context, parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
	// Get options
	resourceParameter, _ := parameters.Get("resource-name")
	samplerParameter, _ := parameters.Get("sampler-name")
	streamNameParameter, _ := parameters.Get("stream-name")

	// Compute list of targeted resources and samplers
	resourceAndSamplers, err := e.controlPlaneClient.getSamplers(ctx, resourceParameter.Value, samplerParameter.Value, streamNameParameter.Value, false)
	if err != nil {
		return err
	}

	// Delete streams one by one
	for resourceAndSamplerEntry, samplerControl := range resourceAndSamplers {

		stream, ok := getEntryByName(samplerControl.Config.Streams, streamNameParameter.Value)
		if !ok {
			writer.WriteStringf("%s.%s: Stream does not exist\n", resourceAndSamplerEntry.resource, resourceAndSamplerEntry.sampler)
			continue
		}

		// Modify sampling rule to existing config
		update := &control.SamplerConfigUpdate{
			StreamUpdates: []control.StreamUpdate{
				{
					Op: control.StreamDelete,
					Stream: control.Stream{
						UID: stream.UID,
					},
				},
			},
		}

		// Propagate new configuration
		err := e.controlPlaneClient.setSamplerConfig(ctx, resourceAndSamplerEntry.sampler, resourceAndSamplerEntry.resource, update)
		if err != nil {
			writer.WriteStringf("%s.%s: Could not delete the sampling rule because %v\n", resourceAndSamplerEntry.resource, resourceAndSamplerEntry.sampler, err)
			continue
		}

		// Write output
		writer.WriteStringf("%s.%s: Rule successfully deleted\n", resourceAndSamplerEntry.resource, resourceAndSamplerEntry.sampler)
	}

	return nil
}

func (e *Executors) setMultipleSamplersConfig(ctx context.Context, parameters interpoler.ParametersWithValue, writer *internal.Writer, updateGen func(*control.Sampler) (*control.SamplerConfigUpdate, error)) error {
	samplerParameter, _ := parameters.Get("sampler-name")
	resourceParameter, _ := parameters.Get("resource-name")
	streamNameValue := "*"
	streamNameParameter, streamNameParameterOk := parameters.Get("stream-name")
	if streamNameParameterOk {
		streamNameValue = streamNameParameter.Value
	}

	resourceAndSamplers, err := e.controlPlaneClient.getSamplers(ctx, resourceParameter.Value, samplerParameter.Value, streamNameValue, false)
	if err != nil {
		return err
	}

	for resourceAndSamplerEntry, sampler := range resourceAndSamplers {
		update, err := updateGen(sampler)
		if err != nil {
			writer.WriteStringf("%s.%s: Could not update sampler config. %v\n", resourceAndSamplerEntry.resource, resourceAndSamplerEntry.sampler, err)
			continue
		}

		if err := e.controlPlaneClient.setSamplerConfig(ctx, resourceAndSamplerEntry.sampler, resourceAndSamplerEntry.resource, update); err != nil {
			writer.WriteStringf("%s.%s: Could not update sampler config. %v\n", resourceAndSamplerEntry.resource, resourceAndSamplerEntry.sampler, err)
			continue
		}

		writer.WriteStringf("%s.%s: Sampler configuration successfully updated\n", resourceAndSamplerEntry.resource, resourceAndSamplerEntry.sampler)
	}

	return nil
}

func (e *Executors) SamplersLimiterInSet(ctx context.Context, parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
	limitParameter, _ := parameters.Get("limit")
	limitInt32, err := limitParameter.AsInt32()
	if err != nil {
		return fmt.Errorf("limit must be an integer")
	}

	updateGen := func(_ *control.Sampler) (*control.SamplerConfigUpdate, error) {
		return &control.SamplerConfigUpdate{
			LimiterIn: &control.LimiterConfig{
				Limit: limitInt32,
			},
		}, nil
	}

	return e.setMultipleSamplersConfig(ctx, parameters, writer, updateGen)
}

func (e *Executors) SamplersLimiterInUnset(ctx context.Context, parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
	updateGen := func(_ *control.Sampler) (*control.SamplerConfigUpdate, error) {
		return &control.SamplerConfigUpdate{
			Reset: control.SamplerConfigUpdateReset{
				LimiterIn: true,
			},
		}, nil
	}

	return e.setMultipleSamplersConfig(ctx, parameters, writer, updateGen)
}

func (e *Executors) SamplersLimiterOutSet(ctx context.Context, parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
	limitParameter, _ := parameters.Get("limit")
	limitInt32, err := limitParameter.AsInt32()
	if err != nil {
		return fmt.Errorf("limit must be an integer")
	}

	updateGen := func(_ *control.Sampler) (*control.SamplerConfigUpdate, error) {
		return &control.SamplerConfigUpdate{
			LimiterOut: &control.LimiterConfig{
				Limit: limitInt32,
			},
		}, nil
	}

	return e.setMultipleSamplersConfig(ctx, parameters, writer, updateGen)
}

func (e *Executors) SamplersLimiterOutUnset(ctx context.Context, parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
	updateGen := func(_ *control.Sampler) (*control.SamplerConfigUpdate, error) {
		return &control.SamplerConfigUpdate{
			Reset: control.SamplerConfigUpdateReset{
				LimiterOut: true,
			},
		}, nil
	}

	return e.setMultipleSamplersConfig(ctx, parameters, writer, updateGen)
}

func (e *Executors) SamplersSamplerInSetDeterministic(ctx context.Context, parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
	sampleRateParameter, _ := parameters.Get("sample_rate")
	sampleRateInt32, err := sampleRateParameter.AsInt32()
	if err != nil {
		return fmt.Errorf("sample_rate must be an integer")
	}

	sampleEmptyDetParameter, _ := parameters.Get("sample_empty_determinant")
	sampleEmptyDetBool, err := sampleEmptyDetParameter.AsBool()
	if err != nil {
		return fmt.Errorf("sample_empty_determinant must be a boolean")
	}

	updateGen := func(_ *control.Sampler) (*control.SamplerConfigUpdate, error) {
		return &control.SamplerConfigUpdate{
			SamplingIn: &control.SamplingConfig{
				SamplingType: control.DeterministicSamplingType,
				DeterministicSampling: control.DeterministicSamplingConfig{
					SampleRate:             sampleRateInt32,
					SampleEmptyDeterminant: sampleEmptyDetBool,
				},
			},
		}, nil
	}

	return e.setMultipleSamplersConfig(ctx, parameters, writer, updateGen)
}

func (e *Executors) SamplersSamplerInUnset(ctx context.Context, parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
	updateGen := func(_ *control.Sampler) (*control.SamplerConfigUpdate, error) {
		return &control.SamplerConfigUpdate{
			Reset: control.SamplerConfigUpdateReset{
				SamplingIn: true,
			},
		}, nil
	}

	return e.setMultipleSamplersConfig(ctx, parameters, writer, updateGen)
}

func (e *Executors) DigestsList(ctx context.Context, parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
	samplerParameter, _ := parameters.Get("sampler-name")
	resourceParameter, _ := parameters.Get("resource-name")

	resourceAndSamplers, err := e.controlPlaneClient.getSamplers(ctx, resourceParameter.Value, samplerParameter.Value, "*", false)
	if err != nil {
		return err
	}
	header := []string{"Resource", "Sampler", "Digest"}
	rows := [][]string{}

	for resourceAndSamplerEntry, samplerData := range resourceAndSamplers {
		for _, stream := range samplerData.Config.Streams {
			for _, digest := range samplerData.Config.Digests {
				if stream.UID == digest.StreamUID {

					var typeInfo string
					switch digest.Type {
					case control.DigestTypeSt:
						typeInfo = fmt.Sprintf("Type: Structure, MaxProcessedFields: %d", digest.St.MaxProcessedFields)
					case control.DigestTypeValue:
						typeInfo = fmt.Sprintf("Type: Value, MaxProcessedFields: %d", digest.Value.MaxProcessedFields)
					}

					digestInfo := fmt.Sprintf("Name: %s, Stream: %s, FlushPeriod: %s, %s",
						digest.Name,
						stream.Name,
						digest.FlushPeriod,
						typeInfo,
					)

					rows = append(rows, []string{
						resourceAndSamplerEntry.resource,
						resourceAndSamplerEntry.sampler,
						digestInfo,
					})
				}
			}
		}
	}

	// Sort rows by resource, rows with the same resource must be ordered by sampler.
	slices.SortStableFunc(rows, func(a []string, b []string) int {
		if a[0] != b[0] {
			// The resource is not the same in the two rows. Order by resource (first entry)
			return cmpStrings(a[0], b[0])
		} else {
			// The resource is the same in the two rows. Order by sampler (second entry)
			return cmpStrings(a[1], b[1])
		}
	})

	// Write table
	writeTable(header, rows, []int{0, 1}, writer)

	if err != nil && errors.Is(err, context.Canceled) {
		writer.WriteStringf("\n\nWarn: internal state was not updated because %s, results could be outdated\n", err)
	}

	return nil
}

func (e *Executors) DigestsStructureCreate(ctx context.Context, parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
	digestNameParameter, _ := parameters.Get("digest-name")

	streamNameParameter, _ := parameters.Get("stream-name")

	flushPeriodParameter, _ := parameters.Get("flush-period")
	flushPeriodInt32, err := flushPeriodParameter.AsInt32()
	if err != nil {
		return fmt.Errorf("flush-period must be an integer")
	}

	maxProcessedFieldsParameter, _ := parameters.Get("max-processed-fields")
	maxProcessedFieldsInt32, err := maxProcessedFieldsParameter.AsInt32()
	if err != nil {
		return fmt.Errorf("flush-period must be an integer")
	}

	updateGen := func(samplerControl *control.Sampler) (*control.SamplerConfigUpdate, error) {

		_, ok := getEntryByName(samplerControl.Config.Digests, digestNameParameter.Value)
		if ok {
			return nil, fmt.Errorf("Digest already exists")
		}

		stream, ok := getEntryByName(samplerControl.Config.Streams, streamNameParameter.Value)
		if !ok {
			return nil, fmt.Errorf("Stream does not exist")
		}

		return &control.SamplerConfigUpdate{
			DigestUpdates: []control.DigestUpdate{
				{
					Op: control.DigestUpsert,
					Digest: control.Digest{
						UID:         control.SamplerDigestUID(uuid.New().String()),
						Name:        digestNameParameter.Value,
						StreamUID:   stream.UID,
						FlushPeriod: time.Second * time.Duration(flushPeriodInt32),
						Type:        control.DigestTypeSt,
						St: control.DigestSt{
							MaxProcessedFields: int(maxProcessedFieldsInt32),
						},
					},
				},
			},
		}, nil
	}

	return e.setMultipleSamplersConfig(ctx, parameters, writer, updateGen)
}

func (e *Executors) DigestsStructureUpdate(ctx context.Context, parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
	digestNameParameter, _ := parameters.Get("digest-name")
	streamNameParameter, _ := parameters.Get("stream-name")

	flushPeriodParameter, _ := parameters.Get("flush-period")
	flushPeriodInt32, err := flushPeriodParameter.AsInt32()
	if err != nil {
		return fmt.Errorf("flush-period must be an integer")
	}

	maxProcessedFieldsParameter, _ := parameters.Get("max-processed-fields")
	maxProcessedFieldsInt32, err := maxProcessedFieldsParameter.AsInt32()
	if err != nil {
		return fmt.Errorf("flush-period must be an integer")
	}

	updateGen := func(samplerControl *control.Sampler) (*control.SamplerConfigUpdate, error) {

		digest, ok := getEntryByName(samplerControl.Config.Digests, digestNameParameter.Value)
		if !ok {
			return nil, fmt.Errorf("Digest does not exist")
		}

		stream, ok := getEntryByName(samplerControl.Config.Streams, streamNameParameter.Value)
		if !ok {
			return nil, fmt.Errorf("Stream does not exist")
		}

		return &control.SamplerConfigUpdate{
			DigestUpdates: []control.DigestUpdate{
				{
					Op: control.DigestUpsert,
					Digest: control.Digest{
						UID:         digest.UID,
						Name:        digestNameParameter.Value,
						StreamUID:   stream.UID,
						FlushPeriod: time.Second * time.Duration(flushPeriodInt32),
						Type:        control.DigestTypeSt,
						St: control.DigestSt{
							MaxProcessedFields: int(maxProcessedFieldsInt32),
						},
					},
				},
			},
		}, nil
	}

	return e.setMultipleSamplersConfig(ctx, parameters, writer, updateGen)
}

func (e *Executors) DigestsValueCreate(ctx context.Context, parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
	digestNameParameter, _ := parameters.Get("digest-name")

	streamNameParameter, _ := parameters.Get("stream-name")

	flushPeriodParameter, _ := parameters.Get("flush-period")
	flushPeriodInt32, err := flushPeriodParameter.AsInt32()
	if err != nil {
		return fmt.Errorf("flush-period must be an integer")
	}

	maxProcessedFieldsParameter, _ := parameters.Get("max-processed-fields")
	maxProcessedFieldsInt32, err := maxProcessedFieldsParameter.AsInt32()
	if err != nil {
		return fmt.Errorf("flush-period must be an integer")
	}

	updateGen := func(samplerControl *control.Sampler) (*control.SamplerConfigUpdate, error) {

		_, ok := getEntryByName(samplerControl.Config.Digests, digestNameParameter.Value)
		if ok {
			return nil, fmt.Errorf("Digest already exists")
		}

		stream, ok := getEntryByName(samplerControl.Config.Streams, streamNameParameter.Value)
		if !ok {
			return nil, fmt.Errorf("Stream does not exist")
		}

		return &control.SamplerConfigUpdate{
			DigestUpdates: []control.DigestUpdate{
				{
					Op: control.DigestUpsert,
					Digest: control.Digest{
						UID:         control.SamplerDigestUID(uuid.New().String()),
						Name:        digestNameParameter.Value,
						StreamUID:   stream.UID,
						FlushPeriod: time.Second * time.Duration(flushPeriodInt32),
						Type:        control.DigestTypeValue,
						Value: control.DigestValue{
							MaxProcessedFields: int(maxProcessedFieldsInt32),
						},
					},
				},
			},
		}, nil
	}

	return e.setMultipleSamplersConfig(ctx, parameters, writer, updateGen)
}

func (e *Executors) DigestsValueUpdate(ctx context.Context, parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
	digestNameParameter, _ := parameters.Get("digest-name")
	streamNameParameter, _ := parameters.Get("stream-name")

	flushPeriodParameter, _ := parameters.Get("flush-period")
	flushPeriodInt32, err := flushPeriodParameter.AsInt32()
	if err != nil {
		return fmt.Errorf("flush-period must be an integer")
	}

	maxProcessedFieldsParameter, _ := parameters.Get("max-processed-fields")
	maxProcessedFieldsInt32, err := maxProcessedFieldsParameter.AsInt32()
	if err != nil {
		return fmt.Errorf("flush-period must be an integer")
	}

	updateGen := func(samplerControl *control.Sampler) (*control.SamplerConfigUpdate, error) {

		digest, ok := getEntryByName(samplerControl.Config.Digests, digestNameParameter.Value)
		if !ok {
			return nil, fmt.Errorf("Digest does not exist")
		}

		stream, ok := getEntryByName(samplerControl.Config.Streams, streamNameParameter.Value)
		if !ok {
			return nil, fmt.Errorf("Stream does not exist")
		}

		return &control.SamplerConfigUpdate{
			DigestUpdates: []control.DigestUpdate{
				{
					Op: control.DigestUpsert,
					Digest: control.Digest{
						UID:         digest.UID,
						Name:        digestNameParameter.Value,
						StreamUID:   stream.UID,
						FlushPeriod: time.Second * time.Duration(flushPeriodInt32),
						Type:        control.DigestTypeValue,
						Value: control.DigestValue{
							MaxProcessedFields: int(maxProcessedFieldsInt32),
						},
					},
				},
			},
		}, nil
	}

	return e.setMultipleSamplersConfig(ctx, parameters, writer, updateGen)
}

func (e *Executors) DigestsDelete(ctx context.Context, parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
	digestNameParameter, _ := parameters.Get("digest-name")

	updateGen := func(samplerControl *control.Sampler) (*control.SamplerConfigUpdate, error) {

		digest, ok := getEntryByName(samplerControl.Config.Digests, digestNameParameter.Value)
		if !ok {
			return nil, fmt.Errorf("Digest does not exist")
		}

		return &control.SamplerConfigUpdate{
			DigestUpdates: []control.DigestUpdate{
				{
					Op: control.DigestDelete,
					Digest: control.Digest{
						UID: digest.UID,
					},
				},
			},
		}, nil
	}

	return e.setMultipleSamplersConfig(ctx, parameters, writer, updateGen)
}

func (e *Executors) EventsList(ctx context.Context, parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
	samplerParameter, _ := parameters.Get("sampler-name")
	resourceParameter, _ := parameters.Get("resource-name")

	resourceAndSamplers, err := e.controlPlaneClient.getSamplers(ctx, resourceParameter.Value, samplerParameter.Value, "*", false)
	if err != nil {
		return err
	}
	header := []string{"Resource", "Sampler", "Events"}
	rows := [][]string{}

	for resourceAndSamplerEntry, samplerData := range resourceAndSamplers {
		for _, stream := range samplerData.Config.Streams {
			for _, event := range samplerData.Config.Events {
				if stream.UID == event.StreamUID {
					eventInfo := fmt.Sprintf("Name: %s, Stream: %s, DataType: %s, Rule: %s",
						event.Name,
						stream.Name,
						event.SampleType,
						event.Rule,
					)
					rows = append(rows, []string{
						resourceAndSamplerEntry.resource,
						resourceAndSamplerEntry.sampler,
						eventInfo,
					})
				}
			}
		}
	}

	// Sort rows by resource, rows with the same resource must be ordered by sampler.
	slices.SortStableFunc(rows, func(a []string, b []string) int {
		if a[0] != b[0] {
			// The resource is not the same in the two rows. Order by resource (first entry)
			return cmpStrings(a[0], b[0])
		} else {
			// The resource is the same in the two rows. Order by sampler (second entry)
			return cmpStrings(a[1], b[1])
		}
	})

	// Write table
	writeTable(header, rows, []int{0, 1}, writer)

	if err != nil && errors.Is(err, context.Canceled) {
		writer.WriteStringf("\n\nWarn: internal state was not updated because %s, results could be outdated\n", err)
	}

	return nil
}

func (e *Executors) EventsCreate(ctx context.Context, parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
	eventNameParameter, _ := parameters.Get("event-name")
	streamNameParameter, _ := parameters.Get("stream-name")
	dataTypeParameter, _ := parameters.Get("sample-type")
	ruleParameter, _ := parameters.Get("rule")

	updateGen := func(samplerControl *control.Sampler) (*control.SamplerConfigUpdate, error) {

		_, ok := getEntryByName(samplerControl.Config.Events, eventNameParameter.Value)
		if ok {
			return nil, fmt.Errorf("Event already exists")
		}

		stream, ok := getEntryByName(samplerControl.Config.Streams, streamNameParameter.Value)
		if !ok {
			return nil, fmt.Errorf("Stream does not exist")
		}

		return &control.SamplerConfigUpdate{
			EventUpdates: []control.EventUpdate{
				{
					Op: control.EventUpsert,
					Event: control.Event{
						UID:        control.SamplerEventUID(uuid.New().String()),
						Name:       eventNameParameter.Value,
						StreamUID:  stream.UID,
						SampleType: control.ParseSampleType(dataTypeParameter.Value),
						Rule: control.Rule{
							Lang:       control.SrlCel,
							Expression: ruleParameter.Value,
						},
					},
				},
			},
		}, nil
	}

	return e.setMultipleSamplersConfig(ctx, parameters, writer, updateGen)
}

func (e *Executors) EventsUpdate(ctx context.Context, parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
	eventNameParameter, _ := parameters.Get("event-name")
	streamNameParameter, _ := parameters.Get("stream-name")
	dataTypeParameter, _ := parameters.Get("sample-type")
	ruleParameter, _ := parameters.Get("rule")

	updateGen := func(samplerControl *control.Sampler) (*control.SamplerConfigUpdate, error) {

		event, ok := getEntryByName(samplerControl.Config.Events, eventNameParameter.Value)
		if !ok {
			return nil, fmt.Errorf("Event does not exist")
		}

		stream, ok := getEntryByName(samplerControl.Config.Streams, streamNameParameter.Value)
		if !ok {
			return nil, fmt.Errorf("Stream does not exist")
		}

		return &control.SamplerConfigUpdate{
			EventUpdates: []control.EventUpdate{
				{
					Op: control.EventUpsert,
					Event: control.Event{
						UID:        event.UID,
						Name:       eventNameParameter.Value,
						StreamUID:  stream.UID,
						SampleType: control.ParseSampleType(dataTypeParameter.Value),
						Rule: control.Rule{
							Lang:       control.SrlCel,
							Expression: ruleParameter.Value,
						},
					},
				},
			},
		}, nil
	}

	return e.setMultipleSamplersConfig(ctx, parameters, writer, updateGen)
}

func (e *Executors) EventsDelete(ctx context.Context, parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
	eventNameParameter, _ := parameters.Get("event-name")

	updateGen := func(samplerControl *control.Sampler) (*control.SamplerConfigUpdate, error) {

		event, ok := getEntryByName(samplerControl.Config.Events, eventNameParameter.Value)
		if !ok {
			return nil, fmt.Errorf("Digest does not exist")
		}

		return &control.SamplerConfigUpdate{
			EventUpdates: []control.EventUpdate{
				{
					Op: control.EventDelete,
					Event: control.Event{
						UID: event.UID,
					},
				},
			},
		}, nil
	}

	return e.setMultipleSamplersConfig(ctx, parameters, writer, updateGen)
}
