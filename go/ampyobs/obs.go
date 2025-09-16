package ampyobs

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"google.golang.org/grpc/credentials"
)

type Config struct {
	ServiceName       string
	ServiceVersion    string
	Environment       string // dev | paper | prod
	CollectorEndpoint string // e.g. "http://localhost:4317" or "localhost:4317"
	TraceProtocol     string // "grpc" | "http" (default: "grpc")
	EnableLogs        bool   // JSON logs via slog (stdout)
	EnableMetrics     bool   // OTLP metrics to collector
	EnableTracing     bool   // OTLP traces to collector
	Sampler           string // "parent" | "ratio"
	SampleRatio       float64
}

var (
	globalCfg       Config
	tracerProvider  *sdktrace.TracerProvider
	meterProvider   *sdkmetric.MeterProvider
	globalResources *resource.Resource
)

// SetErrorHandler sets a custom error handler for OTel errors
func SetErrorHandler(handler func(error)) {
	otel.SetErrorHandler(errorHandlerFunc(handler))
}

type errorHandlerFunc func(error)

func (f errorHandlerFunc) Handle(err error) { f(err) }

// histogramBoundariesMs returns consistent bucket boundaries for latency histograms
func histogramBoundariesMs() []float64 {
	return []float64{1, 2, 5, 10, 20, 50, 100, 200, 500, 1000, 2000} // ms
}

// getMetricViews returns views for customizing histogram buckets
func getMetricViews() []sdkmetric.View {
	return []sdkmetric.View{
		sdkmetric.NewView(
			sdkmetric.Instrument{Name: "ampy.bus.delivery_latency_ms"},
			sdkmetric.Stream{
				Aggregation: sdkmetric.AggregationExplicitBucketHistogram{
					Boundaries: histogramBoundariesMs(),
				},
			},
		),
		sdkmetric.NewView(
			sdkmetric.Instrument{Name: "ampy.oms.order_latency_ms"},
			sdkmetric.Stream{
				Aggregation: sdkmetric.AggregationExplicitBucketHistogram{
					Boundaries: histogramBoundariesMs(),
				},
			},
		),
	}
}

func Init(cfg Config) error {
	globalCfg = cfg

	// ----- Resource -----
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			semconv.DeploymentEnvironmentName(cfg.Environment),
		),
	)
	if err != nil {
		return fmt.Errorf("resource: %w", err)
	}
	globalResources = res

	// ----- Propagation (W3C) -----
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	// ----- Logging -----
	if cfg.EnableLogs {
		setupSlog(res) // JSON stdout with resource attrs; adds trace/span when ctx provided
	}

	// ----- Tracing -----
	if cfg.EnableTracing {
		tp, err := newTracerProvider(cfg, res)
		if err != nil {
			return err
		}
		tracerProvider = tp
		otel.SetTracerProvider(tp)
	}

	// ----- Metrics -----
	if cfg.EnableMetrics {
		mp, err := newMeterProvider(cfg, res)
		if err != nil {
			return err
		}
		meterProvider = mp
		otel.SetMeterProvider(mp)

		// Domain metrics helpers (counters/histograms with safe labels)
		globalMeter = otel.Meter("ampyobs")
		if err := initMetrics(); err != nil {
			return fmt.Errorf("init metrics: %w", err)
		}

		// Runtime metrics (GC, mem, goroutines, etc.)
		_ = runtime.Start(
			runtime.WithMinimumReadMemStatsInterval(10*time.Second),
			runtime.WithMeterProvider(mp),
		)
	}

	return nil
}

func newTracerProvider(cfg Config, res *resource.Resource) (*sdktrace.TracerProvider, error) {
	endpoint, insecure := parseEndpoint(cfg.CollectorEndpoint)
	protocol := strings.ToLower(cfg.TraceProtocol)
	if protocol == "" {
		protocol = "grpc" // default
	}

	var exp sdktrace.SpanExporter
	var err error

	switch protocol {
	case "http":
		// HTTP exporter (port 4318)
		opts := []otlptracehttp.Option{
			otlptracehttp.WithEndpoint(endpoint),
		}
		if insecure {
			opts = append(opts, otlptracehttp.WithInsecure())
		}
		exp, err = otlptracehttp.New(context.Background(), opts...)
		if err != nil {
			return nil, fmt.Errorf("otlptrace http exporter: %w", err)
		}
	case "grpc":
		// gRPC exporter (port 4317)
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(endpoint),
		}
		if insecure {
			opts = append(opts, otlptracegrpc.WithInsecure())
		} else {
			opts = append(opts, otlptracegrpc.WithTLSCredentials(credentials.NewTLS(&tls.Config{})))
		}
		exp, err = otlptracegrpc.New(context.Background(), opts...)
		if err != nil {
			return nil, fmt.Errorf("otlptrace grpc exporter: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported trace protocol: %s (use 'grpc' or 'http')", cfg.TraceProtocol)
	}

	sampler := sdktrace.ParentBased(sdktrace.TraceIDRatioBased(0.25))
	switch strings.ToLower(cfg.Sampler) {
	case "ratio":
		if cfg.SampleRatio >= 0 && cfg.SampleRatio <= 1 {
			sampler = sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.SampleRatio))
		}
	case "parent", "":
		// keep default
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
		sdktrace.WithBatcher(exp,
			sdktrace.WithMaxExportBatchSize(512),
			sdktrace.WithBatchTimeout(5*time.Second)),
	)
	return tp, nil
}

func newMeterProvider(cfg Config, res *resource.Resource) (*sdkmetric.MeterProvider, error) {
	endpoint, insecure := parseEndpoint(cfg.CollectorEndpoint)

	opts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithEndpoint(endpoint),
	}
	if insecure {
		opts = append(opts, otlpmetricgrpc.WithInsecure())
	} else {
		opts = append(opts, otlpmetricgrpc.WithTLSCredentials(credentials.NewTLS(&tls.Config{})))
	}

	exp, err := otlpmetricgrpc.New(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("otlpmetric exporter: %w", err)
	}

	reader := sdkmetric.NewPeriodicReader(exp,
		sdkmetric.WithInterval(10*time.Second),
	)
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(reader),
		sdkmetric.WithView(getMetricViews()...),
	)
	return mp, nil
}

func Shutdown(ctx context.Context) error {
	if meterProvider != nil {
		_ = meterProvider.Shutdown(ctx)
	}
	if tracerProvider != nil {
		return tracerProvider.Shutdown(ctx)
	}
	return nil
}

func parseEndpoint(raw string) (hostport string, insecure bool) {
	if raw == "" {
		return "localhost:4317", true
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" {
		// likely "host:port"
		host, port, _ := net.SplitHostPort(raw)
		if port == "" {
			return raw, true
		}
		if host == "" {
			return "localhost:" + port, true
		}
		return raw, true
	}
	insecure = (u.Scheme == "http")
	return u.Host, insecure
}
