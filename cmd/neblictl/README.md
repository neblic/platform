# Neblictl

<!--learn-start-->
<!-- Not used in docs -->
<!--learn-end-->

<!--how-to-start-->
The `neblictl` command connects with the Neblic `Control Plane` server and allows its user to get information and configure all the `Samplers` in the system.

## Installation

You can download a precompiled binary on the [releases](https://github.com/neblic/platform/releases) page. If you want to build it `go install` won't work because its `go.mod` has replace directives, so you need to clone the entire platform repository and build it in the `neblictl` folder.

It is also bundled with the `otelcol` container image that Neblic distributes so you can run it from inside the container.

## Usage

### Connect to the `Control Plane` server

On first use, it will initialize a configuration file at the path returned by this [function](https://pkg.go.dev/os#UserConfigDir). You will only need to manually edit this file for some advanced usages (e.g. connecting to a `Control Plane` server that uses Bearer authentication, see next section).

You will need to provide the address and port of the `Control Plane` server (default `localhost:8899`) to use it.

``` sh
neblictl -host localhost -control-port 8899
```

`Neblictl` provides an interactive client with auto-completion and built-in help where you will be able to enter commands once connected.

#### Set a Bearer token
- Run `neblictl init` to create the configuration file if necessary and show its path.
- Open the configuration file and set the desired token value in the `Token` field.

### Configure a `Sampler`

Before continuing there are two main concepts that you need to understand: `Samplers` and `Sampling Rules`. Check the concepts [page](https://neblic.github.io/platform/getting-started/concepts/) to learn more about them.

1. Run the command `list samplers`. This will show a list with all the `Samplers` available and some stats for each one of them.
2. Create a rule using the commend `create rule <resource> <sampler> <sampling_rule>`.
3. After a while, you can check the `Sampler` stats to see if the number of exported samples has increased running `list samplers` again.

### Set a `Sampling Rate`

It's usually useful to set a rate limit as a safeguard to avoid exporting too many samples, the sample rate limits the number of samples per second that can be exported, discarding the ones above the limit. By default, a sampling rate is defined when a sampler is initialized in the service code, but this value can be overridden using the client with the command `create rate <resouce> <sampler> <limit>`
<!--how-to-end-->

<!--ref-start-->
# Reference

## Commands

### help

Shows all the available commands

### list

Lists elements: `Samplers`, `Resources`, or `Sampling Rules`. For example, `list samplers` shows the list of registered samplers.

### create/update/delete rule

Allows the definition and modification of `Sampling Rules`. For example: `create rule <resource_name> <sampler_name> <sampling_rule>` (sampling rule as a CEL expression) creates a new `Sampling Rule` on the specified `Sampler`.

### create/update/delete rate

Create/update and delete sampling rates. This configuration limits how many samples can be exported. For example: `create rate <resource> <sampler> <samples_per_second>`

## Using wildcards

In commands where a `<sampler>` or a `<resource>` is set, the special character `*` can be used to match all the entries. For example:

- `list rules * *`: It shows the rules for all resources and `Samplers`
- `list rules * sampler1`: It shows the rules for the resources that have a `sampler1` sampler
- `list rules resource1 *`: It shows all the rules of all the samplers of `resource1` resource
<!--ref-end-->