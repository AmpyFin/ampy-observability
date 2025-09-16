package main

import (
	"context"
	"fmt"
	"time"

	"github.com/AmpyFin/ampy-observability/go/ampyobs"
)

func main() {
	_ = ampyobs.Init(ampyobs.Config{
		ServiceName:       "simple-test",
		ServiceVersion:    "0.1.0",
		Environment:       "dev",
		CollectorEndpoint: "http://localhost:4317",
		EnableLogs:        true,
		EnableMetrics:     false,
		EnableTracing:     true,
		Sampler:           "ratio",
		SampleRatio:       1.0,
	})

	ctx := context.Background()
	ctx, span := ampyobs.StartSpan(ctx, "test-operation", 1) // 1 = SPAN_KIND_INTERNAL
	defer span.End()

	fmt.Println("Created span, waiting for export...")
	time.Sleep(1 * time.Second)
}
