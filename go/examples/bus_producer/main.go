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
		ServiceName:       "demo-producer",
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

	msgID := uuid.NewString()
	attrs := ampyobs.BusAttrs{
		Topic:        "ampy/dev/signals/v1",
		SchemaFQDN:   "ampy.signals.v1.Signal",
		MessageID:    msgID,
		PartitionKey: "AAPL",
		RunID:        "dev_session_1",
	}

	ctx, span := ampyobs.StartBusPublishSpan(ctx, attrs)
	defer span.End()

	ampyobs.C(ctx).Info("publishing signal",
		slog.String("event", "signals.emit"),
		slog.String("symbol", "AAPL"),
	)

	headers := map[string]string{}
	ampyobs.InjectTrace(ctx, headers)

	data, _ := json.MarshalIndent(headers, "", "  ")
	_ = os.WriteFile("bus_headers.json", data, 0o644)
	fmt.Println("Wrote bus_headers.json with headers:", headers)

	time.Sleep(1 * time.Second)

	// Properly shutdown to flush traces
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := ampyobs.Shutdown(shutdownCtx); err != nil {
		fmt.Printf("Shutdown error: %v\n", err)
	}
	fmt.Println("Producer completed - traces should be in Jaeger")
}
