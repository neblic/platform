package controlplane

import (
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

func (e *Executors) doesResourceAndSamplerMatch(resourceParameter, samplerParameter string, resourceAndSamplerEntry resourceAndSampler) bool {
	acceptAllResources := resourceParameter == "*"
	acceptAllSamplers := samplerParameter == "*"

	// This part of the logic explores all four possible combinations.
	// 1) ASSUMPTION: resourceParameter==* and SamplerParameter==*
	allResourcesAndAllSamplers := acceptAllResources && acceptAllSamplers
	// 2) ASSUMPTION: resourceParameter==* and samplerParameter==<sampler> CONDITION resourceAndSamplerEntry.sampler==<sampler>
	allResourcesAndMatchingSampler := acceptAllResources && resourceAndSamplerEntry.sampler == samplerParameter
	// 3) ASSUMPTION resourceAndSamplerEntry.resrouce==<resource> and resourceParameter==* CONDITION resourceParameter==<resource>
	matchingResourceAndAllSamplers := resourceAndSamplerEntry.resource == resourceParameter && acceptAllSamplers
	// 4) ASSUMPTION resourceParameter==<resource> and samplerParameter==<sampler> CONDITION resourceAndSamplerEntry.resource==<resource> and resourceAndSamplerEntry.sampler==<sampler>
	matchingResourceAndMatchingSampler := resourceAndSamplerEntry.resource == resourceParameter && resourceAndSamplerEntry.sampler == samplerParameter
	return allResourcesAndAllSamplers || allResourcesAndMatchingSampler || matchingResourceAndAllSamplers || matchingResourceAndMatchingSampler

}

func (e *Executors) getMatchingSamplers(resourceParameter string, samplerParameter string, cached bool) (map[resourceAndSampler]*data.Sampler, error) {

	// Iterate over all samplers and select the ones matching the input
	resourceAndSamplers := map[resourceAndSampler]*data.Sampler{}
	for resourceAndSamplerEntry, samplerData := range e.controlPlaneClient.getSamplers(cached) {
		if e.doesResourceAndSamplerMatch(resourceParameter, samplerParameter, resourceAndSamplerEntry) {
			resourceAndSamplers[resourceAndSampler{resourceAndSamplerEntry.resource, resourceAndSamplerEntry.sampler}] = samplerData
		}
	}

	// Return error if no matching sampler was found
	if len(resourceAndSamplers) == 0 {
		var err error
		if resourceParameter == "*" || samplerParameter == "*" {
			err = fmt.Errorf("could not find any sampler matching the criteria")
		} else {
			err = fmt.Errorf("sampler does not exist")
		}
		return resourceAndSamplers, err
	}

	return resourceAndSamplers, nil
}

func (e *Executors) ListResources(parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
	// Get all samplers
	samplers := e.controlPlaneClient.getSamplers(false)

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

	return nil
}

func (e *Executors) ListSamplers(parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
	// Get all samplers
	samplers := e.controlPlaneClient.getSamplers(false)

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

	return nil
}

func (e *Executors) ListRules(parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
	// Get options
	samplerParameter, _ := parameters.Get("sampler")
	resourceParameter, _ := parameters.Get("resource")

	// Compute list of targeted resources and samplers
	resourceAndSamplers, err := e.getMatchingSamplers(resourceParameter.Value, samplerParameter.Value, false)
	if err != nil {
		return err
	}

	// Build table rows
	header := []string{"Resource", "Sampler", "Sampling Rule"}
	rows := [][]string{}
	for _, samplerData := range resourceAndSamplers {
		for _, samplingRule := range samplerData.Config.SamplingRules {
			rows = append(rows, []string{samplerData.Resource, samplerData.Name, samplingRule.Rule})
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

	return nil
}

func (e *Executors) CreateRule(parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
	// Get options
	samplerParameter, _ := parameters.Get("sampler")
	resourceParameter, _ := parameters.Get("resource")
	samplingRuleParameter, _ := parameters.Get("sampling_rule")

	// Compute list of targeted resources and samplers
	resourceAndSamplers, err := e.getMatchingSamplers(resourceParameter.Value, samplerParameter.Value, false)
	if err != nil {
		return err
	}

	// Create rules one by one
	for resourceAndSamplerEntry, samplerData := range resourceAndSamplers {
		// Check that the sampling rule does not exist
		ruleExists := false
		for _, samplingRule := range samplerData.Config.SamplingRules {
			if samplingRule.Rule == samplingRuleParameter.Value {
				ruleExists = true
				break
			}
		}
		if ruleExists {
			writer.WriteStringf("%s.%s: Sampling rule already exists\n", resourceAndSamplerEntry.resource, resourceAndSamplerEntry.sampler)
			continue
		}

		update := &data.SamplerConfigUpdate{
			SamplingRuleUpdates: []data.SamplingRuleUpdate{
				{
					Op: data.SamplingRuleUpsert,
					SamplingRule: data.SamplingRule{
						UID:  data.SamplerSamplingRuleUID(uuid.New().String()),
						Lang: data.SrlCel,
						Rule: samplingRuleParameter.Value,
					},
				},
			},
		}

		// Propagate new configuration
		err := e.controlPlaneClient.setSamplerConfig(resourceAndSamplerEntry.sampler, resourceAndSamplerEntry.resource, update)
		if err != nil {
			writer.WriteStringf("%s.%s: Error creating the sampling rule: %v\n", resourceAndSamplerEntry.resource, resourceAndSamplerEntry.sampler, err)
			continue
		}

		writer.WriteStringf("%s.%s: Rule successfully created\n", resourceAndSamplerEntry.resource, resourceAndSamplerEntry.sampler)

	}

	return nil
}

func (e *Executors) CreateRate(parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
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
	resourceAndSamplers, err := e.getMatchingSamplers(resourceParameter.Value, samplerParameter.Value, false)
	if err != nil {
		return err
	}

	// Create rate one by one
	for resourceAndSamplerEntry, samplerData := range resourceAndSamplers {
		// Check that the sampling rate does not exist.
		// If sampling rate is nil, it means the server does not have any configured sampling rate and the sampler default is applied.
		// When the server has configuration, a sampling rate exists if the value is bigger or equal than 0.
		if samplerData.Config.SamplingRate != nil && samplerData.Config.SamplingRate.Limit >= 0 {
			writer.WriteStringf("%s.%s: Sampling rate already exists\n", resourceAndSamplerEntry.resource, resourceAndSamplerEntry.sampler)
			continue
		}

		update := &data.SamplerConfigUpdate{
			SamplingRate: &data.SamplingRate{
				Limit: limitInt64,
				Burst: 0,
			},
		}

		// Propagate new configuration
		err = e.controlPlaneClient.setSamplerConfig(resourceAndSamplerEntry.sampler, resourceAndSamplerEntry.resource, update)
		if err != nil {
			writer.WriteStringf("%s.%s: Error creating the sampling rate: %v\n", resourceAndSamplerEntry.resource, resourceAndSamplerEntry.sampler, err)
			continue
		}

		// Write output
		writer.WriteStringf("%s.%s: Rate successfully created\n", resourceAndSamplerEntry.resource, resourceAndSamplerEntry.sampler)
	}

	return nil
}

func (e *Executors) UpdateRule(parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
	// Get options
	samplerParameter, _ := parameters.Get("sampler")
	resourceParameter, _ := parameters.Get("resource")
	oldSamplingRuleParameter, _ := parameters.Get("old_sampling_rule")
	newSamplingRuleParameter, _ := parameters.Get("new_sampling_rule")

	// Compute list of targeted resources and samplers
	resourceAndSamplers, err := e.getMatchingSamplers(resourceParameter.Value, samplerParameter.Value, false)
	if err != nil {
		return err
	}

	// Update rule one by one
	for resourceAndSamplerEntry, samplerData := range resourceAndSamplers {

		// Check that the sampling rule already exists
		existingUID := data.SamplerSamplingRuleUID("")
		for uid, samplingRule := range samplerData.Config.SamplingRules {
			if samplingRule.Rule == oldSamplingRuleParameter.Value {
				existingUID = uid
				break
			}
		}

		if existingUID == "" {
			writer.WriteStringf("%s.%s: Sampling rule does not exist\n", resourceAndSamplerEntry.resource, resourceAndSamplerEntry.sampler)
			continue

		}

		// Modify sampling rule to existing config
		update := &data.SamplerConfigUpdate{
			SamplingRuleUpdates: []data.SamplingRuleUpdate{
				{
					Op: data.SamplingRuleUpsert,
					SamplingRule: data.SamplingRule{
						UID:  existingUID,
						Lang: data.SrlCel,
						Rule: newSamplingRuleParameter.Value,
					},
				},
			},
		}

		// Propagate new configuration
		err := e.controlPlaneClient.setSamplerConfig(resourceAndSamplerEntry.sampler, resourceAndSamplerEntry.resource, update)
		if err != nil {
			writer.WriteStringf("%s.%s: Error updating the sampling rule: %v\n", resourceAndSamplerEntry.resource, resourceAndSamplerEntry.sampler, err)
			continue
		}

		// Write output
		writer.WriteStringf("%s.%s: Rule successfully updated\n", resourceAndSamplerEntry.resource, resourceAndSamplerEntry.sampler)
	}

	return nil
}

func (e *Executors) UpdateRate(parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
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
	resourceAndSamplers, err := e.getMatchingSamplers(resourceParameter.Value, samplerParameter.Value, false)
	if err != nil {
		return err
	}

	// Update rate one by one
	for resourceAndSamplerEntry, samplerData := range resourceAndSamplers {

		// Check that the sampling rate exists.
		// If sampling rate is nil, it means the server does not have any configured sampling rate and the sampler default is applied.
		// When the server has configuration, a sampling rate does not exist if it forwards without limits (limit is -1)
		if samplerData.Config.SamplingRate == nil || samplerData.Config.SamplingRate.Limit == -1 {
			writer.WriteStringf("%s.%s: Sampling rate does not exist\n", resourceAndSamplerEntry.resource, resourceAndSamplerEntry.sampler)
			continue
		}

		// Modify sampling rate to existing config
		update := &data.SamplerConfigUpdate{
			SamplingRate: &data.SamplingRate{
				Limit: limitInt64,
				Burst: 0,
			},
		}

		// Propagate new configuration
		err = e.controlPlaneClient.setSamplerConfig(resourceAndSamplerEntry.sampler, resourceAndSamplerEntry.resource, update)
		if err != nil {
			writer.WriteStringf("%s.%s: Error updating the sampling rate: %v\n", resourceAndSamplerEntry.resource, resourceAndSamplerEntry.sampler, err)
			continue
		}

		// Write output
		writer.WriteStringf("%s.%s: Rate successfully updated\n", resourceAndSamplerEntry.resource, resourceAndSamplerEntry.sampler)
	}

	return nil
}

func (e *Executors) DeleteRule(parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
	// Get options
	samplerParameter, _ := parameters.Get("sampler")
	resourceParameter, _ := parameters.Get("resource")
	samplingRuleParameter, _ := parameters.Get("sampling_rule")

	// Compute list of targeted resources and samplers
	resourceAndSamplers, err := e.getMatchingSamplers(resourceParameter.Value, samplerParameter.Value, false)
	if err != nil {
		return err
	}

	// Delete rule one by one
	for resourceAndSamplerEntry, samplerData := range resourceAndSamplers {

		// Check that the sampling rule already exists
		existingUID := data.SamplerSamplingRuleUID("")
		for uid, samplingRule := range samplerData.Config.SamplingRules {
			if samplingRule.Rule == samplingRuleParameter.Value {
				existingUID = uid
				break
			}
		}

		if existingUID == "" {
			writer.WriteStringf("%s.%s: Sampling rule does not exist\n", resourceAndSamplerEntry.resource, resourceAndSamplerEntry.sampler)
			continue
		}

		// Modify sampling rule to existing config
		update := &data.SamplerConfigUpdate{
			SamplingRuleUpdates: []data.SamplingRuleUpdate{
				{
					Op: data.SamplingRuleDelete,
					SamplingRule: data.SamplingRule{
						UID: existingUID,
					},
				},
			},
		}

		// Propagate new configuration
		err := e.controlPlaneClient.setSamplerConfig(resourceAndSamplerEntry.sampler, resourceAndSamplerEntry.resource, update)
		if err != nil {
			writer.WriteStringf("%s.%s: Error deleting the sampling rule: %v\n", resourceAndSamplerEntry.resource, resourceAndSamplerEntry.sampler, err)
			continue
		}

		// Write output
		writer.WriteStringf("%s.%s: Rate successfully deleted\n", resourceAndSamplerEntry.resource, resourceAndSamplerEntry.sampler)
	}

	return nil
}

func (e *Executors) DeleteRate(parameters interpoler.ParametersWithValue, writer *internal.Writer) error {
	// Get options
	samplerParameter, _ := parameters.Get("sampler")
	resourceParameter, _ := parameters.Get("resource")

	// Compute list of targeted resources and samplers
	resourceAndSamplers, err := e.getMatchingSamplers(resourceParameter.Value, samplerParameter.Value, false)
	if err != nil {
		return err
	}

	// Delete rate one by one
	for resourceAndSamplerEntry, samplerData := range resourceAndSamplers {

		// Check that the sampling rate exists.
		// If sampling rate is nil, it means the server does not have any configured sampling rate and the sampler default is applied.
		// When the server has configuration, a sampling rate does not exist if it forwards without limits (limit is -1)
		if samplerData.Config.SamplingRate == nil || samplerData.Config.SamplingRate.Limit == -1 {
			writer.WriteStringf("%s.%s: Sampling rate does not exist\n", resourceAndSamplerEntry.resource, resourceAndSamplerEntry.sampler)
			continue

		}

		// Modify sampling rate to existing config
		update := &data.SamplerConfigUpdate{
			SamplingRate: &data.SamplingRate{
				Limit: -1,
				Burst: 0,
			},
		}

		// Propagate new configuration
		err := e.controlPlaneClient.setSamplerConfig(resourceAndSamplerEntry.sampler, resourceAndSamplerEntry.resource, update)
		if err != nil {
			writer.WriteStringf("%s.%s: Error deleting the sampling rate: %v\n", resourceAndSamplerEntry.resource, resourceAndSamplerEntry.sampler, err)
			continue
		}

		// Write output
		writer.WriteStringf("%s.%s: Rate successfully deleted\n", resourceAndSamplerEntry.resource, resourceAndSamplerEntry.sampler)

	}

	return nil
}
