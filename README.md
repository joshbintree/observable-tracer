```markdown
# observable-tracer

This project is a small middleware and helper utility that allows you to install the code and then inject it into your Go HTTP handlers using OpenTelemetry.

## Installation

To install the package, clone the repository and import the `telemetry_helpers` folders:

```go
import "observable_tracer.bintree.io/telemetry_helpers"
```

Ensure that you have OpenTelemetry installed:

```sh
go get -u go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp
```

## Usage

Once the package is installed, look at the `main.go` function to see examples of how to inject these packages into your code. It will allow you to use Logrus and a wrapper to send data to tools like Prometheus or other log aggregation systems for collecting telemetry data.
