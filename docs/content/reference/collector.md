# Neblic Collector

The Neblic collector is based on the OpenTelemetry collector. This page shows an example configuration file for an OpenTelemetry collector built with the Neblic extension.

The options that have a value replaceable with an environment variable, `${env:VAR_NAME}`, are intended to be configured when using the container distribution of the collector.

``` yaml
--8<-- "./dist/otelcol/otelcol.yaml"
```

The `entrypoint.sh` file defines the default values for the environment variables.

``` sh
--8<-- "./dist/otelcol/entrypoint.sh"
```