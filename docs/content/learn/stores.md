# Stores

`Data Samples` are represented as semi-structured data, such as JSON documents, and may not have a fixed schema. Therefore, any document database capable of efficiently storing and, more importantly, querying this data format can be used as a `Data Samples` store. 

Of course, you need some way to insert these `Data Samples` into the database, but since `Data Samples` are encoded as [OpenTelemetry (OTLP) logs](https://opentelemetry.io/docs/reference/specification/logs/data-model), and use the [OTLP/gRPC](https://opentelemetry.io/docs/reference/specification/protocol/otlp/#otlpgrpc) protocol for transport, any pipeline capable of ingesting OTLP logs can be used to store `Data Samples` in a database.

You need to keep in ming account that once you have stored them in a database, you want to be able to explore them easily and efficiently. For example, you want to be able to perform searches using expressions that can target semi-structured nested objects without a fixed schema (for example, if your `Data Sample` has these contents: `{id: "1", product_name: "T-Shirt", price: -10 }`, you should be able to create an expression that targets this data similar to `sample.price < 0`). You may think that this is pretty common, but in practice, not many open-source databases allow you to easily explore data in this way.

Therefore, this page shows a list of databases and UIs that provide a good Neblic experience.

## Grafana Loki

[Grafana Loki](https://grafana.com/oss/loki/), quoting their website, `is a log aggregation system designed to store and query logs from all your applications and infrastructure`. `Data Samples` are similar to logs so `Loki's` architecture and design fit well with Neblic goals.

It is the primary open-source option for storing `Data Samples` due to its ease of deployment and maintenance. In also integrates perfectly with [Grafana](https://grafana.com/grafana/), making it easy and efficient to explore data samples.

### Concepts

!!! note
    For more details in how `Grafana Loki` works, refer to its [documentation](https://grafana.com/docs/loki/latest). This page only describes some concepts relevant to how Neblic uses it.

#### Labels

[Labels](https://grafana.com/docs/loki/latest/fundamentals/labels/) are one of the most important concepts of `Grafana Loki`. See its documentation for more details. To simplify, you can think of labels as Loki's equivalent to indexes. To use Neblic efficiently, it is recommended to set the `Sampler` name and resource (which uniquely identifies a set of samplers) as labels. This way, queries that span one or more `Samplers` are performed fast and efficiently. The default configuration using the Neblic collector stores `Data Samples` this way.

### Configuration

No special configuration is required to use `Grafana Loki` as a Neblic store, other than making sure that the correct labels are set when indexing `Data Samples`. 

### Visualization

To explore `Data Samples` stored in `Grafana Loki`, `Grafana` is the best option. 

First, you need to add `Grafana Loki` as a `Data Source`, the official documentation provides a [guide](https://grafana.com/docs/grafana/latest/datasources/loki/).

Then, you can use the `Explore` tab to perform queries and explore your `Data Samples`. You will need to:

* Select `Loki` on the top-level dropdown to explore its contents.
* Apply a `label filter` selecting one or more `sampler_name` labels to filter what `Samplers` you want to explore.
* Optionally, add a `json` filter so it parses the contents of the `Data Samples`.

![grafana explore order-confirmation config](../assets/imgs/grafana-explore-order-confirmation-config.png)

There are two ways to filter `Data Samples`, [line](https://grafana.com/docs/loki/latest/logql/log_queries/#line-filter-expression) and [label](https://grafana.com/docs/loki/latest/logql/log_queries/#label-filter-expression) filter expressions. A `line filter expression`, as described in `Loki's` documentation, is a distributed grep over the entire `Data Sample` body, decoded as a JSON document. It supports regular expressions.

You can also use a `label filter expression` when adding a `json` filter. This filter will make `Grafana Loki` parse the `Data Sample` body and create a label for each field. Then, you can use predicates to filter them. See the documentation for more details.

![grafana explore order-confirmation data sample](../assets/imgs/grafana-explore-order-confirmation-data-sample.png)

This image shows the labels that get created when you use the `json` filter to parse the `Data Sample`. Given that `Data Samples` are encoded as OpenTelemetry logs, the contents of the `Data Sample` are in the `body` key. This makes all labels to be prepended by `body_`.