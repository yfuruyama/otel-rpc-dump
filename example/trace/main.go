package main

import (
	"context"
	"errors"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/trace"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

var tracer trace.Tracer

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
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint("localhost:4317"),
	)

	if err != nil {
		log.Fatalf("failed to create an exporter: %v", err)
	}

	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource),
	)
	otel.SetTracerProvider(traceProvider)
	tracer = otel.GetTracerProvider().Tracer("my-service")

	for {
		ctx := context.Background()
		doWork(ctx)
		doWork(ctx)
		time.Sleep(time.Second * 5)
	}
}

func doWork(ctx context.Context) {
	ctx, span := tracer.Start(ctx, "doWork", trace.WithAttributes(attribute.String("foo", "bar")))
	defer span.End()
	doWorkChild(ctx)
	span.AddEvent("child work finished")
	span.RecordError(errors.New("some error happened"))
	return
}

func doWorkChild(ctx context.Context) {
	ctx, span := tracer.Start(ctx, "doWorkChild")
	defer span.End()
	doWorkChild2(ctx)
	return
}

func doWorkChild2(ctx context.Context) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent("child work 2 finished")
}
