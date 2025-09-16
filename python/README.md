# Ampy Observability SDK (Python)

Python SDK for Ampy observability stack - logs, metrics, and tracing.

## Installation

```bash
pip install -e ./python  # Editable install for development
```

## Quick Start

```python
from ampyobs import Config, init, shutdown
from ampyobs.logger import L
from ampyobs.tracing import BusAttrs, start_bus_publish
from ampyobs.metrics import init_instruments, bus_produced

# Initialize
init(Config(service_name="my-service", collector_endpoint="localhost:4317"))
init_instruments()

# Logging with trace context
L.info("Processing order", extra={"order_id": "12345"})

# Tracing
attrs = BusAttrs(topic="orders", schema_fqdn="order.v1", message_id="123", partition_key="AAPL", run_id="run1")
ctx, span = start_bus_publish(attrs)
try:
    # Your business logic
    pass
finally:
    span.end()

# Metrics
bus_produced("orders", 1, service="my-service", env="dev")

# Cleanup
shutdown()
```

## Examples

See `python/examples/` for complete producer/consumer and metrics demos.

## Configuration

The SDK sends telemetry to your local OTel Collector (localhost:4317 by default).