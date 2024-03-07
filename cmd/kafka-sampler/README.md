# Neblic Kafka Sampler

<!--learn-start-->
<!-- ### Kafka -->
Neblic provides a standalone service called `kafka-sampler` capable of automatically monitoring your *Apache Kafka* topics and creating *Samplers* that will allow you to inspect all data that flows through them.

#### Supported encodings

| Encoding          | Description                                                                                        |
|-------------------|----------------------------------------------------------------------------------------------------|
| JSON              | A string containing a JSON object.                                                                 |

#### Instrumentation overhead (advanced)

The `kafka-sampler` service is based on the `Go` *Sampler*. Check its [overhead analysis](https://docs.neblic.com/latest/learn/samplers/#instrumentation-overhead-advanced) for details.

<!--learn-end-->

<!--how-to-start-->
## Deployment

See the [releases](https://github.com/neblic/platform/releases) page to download the latest binary or the [packages](https://github.com/neblic/platform/pkgs/container/kafka-sampler) page to see the available containers. It is recommended to use the provided container image to deploy `kafka-sampler`. The following section describes how to deploy it using the container image.

### Container

#### Supported architectures

For now, only `x86-64` builds are offered. If you need another architecture you can build your own container using the files found in [here](https://github.com/neblic/platform/tree/main/dist/kafka-sampler).

#### Examples

##### docker-compose

``` yaml
--8<-- "./dist/kafka-sampler/compose/docker-compose.yaml"
```

##### kubernetes

``` yaml
--8<-- "./dist/kafka-sampler/k8s/deployment.yaml"
```

## Usage

On startup, it will subscribe to all or a subset (based on your configuration, see the [reference](https://docs.neblic.com/latest/reference/kafka-sampler/) page) of your Kafka topics and create a *Sampler* per each one. No other actions are required since it will automatically register the *Samplers* with Neblic's *Control Plane* server and keep the list of *Samplers* updated if topics are added or removed.
<!--how-to-end-->

<!--ref-start-->
## Configuration 

By default, `kafka-sampler* will look for a configuration file at */etc/neblic/kafka-sampler/config.yaml`. This path can be changed using the `--config` flag when executing the service.

All the options defined in the configuration file can be configured/overridden using environment variables. The environment variable name needs to be written in all caps and use `_` to divide nested objects. For example, to configure the Kafka server URL you would need to use the env variable `KAFKA_SERVERS`.

Internally, `kafka-sampler` uses the [*Sarama*](https://github.com/IBM/sarama/) Go library to interact with Kafka and all its [options](https://pkg.go.dev/github.com/IBM/sarama#Config) can be configured under the `kafka.sarama` key. See the following examples section to see advanced configurations.

### Examples

#### Topic monitoring selection

The maximum number of topics to monitor is defined by the key `kafka.topics.max` and by default,  it will monitor the first max number of topics that match the filter rules (in no particular order), the rest will be ignored.

To configure what topics are selected you can use the key `kafka.topics.filter.allow* or the key *kafka.topics.filter.deny`, using both options at the same time is not supported. The value should follow regex RE2 syntax as described in [here](https://github.com/google/re2/wiki/Syntax). For example, to only monitor `topic1* and *topic2` topics:

| Config file YAML key              | Env var                            | Value               |
|-----------------------------------|------------------------------------|---------------------|
| `kafka.topics.filter.allow`       | `KAFKA_TOPICS_FILTER_ALLOW`        | `^(topic1|topic2)$` |

Or to monitor all topics but `topic3`:

| Config file YAML key              | Env var                            | Value               |
|-----------------------------------|------------------------------------|---------------------|
| `kafka.topics.filter.deny`        | `KAFKA_TOPICS_FILTER_DENY`         | `^topic3$` |

Take into account that `kafka-sampler` automatically discovers new topics so if the configuration is not too restrictive it will automatically monitor new topics as they are created. 

#### Apache Kafka authentication

Kafka supports many authentication methods, since its configuration is not straightforward you can use these examples to get started:

##### SASL/PLAIN

| Config file YAML key              | Env var                            | Value       |
|-----------------------------------|------------------------------------|-------------|
| `kafka.sarama.net.sasl.enable`    | `KAFKA_SARAMA_NET_SASL_ENABLE`     | `true`      |
| `kafka.sarama.net.sasl.user`      | `KAFKA_SARAMA_NET_SASL_USER`       | `<username>`|
| `kafka.sarama.net.sasl.password`  | `KAFKA_SARAMA_NET_SASL_PASSWORD`   | `<password>`|

##### SASL/SCRAM

| Config file YAML key              | Env var                            | Value                               |
|-----------------------------------|------------------------------------|-------------------------------------|
| `kafka.sarama.net.sasl.enable`    | `KAFKA_SARAMA_NET_SASL_ENABLE`     | `true`                              |
| `kafka.sarama.net.sasl.mechanism` | `KAFKA_SARAMA_NET_SASL_MECHANISM`  | `SCRAM-SHA-256` or `SCRAM-SHA-512`  |
| `kafka.sarama.net.sasl.user`      | `KAFKA_SARAMA_NET_SASL_USER`       | `<username>`                        |
| `kafka.sarama.net.sasl.password`  | `KAFKA_SARAMA_NET_SASL_PASSWORD`   | `<password>`                        |


<!--ref-end-->

<!-- Link to reference configuration. In the documentation, this file is directly embedded in the reference section -->
### Reference configuration file

A commented complete configuration file is available [here](../../dist/kafka-sampler/config.yaml)
