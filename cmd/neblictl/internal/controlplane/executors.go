package controlplane

import (
	"context"
	"errors"
	"fmt"
	"strconv"

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

func (e *Executors) ListResources(ctx context.Context, parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
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

func (e *Executors) ListSamplers(ctx context.Context, parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
	// Get all samplers
	samplers, err := e.controlPlaneClient.getAllSamplers(ctx, false)

	// Build table rows
	header := []string{"Resource", "Sampler", "Sampling Rate", "Samples Evaluated", "Samples Exported"}
	rows := [][]string{}
	for resourceAndSamplerEntry, samplerData := range samplers {
		// Define sampling rate value
		samplingRate := "sampler default"
		if samplerData.Config.SamplingRate != nil {
			if samplerData.Config.SamplingRate.Limit == -1 {
				samplingRate = "disabled"
			} else {
				samplingRate = fmt.Sprintf("Limit: %d", samplerData.Config.SamplingRate.Limit)
			}
		}

		// Append row
		rows = append(rows, []string{resourceAndSamplerEntry.resource,
			resourceAndSamplerEntry.sampler,
			samplingRate,
			strconv.FormatUint(samplerData.SamplingStats.SamplesEvaluated, 10),
			strconv.FormatUint(samplerData.SamplingStats.SamplesExported, 10),
		})
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
	writeTable(header, rows, []int{0}, writer)

	if err != nil && errors.Is(err, context.Canceled) {
		writer.WriteStringf("\n\nWarn: internal state was not updated because %s, results could be outdated\n", err)
	}

	return nil
}

func (e *Executors) ListStreams(ctx context.Context, parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
	// Get options
	samplerParameter, _ := parameters.Get("sampler")
	resourceParameter, _ := parameters.Get("resource")

	// Compute list of targeted resources and samplers
	resourceAndSamplers, err := e.controlPlaneClient.getSamplers(ctx, resourceParameter.Value, samplerParameter.Value, false)

	// Build table rows
	header := []string{"Resource", "Sampler", "Stream UID", "Stream Rule"}
	rows := [][]string{}
	for _, samplerData := range resourceAndSamplers {
		for _, stream := range samplerData.Config.Streams {
			rows = append(rows, []string{samplerData.Resource, samplerData.Name, string(stream.UID), stream.StreamRule.Rule})
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

func (e *Executors) CreateStreams(ctx context.Context, parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
	resourceParameter, _ := parameters.Get("resource")
	samplerParameter, _ := parameters.Get("sampler")
	streamRuleParameter, _ := parameters.Get("rule")
	streamUIDParameter, streamUIDParameterSet := parameters.Get("uid") // optional

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
					Op: data.StreamRuleUpsert,
					Stream: data.Stream{
						UID: streamUID,
						StreamRule: data.StreamRule{
							UID:  data.SamplerStreamRuleUID(uuid.New().String()),
							Lang: data.SrlCel,
							Rule: streamRuleParameter.Value,
						},
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

func (e *Executors) UpdateStreams(ctx context.Context, parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
	resourceParameter, _ := parameters.Get("resource")
	samplerParameter, _ := parameters.Get("sampler")
	streamUIDParameter, _ := parameters.Get("uid")
	updatedRuleParameter, _ := parameters.Get("updated-rule")

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
					Op: data.StreamRuleUpsert,
					Stream: data.Stream{
						UID: data.SamplerStreamUID(streamUIDParameter.Value),
						StreamRule: data.StreamRule{
							UID:  data.SamplerStreamRuleUID(uuid.New().String()),
							Lang: data.SrlCel,
							Rule: updatedRuleParameter.Value,
						},
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

func (e *Executors) DeleteStreams(ctx context.Context, parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
	// Get options
	resourceParameter, _ := parameters.Get("resource")
	samplerParameter, _ := parameters.Get("sampler")
	streamUIDParameter, _ := parameters.Get("uid")

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
					Op: data.StreamRuleDelete,
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
func (e *Executors) SamplerSamplingSet(ctx context.Context, parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
	// Get options
	samplerParameter, _ := parameters.Get("sampler")
	resourceParameter, _ := parameters.Get("resource")
	limitParameter, _ := parameters.Get("limit")

	// Parse limit and burst parameters
	limitInt64, err := limitParameter.AsInt64()
	if err != nil {
		return fmt.Errorf("limit must be an integer")
	}

	// Compute list of targeted resources and samplers
	resourceAndSamplers, err := e.controlPlaneClient.getSamplers(ctx, resourceParameter.Value, samplerParameter.Value, false)
	if err != nil {
		return err
	}

	// Create rate one by one
	for resourceAndSamplerEntry := range resourceAndSamplers {
		update := &data.SamplerConfigUpdate{
			SamplingRate: &data.SamplingRate{
				Limit: limitInt64,
				Burst: 0,
			},
		}

		// Propagate new configuration
		err = e.controlPlaneClient.setSamplerConfig(ctx, resourceAndSamplerEntry.sampler, resourceAndSamplerEntry.resource, update)
		if err != nil {
			writer.WriteStringf("%s.%s: Could not create the sampling rate because %v\n", resourceAndSamplerEntry.resource, resourceAndSamplerEntry.sampler, err)
			continue
		}

		// Write output
		writer.WriteStringf("%s.%s: Rate successfully created\n", resourceAndSamplerEntry.resource, resourceAndSamplerEntry.sampler)
	}

	return nil
}

func (e *Executors) SamplerSamplingUnset(ctx context.Context, parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
	// Get options
	samplerParameter, _ := parameters.Get("sampler")
	resourceParameter, _ := parameters.Get("resource")

	// Compute list of targeted resources and samplers
	resourceAndSamplers, err := e.controlPlaneClient.getSamplers(ctx, resourceParameter.Value, samplerParameter.Value, false)
	if err != nil {
		return err
	}

	// Delete rate one by one
	for resourceAndSamplerEntry := range resourceAndSamplers {
		// Modify sampling rate to existing config
		update := &data.SamplerConfigUpdate{
			SamplingRate: &data.SamplingRate{
				Limit: -1,
				Burst: 0,
			},
		}

		// Propagate new configuration
		err := e.controlPlaneClient.setSamplerConfig(ctx, resourceAndSamplerEntry.sampler, resourceAndSamplerEntry.resource, update)
		if err != nil {
			writer.WriteStringf("%s.%s: Could not delete the sampling rate because %v\n", resourceAndSamplerEntry.resource, resourceAndSamplerEntry.sampler, err)
			continue
		}

		// Write output
		writer.WriteStringf("%s.%s: Rate successfully deleted\n", resourceAndSamplerEntry.resource, resourceAndSamplerEntry.sampler)
	}

	return nil
}
