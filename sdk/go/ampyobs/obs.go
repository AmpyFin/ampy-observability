package ampyobs

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	CollectorGRPC  string // "127.0.0.1:4317" (your Collector)
}

type Handle struct {
	cfg     Config
	tp      *sdktrace.TracerProvider
	Logger  Logger
	Metrics *Metrics
}

func Init(ctx context.Context, cfg Config) (*Handle, error) {
	if cfg.CollectorGRPC == "" {
		cfg.CollectorGRPC = "127.0.0.1:4317"
	}
	res, err := resource.New(ctx,
		resource.WithAttributes(
			attribute.String("service.name", cfg.ServiceName),
			attribute.String("service.version", cfg.ServiceVersion),
			attribute.String("deployment.environment", cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("resource: %w", err)
	}

	exp, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(cfg.CollectorGRPC),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("otlp exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return &Handle{
		cfg:     cfg,
		tp:      tp,
		Logger:  newLogger(cfg),
		Metrics: NewMetrics(),
	}, nil
}

func (h *Handle) Tracer(name string) trace.Tracer {
	return h.tp.Tracer(name)
}

func (h *Handle) Shutdown(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return h.tp.Shutdown(ctx)
}
