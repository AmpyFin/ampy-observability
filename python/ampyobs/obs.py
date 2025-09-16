from dataclasses import dataclass
from typing import Optional

from opentelemetry import metrics as otel_metrics, trace as otel_trace
from opentelemetry.sdk.resources import Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter

from opentelemetry.sdk.metrics import MeterProvider
from opentelemetry.sdk.metrics.export import PeriodicExportingMetricReader
from opentelemetry.exporter.otlp.proto.grpc.metric_exporter import OTLPMetricExporter

from .logger import setup_json_logging

@dataclass
class Config:
    service_name: str = "ampy-python"
    service_version: str = "0.1.0"
    environment: str = "dev"
    collector_endpoint: str = "localhost:4317"  # gRPC
    enable_logs: bool = True
    enable_tracing: bool = True
    enable_metrics: bool = True
    sample_always: bool = True

_tracer_provider: Optional[TracerProvider] = None
_meter_provider: Optional[MeterProvider] = None

def init(cfg: Config) -> None:
    resource = Resource.create({
        "service.name": cfg.service_name,
        "service.version": cfg.service_version,
        "deployment.environment": cfg.environment,
    })

    # ----- Logs -----
    if cfg.enable_logs:
        setup_json_logging(cfg.service_name, cfg.environment, cfg.service_version)

    # ----- Tracing -----
    if cfg.enable_tracing:
        span_exporter = OTLPSpanExporter(endpoint=cfg.collector_endpoint, insecure=True)
        sp = BatchSpanProcessor(span_exporter)

        global _tracer_provider
        _tracer_provider = TracerProvider(resource=resource)
        _tracer_provider.add_span_processor(sp)
        if cfg.sample_always:
            # AlwaysOn is the default sampler for Python when not overridden; BatchSpanProcessor handles buffering.
            pass
        otel_trace.set_tracer_provider(_tracer_provider)

    # ----- Metrics -----
    if cfg.enable_metrics:
        metric_exporter = OTLPMetricExporter(endpoint=cfg.collector_endpoint, insecure=True)
        reader = PeriodicExportingMetricReader(metric_exporter)
        global _meter_provider
        _meter_provider = MeterProvider(resource=resource, metric_readers=[reader])
        otel_metrics.set_meter_provider(_meter_provider)

def shutdown() -> None:
    # Flush providers if present
    if _tracer_provider is not None:
        _tracer_provider.shutdown()
    if _meter_provider is not None:
        _meter_provider.shutdown()
