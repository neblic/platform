# Collector

Neblic uses a custom build of the [OpenTelemetry Collector](https://opentelemetry.io/docs/collector/). The OpenTelemetry Collector is an amazing modular collector built by the OpenTelemetry community that can be customized by adding components at build time. Neblic's `Control Plane` server is packaged in an [OpenTelemetry Collector processor](https://github.com/neblic/platform/tree/main/controlplane/server/otelcolext) so that it can be bundled into a custom OpenTelemetry Collector build using the [OpenTelemetry Collector Builder (ocb)](../how-to/build-your-own-collector.md).

An OpenTelemtry Collector with the `Control Plane` server built-in is provided in each Neblic Platform release.

* Binaries can be found on the [GitHub releases](https://github.com/neblic/platform/releases) page.
* Containers are available on the [GitHub packages](https://github.com/neblic/platform/pkgs/container/otelcol) page.

You can see what other components are included in its ocb configuration file (see [Appendix A](#appendix-a)). If you want to build your own OpenTelemetry Collector with the Neblic `Control Plane` server and additional components, you can check [this](../how-to/build-your-own-collector.md) page.

!!! note
    Since Neblic's `Control Plane` server is the central point where all `Samplers` register and where clients connect to configure them, it is not recommended to run multiple collectors (e.g. as an agent in each host) running in the same cluster. If you do so, you will have to connect to multiple locations to configure your `Samplers`.

## Deployment

The recommended approach is to deploy the Collector using the provided container. You can find the latest release in [here](https://github.com/neblic/platform/pkgs/container/otelcol).

## Configuration

The collector is configured using a YAML configuration file. However, if you are using a container, you can configure most options using environment variables. To use a custom configuration file run use the `--config` flag.

```bash
otelcol --config /path/to/config.yaml
```

[Here](../reference/collector.md) you can find an up-to-date complete configuration file that you can use as a reference, the reference configuration file is also the configuration that is shipped with the container. There are three main sections, configuring the `Data Plane`, configuring the `Control Plane`, and configuring the `Data Samples` exporter.

You may also find useful the example `Docker Compose` and `Kubernetes` configurations included in [Appendix B](#appendix-b).

### Control plane

!!! Note
    Even if you are only deploying the collector to use the included `Control Plane` server, you still need to configure and enable a data pipeline with a receiver and an exporter.

See the annotated [reference](../reference/collector.md) documentation (`processor.neblic` section) to see what options are required to configure the `Control Plane` server.

To communicate with the `Collector Control Plane` you can use the CLI command `neblictl`. This [page](../how-to/configure-samplers-using-neblictl.md) shows how to use it to configure `Samplers`.

### Data Plane

The `Data Plane` uses the standard [OTLP logging receiver](https://github.com/open-telemetry/opentelemetry-collector/blob/main/receiver/otlpreceiver/README.md). Neblic doesn't require any special configuration, so it is enough to simply enable it by setting up an endpoint.

Neblic also provides a custom Bearer authenticator that can be used to authenticate `Sampler` connections when TLS is enabled. You will only need it if you want to connect using a Bearer token to authenticate with the `Data Plane` server.

## Appendix A:

``` yaml
--8<-- "./dist/otelcol/ocb.yaml"
```

## Appendix B:

### docker-compose

``` yaml
--8<-- "./dist/otelcol/compose/docker-compose.yaml"
```

### kubernetes

Create a secret with your exporter API token

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: neblic-otelcol
stringData:
  exporter-bearer-token: <your-api-token>
```

and then a StatefulSet and a Service definition to deploy the `Collector`.

``` yaml
--8<-- "./dist/otelcol/k8s/statefulset.yaml"
```
