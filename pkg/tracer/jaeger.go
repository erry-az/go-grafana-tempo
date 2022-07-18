package tracer

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"time"
)

const (
	environment = "production"
	id          = 1
)

func NewJaeger(agentHost, agentPort, serviceName, appName string) (func(), error) {
	ctx := context.Background()

	// Create the Jaeger exporter
	exp, err := jaeger.New(jaeger.WithAgentEndpoint(jaeger.WithAgentHost(agentHost), jaeger.WithAgentPort(agentPort)))
	if err != nil {
		return nil, err
	}

	tracerProvider := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
			attribute.String("environment", environment),
			attribute.Int64("ID", id),
		)),
	)

	otel.SetTracerProvider(tracerProvider)

	traceTracer = tracerProvider.Tracer(appName)

	return func() {
		cxt, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()

		if err := exp.Shutdown(cxt); err != nil {
			otel.Handle(err)
		}
	}, nil
}
