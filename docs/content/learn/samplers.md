# Samplers

*Samplers* are available as libraries to be imported into your services or are created by standalone components that retrieve *Data Samples* from a system (such as a data broker or a database) by themselves such as the [`kafka-sampler`](../how-to/data-from-kafka.md).

They are designed to not interfere with the normal operation of your systems and to not impact your application performace. Go to the [Data Collection](/learn/data-collection/#sampler) page to get a deeper understanding of how a *Sampler* works.

*Samplers* need to be able to decode the data that they intercept so that it can be evaluated by their configured *Streams*, which decide whether or not that *Data Sample* needs to be processed. Therefore, you need to choose a *Sampler* that is compatible with the message encoding (e.g. Protocol Buffers, JSON...) that your service works with.

!!! note
    All *Samplers* are able to process *JSON* messages. Since it is a self-describing language, it is enough with the message itself (no external schema required) to be able to decode its contents. And since, at least when using *Samplers* within your services, it is usually possible to convert any object to *JSON*, this option works as a fallback in case the encoding your service uses is not supported. Of course, there is a performance penalty to consider when converting messages to *JSON*. 

## Best practices

Because their performance impact is negligible when no *Streams* are configured, it is recommended to add them wherever data is transformed or exchanged. This will allow you to track how your data evolves throughout your system. 

!!! note
    Unlike logs, where it is usually recommended to not add logging in the critical path to avoid too much noise and increased costs, *Samplers* can be dynamically configured so you can add them anywhere without worrying about impacting your application or costs. 

*Samplers* have sensible defaults to protect your services and never export large amounts of data without your permission. By default, they have a rate limit that puts a ceiling on how many *Data Samples* they export per second. All of these default options can be adjusted at runtime using a Neblic client, like *neblictl*, if you need it (e.g. temporarily increase the rate limit while troubleshooting an issue).

For example, it is common to add *Samplers* in all service or module boundaries:

* Requests and responses sent between services (e.g. HTTP API requests, gRPC calls...)
* Data passed to/from modules/libraries used within the same service
* Requests/responses or messages to external systems (e.g. DBs, Apache Kafka...)

Other interesting places could be:

* Before/after relevant data transformation
* When a service starts, to register its configuration

To make it easier to get *Data Samples* from multiple places, Neblic provides helpers, wrappers, and components that can automatically add *Samplers* in multiple places in your system e.g. in all gRPC requests/responses or in multiple Kafka topics. Check out the next sections to see what *Samplers* Neblic provides.

## Configuration

The pair *Sampler* name and resource id is what identifies a particular set of *Samplers*. For example, if you have multiple replicas of the same service, each replica will register a *Sampler* with the same name and resource id. All of these *Samplers* are treated as a group and you can configure them all together.

The [Data Collection](/learn/data-collection/) page shows all the operations performed by a *Sampler* and the *Collector* and provide the foundation to understand how to configure the *Sampler* behaviour. See this [how-to](../how-to/configure-samplers-using-neblictl.md) page to learn how to configure *Samplers* using *neblictl*.

### Streams

*Streams* are the initial configuration that *Samplers* need to be able to start generating *Data Telemetry*. To create a *Stream* you need to provide a target *Sampler* and a [rule](/reference/rules) which will select which *Data Samples* are part of the *Stream*.

*Data Samples* can have a key associated with them. By setting up a *Stream* as keyed, you can gather *Data Telemetry* separately for each key value. This feature is handy, for example, for collecting independent *Data Telemetry* for different customers (by using the customer ID as the key) or for various *Event* types (using the event type ID as the key). Currently, only *Events* are compatible with keyed *Streams*. For details on which functions are compatible with keyed *Streams*, please check the [reference table](/reference/rules).

### Digests and Metrics

*Digests* are generated at the *Stream* level. First, you need to create a *Stream* and then you will be able to generate the required *Digests*. *Metrics* are generated from *Digests* so you first need to create a *Digest* and then the *Collector* will automatically generate and export *Metrics* based on its contents.

### Events

*Events* are generated in the *Collector* and also work at the *Stream* level. To generate *Events* you will need to create a *Stream* and configure it to export *Raw Data* to the *Collector*. 

Then, you can create *Events* by specifying the target *Stream* and a [rule](/reference/rules) that will trigger the generation of the *Event*.

## Available Samplers

### Go

!!! info
    Check [this guide](../how-to/data-from-go-svc.md) for an example of how to use it and the [Godoc](https://pkg.go.dev/github.com/neblic/platform/sampler) page for reference.

{%
   include-markdown "../../../sampler/README.md"
   start="<!--learn-start-->"
   end="<!--learn-end-->"
%}


### Kafka

{%
   include-markdown "../../../cmd/kafka-sampler/README.md"
   start="<!--learn-start-->"
   end="<!--learn-end-->"
%}
 
Check [this guide](../how-to/data-from-kafka.md) to learn how to use it.

## Advanced

### Using OpenTelemetry SDK

The Neblic collector is built on top of OpenTelemetry stack, and as a result, the neblic collector is capable of understanding and processing samples encoded
as OpenTelemetry logs if they are correctly formatted. Any [OpenTelemetry SDK implementation supporting logs](https://opentelemetry.io/docs/languages/#status-and-releases) can be used to generate samples that neblic
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

- Create a [LoggerProvider](https://opentelemetry.io/docs/specs/otel/logs/bridge-api/#loggerprovider) with the desired *Resource* name.
- Create a [Logger](https://opentelemetry.io/docs/specs/otel/logs/bridge-api/#logger) with the desired sampler name as the *InstrumentationScope* name.
- Emit a log with:
    - Attribute `com.neblic.sample.stream.names* with value *all`
    - Attribute `com.neblic.sample.key` with the desired key value
    - Attribute `com.neblic.sample.type* with value *raw`
    - Attribute `com.neblic.sample.encoding* with value *json`
    - Body with the serialized version of the data

Once the collector receives the first sample, the sampler will appear to the controlplane as any other sampler (but with limited functionality) 
