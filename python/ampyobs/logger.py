import logging
import sys
from typing import Any, Optional

from opentelemetry import trace as otel_trace

# Global logger instance
L: Optional[logging.LoggerAdapter] = None

# Fallback logger initialization
def _get_logger():
    if L is None:
        # Create a basic logger if not initialized
        logger = logging.getLogger("ampyobs")
        if not logger.handlers:
            handler = logging.StreamHandler(sys.stdout)
            handler.setFormatter(JsonFormatter())
            logger.addHandler(handler)
            logger.setLevel(logging.INFO)
        return logger
    return L

class TraceContextFilter(logging.Filter):
    def filter(self, record: logging.LogRecord) -> bool:
        span = otel_trace.get_current_span()
        span_ctx = span.get_span_context()
        if span_ctx and span_ctx.is_valid:
            record.trace_id = span_ctx.trace_id
            record.span_id = span_ctx.span_id
            # hex format
            record.trace_id_hex = format(span_ctx.trace_id, "032x")
            record.span_id_hex = format(span_ctx.span_id, "016x")
        else:
            record.trace_id = None
            record.span_id = None
            record.trace_id_hex = None
            record.span_id_hex = None
        return True

class JsonFormatter(logging.Formatter):
    # Minimal JSON without extra deps; relies on default LogRecord.__dict__
    def format(self, record: logging.LogRecord) -> str:
        # Build a dict of stable fields
        d: dict[str, Any] = {
            "ts": self.formatTime(record, "%Y-%m-%dT%H:%M:%S.%fZ"),
            "level": record.levelname.lower(),
            "message": record.getMessage(),
            "service": getattr(record, "service", None),
            "env": getattr(record, "env", None),
            "service_version": getattr(record, "service_version", None),
            "trace_id": getattr(record, "trace_id_hex", None),
            "span_id": getattr(record, "span_id_hex", None),
        }
        # Include any extra attributes from `logger.bind(...)` style calls
        for k, v in record.__dict__.items():
            if k in d or k.startswith("_"):
                continue
            if k in ("args", "msg", "exc_text", "exc_info", "stack_info", "lineno", "pathname", "filename", "module", "funcName", "created", "msecs", "relativeCreated", "thread", "threadName", "process", "processName"):
                continue
            d[k] = v
        import json
        return json.dumps(d, separators=(",", ":"), ensure_ascii=False)

def setup_json_logging(service: str, env: str, version: str) -> None:
    root = logging.getLogger()
    root.setLevel(logging.INFO)

    handler = logging.StreamHandler(sys.stdout)
    fmt = JsonFormatter()
    handler.setFormatter(fmt)

    # Attach default fields to every record
    class ServiceAdapter(logging.LoggerAdapter):
        def process(self, msg, kwargs):
            extra = kwargs.setdefault("extra", {})
            extra.setdefault("service", service)
            extra.setdefault("env", env)
            extra.setdefault("service_version", version)
            return msg, kwargs

    # Add trace filter
    handler.addFilter(TraceContextFilter())
    root.handlers = [handler]

    # Expose a module-level logger adapter
    global L
    L = ServiceAdapter(logging.getLogger("ampyobs"), {})
    
    # Also set up a fallback logger if L is still None
    if L is None:
        L = ServiceAdapter(logging.getLogger("ampyobs"), {})
