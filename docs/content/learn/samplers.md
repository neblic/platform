# Samplers

`Samplers` are available as libraries to be imported into your services or are created by standalone components that retrieve `Data Samples` from a system (such as a data broker or a database) by themselves such as the [`kafka-sampler`](../how-to/data-from-kafka.md).

They are designed to not interfere with the normal operation of your systems and to have a negligible impact on performace. `Streams` of `Data Samples` are created using rules defined with the [Google's CEL language](https://opensource.google.com/projects/cel), which, quoting their documentation is `designed for simplicity, speed, safety, and portability`.

`Samplers` need to be able to decode the data that they intercept so that it can be evaluated by their configured `Streams`, which decide whether or not that `Data Sample` needs to be processed. Therefore, you need to choose a `Sampler` that is compatible with the message encoding (e.g. Protocol Buffers, JSON...) that your service works with.

!!! note
    All `Samplers` are able to process `JSON` messages. Since it is a self-describing language, it is enough with the message itself (no external schema required) to be able to decode its contents. And since, at least when using `Samplers` within your services, it is usually possible to convert any object to `JSON`, this option works as a fallback in case the encoding your service uses is not supported. Of course, there is a performance penalty to consider when converting messages to `JSON`. 

The `Data Samples` selected to be processed are then parsed so they can be incorporated to the `Sampler` enabled `Digests` or forwarded to the `Collector` as-is so they can be processed there. This depends on how are they configured.

## Best practices

Because their performance impact is negligible when no `Streams` are configured, it is recommended to add them wherever data is transformed or exchanged. This will allow you to track how your data evolves throughout your system. 

!!! note
    Unlike logs, where it is usually recommended to not add logging in the critical path to avoid too much noise and increased costs, `Samplers` can be dynamically configured so you can add them anywhere without worrying about impacting your application or costs. 

`Samplers` have sensible defaults to protect your services and never export large amounts of data without your permission. By default, they have a rate limit that puts a ceiling on how many `Data Samples` they export per second. All of these default options can be adjusted at runtime using a Neblic client, like `neblictl`, if the user needs it (e.g. temporarily increase the rate limit while troubleshooting an issue).

For example, it is common to add `Samplers` in all service or module boundaries:

* Requests and responses sent between services (e.g. HTTP API requests, gRPC calls...)
* Data passed to/from modules/libraries used within the same service
* Requests/responses or messages to external systems (e.g. DBs, Apache Kafka...)

Other interesting places could be:

* Before/after relevant data transformation
* When a service starts to register its configuration

To make it easier to get `Data Samples` from multiple places, Neblic provides helpers, wrappers, and components that can automatically add `Samplers` in multiple places in your system e.g. in all gRPC requests/responses or in multiple Kafka topics. Check out the next sections to see what `Samplers` Neblic provides.

## Configuration

The pair `Sampler` name and resource id is what identifies a particular set of `Samplers`. For example, if you have multiple replicas of the same service, each replica will register a `Sampler` with the same name and resource id. All of these `Samplers` are treated as a group and you can configure them all together. However, each `Sampler` has a unique id in case you want to send a configuration to only one of the `Samplers`.

### Streams

`Streams` are the initial configuration that `Samplers` need to be able to start exporting `Data Samples`, generating `Digests` or `Events`. Clients (such as the CLI client `neblictl`) connect to the `Control Plane` server, usually running in your collector, and create `Streams` in the registered `Samplers`.

See this [how-to](../how-to/configure-samplers-using-neblictl.md) page to learn how to configure `Samplers` using `neblictl` and the [rules reference](../reference/rules.md) to learn what expressions you can use.

## Available Samplers

### Go

{%
   include-markdown "../../../sampler/README.md"
   start="<!--learn-start-->"
   end="<!--learn-end-->"
%}

Check [this guide](../how-to/data-from-go-svc.md) for an example of how to use it and the [Godoc](https://pkg.go.dev/github.com/neblic/platform/sampler) page for reference.

### Kafka

{%
   include-markdown "../../../cmd/kafka-sampler/README.md"
   start="<!--learn-start-->"
   end="<!--learn-end-->"
%}
 
Check [this guide](../how-to/data-from-kafka.md) to learn how to use it.

# Advanced

## Using OpenTelemetry SDK

The Neblic collector is built on top of OpenTelemetry stack, and as a result, the neblic collector is capable of understanding and processing samples encoded
as opentelemetry logs if they are correctly formatted. Any [OpenTelemetry SDK implementation supporting logs](https://opentelemetry.io/docs/languages/#status-and-releases) can be used to generate samples that neblic
will process.

Concept match beetween OpenTelemetry and Neblic:

| OpenTelemetry                                                                                    | Neblic                                             |
| ------------------------------------------------------------------------------------------------ | -------------------------------------------------- |
| [Resource](https://opentelemetry.io/docs/specs/otel/resource/sdk/)                               | Resource                                           |
| [InstrumentationScope](https://opentelemetry.io/docs/specs/otel/glossary/#instrumentation-scope) | [Sampler](../getting-started/concepts.md#sampler)  |
| Attribute `com.neblic.sample.stream.names`                                                       | [Stream](../getting-started/concepts.md#stream)    |
| Attribute `com.neblic.sample.key`                                                                | [Key](../getting-started/concepts.md#keyed-stream) |

!!! note
    OpenTelemetry recommends using appenders to propagate logs, for that use case it does not work, and the Logs API is used instead.

Steps to follow:

- Create a [LoggerProvider](https://opentelemetry.io/docs/specs/otel/logs/bridge-api/#loggerprovider) with the desired `Resource` name.
- Create a [Logger](https://opentelemetry.io/docs/specs/otel/logs/bridge-api/#logger) with the desired sampler name as the `InstrumentationScope` name.
- Emit a log with:
    - Attribute `com.neblic.sample.stream.names` with value `all`
    - Attribute `com.neblic.sample.key` with the desired key value
    - Attribute `com.neblic.sample.type` with value `raw`
    - Attribute `com.neblic.sample.encoding` with value `json`
    - Body with the serialized version of the data

Once the collector receives the first sample, the sampler will appear to the controlplane as any other sampler (but with limited functionality) 
