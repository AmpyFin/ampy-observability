package main

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type errHandler struct{}

func (errHandler) Handle(err error) { fmt.Println("OTEL-ERROR:", err.Error()) }

func main() {
	// Set error handler to catch exporter errors
	otel.SetErrorHandler(errHandler{})

	ctx := context.Background()

	// Create gRPC exporter (port 4317)
	exp, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint("localhost:4317"), // gRPC port
		otlptracegrpc.WithInsecure(),                 // no TLS locally
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create gRPC exporter: %v", err))
	}

	// Create tracer provider with gRPC exporter
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exp,
			sdktrace.WithMaxExportBatchSize(64),
			sdktrace.WithBatchTimeout(200*time.Millisecond),
		),
	)
	otel.SetTracerProvider(tp)

	// Create a test span
	tr := otel.Tracer("grpc-test")
	_, span := tr.Start(ctx, "demo.grpc.span",
		trace.WithAttributes(
			attribute.String("test.type", "grpc_exporter"),
			attribute.String("protocol", "grpc"),
			attribute.Int("test.id", 456),
		),
	)

	// Simulate some work
	time.Sleep(100 * time.Millisecond)
	span.End()

	// Force flush and shutdown
	flushCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Println("Flushing traces...")
	if err := tp.ForceFlush(flushCtx); err != nil {
		fmt.Printf("ForceFlush error: %v\n", err)
	}

	fmt.Println("Shutting down tracer provider...")
	if err := tp.Shutdown(flushCtx); err != nil {
		fmt.Printf("Shutdown error: %v\n", err)
	}

	fmt.Println("gRPC exporter test completed - check Jaeger for 'demo.grpc.span'")
}
