# Real-time data streaming

## What are real-time data streaming applications?

Real-time data streaming applications are designed to process and analyze data as it arrives, without significant delay. These applications are designed to handle high-velocity, high-volume data streams, and provide sub-second latency, enabling real-time insights and decision-making.

These characteristics make them challenging to debug since it is common to have intermediate states between operations and local caches that are not externally observable.

## How can Neblic help?

### Monitor data arrival patterns

Keeping track of when and in what order data arrives is often very important in the realm of real-time data streaming. The presence of late, duplicate, or out-of-sequence data can adversely affect the accuracy and reliability of the data pipeline and it can be challenging to notice. Neblic offers tools to help you monitor these situations and get alerted if they occur.

* Late data: Observe the evolution of the data timestamp to detect situations where it doesn't follow the expected pattern
* Out-of-Order Data: Similar to tracking late data, you can also monitor whether the data arrives in the expected sequence. This can be achieved by looking at the sequence number or timestamp within the data.
* Lost data: Ensure that the data sequence id or timestamp is an ever-increasing sequence or compare data received at different points of your application to easily determine if there has been a significant loss of messages.
* Duplicate data: Keep track of the number of unique  processed so you are aware if there has been a significant number of duplicates.

### Observe how data evolves across the pipeline

A real-time data processing pipeline is composed of a sequence of operations such as transformations, enrichments, and aggregations. These pipelines often perform multiple, consecutive in-memory operations that pull from various sources, ultimately persisting only the final outcome.

Troubleshooting is challenging when errors emerge, typically noticeable only at the pipeline's output stage, especially if it requires pulling data from different sources and the only available clues come from the pipeline inputs and outputs.

Using Neblic, you can get visibility into the structure of the data and statistics about its contents in between each operation. When an incorrect output is found, having information about how the data has evolved across the pipeline can help you pin-point the root cause. For example, common situations that Neblic can detect are:

* Failed transformations and aggregations: By observing changes in the data structure and its contents at various stages, you can identify which operation might be failing.
* Failed enrichments: Get metrics about the number of nulls and empty values that a certain field or fields have after an enrichment or track its value distribution to understand if the enrichment has failed or is behaving incorrectly.

### Get alerted on data invariants

You can also get proactively alerted by setting up alert conditions that trigger when specific situations related to your data pipeline business domain occur. For example, you can get alerted when there are:

* Outlier values
* A significant amount of null or empty values 
* Statistical outliers such as variations on the number of distinct values

The alert expression language is very powerful and allows you to set up complex rules that can catch advanced situations. 

!!! note
    Some features may be currently being developed and still not have all the capabilities described in here. Check the [project roadmap](https://github.com/orgs/neblic/projects/3) for details about what improvements are we working on.

    Also, do not hesitate to suggest features, improvements or use cases by creating a [feature request](https://github.com/neblic/platform/issues/new?assignees=&labels=&projects=&template=feature_request.md&title=) in our github!
