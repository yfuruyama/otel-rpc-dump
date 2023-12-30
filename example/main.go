package main

import (
	"context"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/resource"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

func main() {
	resource, err := resource.Merge(resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceName("my-service"),
			semconv.ServiceVersion("0.1.0"),
		),
	)
	if err != nil {
		log.Fatalf("failed to create a resource: %v", err)
	}

	ctx := context.Background()
	exporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint("localhost:4317"),
	)
	if err != nil {
		log.Fatalf("failed to create an exporter: %v", err)
	}

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(resource),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter,
			sdkmetric.WithInterval(3*time.Second))),
	)
	defer provider.Shutdown(ctx)
	otel.SetMeterProvider(provider)

	meter := otel.Meter("my-service-meter")

	counter, err := meter.Int64Counter("tick-counter")
	if err != nil {
		log.Fatalf("failed to create a counter: %v", err)
	}

	exportCounter, err := meter.Int64ObservableCounter("export-counter")
	var exportCount int64
	meter.RegisterCallback(func(_ context.Context, o metric.Observer) error {
		exportCount++
		o.ObserveInt64(exportCounter, exportCount)
		return nil
	}, exportCounter)

	for {
		counter.Add(ctx, 1)
		time.Sleep(time.Second)
	}
}
