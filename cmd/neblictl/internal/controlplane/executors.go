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
	"github.com/neblic/platform/controlplane/data"
	"github.com/olekukonko/tablewriter"
	"golang.org/x/exp/slices"
)

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
	slices.SortStableFunc(rows, func(a []string, b []string) bool {
		return a[0] < b[0]
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

	resourceAndSamplers, err := e.controlPlaneClient.getSamplers(ctx, resourceParameter.Value, samplerParameter.Value, false)
	if err != nil {
		return err
	}

	header := []string{"Resource", "Sampler", "Stats"}
	rows := [][]string{}
	for resourceAndSamplerEntry, samplerData := range resourceAndSamplers {
		rows = append(rows, []string{resourceAndSamplerEntry.resource,
			resourceAndSamplerEntry.sampler,
			samplerData.SamplingStats.CLIInfo(),
		})
	}

	slices.SortStableFunc(rows, func(a []string, b []string) bool {
		if a[0] != b[0] {
			return a[0] < b[0]
		} else {
			return a[1] < b[1]
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

	resourceAndSamplers, err := e.controlPlaneClient.getSamplers(ctx, resourceParameter.Value, samplerParameter.Value, false)
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
			samplingIn = fmt.Sprintf("%s", samplerData.Config.SamplingIn.CLIInfo())
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

	slices.SortStableFunc(rows, func(a []string, b []string) bool {
		if a[0] != b[0] {
			return a[0] < b[0]
		} else {
			return a[1] < b[1]
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
	resourceAndSamplers, err := e.controlPlaneClient.getSamplers(ctx, resourceParameter.Value, samplerParameter.Value, false)

	// Build table rows
	header := []string{"Resource", "Sampler", "Stream"}
	rows := [][]string{}
	for _, samplerData := range resourceAndSamplers {
		for _, stream := range samplerData.Config.Streams {
			rows = append(rows, []string{samplerData.Resource, samplerData.Name, stream.CLIInfo()})
		}
	}

	// Sort rows first by resource and then by sampler.
	slices.SortStableFunc(rows, func(a []string, b []string) bool {
		if a[0] != b[0] {
			return a[0] < b[0]
		} else {
			return a[1] < b[1]
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
	streamUIDParameter, streamUIDParameterSet := parameters.Get("stream-uid") // optional

	exportRawParameter, _ := parameters.Get("export-raw")
	exportRawBool, err := strconv.ParseBool(exportRawParameter.Value)
	if err != nil {
		return fmt.Errorf("export-raw must be a boolean")
	}

	// Compute list of targeted resources and samplers
	resourceAndSamplers, err := e.controlPlaneClient.getSamplers(ctx, resourceParameter.Value, samplerParameter.Value, false)
	if err != nil {
		return err
	}

	// If multiple streans are created at once, they will all have the same UID
	var streamUID data.SamplerStreamUID
	if streamUIDParameterSet {
		streamUID = data.SamplerStreamUID(streamUIDParameter.Value)
	} else {
		streamUID = data.SamplerStreamUID(uuid.New().String())
	}

	// Create rules one by one
	for resourceAndSamplerEntry, samplerData := range resourceAndSamplers {
		// Check that the stream does not exist
		streamExists := false
		for _, stream := range samplerData.Config.Streams {
			if stream.UID == streamUID {
				streamExists = true
				break
			}
		}
		if streamExists {
			writer.WriteStringf("%s.%s: Stream already exists\n", resourceAndSamplerEntry.resource, resourceAndSamplerEntry.sampler)
			continue
		}
		update := &data.SamplerConfigUpdate{
			StreamUpdates: []data.StreamUpdate{
				{
					Op: data.StreamUpsert,
					Stream: data.Stream{
						UID: streamUID,
						StreamRule: data.StreamRule{
							Lang: data.SrlCel,
							Rule: streamRuleParameter.Value,
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
	streamUIDParameter, _ := parameters.Get("stream-uid")
	updatedRuleParameter, _ := parameters.Get("updated-rule")

	exportRawParameter, _ := parameters.Get("export-raw")
	exportRawBool, err := strconv.ParseBool(exportRawParameter.Value)
	if err != nil {
		return fmt.Errorf("export-raw must be a boolean")
	}

	// Compute list of targeted resources and samplers
	resourceAndSamplers, err := e.controlPlaneClient.getSamplers(ctx, resourceParameter.Value, samplerParameter.Value, false)
	if err != nil {
		return err
	}

	// Update streams one by one
	for resourceAndSamplerEntry, samplerData := range resourceAndSamplers {
		// Check if the stream exists
		found := false
		for _, stream := range samplerData.Config.Streams {
			if stream.UID == data.SamplerStreamUID(streamUIDParameter.Value) {
				found = true
				break
			}
		}

		if !found {
			writer.WriteStringf("%s.%s: Stream with UID '%s' does not exist\n", resourceAndSamplerEntry.resource, resourceAndSamplerEntry.sampler, streamUIDParameter.Value)
			continue
		}

		// Modify sampling rule to existing config
		update := &data.SamplerConfigUpdate{
			StreamUpdates: []data.StreamUpdate{
				{
					Op: data.StreamUpsert,
					Stream: data.Stream{
						UID: data.SamplerStreamUID(streamUIDParameter.Value),
						StreamRule: data.StreamRule{
							Lang: data.SrlCel,
							Rule: updatedRuleParameter.Value,
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
	streamUIDParameter, _ := parameters.Get("stream-uid")

	// Compute list of targeted resources and samplers
	resourceAndSamplers, err := e.controlPlaneClient.getSamplers(ctx, resourceParameter.Value, samplerParameter.Value, false)
	if err != nil {
		return err
	}

	// Delete streams one by one
	for resourceAndSamplerEntry, samplerData := range resourceAndSamplers {

		// Check if the stream exists
		found := false
		for _, stream := range samplerData.Config.Streams {
			if stream.UID == data.SamplerStreamUID(streamUIDParameter.Value) {
				found = true
				break
			}
		}

		if !found {
			writer.WriteStringf("%s.%s: Stream with UID '%s' does not exist\n", resourceAndSamplerEntry.resource, resourceAndSamplerEntry.sampler, streamUIDParameter.Value)
			continue
		}

		// Modify sampling rule to existing config
		update := &data.SamplerConfigUpdate{
			StreamUpdates: []data.StreamUpdate{
				{
					Op: data.StreamDelete,
					Stream: data.Stream{
						UID: data.SamplerStreamUID(streamUIDParameter.Value),
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

func (e *Executors) setMultipleSamplersConfig(ctx context.Context, parameters interpoler.ParametersWithValue, writer *internal.Writer, update *data.SamplerConfigUpdate) error {
	samplerParameter, _ := parameters.Get("sampler-name")
	resourceParameter, _ := parameters.Get("resource-name")

	resourceAndSamplers, err := e.controlPlaneClient.getSamplers(ctx, resourceParameter.Value, samplerParameter.Value, false)
	if err != nil {
		return err
	}

	for resourceAndSamplerEntry := range resourceAndSamplers {
		if err := e.controlPlaneClient.setSamplerConfig(ctx, resourceAndSamplerEntry.sampler, resourceAndSamplerEntry.resource, update); err != nil {
			writer.WriteStringf("%s.%s: Could not update sampler config%v\n", resourceAndSamplerEntry.resource, resourceAndSamplerEntry.sampler, err)
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

	update := &data.SamplerConfigUpdate{
		LimiterIn: &data.LimiterConfig{
			Limit: limitInt32,
		},
	}

	return e.setMultipleSamplersConfig(ctx, parameters, writer, update)
}

func (e *Executors) SamplersLimiterInUnset(ctx context.Context, parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
	update := &data.SamplerConfigUpdate{
		Reset: data.SamplerConfigUpdateReset{
			LimiterIn: true,
		},
	}

	return e.setMultipleSamplersConfig(ctx, parameters, writer, update)
}

func (e *Executors) SamplersLimiterOutSet(ctx context.Context, parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
	limitParameter, _ := parameters.Get("limit")
	limitInt32, err := limitParameter.AsInt32()
	if err != nil {
		return fmt.Errorf("limit must be an integer")
	}

	update := &data.SamplerConfigUpdate{
		LimiterOut: &data.LimiterConfig{
			Limit: limitInt32,
		},
	}

	return e.setMultipleSamplersConfig(ctx, parameters, writer, update)
}

func (e *Executors) SamplersLimiterOutUnset(ctx context.Context, parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
	update := &data.SamplerConfigUpdate{
		Reset: data.SamplerConfigUpdateReset{
			LimiterOut: true,
		},
	}

	return e.setMultipleSamplersConfig(ctx, parameters, writer, update)
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

	update := &data.SamplerConfigUpdate{
		SamplingIn: &data.SamplingConfig{
			SamplingType: data.DeterministicSamplingType,
			DeterministicSampling: data.DeterministicSamplingConfig{
				SampleRate:             sampleRateInt32,
				SampleEmptyDeterminant: sampleEmptyDetBool,
			},
		},
	}

	return e.setMultipleSamplersConfig(ctx, parameters, writer, update)
}

func (e *Executors) SamplersSamplerInUnset(ctx context.Context, parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
	update := &data.SamplerConfigUpdate{
		Reset: data.SamplerConfigUpdateReset{
			SamplingIn: true,
		},
	}

	return e.setMultipleSamplersConfig(ctx, parameters, writer, update)
}

func (e *Executors) DigestsList(ctx context.Context, parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
	samplerParameter, _ := parameters.Get("sampler-name")
	resourceParameter, _ := parameters.Get("resource-name")

	resourceAndSamplers, err := e.controlPlaneClient.getSamplers(ctx, resourceParameter.Value, samplerParameter.Value, false)
	if err != nil {
		return err
	}
	header := []string{"Resource", "Sampler", "Digest"}
	rows := [][]string{}

	for resourceAndSamplerEntry, samplerData := range resourceAndSamplers {
		for _, stream := range samplerData.Config.Streams {
			for _, digest := range samplerData.Config.Digests {
				if stream.UID == digest.StreamUID {
					rows = append(rows, []string{
						resourceAndSamplerEntry.resource,
						resourceAndSamplerEntry.sampler,
						digest.CLIInfo(),
					})
				}
			}
		}
	}

	// Sort rows first by resource and then by sampler.
	slices.SortStableFunc(rows, func(a []string, b []string) bool {
		if a[0] != b[0] {
			return a[0] < b[0]
		} else {
			return a[1] < b[1]
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
	digestUID := uuid.NewString()
	digestUIDParameter, digestUIDParameterOk := parameters.Get("digest-uid")
	if digestUIDParameterOk {
		digestUID = digestUIDParameter.Value
	}

	streamUIDParameter, _ := parameters.Get("stream-uid")

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

	update := &data.SamplerConfigUpdate{
		DigestUpdates: []data.DigestUpdate{
			{
				Op: data.DigestUpsert,
				Digest: data.Digest{
					UID:         data.SamplerDigestUID(digestUID),
					StreamUID:   data.SamplerStreamUID(streamUIDParameter.Value),
					FlushPeriod: time.Second * time.Duration(flushPeriodInt32),
					Type:        data.DigestTypeSt,
					St: data.DigestSt{
						MaxProcessedFields: int(maxProcessedFieldsInt32),
					},
				},
			},
		},
	}

	return e.setMultipleSamplersConfig(ctx, parameters, writer, update)
}

func (e *Executors) DigestsStructureUpdate(ctx context.Context, parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
	digestUIDParameter, _ := parameters.Get("digest-uid")
	streamUIDParameter, _ := parameters.Get("stream-uid")

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

	update := &data.SamplerConfigUpdate{
		DigestUpdates: []data.DigestUpdate{
			{
				Op: data.DigestUpsert,
				Digest: data.Digest{
					UID:         data.SamplerDigestUID(digestUIDParameter.Value),
					StreamUID:   data.SamplerStreamUID(streamUIDParameter.Value),
					FlushPeriod: time.Second * time.Duration(flushPeriodInt32),
					Type:        data.DigestTypeSt,
					St: data.DigestSt{
						MaxProcessedFields: int(maxProcessedFieldsInt32),
					},
				},
			},
		},
	}

	return e.setMultipleSamplersConfig(ctx, parameters, writer, update)
}

func (e *Executors) DigestsStructureDelete(ctx context.Context, parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
	digestUIDParameter, _ := parameters.Get("uid")

	update := &data.SamplerConfigUpdate{
		DigestUpdates: []data.DigestUpdate{
			{
				Op: data.DigestDelete,
				Digest: data.Digest{
					UID: data.SamplerDigestUID(digestUIDParameter.Value),
				},
			},
		},
	}

	return e.setMultipleSamplersConfig(ctx, parameters, writer, update)
}
