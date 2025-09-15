package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
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

	logMu   sync.Mutex
	logFile *os.File
	logBuf  *bufio.Writer
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
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tp.Shutdown, nil
}

// --- Exemplars helpers (safe fallback if exemplar APIs are absent) ---
func recordExemplars(ctx context.Context, durationMs float64) {
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
	lat.Observe(durationMs)
}

func incRequests(ctx context.Context) {
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
	reqs.Inc()
}

// --- Log to ./logs/app.log in logfmt, including trace/span IDs ---
func openLogFile() error {
	if err := os.MkdirAll("./logs", 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile("./logs/app.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	logFile = f
	logBuf = bufio.NewWriterSize(f, 64*1024)
	return nil
}

func closeLogFile() {
	if logBuf != nil {
		_ = logBuf.Flush()
	}
	if logFile != nil {
		_ = logFile.Close()
	}
}

func logLine(ctx context.Context, level, msg string, kv map[string]any) {
	// pull trace/span from context
	sc := trace.SpanContextFromContext(ctx)
	traceID := ""
	spanID := ""
	if sc.IsValid() {
		traceID = sc.TraceID().String()
		spanID = sc.SpanID().String()
	}

	// logfmt: ts=<rfc3339> level=info msg="..." trace_id=... span_id=... k=v ...
	ts := time.Now().UTC().Format(time.RFC3339Nano)

	logMu.Lock()
	defer logMu.Unlock()

	fmt.Fprintf(logBuf, "ts=%s level=%s msg=%q trace_id=%s span_id=%s",
		ts, level, msg, traceID, spanID)
	for k, v := range kv {
		fmt.Fprintf(logBuf, " %s=%v", k, v)
	}
	fmt.Fprint(logBuf, "\n")
	_ = logBuf.Flush()
}

// --- HTTP plumbing with tracing + metrics + logs ---
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

		// metrics (with exemplars)
		incRequests(ctx)
		recordExemplars(ctx, durMs)

		// structured log line (picked up by filelog receiver)
		logLine(ctx, "info", "request handled", map[string]any{
			"route":       r.URL.Path,
			"status":      200,
			"latency_ms":  int(durMs),
			"user_agent":  r.UserAgent(),
			"remote_addr": r.RemoteAddr,
		})
	})
}

func main() {
	ctx := context.Background()
	shutdown, err := initTracer(ctx)
	if err != nil {
		log.Fatalf("tracing init: %v", err)
	}
	defer func() { _ = shutdown(context.Background()) }()

	if err := openLogFile(); err != nil {
		log.Fatalf("open log file: %v", err)
	}
	defer closeLogFile()

	mux := http.NewServeMux()
	mux.HandleFunc("/work", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Duration(5+rand.Intn(50)) * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})

	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	}))

	addr := ":9464"
	log.Printf("demo: serving on http://localhost%v  (GET /work, /metrics)\n", addr)
	if err := http.ListenAndServe(addr, withTracing(mux)); err != nil {
		log.Fatal(err)
	}
}

