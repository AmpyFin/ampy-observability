import json, time
from uuid import uuid4
from ampyobs import Config, init, shutdown
from ampyobs.logger import L, _get_logger
from ampyobs.tracing import BusAttrs, start_bus_consume

def main():
    init(Config(service_name="py-consumer", collector_endpoint="localhost:4317"))
    with open("bus_headers.json") as f:
        headers = json.load(f)

    a = BusAttrs(topic="ampy/dev/signals/v1", schema_fqdn="ampy.signals.v1.Signal",
                 message_id=str(uuid4()), partition_key="AAPL", run_id="dev_session_1")

    with start_bus_consume(headers, a) as span:
        logger = L if L is not None else _get_logger()
        logger.info("consumed signal", extra={"event":"signals.consume","action":"forward_to_oms"})
        time.sleep(0.5)

    shutdown()

if __name__ == "__main__":
    main()
