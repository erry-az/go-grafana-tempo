package tracer

import (
	"context"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

var traceTracer trace.Tracer

// NewOtel initializes an OTLP exporter, and configures the corresponding trace providers.
func NewOtel(agentAddr, serviceName, appName string, grpcDialOpts ...grpc.DialOption) func() {
	ctx := context.Background()

	if agentAddr == "" {
		agentAddr = "localhost:4317"
	}

	// The exporter is the component in SDK responsible for exporting the telemetry signal (trace) out of the
	// application to a remote backend, log to a file, stream to stdout. etc.
	traceClient := otlptracegrpc.NewClient(
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(agentAddr),
		otlptracegrpc.WithDialOption(grpcDialOpts...),
	)
	traceExp, err := otlptrace.New(ctx, traceClient)
	handleErr(err, "Failed to create the collector trace exporter")

	// The resource describes the object that generated the telemetry signals.
	// use resource.WithProcess(), to send detail proses
	// use resource.WithTelemetrySDK(), to send info telemetry sdk
	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithHost(),
		resource.WithAttributes(
			// the service name used to display traces in backends
			semconv.ServiceNameKey.String(serviceName),
			semconv.TelemetrySDKLanguageGo,
		),
	)
	handleErr(err, "failed to create resource")

	// Span processors are responsible for CRUD operations, batching of the requests for
	// better QoS, Sampling the span data based on certain conditions.
	bsp := sdktrace.NewBatchSpanProcessor(traceExp)

	// sdktrace.WithSampler(sdktrace.AlwaysSample()),
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	// Propagators are used to extract and inject context data from and into messages exchanged by applications.
	propagator := propagation.NewCompositeTextMapPropagator(propagation.Baggage{}, propagation.TraceContext{})

	// Set global propagator to trace context (the default is no-op).
	otel.SetTextMapPropagator(propagator)
	otel.SetTracerProvider(tracerProvider)

	traceTracer = otel.Tracer(appName)

	return func() {
		cxt, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		if err := traceExp.Shutdown(cxt); err != nil {
			otel.Handle(err)
		}
	}
}

func handleErr(err error, message string) {
	if err != nil {
		log.Fatalf("%s: %v", message, err)
	}
}
