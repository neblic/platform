# Go sampler module

<!--learn-start-->
<!-- ### Go  -->
The Go `Sampler` module allows you to get `Data Samples` from your Go services.

#### gRPC interceptors

The Go library also provides a package to easily add `Samplers` as gRPC interceptors. This will automatically add `Samplers` in each gRPC method that will intercept all requests and responses.

#### Supported encodings

| Encoding          | Description                                                                                        |
|-------------------|----------------------------------------------------------------------------------------------------|
| Protobuf          | It can efficiently process `proto.Message` objects if the Protobuf message definition is provided. |
| Go native object  | Only capable of accessing exported fields. This is a limitation of the Go language.                |
| JSON              | A string containing a JSON object.                                                                 |
<!--learn-end-->

> **Warning**
> The next section is better read in the [documentation page](https://docs.neblic.com/latest/how-to/data-from-go-svc/) since it is post-processed to include code snippets from tests

<!--how-to-start-->
## Usage

!!! note
    All code snippets have error handling omitted for brevity

To add `Samplers` in your code that export `Data Samples` you first need to initialize a `Provider`. A `Provider` receives a configuration that is then used to configure and initialize all the `Samplers` in your application.

``` go
import (
--8<-- "./sampler/test/docs/provider_test.go:ProviderInitImport"
)

--8<-- "./sampler/test/docs/provider_test.go:ProviderInit"
```

To see details about the required settings and available options, see this [page](https://pkg.go.dev/github.com/neblic/platform/sampler#pkg-types).

Once the `Provider` is initialized, you can use it to initialize `Samplers`. If you have registered the provider as global, you can simply call the `Sampler` method from anywhere in your application. If not, you will need to call the `Sampler` method of your provider. They both have the same signature so the following explanation works for both options.

!!! Info
    It is not required to first initialize and register the `Provider` as global before creating `Samplers`. If a `Sampler` is initialized using the global provider before a `Provider` is registered, it will return a stubbed `Sampler` with no-op methods. Once the `Provider` is registered, it will internally replace the no-op stub with the real `Sampler`. This happens transparently for the application.

``` go
import (
--8<-- "./sampler/test/docs/sampler_test.go:SamplerInitImport"
)

--8<-- "./sampler/test/docs/sampler_test.go:SamplerInit"
```

To see what other schemas the Go Sampler supports, check this [Godoc](https://pkg.go.dev) page.

Once you have initialized the `Sampler`, you can call any of its methods to have it evaluate a `Data Sample`. It will then be evaluated by any configured `Sampling Rule` and exported if there is a match.

!!! Warning
    You need to be mindful of what methods you use to sample data. Depending on the schema provided when the `Sampler` is initialized, some methods will work better or faster than others. 
    
    As a rule of thumb, you want to provide a schema if you have it since this allows the `Sampler` to internally optimize how it evaluates the `Sampling Rules`. If you do not have it, a sampler configured with a `DynamicSchema` is capable of processing any type data using any of the sampling methods. See the [Godoc](https://pkg.go.dev/github.com/neblic/platform/sampler/defs) documentation for details.

``` go
--8<-- "./sampler/test/docs/sampler_test.go:SampleData"
```

In this example, since the `Sampler` was initialized with a `DynamicSchema`, it is best to use the method `SampleJson()` or `SampleNative()`. These sampling methods are designed to work with samples that do not have a fixed or known schema.


## gRPC interceptor

If you use gRPC servers or clients in your services, you can make use of a gRPC [interceptor](https://github.com/neblic/platform/tree/main/sampler/instrumentation/google.golang.org/grpc). They will automatically create `Samplers` that will efficiently intercept all requests and responses. 

Internally, they create `Samplers` with a `ProtoSchema` so they do not need to deserialize the `Protobuf` message to evaluate its contents.

To use it, you need to initialize the interceptor and provide it when initializing the gRPC connection

``` go
import (
--8<-- "./sampler/test/docs/interceptor_test.go:InterceptorInitImport"
)

--8<-- "./sampler/test/docs/interceptor_test.go:InterceptorInit"
```

<!--how-to-end-->

<!--ref-start-->
<!-- Godoc page ->
<!--ref-end-->