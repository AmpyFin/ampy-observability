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
	exp, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint("127.0.0.1:4317"), // Collector exposed in docker-compose
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
	// Prometheus metrics
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

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	mux.HandleFunc("/work", func(w http.ResponseWriter, _ *http.Request) {
		start := time.Now()
		// simulate a bit of work
		time.Sleep(time.Duration(5+rand.Intn(50)) * time.Millisecond)
		reqs.Inc()
		lat.Observe(float64(time.Since(start).Milliseconds()))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	go func() {
		log.Println("demo: serving on http://localhost:9464  (GET /work, /metrics)")
		if err := http.ListenAndServe(":9464", mux); err != nil {
			log.Fatalf("metrics server: %v", err)
		}
	}()

	// background traffic so you see time-series immediately
	go func() {
		rand.Seed(time.Now().UnixNano())
		for {
			start := time.Now()
			time.Sleep(time.Duration(5+rand.Intn(50)) * time.Millisecond)
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
	defer func() { _ = shutdown(context.Background()) }()

	startMetricsServer()

	tr := otel.Tracer("ampy-demo/worker")
	runID := fmt.Sprintf("demo_run_%d", time.Now().Unix())
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

	<-ctx.Done()
	fmt.Println("\nshutting down…")
}
