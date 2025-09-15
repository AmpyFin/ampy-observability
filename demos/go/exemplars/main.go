package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var (
	reqs = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "ampy_demo",
		Name:      "requests_total",
		Help:      "Number of demo requests processed.",
	})
	lat = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "ampy_demo",
		Name:      "request_latency_ms",
		Help:      "Latency of demo requests in milliseconds.",
		Buckets:   []float64{1, 2, 5, 10, 20, 50, 100, 200, 500},
	})
	reg = prometheus.NewRegistry()
)

func init() {
	reg.MustRegister(reqs, lat)
}

func initTracer(ctx context.Context) (func(context.Context) error, error) {
	exp, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint("127.0.0.1:4317"),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("create otlp exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			attribute.String("service.name", "ampy-demo-svc"),
			attribute.String("service.version", "0.1.0"),
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
	// Ensures W3C TraceContext is used end-to-end (good practice).
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tp.Shutdown, nil
}

// recordExemplars safely attaches trace/span IDs to metrics if the client library supports it.
func recordExemplars(ctx context.Context, durationMs float64) {
	// Try histogram exemplar
	if eo, ok := lat.(interface {
		ObserveWithExemplar(float64, prometheus.Labels)
	}); ok {
		if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
			eo.ObserveWithExemplar(durationMs, prometheus.Labels{
				"trace_id": sc.TraceID().String(),
				"span_id":  sc.SpanID().String(),
			})
			return
		}
	}
	// Fallback if no exemplar support or no span: record normally
	lat.Observe(durationMs)
}

func incRequests(ctx context.Context) {
	// Try counter exemplar
	if ea, ok := reqs.(interface {
		AddWithExemplar(float64, prometheus.Labels)
	}); ok {
		if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
			ea.AddWithExemplar(1, prometheus.Labels{
				"trace_id": sc.TraceID().String(),
				"span_id":  sc.SpanID().String(),
			})
			return
		}
	}
	// Fallback
	reqs.Inc()
}

func withTracing(next http.Handler) http.Handler {
	tr := otel.Tracer("ampy-demo/handler")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tr.Start(r.Context(), "http.request",
			trace.WithAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.route", r.URL.Path),
			),
		)
		defer span.End()

		start := time.Now()
		next.ServeHTTP(w, r.WithContext(ctx))
		durMs := float64(time.Since(start).Milliseconds())

		// Record metrics + exemplars
		incRequests(ctx)
		recordExemplars(ctx, durMs)
	})
}

func main() {
	ctx := context.Background()
	shutdown, err := initTracer(ctx)
	if err != nil {
		log.Fatalf("tracing init: %v", err)
	}
	defer func() { _ = shutdown(context.Background()) }()

	mux := http.NewServeMux()

	// /work simulates variable work and produces traces + metrics
	mux.HandleFunc("/work", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Duration(5+rand.Intn(50)) * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})

	// /metrics with OpenMetrics enabled so exemplars are exposed
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	}))

	addr := ":9464"
	log.Printf("demo: serving on http://localhost%v  (GET /work, /metrics)\n", addr)
	if err := http.ListenAndServe(addr, withTracing(mux)); err != nil {
		log.Fatal(err)
	}
}
