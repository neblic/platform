# OpenTelemetry collector integration

This package implements an `OpenTelemetry Collector` [processor](https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/processing.md) that can be easily integrated with a custom OpenTelemetry collector build as an extension.

<!--how-to-start-->
To build your own OpenTelemetry Collector that includes Neblic's `Control Plane` server and its data processing functionalities, you need to follow [this](https://opentelemetry.io/docs/collector/custom-collector/) guide to install the required tool and prepare a configuration file that includes all your other required components. Then, you need to include the `otelcolext` processor included in Neblic's platform source code:

``` yaml
processors:
  - import: github.com/neblic/platform/controlplane/server/otelcolext
    gomod: github.com/neblic/platform vX.X.X # Set the proper version
    # Optional: To be able to support `Sampler` `Data Plane` Bearer authentication
  - import: github.com/neblic/platform/controlplane/server/otelcolext/bearerauthextension
    gomod: github.com/neblic/platform vX.X.X # Set the proper version
```

You can use as a reference the [configuration file](https://github.com/neblic/platform/blob/main/dist/otelcol/ocb.yaml) used to build the collector that Neblic distributes.

Once built, you can configure it as described [here](https://docs.neblic.com/latest/learn/collector/)
<!--how-to-end-->
