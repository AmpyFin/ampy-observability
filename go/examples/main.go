package main

import (
	"context"
	"math/rand"
	"time"

	"github.com/AmpyFin/ampy-observability/go/ampyobs"
	"log/slog"
)

func main() {
	_ = ampyobs.Init(ampyobs.Config{
		ServiceName:       "demo-go",
		ServiceVersion:    "0.1.0",
		Environment:       "dev",
		CollectorEndpoint: "http://localhost:4317",
		EnableLogs:        true,
		EnableMetrics:     true,
		EnableTracing:     true,
		Sampler:           "ratio",
		SampleRatio:       1.0, // always sample for demo
	})

	ctx := context.Background()
	ampyobs.C(ctx).Info("hello from go",
		slog.String("event", "bars.ingest"),
		slog.String("symbol", "AAPL"),
		slog.String("mic", "XNAS"),
	)

	// Emit demo domain metrics
	ampyobs.BusProducedAdd(ctx, "ampy/dev/bars/v1", 5)
	ampyobs.BusConsumedAdd(ctx, "ampy/dev/bars/v1", 5)
	ampyobs.BusDeliveryLatencyMs(ctx, "ampy/dev/bars/v1", 12.3)

	ampyobs.OMSOrderSubmitAdd(ctx, "alpaca", ampyobs.OutcomeOK)
	ampyobs.OMSOrderLatencyMs(ctx, "alpaca", 42.0)
	ampyobs.OMSRejectAdd(ctx, "alpaca", "risk_check")

	for i := 0; i < 10; i++ {
		ampyobs.BusDeliveryLatencyMs(ctx, "ampy/dev/bars/v1", 5+rand.Float64()*100)
		ampyobs.OMSOrderLatencyMs(ctx, "alpaca", 20+rand.Float64()*80)
		time.Sleep(200 * time.Millisecond)
	}

	// Allow at least one export interval (10s) plus Prometheus scrape (15s).
	time.Sleep(20 * time.Second)

	_ = ampyobs.Shutdown(context.Background())
}
