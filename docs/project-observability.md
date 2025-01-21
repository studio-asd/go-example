# Observability

Have you ever heard [three pillars of observability](https://en.wikipedia.org/wiki/Observability_(software)#%22Pillars_of_observability%22)?

The three pillars usually consists of:

1. Logs 
2. Metrics
3. Traces

These three things are mainly the most used instrumentation that widely used to ensure we can get something to look/debug at when we have problems with our applications. In other words, making hard things easy. Debugging a running software in production is indeed hard because we can't attatch a debugger and problems in production environment might somehow cannot be replicable. So, having as much instrumentation as possible would be helpful to help us to **observe** on what really happen with our software.

Let's dig a bit about each one of them.

1. Log

	Log is maybe one of the most common instrumentation that exist in the computer programming. Programmers are used to write the log into the screen/files for quite a long time. But since its inception, the technique and standard of logging for instrumentation changed quite a bit. For instrumentation, softwares are usually emitted a structured logs to it can be parsed further into something "meaningful". For example:

	```text
	This is unstructured log that contains id:123 and path:/v1/some/api
	```

	And compare this with:

	```json
	{
		"message": "this is structured log",
		"level": "error",
		"fields": {
			"id": 123,
			"path": "/v1/some/api"
		}
	}
	```

	The prior log message is unstructured and we might need to define a pattern using `regex` to parse the meaningful field inside of it. While we can esily decode the `json` one into a meaningful information.

	While the `json` formatted log is friendly for computer, it usually not really friendly for the end user in the local environment where logs are not parsed into a friendly UI. So people are usually using a different log format in the local and production environment.

	```text
	this is a structured log level=error fields.id=123 fields.path=/v1/some/api
	```

2. Metrics

	Metrics events usually contains information about specific metrics type and information that we want to send to the metrics backend. There are several types of metrics, depending on the platform. But in general these metrics are being used the most:

	1. Count
	1. Histogram
	1. Gauge

	References aout metrics can be found in different-different resources based on the platform:

	- [Prometheus](https://prometheus.io/docs/concepts/metric_types/)
	- [Datadog](https://docs.datadoghq.com/metrics/types/?tab=count)

	These platforms have different characteristics, as Prometheus is a [pull-based]() platform that we can host on our own, and [Datadog]() is a push based system that we need to pay to use. We will not go to deep into details on what is the difference of `pull-based` and `push-based` system. But, from our experience, the `pull-based` system is easier to be maintained from the end user perspective as the user doesn't have to deal with a lot of scaling techniques that `push-based` system need to deal with. A `push-based` system receives traffic, so it needs to have some backpressure mechanism to ensure the system is not overloaded because of [thundering-herd]() problem.

3. Traces

	Traces are events that contains the time being spent in a process inside the program and usually contains metadata or information about it. The traces are usually correlated so its easier for the end user to see the complete view of chained events thus leading to better experience of debugging and problem analysis. Traces are usually being correlated by using a unique `trace-id`, and inside of each trace we can have multiple number of `span`. Lets now dig deeper into `trace` and `span` so you understand the differences.

While the three pillars will help us a lot to debug our applications from the "blackbox" perspective, these three are lack of internal programs information. The "whitebox" instrumentation from the programming language is extremely useful to understand further on how the behavior of our programs affects the programs and vice-versa. For this kind of thing, profiling is one of the widely used "whitebox" instrumentation.

## Costs & Overhead

### Instrumentation Overhead

Instrumentations are not without a costs. By default, there are "price" we need to pay inside of our program to record and send the instrumentations. And of course, the costs of each instrumentation is different as the techniques and data that being collected for each instrumentation is different.

As we are using the Go programming language, lets take a look of each instrumentation overhead in the Go programming language.

1. Log
1. Metrics
1. Trace
1. Profiles

### Instrumentation Costs

As we store the instrumentation data to the data storage and the data need to be processed into something meaningful for us to look at, there will be costs for all of that.

## Open Telemetry

Now that you have enough information about `observability`, lets take a look into the platform and protocol that we use to record and send instrumentation to our observability backends. In this project, we use OTel(Open Telemetry) as the open standard for observability.

## Grafana Stack

In this project, we are using [Grafana]() stack as the observability platform. If you are new to Grafana, then you can look some of their repositories:

1. [Grafana]() for the UI.
2. [Loki]() for logging.
3. [Prometheus]() for metrics.
4. [Tempo]() for traces.
5. [Pyroscope]() for profiles.