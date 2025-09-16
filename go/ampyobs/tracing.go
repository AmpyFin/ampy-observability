package ampyobs

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// BusAttrs captures stable, low-cardinality attributes for bus spans.
type BusAttrs struct {
	Topic        string
	SchemaFQDN   string
	MessageID    string
	PartitionKey string
	RunID        string
}

// StartSpan creates a span with a conventional name and kind.
func StartSpan(ctx context.Context, name string, kind trace.SpanKind, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	tr := otel.Tracer("ampyobs")
	opts := []trace.SpanStartOption{
		trace.WithSpanKind(kind),
		trace.WithAttributes(attrs...),
	}
	return tr.Start(ctx, name, opts...)
}

// StartBusPublishSpan creates a `bus.publish` span with standardized attributes.
func StartBusPublishSpan(ctx context.Context, a BusAttrs) (context.Context, trace.Span) {
	return StartSpan(ctx, "bus.publish", trace.SpanKindProducer,
		attribute.String("topic", a.Topic),
		attribute.String("schema_fqdn", a.SchemaFQDN),
		attribute.String("message_id", a.MessageID),
		attribute.String("partition_key", a.PartitionKey),
		attribute.String("run_id", a.RunID),
	)
}

// StartBusConsumeSpan extracts W3C context from headers and starts `bus.consume`
// as a child of the upstream span. It also adds a span link to the upstream context.
func StartBusConsumeSpan(parent context.Context, headers map[string]string, a BusAttrs) (context.Context, trace.Span) {
	remoteCtx := ExtractTrace(parent, headers) // from propagation.go
	link := trace.LinkFromContext(remoteCtx)

	tr := otel.Tracer("ampyobs")
	return tr.Start(remoteCtx, "bus.consume",
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(
			attribute.String("topic", a.Topic),
			attribute.String("schema_fqdn", a.SchemaFQDN),
			attribute.String("message_id", a.MessageID),
			attribute.String("partition_key", a.PartitionKey),
			attribute.String("run_id", a.RunID),
		),
		trace.WithLinks(link),
	)
}
