package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func initTracer(ctx context.Context) (func(context.Context) error, error) {
	// OpenTelemetry Collector is listening on host:4317 (OTLP gRPC) — from your docker-compose.
	exp, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint("127.0.0.1:4317"),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("create otlp exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			attribute.String("service.name", "ampy-demo"),
			attribute.String("service.version", "0.0.1"),
			attribute.String("deployment.environment", "dev"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("create resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tp.Shutdown, nil
}

func startMetricsServer() {
	// Simple native Prometheus metrics (keeps this demo minimal & robust).
	reqs := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "ampy_demo",
		Name:      "requests_total",
		Help:      "Number of demo requests processed.",
	})
	lat := prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "ampy_demo",
		Name:      "request_latency_ms",
		Help:      "Latency of demo requests in milliseconds.",
		Buckets:   []float64{1, 2, 5, 10, 20, 50, 100, 200, 500},
	})
	reg := prometheus.NewRegistry()
	reg.MustRegister(reqs, lat)

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	go func() {
		log.Println("metrics: serving on :9464/metrics")
		if err := http.ListenAndServe(":9464", nil); err != nil {
			log.Fatalf("metrics server: %v", err)
		}
	}()

	// Emit some traffic so you immediately see time-series.
	go func() {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		for {
			start := time.Now()
			// Simulate variable work
			time.Sleep(time.Duration(5+r.Intn(50)) * time.Millisecond)
			reqs.Inc()
			lat.Observe(float64(time.Since(start).Milliseconds()))
			time.Sleep(500 * time.Millisecond)
		}
	}()
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	shutdown, err := initTracer(ctx)
	if err != nil {
		log.Fatalf("tracing init: %v", err)
	}
	defer func() {
		_ = shutdown(context.Background())
	}()

	startMetricsServer()

	tr := otel.Tracer("ampy-demo/worker")
	runID := fmt.Sprintf("demo_run_%d", time.Now().Unix())

	// Create one root span and a few child spans.
	ctx, root := tr.Start(ctx, "demo.run")
	root.SetAttributes(
		attribute.String("run_id", runID),
		attribute.String("as_of", time.Now().UTC().Format(time.RFC3339)),
	)

	for i := 0; i < 5; i++ {
		_, child := tr.Start(ctx, "demo.work")
		child.SetAttributes(
			attribute.Int("iteration", i),
			attribute.String("symbol", "AAPL"),
			attribute.String("mic", "XNAS"),
		)
		time.Sleep(80 * time.Millisecond)
		child.End()
	}
	root.End()

	traceID := root.SpanContext().TraceID().String()
	fmt.Println("✅ trace sent to Tempo. TraceID:", traceID)
	fmt.Println("   Open Grafana → Explore → Tempo, search for service.name=ampy-demo, or paste TraceID above.")

	// Keep running to serve /metrics and keep the process alive.
	<-ctx.Done()
	fmt.Println("\nshutting down…")
}
