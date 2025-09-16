from typing import Dict, Optional
from opentelemetry.context import get_current
from opentelemetry.propagate import get_global_textmap

def inject_trace(headers: Dict[str, str], context=None) -> None:
    ctx = context or get_current()
    get_global_textmap().inject(headers, context=ctx)

def extract_trace(headers: Dict[str, str]):
    return get_global_textmap().extract(headers)
