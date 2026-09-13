package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
)

var tracerProvider *sdktrace.TracerProvider

// otlpEndpoint: if empty, no OTLP exporter is added (e.g. running without
// the Alloy stack up) - the app still works, spans are just discarded.
func Init(serviceName, env, otlpEndpoint string, consoleExport bool) error {
	// set resource
	res, err := resource.Merge(
		resource.Default(),
		resource.NewSchemaless(
			semconv.ServiceName(serviceName),
			attribute.String("deployment.environment", env),
		),
	)
	if err != nil {
		return err
	}

	opts := []sdktrace.TracerProviderOption{
		sdktrace.WithResource(res),
	}

	// real backend: push spans to Alloy, batched (not per-span, unlike the
	// console exporter below - this one is a real network call).
	if otlpEndpoint != "" {
		otlpExporter, err := otlptracegrpc.New(context.Background(),
			otlptracegrpc.WithEndpoint(otlpEndpoint),
			otlptracegrpc.WithInsecure(),
		)
		if err != nil {
			return err
		}
		opts = append(opts, sdktrace.WithBatcher(otlpExporter))
	}

	// local debugging only: print spans to stdout too, if enabled.
	if consoleExport {
		exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return err
		}
		opts = append(opts, sdktrace.WithSyncer(exporter))
	}

	// create sdk for tracing from option
	tracerProvider = sdktrace.NewTracerProvider(opts...)

	// plug with global provider with these sdk setting
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return nil
}

func Shutdown(ctx context.Context) error {
	if tracerProvider == nil {
		return nil
	}
	return tracerProvider.Shutdown(ctx)
}
