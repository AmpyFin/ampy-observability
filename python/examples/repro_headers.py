import json, time
from ampyobs import Config, init, shutdown
from ampyobs.tracing import start_span
from ampyobs.propagation import inject_trace
from opentelemetry.context import get_current

def main():
    init(Config(service_name="py-repro", collector_endpoint="localhost:4317"))
    with start_span("demo.span"):
        headers = {}
        inject_trace(headers, context=get_current())
        print("headers:", headers)  # should include 'traceparent'
        time.sleep(0.2)
    shutdown()

if __name__ == "__main__":
    main()
