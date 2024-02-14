package telemetry_helpers

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type TelemetryClientWrapper struct {
	Client    *http.Client
	TraceName string
	Endpoint  string
	LogInfo   string
}

func SetupTracer(ctx context.Context, serviceName string) (*trace.TracerProvider, error) {

	exporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint("localhost:4317"),
	)
	if err != nil {
		return nil, err
	}

	// labels/tags/resources that are common to all traces.
	resource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(serviceName),
		//example if you want an attributes
		//attribute.String("some-attribute", "some-value"),
	)

	provider := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource),
		// set the sampling rate based on the parent span example below is 60%
		//trace.WithSampler(trace.ParentBased(trace.TraceIDRatioBased(0.6))),
	)

	otel.SetTracerProvider(provider)

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{}, // W3C Trace Context format; https://www.w3.org/TR/trace-context/
		),
	)

	return provider, nil
}

func TelemetryHttpWrapServer(handler http.Handler, tracerName string, logInfo string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//ctx := propagation.ExtractHTTP(r.Context(), propagation.HeaderCarrier(r.Header))

		spanName := r.Method + " " + r.URL.Path
		ctx, span := otel.Tracer(tracerName).Start(r.Context(), spanName)
		defer span.End()
		log := NewLogrus(ctx)
		log.Info(logInfo)
		// setting specific attributes can be done below. see example
		//span.SetAttributes(
		//	attribute.String("<attribute string> i.e.http.method", r.Method),
		//	attribute.String("http.url", r.URL.Path),
		//)

		handler.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (client *TelemetryClientWrapper) Do(req *http.Request) (*http.Response, error) {
	tracer := otel.Tracer(client.TraceName)
	ctx := req.Context()

	// Start a new span for the outgoing request
	ctx, span := tracer.Start(ctx, req.URL.Path)
	defer span.End()

	// Set span attributes example
	span.SetAttributes(
		attribute.String("http.method", req.Method),
		attribute.String("http.url", req.URL.String()),
		attribute.String("my_parameter", client.Endpoint), // Set the additional parameter

	)

	log := NewLogrus(ctx)
	log.Info(client.LogInfo)
	// Inject the span context into the request headers
	req = req.WithContext(ctx)

	// Send the request using the underlying client
	resp, err := client.Client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
