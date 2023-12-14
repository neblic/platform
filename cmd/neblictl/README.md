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

#### Authentication

If authentication is enabled in the `Control Plane` server, you will need to set a `Bearer token`:

- Run `neblictl init` to create the configuration file if necessary and show its path.
- Open the configuration file and set the desired token value in the `Token` field.

### Configure a `Sampler`

Before continuing, it's useful to understand the concepts defined in this [page](https://docs.neblic.com/latest/getting-started/concepts/).

The format of the commands is `namespace:command`. Run the command `help` to see all of them and their commnds. The 4 main namespaces are:

* `sampler`: Configures options at the `sampler` level. This means they affect all `streams` defined in the `sampler`. For example, you can configure a maximum amount of samples processed or exported.
* `streams`: Creates a pipeline of data containing a subset of the `Data Samples`. On creation, you will need to provide a [rule](https://docs.neblic.com/latest/reference/rules/) that filters the `Data Samples` that will be part of that `Stream`.
* `digests`: Enables and configures `Digests`. On creation, a `Digest` requires a target `Stream`.
* `events`: Configure the generation of `Events`. Similarly to `Digests` they require a target `Stream` and additionally, they require a [rule](https://docs.neblic.com/latest/reference/rules/) that determines when has occurred the `Event`.

<!--how-to-end-->

<!--ref-start-->
# Reference

## Commands

Format `namespace:command` (e.g `samplers:list`)

```
resources: A resource identifies a service or a group of services that are part of the same logical application.
   o list: List all resources

samplers: A sampler is a component that collects samples from a resource.
   o list: List all samplers
   o list:config: List all samplers configurations
   o limiterin:set: Sets the maximum number of samples processed per second by a sampler
   o limiterin:unset: Unsets the maximum number of samples per second processed by a sampler
   o samplerin:set:deterministic: Sets a deterministic samplerin configuration
   o samplerin:unset: Unsets any samplerin configuration set
   o limiterout:set: Sets the maximum number of samples exported per second by a sampler
   o limiterout:unset: Unsets the maximum number of samples per second exported by a sampler

streams: A stream is a sequence of samples collected from a resource by a sampler.
   o list: List streams
   o create: Create streams
   o update: Update streams
   o delete: Delete streams

digests:
   o list: List configured digests
   o structure:create: Generate structure digests
   o structure:update: Update structure digests
   o value:create: Configure generation of value digests
   o value:update: Update value digests
   o delete: Delete digest

events:
   o list: List configured events
   o create: Create events
   o update: Update events
   o delete: Delete events
```

<!--ref-end-->
