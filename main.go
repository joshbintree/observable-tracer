package main

import (
	"context"
	"fmt"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"log"
	"net/http"
	"observable_tracer.bintree.io/telemetry_helpers"
)

func main() {
	ctx := context.Background()
	tp, err := telemetry_helpers.SetupTracer(ctx, "woo")

	if err != nil {
		panic(err)
	}
	defer tp.Shutdown(ctx)

	telemetry_helpers.NewLogrus(ctx)
	fmt.Println("hello")
	///telemetry_helpers.RunTracer("my service", ctx)
	serviceA(ctx, 8081)

}

func serviceA(ctx context.Context, port int) {
	mux := http.NewServeMux()

	// create a decorated handler using the telemetry tracer TelemetryHttpWrapServer and a name for the trace.
	// for example if you had a handler like this
	// mux.HandleFunc("/serviceA", serviceB_HttpHandler)
	// Wrap it with the helper function, pass in your function and attach it to a trace name
	// if you want it part of the same trace make sure the names are uniform
	testDecorator := telemetry_helpers.TelemetryHttpWrapServer(http.HandlerFunc(serviceA_HttpHandler), "testTracer", "my serverLog running on this")
	// lastly pass in the wrraper to your mux handler
	// you now have tracing
	mux.HandleFunc("/servicea", serviceA_HttpHandler)

	mux.Handle("/serviceA", testDecorator)
	serverPort := fmt.Sprintf(":%d", port)
	server := &http.Server{Addr: serverPort, Handler: mux}

	fmt.Println("serviceA listening on", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}

func serviceA_HttpHandler(w http.ResponseWriter, r *http.Request) {

	client := &telemetry_helpers.TelemetryClientWrapper{
		Client: &http.Client{
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		},
		Endpoint:  "example.com",    // Set the additional parameter
		TraceName: "my-http-client", // Set the additional argument
		LogInfo:   "my agent",       // if you want logs with your traces
	}

	req, err := http.NewRequest("GET", "https://google.com/", nil)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp.Status)
}
