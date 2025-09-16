from opentelemetry import metrics as otel_metrics
from opentelemetry.metrics import Counter, Histogram
from opentelemetry.sdk.metrics import MeterProvider

# Instruments (module-level singletons after init)
_bus_produced: Counter
_bus_consumed: Counter
_bus_delivery_latency: Histogram
_oms_order_submit: Counter
_oms_order_latency: Histogram
_oms_rejections: Counter

def init_instruments() -> None:
    meter = otel_metrics.get_meter_provider().get_meter("ampyobs")

    global _bus_produced, _bus_consumed, _bus_delivery_latency
    global _oms_order_submit, _oms_order_latency, _oms_rejections

    _bus_produced = meter.create_counter(
        name="ampy.bus.produced_total",
        description="Messages produced to ampy-bus",
    )
    _bus_consumed = meter.create_counter(
        name="ampy.bus.consumed_total",
        description="Messages consumed from ampy-bus",
    )
    _bus_delivery_latency = meter.create_histogram(
        name="ampy.bus.delivery_latency_ms",
        unit="ms",
        description="Bus end-to-end delivery latency (ms)",
    )

    _oms_order_submit = meter.create_counter(
        name="ampy.oms.order_submit_total",
        description="Order submissions by outcome",
    )
    _oms_order_latency = meter.create_histogram(
        name="ampy.oms.order_latency_ms",
        unit="ms",
        description="OMS order latency (submit→ack) in ms",
    )
    _oms_rejections = meter.create_counter(
        name="ampy.oms.rejections_total",
        description="Order rejections by reason",
    )

def bus_produced(topic: str, n: int = 1, service: str | None = None, env: str | None = None) -> None:
    _bus_produced.add(n, {"topic": topic, "service": service or "", "env": env or ""})

def bus_consumed(topic: str, n: int = 1, service: str | None = None, env: str | None = None) -> None:
    _bus_consumed.add(n, {"topic": topic, "service": service or "", "env": env or ""})

def bus_delivery_latency_ms(topic: str, ms: float, service: str | None = None, env: str | None = None) -> None:
    _bus_delivery_latency.record(ms, {"topic": topic, "service": service or "", "env": env or ""})

def oms_order_submit(broker: str, outcome: str, service: str | None = None, env: str | None = None) -> None:
    _oms_order_submit.add(1, {"broker": broker, "outcome": outcome, "service": service or "", "env": env or ""})

def oms_order_latency_ms(broker: str, ms: float, service: str | None = None, env: str | None = None) -> None:
    _oms_order_latency.record(ms, {"broker": broker, "service": service or "", "env": env or ""})

def oms_reject(broker: str, reason: str, service: str | None = None, env: str | None = None) -> None:
    _oms_rejections.add(1, {"broker": broker, "reason": reason, "service": service or "", "env": env or ""})
