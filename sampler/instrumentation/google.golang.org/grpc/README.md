# Golang gRPC interceptors

This module contains methods to automatically instrment gRPC clients and servers. When adding these interceptors, two independent `Samplers` will be created that will intercept all `Requests` and `Responses`. 

The name of these `Samplers` will be: `${FullMethod}Req` and `${FullMethod}Res`.
