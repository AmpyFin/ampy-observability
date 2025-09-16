# Ampy Observability SDK

[![Version](https://img.shields.io/badge/version-0.0.1-blue.svg)](https://github.com/ampyfin/ampy-observability/releases/tag/v0.0.1)
[![Go Version](https://img.shields.io/badge/go-1.22%2B-blue.svg)](https://golang.org/dl/)
[![Python Version](https://img.shields.io/badge/python-3.10%2B-blue.svg)](https://www.python.org/downloads/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![OpenTelemetry](https://img.shields.io/badge/OpenTelemetry-1.37.0-orange.svg)](https://opentelemetry.io/)

A comprehensive observability solution providing logs, metrics, and distributed tracing for Go and Python applications. Built on OpenTelemetry with production-ready collector configuration, SLO monitoring, and alerting.

## Release v0.0.1 - Initial Release

**Release Date**: September 2024

### Features
- **Go SDK**: Complete observability SDK with HTTP/gRPC exporters
- **Python SDK**: Full-featured Python SDK with context managers and proper tracing
- **Distributed Tracing**: W3C trace context propagation across services
- **Structured Logging**: JSON logging with automatic trace context enrichment
- **Domain Metrics**: Bus and OMS specific counters and histograms
- **Production-Ready Collector**: Hardened with sampling, redaction, and memory limits
- **SLO Monitoring**: Built-in Prometheus alert rules for latency and error rates
- **CI/CD Pipeline**: Comprehensive testing with integration smoke tests

### Components
- **Observability Stack**: Jaeger, Prometheus, Grafana, Loki, OpenTelemetry Collector
- **Go SDK**: `github.com/ampyfin/ampy-observability/go/ampyobs`
- **Python SDK**: `ampyobs` package on PyPI
- **Docker Compose**: One-command stack deployment
- **GitHub Actions**: Automated testing and publishing

### Installation
```bash
# Go
go get github.com/ampyfin/ampy-observability/go/ampyobs

# Python
pip install ampyobs
```

### Quick Start
```bash
# Start observability stack
cd deploy && docker compose up -d

# Run examples
cd go/examples && go run ./bus_producer
cd python/examples && python producer.py
```

## Problem Statement

Modern distributed systems require comprehensive observability to understand system behavior, debug issues, and maintain service level objectives (SLOs). Traditional logging and monitoring approaches often fall short because they:

- Lack correlation between logs, metrics, and traces
- Provide incomplete visibility into distributed request flows
- Don't scale well with microservices architectures
- Require complex setup and maintenance
- Have inconsistent instrumentation across different languages

This project addresses these challenges by providing:

- **Unified Observability**: Consistent logging, metrics, and tracing across Go and Python
- **Distributed Tracing**: End-to-end request flow visibility with W3C trace context propagation
- **Production-Ready**: Hardened collector with sampling, redaction, and memory limits
- **SLO Monitoring**: Built-in alerting for latency and error rate thresholds
- **Easy Integration**: Simple SDKs with minimal configuration required

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Go Services   │    │ Python Services │    │  Other Services │
│                 │    │                 │    │                 │
│  ampyobs SDK    │    │  ampyobs SDK    │    │  OTLP Clients   │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌─────────────▼─────────────┐
                    │  OpenTelemetry Collector  │
                    │                           │
                    │  • Sampling (10%)         │
                    │  • PII Redaction          │
                    │  • Memory Limits          │
                    │  • Batch Processing       │
                    └─────────────┬─────────────┘
                                  │
                    ┌─────────────▼─────────────┐
                    │     Observability Stack   │
                    │                           │
                    │  • Jaeger (Traces)        │
                    │  • Prometheus (Metrics)   │
                    │  • Grafana (Dashboards)   │
                    │  • Loki (Logs)            │
                    └───────────────────────────┘
```

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.22+ (for Go SDK)
- Python 3.10+ (for Python SDK)

### 1. Start the Observability Stack

```bash
cd deploy
docker compose up -d
```

This starts:
- OpenTelemetry Collector (ports 4317, 4318)
- Jaeger UI (http://localhost:16686)
- Prometheus (http://localhost:9090)
- Grafana (http://localhost:3000)
- Loki (http://localhost:3100)

### 2. Verify Stack is Running

```bash
# Check all services are up
docker compose ps

# Verify endpoints
curl http://localhost:16686/api/services  # Jaeger
curl http://localhost:9090/api/v1/query?query=up  # Prometheus
```

## Go SDK Installation and Usage

### Installation

```bash
go get github.com/ampyfin/ampy-observability/go/ampyobs
```

### Basic Setup

```go
package main

import (
    "context"
    "time"
    
    "github.com/ampyfin/ampy-observability/go/ampyobs"
)

func main() {
    // Initialize observability
    cfg := ampyobs.Config{
        ServiceName:    "my-go-service",
        ServiceVersion: "1.0.0",
        Environment:    "production",
        CollectorEndpoint: "localhost:4317", // gRPC
        TraceProtocol:  "grpc", // or "http"
    }
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := ampyobs.Init(ctx, cfg); err != nil {
        log.Fatal("Failed to initialize observability:", err)
    }
    defer ampyobs.Shutdown(ctx)
    
    // Your application code here
}
```

### Logging

```go
import "github.com/ampyfin/ampy-observability/go/ampyobs"

// Structured logging with trace context
ampyobs.Logger.Info("Processing order",
    "order_id", "12345",
    "customer_id", "67890",
    "amount", 99.99,
)

// Error logging
ampyobs.Logger.Error("Payment failed",
    "error", err,
    "order_id", "12345",
    "payment_method", "credit_card",
)
```

### Metrics

```go
import "github.com/ampyfin/ampy-observability/go/ampyobs"

// Counter metrics
ampyobs.BusProduced("orders", 1, "my-service", "production")
ampyobs.OMSOrderSubmit("alpaca", "success", "my-service", "production")

// Histogram metrics (latency)
ampyobs.BusDeliveryLatency("orders", 150.5, "my-service", "production")
ampyobs.OMSOrderLatency("alpaca", 89.2, "my-service", "production")

// Error metrics
ampyobs.OMSReject("alpaca", "insufficient_funds", "my-service", "production")
```

### Distributed Tracing

```go
import (
    "context"
    "github.com/ampyfin/ampy-observability/go/ampyobs"
)

// Create a span
ctx, span := ampyobs.StartSpan(ctx, "process.payment")
defer span.End()

// Add attributes
span.SetAttributes(
    attribute.String("payment.method", "credit_card"),
    attribute.String("order.id", "12345"),
    attribute.Float64("amount", 99.99),
)

// Create child spans
ctx, childSpan := ampyobs.StartSpan(ctx, "validate.card")
defer childSpan.End()

// Your business logic here
```

### Bus Message Tracing

```go
// Producer
attrs := ampyobs.BusAttrs{
    Topic:        "orders.v1",
    SchemaFQDN:   "com.ampy.orders.v1.OrderCreated",
    MessageID:    "msg-123",
    PartitionKey: "customer-456",
    RunID:        "run-789",
}

ctx, span := ampyobs.StartBusPublish(ctx, attrs)
defer span.End()

// Inject trace context into message headers
headers := make(map[string]string)
ampyobs.InjectTrace(ctx, headers)

// Consumer
ctx, span := ampyobs.StartBusConsume(ctx, headers, attrs)
defer span.End()
```

## Python SDK Installation and Usage

### Installation

```bash
pip install -e ./python
```

### Basic Setup

```python
from ampyobs import Config, init, shutdown

# Initialize observability
cfg = Config(
    service_name="my-python-service",
    service_version="1.0.0",
    environment="production",
    collector_endpoint="localhost:4317"  # gRPC
)

init(cfg)

try:
    # Your application code here
    pass
finally:
    shutdown()
```

### Logging

```python
from ampyobs.logger import L, _get_logger

# Get logger (with fallback)
logger = L if L is not None else _get_logger()

# Structured logging with trace context
logger.info("Processing order", extra={
    "order_id": "12345",
    "customer_id": "67890",
    "amount": 99.99
})

# Error logging
logger.error("Payment failed", extra={
    "error": str(err),
    "order_id": "12345",
    "payment_method": "credit_card"
})
```

### Metrics

```python
from ampyobs.metrics import (
    init_instruments,
    bus_produced, bus_consumed, bus_delivery_latency_ms,
    oms_order_submit, oms_order_latency_ms, oms_reject,
)

# Initialize metrics instruments
init_instruments()

# Counter metrics
bus_produced("orders", 1, service="my-service", env="production")
oms_order_submit("alpaca", "success", service="my-service", env="production")

# Histogram metrics (latency)
bus_delivery_latency_ms("orders", 150.5, service="my-service", env="production")
oms_order_latency_ms("alpaca", 89.2, service="my-service", env="production")

# Error metrics
oms_reject("alpaca", "insufficient_funds", service="my-service", env="production")
```

### Distributed Tracing

```python
from ampyobs.tracing import start_span

# Create a span
with start_span("process.payment") as span:
    span.set_attribute("payment.method", "credit_card")
    span.set_attribute("order.id", "12345")
    span.set_attribute("amount", 99.99)
    
    # Create child spans
    with start_span("validate.card") as child_span:
        # Your business logic here
        pass
```

### Bus Message Tracing

```python
from ampyobs.tracing import BusAttrs, start_bus_publish, start_bus_consume
from ampyobs.propagation import inject_trace, extract_trace

# Producer
attrs = BusAttrs(
    topic="orders.v1",
    schema_fqdn="com.ampy.orders.v1.OrderCreated",
    message_id="msg-123",
    partition_key="customer-456",
    run_id="run-789"
)

with start_bus_publish(attrs) as span:
    # Inject trace context into message headers
    headers = {}
    inject_trace(headers)
    
    # Send message with headers

# Consumer
with start_bus_consume(headers, attrs) as span:
    # Process message
    pass
```

## Example Use Cases

### 1. E-commerce Order Processing

**Go Example:**
```go
func ProcessOrder(ctx context.Context, order Order) error {
    ctx, span := ampyobs.StartSpan(ctx, "process.order")
    defer span.End()
    
    span.SetAttributes(
        attribute.String("order.id", order.ID),
        attribute.String("customer.id", order.CustomerID),
        attribute.Float64("order.amount", order.Amount),
    )
    
    // Validate order
    ctx, validateSpan := ampyobs.StartSpan(ctx, "validate.order")
    if err := validateOrder(ctx, order); err != nil {
        validateSpan.End()
        ampyobs.OMSReject("internal", "validation_failed", "order-service", "prod")
        return err
    }
    validateSpan.End()
    
    // Process payment
    ctx, paymentSpan := ampyobs.StartSpan(ctx, "process.payment")
    if err := processPayment(ctx, order); err != nil {
        paymentSpan.End()
        ampyobs.OMSReject("stripe", "payment_failed", "order-service", "prod")
        return err
    }
    paymentSpan.End()
    
    // Publish order created event
    attrs := ampyobs.BusAttrs{
        Topic: "orders.created",
        SchemaFQDN: "com.ampy.orders.v1.OrderCreated",
        MessageID: generateID(),
        PartitionKey: order.CustomerID,
        RunID: "order-processing-v1",
    }
    
    ctx, publishSpan := ampyobs.StartBusPublish(ctx, attrs)
    defer publishSpan.End()
    
    headers := make(map[string]string)
    ampyobs.InjectTrace(ctx, headers)
    
    if err := publishOrderCreated(ctx, order, headers); err != nil {
        return err
    }
    
    ampyobs.OMSOrderSubmit("stripe", "success", "order-service", "prod")
    ampyobs.OMSOrderLatency("stripe", float64(time.Since(start).Milliseconds()), "order-service", "prod")
    
    return nil
}
```

**Python Example:**
```python
def process_order(order: Order) -> None:
    with start_span("process.order") as span:
        span.set_attribute("order.id", order.id)
        span.set_attribute("customer.id", order.customer_id)
        span.set_attribute("order.amount", order.amount)
        
        # Validate order
        with start_span("validate.order") as validate_span:
            if not validate_order(order):
                oms_reject("internal", "validation_failed", "order-service", "prod")
                return
            validate_span.end()
        
        # Process payment
        with start_span("process.payment") as payment_span:
            if not process_payment(order):
                oms_reject("stripe", "payment_failed", "order-service", "prod")
                return
            payment_span.end()
        
        # Publish order created event
        attrs = BusAttrs(
            topic="orders.created",
            schema_fqdn="com.ampy.orders.v1.OrderCreated",
            message_id=generate_id(),
            partition_key=order.customer_id,
            run_id="order-processing-v1"
        )
        
        with start_bus_publish(attrs) as publish_span:
            headers = {}
            inject_trace(headers)
            publish_order_created(order, headers)
        
        oms_order_submit("stripe", "success", "order-service", "prod")
        oms_order_latency("stripe", 89.2, "order-service", "prod")
```

### 2. Message Queue Processing

**Go Example:**
```go
func ProcessMessage(ctx context.Context, msg Message, headers map[string]string) error {
    attrs := ampyobs.BusAttrs{
        Topic: msg.Topic,
        SchemaFQDN: msg.Schema,
        MessageID: msg.ID,
        PartitionKey: msg.PartitionKey,
        RunID: msg.RunID,
    }
    
    ctx, span := ampyobs.StartBusConsume(ctx, headers, attrs)
    defer span.End()
    
    ampyobs.Logger.Info("Processing message",
        "message_id", msg.ID,
        "topic", msg.Topic,
        "partition_key", msg.PartitionKey,
    )
    
    start := time.Now()
    
    // Process message
    if err := processMessageContent(ctx, msg); err != nil {
        ampyobs.Logger.Error("Message processing failed",
            "error", err,
            "message_id", msg.ID,
        )
        return err
    }
    
    // Record metrics
    latency := time.Since(start).Milliseconds()
    ampyobs.BusConsumed(msg.Topic, 1, "message-processor", "prod")
    ampyobs.BusDeliveryLatency(msg.Topic, float64(latency), "message-processor", "prod")
    
    ampyobs.Logger.Info("Message processed successfully",
        "message_id", msg.ID,
        "processing_time_ms", latency,
    )
    
    return nil
}
```

**Python Example:**
```python
def process_message(msg: Message, headers: dict) -> None:
    attrs = BusAttrs(
        topic=msg.topic,
        schema_fqdn=msg.schema,
        message_id=msg.id,
        partition_key=msg.partition_key,
        run_id=msg.run_id
    )
    
    with start_bus_consume(headers, attrs) as span:
        logger.info("Processing message", extra={
            "message_id": msg.id,
            "topic": msg.topic,
            "partition_key": msg.partition_key
        })
        
        start_time = time.time()
        
        # Process message
        try:
            process_message_content(msg)
        except Exception as e:
            logger.error("Message processing failed", extra={
                "error": str(e),
                "message_id": msg.id
            })
            raise
        
        # Record metrics
        latency_ms = (time.time() - start_time) * 1000
        bus_consumed(msg.topic, 1, "message-processor", "prod")
        bus_delivery_latency_ms(msg.topic, latency_ms, "message-processor", "prod")
        
        logger.info("Message processed successfully", extra={
            "message_id": msg.id,
            "processing_time_ms": latency_ms
        })
```

### 3. API Request Handling

**Go Example:**
```go
func HandleAPIRequest(w http.ResponseWriter, r *http.Request) {
    ctx, span := ampyobs.StartSpan(r.Context(), "api.handle_request")
    defer span.End()
    
    span.SetAttributes(
        attribute.String("http.method", r.Method),
        attribute.String("http.url", r.URL.String()),
        attribute.String("http.user_agent", r.UserAgent()),
    )
    
    start := time.Now()
    
    // Process request
    result, err := processRequest(ctx, r)
    if err != nil {
        span.RecordError(err)
        ampyobs.Logger.Error("API request failed",
            "error", err,
            "method", r.Method,
            "url", r.URL.String(),
        )
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }
    
    // Record metrics
    latency := time.Since(start).Milliseconds()
    ampyobs.APIRequestLatency(r.Method, r.URL.Path, float64(latency), "api-service", "prod")
    ampyobs.APIRequestCount(r.Method, r.URL.Path, "success", "api-service", "prod")
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}
```

**Python Example:**
```python
from flask import Flask, request, jsonify

app = Flask(__name__)

@app.route('/api/orders', methods=['POST'])
def create_order():
    with start_span("api.handle_request") as span:
        span.set_attribute("http.method", request.method)
        span.set_attribute("http.url", request.url)
        span.set_attribute("http.user_agent", request.headers.get('User-Agent', ''))
        
        start_time = time.time()
        
        try:
            # Process request
            result = process_request(request)
            
            # Record metrics
            latency_ms = (time.time() - start_time) * 1000
            api_request_latency(request.method, request.path, latency_ms, "api-service", "prod")
            api_request_count(request.method, request.path, "success", "api-service", "prod")
            
            return jsonify(result)
            
        except Exception as e:
            span.record_exception(e)
            logger.error("API request failed", extra={
                "error": str(e),
                "method": request.method,
                "url": request.url
            })
            return jsonify({"error": "Internal Server Error"}), 500
```

## Configuration

### Collector Configuration

The OpenTelemetry Collector is configured with production-ready settings:

- **Sampling**: 10% probabilistic sampling to control volume
- **Redaction**: Automatic PII removal (passwords, tokens, secrets)
- **Memory Limits**: 80% memory limit with 25% spike protection
- **Batch Processing**: 512 batch size with 2s timeout

### SLO Monitoring

Built-in Prometheus alert rules monitor:

- **Latency SLOs**: p95 > 250ms for OMS orders, p99 > 150ms for bus delivery
- **Error Rate SLOs**: Rejection rate > 0.2/s for 5 minutes
- **Availability**: Service uptime and health checks

### Environment Variables

```bash
# Collector endpoint
OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317

# Service identification
OTEL_SERVICE_NAME=my-service
OTEL_SERVICE_VERSION=1.0.0
OTEL_DEPLOYMENT_ENVIRONMENT=production

# Sampling
OTEL_TRACES_SAMPLER=probabilistic
OTEL_TRACES_SAMPLER_ARG=0.1
```

## Monitoring and Alerting

### Jaeger (Traces)
- **URL**: http://localhost:16686
- **Features**: Distributed tracing, service dependency graphs, latency analysis
- **Use Cases**: Debug request flows, identify bottlenecks, understand service interactions

### Prometheus (Metrics)
- **URL**: http://localhost:9090
- **Features**: Time-series metrics, alerting, SLO monitoring
- **Use Cases**: Monitor system health, track business metrics, alert on thresholds

### Grafana (Dashboards)
- **URL**: http://localhost:3000
- **Features**: Visualization, dashboards, alerting
- **Use Cases**: System overview, business metrics, operational dashboards

### Alert Rules

The system includes pre-configured alert rules:

```yaml
# OMS order latency p95 > 250ms for 5m
- alert: AmpyOMSHighLatencyP95
  expr: histogram_quantile(0.95, sum by (le) (rate(ampy_oms_order_latency_ms_bucket[5m]))) > 250
  for: 5m
  labels:
    severity: warning

# Bus delivery latency p99 > 150ms for 5m
- alert: AmpyBusHighLatencyP99
  expr: histogram_quantile(0.99, sum by (le) (rate(ampy_bus_delivery_latency_ms_bucket[5m]))) > 150
  for: 5m
  labels:
    severity: warning

# OMS rejections spike
- alert: AmpyOMSRejectionsSpike
  expr: sum by (reason) (rate(ampy_oms_rejections_total[5m])) > 0.2
  for: 5m
  labels:
    severity: page
```

## Development

### Running Tests

```bash
# Go tests
cd go
go test -race ./...

# Python tests
cd python
pytest

# Integration tests
cd deploy
docker compose up -d
# Run examples and verify telemetry
```

### CI/CD

The project includes a comprehensive CI workflow that:

- Builds and tests Go and Python SDKs
- Runs integration tests with the full observability stack
- Validates telemetry flow from SDKs to collector to backends
- Checks for code formatting and linting issues

### Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Ensure CI passes
6. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For questions, issues, or contributions:

- **Issues**: GitHub Issues
- **Discussions**: GitHub Discussions
- **Documentation**: This README and inline code documentation

## Roadmap

- [ ] Additional language SDKs (Java, .NET, Node.js)
- [ ] Advanced sampling strategies
- [ ] Custom dashboard templates
- [ ] Kubernetes deployment manifests
- [ ] Performance benchmarking suite
- [ ] Security audit and compliance
