package telemetry

import (
	"context"
	"fmt"

	"github.com/skomaroh1845/crypto_telemetry/notifier-service/internal/config"
	"go.opentelemetry.io/otel"
	otlpmetricgrpc "go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.19.0"
)

type TelemetryProviders struct {
	TracerProvider *sdktrace.TracerProvider
	MeterProvider  *metric.MeterProvider
}

func InitTelemetry(cfg *config.Config) (*sdktrace.TracerProvider, *metric.MeterProvider, error) {
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName("notifier-service"),
			semconv.ServiceVersion("1.0.0"),
			semconv.DeploymentEnvironment(cfg.Environment),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create resource: %w", err)
	}

	traceExporter, err := otlptracegrpc.New(
		context.Background(),
		otlptracegrpc.WithEndpoint(cfg.OTLPCollectorURL),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	metricExporter, err := otlpmetricgrpc.New(
		context.Background(),
		otlpmetricgrpc.WithEndpoint(cfg.OTLPCollectorURL),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create metric exporter: %w", err)
	}

	mp := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(metricExporter)),
	)

	otel.SetMeterProvider(mp)

	return tp, mp, nil
}
