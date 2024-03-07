# How does Neblic work

{%
   include-markdown "../../../README.md"
   start="<!--how-does-neblic-work-start-->"
   end="<!--how-does-neblic-work-end-->"
%}

## Data collection

As you can imagine, the fundamental piece that Neblic needs to do its magic is your application data. And it's natural that one of the first questions that come to mind is: how can Neblic get it?

This is a hard and complex problem given that Neblic is intended to be an observability platform that can be used in any application and cover it end-to-end. All architectures are different and their diverse characteristics make it so that what works in one may be unthinkable to use in another: data volume, message size, synchronous/asynchronous communication, different technologies...

To try to cover as much architectures as possible Neblic instrumentation has been designed to be flexible and to adapt as needed. See the [Samplers](/learn/samplers) page to get a deep dive into how *Samplers* work and how you can use them for your application.

### Instrumentation overhead

Another common concern when adding observability to an application is how much overhead will it introduce. Neblic has been designed with two principles in mind: be dynamically configurable at runtime, so it can be easily adjusted when needed, and to provide many levers, so you can adjust how much and where does this overhead may impact your application. 

In any case, keep in mind that unless your application processes large volumes of data, is very resource constrained, or very time sensitive, the small overhead introduced by Neblic outsets its benefits.

Neblic performance is continuously monitored with a benchmark suite and there are several initiatives to optimize it. See [this section](/learn/data-pipeline) to learn about the data collection pipeline and how you can adjust it.

## Try Neblic

If you would like to get a feel of how does Neblic work using an easy to spin up local playground, you can follow the guide [here](/quickstart/playground).

On the other hand, if you would like to learn about Neblic inner details and how can you set it up in your own infrastructure please go to the *Getting Started* section. The [concepts](/getting-started/concepts) page describe the foundational concepts that Neblic is built on and the [usage](/getting-started/usage)
