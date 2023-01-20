# Neblictl

This command connects with the Neblic control plane server and allows its user to get information and configure all the available samplers in the system.

There are two main elements:
* samplers: As defined in the instrumented applications.
* Rules: A sampling rule that is evaluated per each sampled message. If the rule evaluates as `true`, the message is exported. Rules are defined using [Google CEL](https://github.com/google/cel-spec). The sample data is available under the key `sample`. For example, to extract messages matching:
  * A particular id `sample.id == 1234`
  * A particular string `sample.field == "some_string"`
  * A complex expression: `sample.field1 > 4 && sampling.field2 > 5 || sampling.field3 != null`
  * Anything: `true`

Neblictl implements an interactive cli with autocompletion and each command shows a usage message if used incorrectly.

## Configure neblictl
### Set token
- Run `neblictl init` to create the configuration file. The command will output its location.
- Open the configuration file and set the desired token value in the `Token` field.

## Main commands

Commands where a `<sampler>` or a `<resource>` is set, the special character `*` can be used to match all the entries.
For example:
- `list rules * *`: It shows the rules for all resources and samplers
- `list rules * sampler1`: It shows the rules for the resources that have a `sampler1` sampler
- `list rules resource1 *`: It shows all the rules of all the samplers of `resource1` resource

### help

Shows available commands

### list

Lists elements as samplers, resources or sampling rules. For example: `list samplers` shows the list of registered samplers.

### create/update/delete rule

Allows the definition and modification of sampling rules. For example: `create rule <resource_name> <sampler_name> <sampling_rule>` (sampling rule as a CEL expression) create a new sampling rule at the specified sampler.

### create/update/delete rate

Create/update and delete sampling rates. This configuration limits how much samples can be exported. For example: `create rate <resource> <sampler> <samples_per_second>`

## Quickstart

### Check global status
Run `list samplers`. That will show the configured samplers, the samples evaluated (times the sample function of a specific sampler was called), and the samples exported (times the sample function of a specific sampler was called, and the evaluation was positive, as a result the data has been exported).

### Create a sampling rule
If the sampler that that's being targeted does not have any sampling rule, it is usually useful to start checking the global status of the sampler running. `list samplers`, that shows the number of samples exported.

Create a rule running `create rule <resource> <sampler> <sampling_rule>` (`sampling_rule` can be set to `true` to forward all the data, doing that is just recommended to test the client).

Check the sampler stats to see if the number of exported samples has increased. Run `list samplers` again, check if the numbers changed from the last time

### Create a sampling rate
It's usually useful to set a rate limit as a safeguard to avoid exporting too many samples, the sample rate limits the number of samples per second that can be exported, discarding the ones above the limit. By default, a sampling rate is defined when a sampler is initialized in the service code, but this value can be overridden using the client running `create rate <resouce> <sampler> <limit>`
