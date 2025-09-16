# Minimal JSON-structured logging facade (scaffold).
# Real JSON formatting + trace/metric correlation to be added later.

from __future__ import annotations
import json, sys, time
from typing import Any, Protocol

class _GlobalLogger(Protocol):
    def debug(self, msg: str, **kv: Any) -> None: ...
    def info(self, msg: str, **kv: Any) -> None: ...
    def warn(self, msg: str, **kv: Any) -> None: ...
    def error(self, msg: str, **kv: Any) -> None: ...

class _NopLogger:
    def debug(self, msg: str, **kv: Any) -> None: pass
    def info(self, msg: str, **kv: Any) -> None: pass
    def warn(self, msg: str, **kv: Any) -> None: pass
    def error(self, msg: str, **kv: Any) -> None: pass

def _emit(level: str, msg: str, kv: dict[str, Any]) -> None:
    rec = {"ts": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()), "level": level, "message": msg}
    if kv:
        rec.update(kv)
    sys.stdout.write(json.dumps(rec, separators=(",", ":"), ensure_ascii=False) + "\n")
    sys.stdout.flush()

class _StdoutLogger:
    def debug(self, msg: str, **kv: Any) -> None: _emit("debug", msg, kv)
    def info(self, msg: str, **kv: Any) -> None: _emit("info", msg, kv)
    def warn(self, msg: str, **kv: Any) -> None: _emit("warn", msg, kv)
    def error(self, msg: str, **kv: Any) -> None: _emit("error", msg, kv)

_global_logger: _GlobalLogger = _StdoutLogger()

def get_logger() -> _GlobalLogger:
    return _global_logger
