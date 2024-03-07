# Neblic Platform

[![CI](https://github.com/neblic/platform/actions/workflows/ci_checks.yaml/badge.svg?branch=main)](https://github.com/neblic/platform/actions/workflows/ci_checks.yaml?query=branch%3Amain)
[![PkgGoDev](https://pkg.go.dev/badge/platform)](https://pkg.go.dev/github.com/neblic/platform)


## What is Neblic

<!--what-is-neblic-start-->
Neblic provides application observability through data. Application data is monitored continuously, and it’s done at many points of the application as a way to have an understanding of how each component behaves and how that changes across time.  

Some of the core challenges of making data usable in a systematic way are: accessing the data, past and future, understanding the data including the business domain and its constraints, and knowing what to prioritize rather than getting overwhelmed by the sheer volume and complexity and all the possible dimensions that the data itself can take. 

To make the data practical and useful for application troubleshooting Neblic is built on analyzing the data based on three principles: Data Value Statistics, Data Structure and Business Logic and Data Validation.

### Data Values Statistics

Generate statistics about the data at field level that can be tailored to what is relevant for each field within a specific application. For instance, min-max-avg to understand numerical distributions across time, cardinality to understand distinct events, and nulls and zero-values as a way to find if data is missing unexpectedly. 

### Data Structure Analysis

Get field level structural digests that lets you understand the schema, field types and field presence for every period at every point of your application. This lets you visualize the actual schema seen, know the field type for each field and figure out if it has changed over time, and monitor field presence relative to each event. 

### Business Logic and Data Validation

Validate incoming data to get proactively alerted when something goes wrong. For instance, validate that the timestamps you’re receiving are current as a way to detect late events. Or correlate two events from two different fields or samplers to make sure they’re working correctly. 

We sometimes refer to the combination of these three elements, Value Stats, Structure Analysis and, Logic/Data Validation as Telemetry, or *Data Telemetry*.
<!--what-is-neblic-end-->

## How does Neblic Work

<!--how-does-neblic-work-start-->
Neblic requires two main components, *Samplers* and a *Collector*. *Samplers* sample data from each component in your application, the *Collector* aggregates that data, detects events and creates the *Data Telemetry*. Both components are built on top of OpenTelemetry as a way to ensure vendor neutrality and support standardization across the observability stack.

Operationally, its design allows for dynamic rule setting, dynamic control of the sampling rules and metrics generated.

![Architecture overview](./docs/content/assets/imgs/arch-overview.png)
<!--how-does-neblic-work-end-->

## Learn more

* Quickly spin up an instrumented e-commerce application and play with Neblic using our [playground](https://github.com/neblic/playground). See this [page](https://docs.neblic.com/latest/quickstart/playground/) for a step-by-step guide.
* Check Neblic [documentation](https://docs.neblic.com) to learn more about how it works and how to use it.

## Contributing

See the [contributing](./CONTRIBUTING.md) guide.
