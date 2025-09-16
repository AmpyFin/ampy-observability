package ampyobs

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

const (
	HeaderTraceParent = "traceparent"
	HeaderTraceState  = "tracestate"

	// Optional AmpyFin correlation headers (only if you choose to send them via ampy-bus):
	HeaderRunID      = "run_id"
	HeaderUniverseID = "universe_id"
	HeaderAsOf       = "as_of"
)

// InjectTrace injects W3C trace context into key/value headers.
func InjectTrace(ctx context.Context, headers map[string]string) {
	otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(headers))
}

// ExtractTrace extracts W3C trace context from headers and returns a child context.
func ExtractTrace(parent context.Context, headers map[string]string) context.Context {
	return otel.GetTextMapPropagator().Extract(parent, propagation.MapCarrier(headers))
}
