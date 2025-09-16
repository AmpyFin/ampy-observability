# Minimal bootstrap (scaffold). Real implementations will hook OTel & JSON logs.
from typing import Optional, Dict, Any
from .logging import _GlobalLogger, _NopLogger

_global_cfg: Dict[str, Any] = {}
_global_logger: _GlobalLogger = _NopLogger()

def init(
    service_name: str,
    service_version: str = "",
    environment: str = "dev",
    collector_endpoint: str = "http://localhost:4317",
    enable_logs: bool = True,
    enable_metrics: bool = True,
    enable_tracing: bool = True,
    sampler: str = "parent",
    sample_ratio: float = 0.25,
) -> None:
    """Initialize global observability (scaffold).

    All params will be wired to actual backends later.
    """
    global _global_cfg, _global_logger
    _global_cfg = {
        "service_name": service_name,
        "service_version": service_version,
        "environment": environment,
        "collector_endpoint": collector_endpoint,
        "enable_logs": enable_logs,
        "enable_metrics": enable_metrics,
        "enable_tracing": enable_tracing,
        "sampler": sampler,
        "sample_ratio": sample_ratio,
    }
    _global_logger = _NopLogger()

def shutdown() -> None:
    """Flush/close exporters (no-op scaffold)."""
    pass
