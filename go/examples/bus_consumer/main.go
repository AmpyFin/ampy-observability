package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/AmpyFin/ampy-observability/go/ampyobs"
	"github.com/google/uuid"
	"log/slog"
)

func main() {
	// Set error handler to catch OTel errors
	ampyobs.SetErrorHandler(func(err error) {
		fmt.Printf("OTEL-ERROR: %v\n", err)
	})

	err := ampyobs.Init(ampyobs.Config{
		ServiceName:       "demo-consumer",
		ServiceVersion:    "0.1.0",
		Environment:       "dev",
		CollectorEndpoint: "localhost:4318", // Use HTTP port
		TraceProtocol:     "http",           // Use HTTP exporter
		EnableLogs:        true,
		EnableMetrics:     true,
		EnableTracing:     true,
		Sampler:           "ratio",
		SampleRatio:       1.0,
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to init ampyobs: %v", err))
	}

	ctx := context.Background()

	raw, err := os.ReadFile("bus_headers.json")
	if err != nil {
		panic("run the producer first to create bus_headers.json")
	}
	headers := map[string]string{}
	_ = json.Unmarshal(raw, &headers)
	fmt.Println("Read headers:", headers)

	attrs := ampyobs.BusAttrs{
		Topic:        "ampy/dev/signals/v1",
		SchemaFQDN:   "ampy.signals.v1.Signal",
		MessageID:    uuid.NewString(),
		PartitionKey: "AAPL",
		RunID:        "dev_session_1",
	}

	ctx, span := ampyobs.StartBusConsumeSpan(ctx, headers, attrs)
	defer span.End()

	ampyobs.C(ctx).Info("consumed signal",
		slog.String("event", "signals.consume"),
		slog.String("action", "forward_to_oms"),
	)

	time.Sleep(1 * time.Second)

	// Properly shutdown to flush traces
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := ampyobs.Shutdown(shutdownCtx); err != nil {
		fmt.Printf("Shutdown error: %v\n", err)
	}
	fmt.Println("Consumer completed - traces should be in Jaeger")
}
