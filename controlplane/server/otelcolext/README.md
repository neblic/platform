# OpenTelemetry collector integration

This package implements an `OpenTelemetry Collector` [connector](https://opentelemetry.io/docs/collector/configuration/#connectors) that can be easily integrated with a custom OpenTelemetry collector build as an extension.

<!--how-to-start-->
To build your own OpenTelemetry Collector that includes Neblic's `Control Plane` server and its `Data Plane`, you need to follow [this](https://opentelemetry.io/docs/collector/custom-collector/) guide to install the required tool and prepare a configuration file that includes all your other required components. Then, you need to include the `otelcolext` processor included in Neblic's platform source code:

``` yaml

...

connectors:
  - import: github.com/neblic/platform/controlplane/server/otelcolext
    gomod: github.com/neblic/platform vX.X.X # Set the proper version

...

extensions:
  # Optional: To be able to support `Sampler` `Data Plane` Bearer authentication
  - import: github.com/neblic/platform/controlplane/server/otelcolext/bearerauthextension
    gomod: github.com/neblic/platform vX.X.X # Set the proper version

...

```

You can use as a reference the [configuration file](https://github.com/neblic/platform/blob/main/dist/otelcol/ocb.yaml) used to build the collector that Neblic distributes.

Once built, you can learn more about how to use it and configure it in [here](https://docs.neblic.com/latest/learn/collector/)
<!--how-to-end-->
