# Get Data Samples from Java services

!!! warning
    There is no Neblic SDK for Java, this guide shows how to send *Data Samples* as OpenTelemetry logs directly to the *Collector*

To add *Samplers* in your code that export *Data Samples* you first need to initialize a *SdkLoggerProvider*. A *SdkLoggerProvider* receives a configuration that is then used to configure and initialize all the *Loggers* in your application.

``` java
--8<-- "./sampler/test/docs/java/sampler_example/src/main/java/sampler_example/SamplerExample.java:InitImport"

public final class SamplerExample {

  ...

--8<-- "./sampler/test/docs/java/sampler_example/src/main/java/sampler_example/SamplerExample.java:ProviderInit"

  ...

}
```

Once the *SdkLoggerProvider* is initialized, you can use it to initialize a *Logger*. To do that, you will need to call the *get* method of your provider.

``` java
public final class SamplerExample {

  ...

--8<-- "./sampler/test/docs/java/sampler_example/src/main/java/sampler_example/SamplerExample.java:SamplerInit"

  ...

}
```

Once you have initialized the *Logger*, you have to create a *LogRecordBuilder*, set the required attributes and body, and call the *emit* function to send the *Data Sample*.

!!! Note
    Full list of parameters to provide to the *LogRecord* can be found in [*learn about opentelemetry samplers*](../learn/samplers.md#using-opentelemetry-sdk)

``` java
public final class SamplerExample {

  ...

--8<-- "./sampler/test/docs/java/sampler_example/src/main/java/sampler_example/SamplerExample.java:SampleData"

  ...

}
```

## Full example
If we put everything together, we get the following.

``` java
{%
   include "../../../sampler/test/docs/java/sampler_example/src/main/java/sampler_example/SamplerExample.java"
%}
```
