# Storage

All the *Data Samples* and *Data Telemetry* are represented as semi-structured data, such as JSON documents, and may not have a fixed schema. Therefore, any document database capable of efficiently storing and, more importantly, querying this data format can be used as a *Data Sample* and *Data Telemetry* storage.

At the same time, *Value Digests* *Data Telemetry* can also be found in the format of metrics.

## Store OTLP logs
*Data Samples* and *Data Telemetry* are encoded as [OpenTelemetry (OTLP) logs](https://opentelemetry.io/docs/reference/specification/logs/data-model), you have two main options to do store those:

- If the storage supports ingestion of OTLP logs, use the [OTLP/gRPC](https://opentelemetry.io/docs/reference/specification/protocol/otlp/#otlpgrpc) to send those.
- Use one of the supported [contrib exporters](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter) to automatically transform OTLP logs to the representation supported by your storage (for example [clickhouse](https://clickhouse.com/), [elasticsearch](https://www.elastic.co/), etc.). It's possible that the provided Neblic collector does not include the exporter you want, if that's the case you will need to [build your own collector](../how-to/build-your-own-collector.md).

You need to take into account that once you have stored them in a database, you want to be able to explore them easily and efficiently. For example, you want to be able to perform searches using expressions that can target semi-structured nested objects without a fixed schema (for example, if you have *Data Sample* with these contents: `{id: "1", product_name: "T-Shirt", price: -10 }`, you should be able to create an expression that targets this data similar to `sample.price < 0`). You may think that this is pretty common, but in practice, not many open-source databases allow you to easily explore data in this way.

### Details

| OpenTelemetry                                                                                    | Neblic                                                                                            |
| ------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------- |
| [Resource](https://opentelemetry.io/docs/specs/otel/resource/sdk/)                               | Resource                                                                                          |
| [InstrumentationScope](https://opentelemetry.io/docs/specs/otel/glossary/#instrumentation-scope) | [Sampler](../getting-started/concepts.md#sampler)                                                 |
| Attribute `com.neblic.sample.stream.uids`                                                        | [Stream](../getting-started/concepts.md#stream)                                                   |
| Attribute `com.neblic.sample.key`                                                                | [Key](../getting-started/concepts.md#keyed-stream)                                                |
| Attribute `com.neblic.sample.type`                                                               | `config`, `raw`, `struct-digest`, `value-digest`, `event`                                         |
| Attribute `com.neblic.sample.encoding`                                                           | `json`                                                                                            |
| Attribute `com.neblic.event.uid`                                                                 | Event UID in case `com.neblic.sample.type` is `event`. Empty otherwise                            |
| Attribute `com.neblic.event.rule`                                                                | Event rule in case `com.neblic.sample.type` is `event`. Empty otherwise                           |
| Attribute `com.neblic.digest.uid`                                                                | Digest UID in case `com.neblic.sample.type` is `value-digest` or `struct-digest`. Empty otherwise |
| Body                                                                                             | Sample contents                                                                                   |

## Store OTLP metrics
*Value Digests* are encoded as [OpenTelemetry (OTLP) metrics](https://opentelemetry.io/docs/specs/otel/metrics/data-model/). you have two main options to store those:

- If the storage supports ingestion of OTLP metrics, use the [OTLP/gRPC](https://opentelemetry.io/docs/reference/specification/protocol/otlp/#otlpgrpc) to send those.
- Use one of the supported [contrib exporters](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter) to automatically transform OTLP metrics to the representation supported by your storage (for example [clickhouse](https://clickhouse.com/), [influx](https://www.influxdata.com/), etc.). It's possible that the provided Neblic collector does not include the exporter you want, if that's the case you will need to [build your own collector](../how-to/build-your-own-collector.md).

### Details

| OpenTelemetry                                                                                    | Neblic                                                                                                         |
| ------------------------------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------- |
| [Resource](https://opentelemetry.io/docs/specs/otel/resource/sdk/)                               | Resource                                                                                                       |
| [InstrumentationScope](https://opentelemetry.io/docs/specs/otel/glossary/#instrumentation-scope) | [Sampler](../getting-started/concepts.md#sampler)                                                              |
| Attribute `com.neblic.sample.stream.uids`                                                        | [Stream](../getting-started/concepts.md#stream)                                                                |
| Attribute `com.neblic.sample.type`                                                               | `value-digest`                                                                                                 |
| Attribute `com.neblic.sample.field.path`                                                         | Path to a field in a `Data Sample` (e.g `["$":obj]["order":obj]["shipping_cost":obj]["currency_code":string]`) |
| Attribute `com.neblic.sample.name`                                                               | Metric name (e.g. `cardinality`)                                                                               |
| Attribute `com.neblic.event.uid`                                                                 | Event UID in case `com.neblic.sample.type` is `event`. Empty otherwise                                         |
| Attribute `com.neblic.digest.uid`                                                                | Digest UID in case `com.neblic.sample.type` is `value-digest` or `struct-digest`. Empty otherwise              |
| Body                                                                                             | Sample contents                                                                                                |