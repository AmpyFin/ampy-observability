package ampyobs

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Global meter assigned after Init sets the meter provider.
var globalMeter metric.Meter

// Instruments
var (
	busProduced        metric.Int64Counter
	busConsumed        metric.Int64Counter
	busDeliveryLatency metric.Float64Histogram

	omsOrderSubmit  metric.Int64Counter
	omsOrderLatency metric.Float64Histogram
	omsRejections   metric.Int64Counter
)

// Public enums (bounded label values)
const (
	OutcomeOK     = "ok"
	OutcomeRetry  = "retry"
	OutcomeDLQ    = "dlq"
	OutcomeReject = "reject"
)

// initMetrics constructs instruments. Call once after MeterProvider is set.
func initMetrics() error {
	if globalMeter == nil {
		globalMeter = otel.Meter("ampyobs")
	}

	var err error

	// Bus
	busProduced, err = globalMeter.Int64Counter(
		"ampy.bus.produced_total",
		metric.WithDescription("Messages produced to ampy-bus"),
	)
	if err != nil {
		return err
	}

	busConsumed, err = globalMeter.Int64Counter(
		"ampy.bus.consumed_total",
		metric.WithDescription("Messages consumed from ampy-bus"),
	)
	if err != nil {
		return err
	}

	busDeliveryLatency, err = globalMeter.Float64Histogram(
		"ampy.bus.delivery_latency_ms",
		metric.WithDescription("Bus end-to-end delivery latency in milliseconds"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return err
	}

	// OMS
	omsOrderSubmit, err = globalMeter.Int64Counter(
		"ampy.oms.order_submit_total",
		metric.WithDescription("Order submissions by outcome"),
	)
	if err != nil {
		return err
	}

	omsOrderLatency, err = globalMeter.Float64Histogram(
		"ampy.oms.order_latency_ms",
		metric.WithDescription("OMS order latency (submit→ack) in milliseconds"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return err
	}

	omsRejections, err = globalMeter.Int64Counter(
		"ampy.oms.rejections_total",
		metric.WithDescription("Order rejections by reason"),
	)
	if err != nil {
		return err
	}

	return nil
}

// ----------- Helper Recording Functions (safe labels only) -----------

// BusProducedAdd increments produced counter for a topic.
func BusProducedAdd(ctx context.Context, topic string, n int64) {
	busProduced.Add(ctx, n,
		metric.WithAttributes(
			attribute.String("topic", topic),
			attribute.String("service", globalCfg.ServiceName),
			attribute.String("env", globalCfg.Environment),
		),
	)
}

// BusConsumedAdd increments consumed counter for a topic.
func BusConsumedAdd(ctx context.Context, topic string, n int64) {
	busConsumed.Add(ctx, n,
		metric.WithAttributes(
			attribute.String("topic", topic),
			attribute.String("service", globalCfg.ServiceName),
			attribute.String("env", globalCfg.Environment),
		),
	)
}

// BusDeliveryLatencyMs records bus delivery latency for a topic.
func BusDeliveryLatencyMs(ctx context.Context, topic string, ms float64) {
	busDeliveryLatency.Record(ctx, ms,
		metric.WithAttributes(
			attribute.String("topic", topic),
			attribute.String("service", globalCfg.ServiceName),
			attribute.String("env", globalCfg.Environment),
		),
	)
}

// OMSOrderSubmitAdd increments order submit counter for a broker+outcome.
func OMSOrderSubmitAdd(ctx context.Context, broker string, outcome string) {
	omsOrderSubmit.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("broker", broker),
			attribute.String("outcome", outcome),
			attribute.String("service", globalCfg.ServiceName),
			attribute.String("env", globalCfg.Environment),
		),
	)
}

// OMSOrderLatencyMs records order latency for a broker.
func OMSOrderLatencyMs(ctx context.Context, broker string, ms float64) {
	omsOrderLatency.Record(ctx, ms,
		metric.WithAttributes(
			attribute.String("broker", broker),
			attribute.String("service", globalCfg.ServiceName),
			attribute.String("env", globalCfg.Environment),
		),
	)
}

// OMSRejectAdd increments rejection counter for a broker+reason.
func OMSRejectAdd(ctx context.Context, broker string, reason string) {
	omsRejections.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("broker", broker),
			attribute.String("reason", reason),
			attribute.String("service", globalCfg.ServiceName),
			attribute.String("env", globalCfg.Environment),
		),
	)
}
