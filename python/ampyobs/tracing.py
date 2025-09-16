from dataclasses import dataclass
from contextlib import contextmanager
from typing import Dict, Iterator

from opentelemetry import trace as otel_trace
from opentelemetry.trace import Span, SpanKind

@dataclass
class BusAttrs:
    topic: str
    schema_fqdn: str
    message_id: str
    partition_key: str
    run_id: str

@contextmanager
def start_span(name: str, kind: SpanKind = SpanKind.INTERNAL, **attrs) -> Iterator[Span]:
    tracer = otel_trace.get_tracer("ampyobs")
    with tracer.start_as_current_span(name, kind=kind) as span:
        for k, v in attrs.items():
            span.set_attribute(k, v)
        yield span  # span will end on exiting the 'with'

@contextmanager
def start_bus_publish(a: BusAttrs) -> Iterator[Span]:
    with start_span(
        "bus.publish",
        SpanKind.PRODUCER,
        topic=a.topic,
        schema_fqdn=a.schema_fqdn,
        message_id=a.message_id,
        partition_key=a.partition_key,
        run_id=a.run_id,
    ) as span:
        yield span

@contextmanager
def start_bus_consume(headers: Dict[str, str], a: BusAttrs) -> Iterator[Span]:
    from .propagation import extract_trace
    remote_ctx = extract_trace(headers)
    tracer = otel_trace.get_tracer("ampyobs")
    with tracer.start_as_current_span("bus.consume", context=remote_ctx, kind=SpanKind.CONSUMER) as span:
        for k, v in dict(topic=a.topic, schema_fqdn=a.schema_fqdn, message_id=a.message_id, partition_key=a.partition_key, run_id=a.run_id).items():
            span.set_attribute(k, v)
        yield span
