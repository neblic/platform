# OpenTelemetry collector integration

This package implements an OpenTelemetry collector [extension](https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/service-extensions.md) that can be easily integrated into a custom OpenTelemetry collector build.

To build a custom OpenTelemetry collector you'll need the [OpenTelemetry Collector Builder](https://github.com/open-telemetry/opentelemetry-collector/tree/main/cmd/builder) (OCB). To install it run:

```shell
$ go install go.opentelemetry.io/collector/cmd/builder@latest
```

Once installed, add the following lines to the OCB builder config file (in the example `ocb.yaml`) to include the Neblic and the Bearer Auth extensions (if needed):

```
extensions:
  - gomod: github.com/neblic/platform/controlplane/server/otelcolext [VERSION]
  - gomod: github.com/neblic/platform/controlplane/server/otelcolext/bearerauthextension [VERSION]
```

Replace `[VERSION]` with the target Neblic Platform version (e.g. v1.0.0). Then, build the OpenTelemetry collector.


```shell
$ builder --config=ocb.yaml
```

To enable the extension check the example configuration file `otelcol.yaml.example`. There are two important sections. To configure it:

```yaml
extensions:
  neblic:
    listen_addr: localhost:8899
```

And to enable it:

```yaml
service:
  extensions: [neblic]
```

Finally, to run the collector with this configuration execute:

```shell
$ ./dist/otelcol --config=./otelcol.yaml
```