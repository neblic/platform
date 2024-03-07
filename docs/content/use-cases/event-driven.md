# Event-Driven applications

## What are event-driven applications?

Event-Driven Architecture (EDA) is a design paradigm that allows different parts of a system to communicate and perform actions in response to asynchronous events. In EDA, an event is a significant change or occurrence within a system. When this change happens, it triggers the sending of an event message to parts of the system that are designed to listen for and respond to such messages.

## How can Neblic help?

### Monitor event delivery reliability

Reliable event delivery is a critical part for the smooth operations of EDA systems. Common problems that you can monitor and get alerted on using Neblic:

* Out-of-order events: Monitor the sequence in which events are supposed to occur. This can be done by looking at the sequence number or timestamp included in the event itself.
* Lost events: To guard against losing events, you can track if the event sequence identifier is consistently increasing or compare the events received at different points within your application. This comparison helps in quickly identifying significant losses of events.
* Duplicate events: Keep track of the number of unique elements processed so you are aware if there has been a significant number of duplicates.

### Enhance dead-letter queues

Dead-letter queues (DLQs) are commonly used to handle messages that couldn’t be processed. They act as a holding area for these events, enabling you to look into them later to figure out the issues. However, DLQs often don't get the attention they need.

When events end up in your DLQs, Neblic can automatically notify you and provide detailed insights into the structure of these events, along with analytics on their content. This approach simplifies the process of understanding what went wrong, making it easier to identify and address problems without manual extraction and analysis. 

In some scenarios, replacing DLQs entirely with Neblic samplers can be a more efficient strategy. This is particularly applicable when there's no intention of reprocessing the failed events, and where Neblic's analysis provides all the information you need to understand and potentially rectify the issue.

### Monitor events structure and their contents

In EDA, bugs often manifest within the events exchanged throughout the system. These errors are often application-specific, making it challenging to offer examples here. However, using Neblic you can get information about the structure and contents of the events produced in any part of your application. And then, you can visualize everything in a single interface and without having to manually extract them from different systems making it easier to correlate and understand what went wrong.

This allows you to easily troubleshoot the application so it is easier to identify the root cause. For example, it can help identify problems related to:

* Empty fields: Obtain statistics on null and empty values for each field.
* Incorrect or deprecated field usage: Track the usage patterns of event fields.
* Invalid values: Get statistics about the value distribution of each field and identify outliers.

You can also set up rules that proactively trigger alerts when these situations occur.

!!! note
    Some features may be currently being developed and still not have all the capabilities described in here. Check the [project roadmap](https://github.com/orgs/neblic/projects/3) for details about what improvements are we working on.

    Also, do not hesitate to suggest features, improvements or use cases by creating a [feature request](https://github.com/neblic/platform/issues/new?assignees=&labels=&projects=&template=feature_request.md&title=) in our github!
