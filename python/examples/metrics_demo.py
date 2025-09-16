import random, time

from ampyobs import Config, init, shutdown
from ampyobs.metrics import (
    init_instruments,
    bus_produced, bus_consumed, bus_delivery_latency_ms,
    oms_order_submit, oms_order_latency_ms, oms_reject,
)

def main():
    cfg = Config(service_name="py-metrics", collector_endpoint="localhost:4317")
    init(cfg)
    init_instruments()

    topic = "ampy/dev/bars/v1"
    for i in range(10):
        bus_produced(topic, 5, service=cfg.service_name, env=cfg.environment)
        bus_consumed(topic, 5, service=cfg.service_name, env=cfg.environment)
        bus_delivery_latency_ms(topic, 5 + random.random()*80, service=cfg.service_name, env=cfg.environment)

        oms_order_submit("alpaca", "ok", service=cfg.service_name, env=cfg.environment)
        oms_order_latency_ms("alpaca", 20 + random.random()*60, service=cfg.service_name, env=cfg.environment)
        if i % 4 == 0:
            oms_reject("alpaca", "risk_check", service=cfg.service_name, env=cfg.environment)
        time.sleep(0.2)

    time.sleep(1.0)
    shutdown()

if __name__ == "__main__":
    main()
