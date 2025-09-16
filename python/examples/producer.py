import json, time
from uuid import uuid4
from ampyobs import Config, init, shutdown
from ampyobs.logger import L, _get_logger
from ampyobs.tracing import BusAttrs, start_bus_publish
from ampyobs.propagation import inject_trace
from opentelemetry.context import get_current

def main():
    init(Config(service_name="py-producer", collector_endpoint="localhost:4317"))
    a = BusAttrs(topic="ampy/dev/signals/v1", schema_fqdn="ampy.signals.v1.Signal",
                 message_id=str(uuid4()), partition_key="AAPL", run_id="dev_session_1")

    with start_bus_publish(a) as span:
        logger = L if L is not None else _get_logger()
        logger.info("publishing signal", extra={"event":"signals.emit","symbol":"AAPL"})
        headers = {}
        inject_trace(headers, context=get_current())  # ensure we use the active context
        with open("bus_headers.json","w") as f: json.dump(headers, f, indent=2)
        print("wrote bus_headers.json:", headers)
        time.sleep(0.5)

    shutdown()

if __name__ == "__main__":
    main()
