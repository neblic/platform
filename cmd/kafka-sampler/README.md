# Neblic Kafka Sampler

<!--learn-start-->
<!-- ### Kafka -->
Neblic provides a standalone service called `kafk-sampler` capable of automatically monitoring your `Apache Kafka` topics and creating `Samplers` that will allow you to inspect all data that flows through them.

#### Supported encodings

| Encoding          | Description                                                                                        |
|-------------------|----------------------------------------------------------------------------------------------------|
| JSON              | A string containing a JSON object.                                                                 |
<!--learn-end-->

<!--how-to-start-->
## Deployment

See the [releases](https://github.com/neblic/platform/releases) page to download the latest binary or the [packages](https://github.com/orgs/neblic/packages?repo_name=platform) page to see the available containers. It is recommended to use the provided container image to deploy `kafka-sampler`.

## Usage

On startup, it will subscribe to all or a subset (based on your configuration) of your Kafka topics and create a `Sampler` per each one. No other actions are required since it will automatically register the `Samplers` with the `Control Plane` server and keep the list of `Samplers` updated if topics are added or removed.
<!--how-to-end-->

<!--ref-start-->
## Configuration 

By default, `kafka-sampler` will look for a config file at `/etc/neblic/kafka-sampler/config.yaml`.

All the options defined in the configuration file can be configured/overridden using environment variables. The environment variable name will be written in all caps and using `_` to divide nested objects. For example, to configure the Kafka server URL you would need to use the env variable `KAFKA_SERVERS`.
<!--ref-end-->

<!-- Link to reference configuration. In the documentation, this file is directly embedded in the reference section -->
## Reference configuration file

A commented complete configuration file is available [here](../../dist/kafka-sampler/config.yaml)