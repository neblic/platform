package sampler_example;

import java.util.Arrays;

// --8<-- [start:InitImport]
import io.opentelemetry.api.common.AttributeKey;
import io.opentelemetry.api.logs.Logger;
import io.opentelemetry.exporter.otlp.logs.OtlpGrpcLogRecordExporter;
import io.opentelemetry.sdk.OpenTelemetrySdk;
import io.opentelemetry.sdk.logs.SdkLoggerProvider;
import io.opentelemetry.sdk.logs.export.BatchLogRecordProcessor;
import io.opentelemetry.sdk.resources.Resource;
import io.opentelemetry.semconv.resource.attributes.ResourceAttributes;
// --8<-- [end:InitImport]

public final class SamplerExample {

  // --8<-- [start:ProviderInit]
  private static SdkLoggerProvider initProvider() {
    OpenTelemetrySdk otelSdk = OpenTelemetrySdk.builder()
        .setLoggerProvider(
            SdkLoggerProvider.builder()
                .setResource(
                    Resource.getDefault().toBuilder()
                        .put(ResourceAttributes.SERVICE_NAME, "service-name")
                        .build())
                .addLogRecordProcessor(
                    BatchLogRecordProcessor.builder(
                        OtlpGrpcLogRecordExporter.builder()
                            .setEndpoint("http://" + System.getenv("COLLECTOR_SERVICE_ADDR"))
                            .build())
                        .build())
                .build())
        .buildAndRegisterGlobal();

    return otelSdk.getSdkLoggerProvider();
  }
  // --8<-- [end:ProviderInit]

  // --8<-- [start:SamplerInit]
  private static Logger initSampler(SdkLoggerProvider provider) {
    return provider.get("sampler-name");
  }
  // --8<-- [end:SamplerInit]

  // --8<-- [start:SampleData]
  private static void sampleData(Logger sampler, String data) {
    sampler.logRecordBuilder()
        .setAttribute(AttributeKey.stringArrayKey("com.neblic.sample.stream.names"), Arrays.asList("all"))
        .setAttribute(AttributeKey.stringKey("com.neblic.sample.type"), "raw")
        .setAttribute(AttributeKey.stringKey("com.neblic.sample.key"), "")
        .setAttribute(AttributeKey.stringKey("com.neblic.sample.encoding"), "json")
        .setBody(data)
        .emit();
  }
  // --8<-- [end:SampleData]

  public static void main(String[] args) throws InterruptedException {

    SdkLoggerProvider provider = initProvider();
    Logger sampler = initSampler(provider);
    sampleData(sampler, "{\"foo\": 1, \"bar\": \"baz\"}");
  }
}
