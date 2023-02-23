# Collector

Neblic uses a custom build of the [OpenTelemetry Collector](https://opentelemetry.io/docs/collector/). The OpenTelemetry Collector is an amazing modular collector built by the OpenTelemetry community that can be customized by adding components at build time. Neblic's `Control Plane` server is packaged in an [OpenTelemetry Collector extension](https://github.com/neblic/platform/tree/main/controlplane/server/otelcolext) so that it can be bundled into a custom OpenTelemetry Collector build using the [OpenTelemetry Collector Builder (ocb)](../how-to/build-your-own-collector.md).

An OpenTelemtry Collector with the `Control Plane` server built-in is provided in each Neblic Platform release.

* Binaries can be found on the [GitHub releases](https://github.com/neblic/platform/releases) page.
* Containers are available on the [GitHub packages](https://github.com/neblic/platform/pkgs/container/otelcol) page.

You can see what other components are included in its ocb configuration file (see [Appendix A](#appendix-a)). If you want to build your own OpenTelemetry Collector with the Neblic `Control Plane` server and additional components, you can check [this](../how-to/build-your-own-collector.md) page.

!!! note
    Since Neblic's `Control Plane` server is the central point where all `Samplers` register and where clients connect to configure them, it is not recommended to run multiple collectors (e.g. as an agent in each host) running in the same cluster. If you do so, you will have to connect to multiple locations to configure your `Samplers`.

## Installation

The recommended approach is to deploy the Collector using the provided container. You can find the latest release in [here](https://github.com/neblic/platform/pkgs/container/otelcol).

## Configuration

The collector is configured using a YAML configuration file. However, if you are using a container, you can configure most options using environment variables. To use a custom configuration file run use the `--config` flag.

```bash
otelcol --config /path/to/config.yaml
```

 [Here](../reference/collector.md) you can find an up-to-date complete configuration file that you can use as a reference, the reference configuration file is also the configuration that is shipped with the container. There are three main sections, configuring the `Data Plane`, configuring the `Control Plane`, and configuring the `Data Samples` exporter.

### Data Plane

The `Data Plane` uses the standard [OTLP logging receiver](https://github.com/open-telemetry/opentelemetry-collector/blob/main/receiver/otlpreceiver/README.md). Neblic doesn't require any special configuration, so it is enough to simply enable it by setting up an endpoint.

Neblic also provides a custom Bearer authenticator that can be used to authenticate `Sampler` connections when TLS is enabled. You will only need it if you want to connect using a Bearer token to authenticate with the `Data Plane` server.

### Control plane

!!! Note
    Even if you are only deploying the collector to use the included `Control Plane` server, you still need to configure and enable a data pipeline with a receiver and an exporter.

See the annotated [reference](../reference/collector.md) documentation (`extensions.neblic` section) to see what options are required to configure the `Control Plane` server.

### Exporter

#### Loki exporter

In addition to configuring and enabling the [Loki exporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/lokiexporter), you should add a resource attributes process to set the name of the `Sampler` as a Loki label. This will create an index that will allow you to efficiently explore `Data Samples` in Loki. You can read more about why this is necessary on the [stores](../learn/stores.md#labels) documentation page.

#### Other exporters

Refer to your exporter documentation to learn how to configure it to save  `Data Samples` in your preferred store.

## Appendix A:

``` yaml
--8<-- "./dist/otelcol/ocb.yaml"
```