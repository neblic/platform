# Get data from Apache Kafka

Neblic provides a standalone service called `kafka-sampler* capable of monitoring your *Apache Kafka* topics by automatically creating *Samplers` that will read the data that flows through them.

!!! Note
    The *Samplers* created by `kafka-sampler` require a [*Collector*](../learn/collector.md) to work so it is necessary to deploy a *Collector* as well. Follow [this](../getting-started/usage.md#collector) guide first to deploy it.

{%
   include-markdown "../../../cmd/kafka-sampler/README.md"
   start="<!--how-to-start-->"
   end="<!--how-to-end-->"
%}

## Configuration

To see all the configuration options available, check the `kafka-sampler` reference [page](../reference/kafka-sampler.md).
