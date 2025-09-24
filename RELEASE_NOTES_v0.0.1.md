# Ampy Observability SDK v0.0.1 Release Notes

**Release Date**: September 15, 2024  
**Version**: 0.0.1  
**Type**: Initial Release

## 🎉 What's New

This is the initial release of the Ampy Observability SDK, providing comprehensive observability capabilities for Go and Python applications. Built on OpenTelemetry with production-ready configurations and monitoring.

## ✨ Key Features

### 🔧 Multi-Language Support
- **Go SDK**: Complete observability SDK with HTTP/gRPC exporters
- **Python SDK**: Full-featured Python SDK with context managers and proper tracing
- **Unified API**: Consistent logging, metrics, and tracing across both languages

### 📊 Observability Stack
- **Distributed Tracing**: W3C trace context propagation across services
- **Structured Logging**: JSON logging with automatic trace context enrichment
- **Domain Metrics**: Bus and OMS specific counters and histograms
- **Production-Ready Collector**: Hardened with sampling, redaction, and memory limits

### 🚀 Production Features
- **SLO Monitoring**: Built-in Prometheus alert rules for latency and error rates
- **CI/CD Pipeline**: Comprehensive testing with integration smoke tests
- **Docker Compose**: One-command stack deployment
- **GitHub Actions**: Automated testing and publishing

## 📦 Components Included

### Observability Stack
- **Jaeger** (Traces) - Distributed tracing UI
- **Prometheus** (Metrics) - Metrics collection and alerting
- **Grafana** (Dashboards) - Visualization and monitoring
- **Loki** (Logs) - Log aggregation
- **OpenTelemetry Collector** - Telemetry data processing

### SDKs
- **Go SDK**: `github.com/ampyfin/ampy-observability/go/ampyobs`
- **Python SDK**: `ampyobs` package (available on PyPI)

## 🛠 Installation

### Go SDK
```bash
go get github.com/ampyfin/ampy-observability/go/ampyobs
```

### Python SDK
```bash
pip install ampyobs
```

### Observability Stack
```bash
cd deploy
docker compose up -d
```

## 📋 Quick Start

### Go Example
```go
import "github.com/ampyfin/ampy-observability/go/ampyobs"

// Initialize
cfg := ampyobs.Config{
    ServiceName: "my-service",
    CollectorEndpoint: "localhost:4317",
}
ampyobs.Init(ctx, cfg)

// Logging with trace context
ampyobs.Logger.Info("Processing order", "order_id", "12345")

// Metrics
ampyobs.BusProduced("orders", 1, "my-service", "production")

// Tracing
ctx, span := ampyobs.StartSpan(ctx, "process.payment")
defer span.End()
```

### Python Example
```python
from ampyobs import Config, init
from ampyobs.logger import L
from ampyobs.metrics import bus_produced

# Initialize
init(Config(service_name="my-service", collector_endpoint="localhost:4317"))

# Logging with trace context
L.info("Processing order", extra={"order_id": "12345"})

# Metrics
bus_produced("orders", 1, service="my-service", env="production")
```

## 🎯 Domain-Specific Metrics

### Bus Metrics
- `ampy_bus_produced_total` - Messages produced to topics
- `ampy_bus_delivery_latency_seconds` - Message delivery latency
- `ampy_bus_consumed_total` - Messages consumed from topics

### OMS Metrics
- `ampy_oms_order_submit_total` - Order submission attempts
- `ampy_oms_order_latency_seconds` - Order processing latency
- `ampy_oms_reject_total` - Order rejections by reason

## 🔧 Configuration

### Collector Features
- **Sampling**: 10% trace sampling for production
- **PII Redaction**: Automatic sensitive data removal
- **Memory Limits**: Bounded memory usage
- **Batch Processing**: Efficient telemetry batching

### Supported Protocols
- **gRPC**: Primary protocol for traces and metrics
- **HTTP**: Alternative protocol support
- **OTLP**: OpenTelemetry Protocol for maximum compatibility

## 🧪 Testing & Quality

### CI/CD Pipeline
- **Go Testing**: Unit tests with race detection
- **Python Testing**: Comprehensive test suite
- **Integration Tests**: End-to-end observability stack validation
- **Automated Publishing**: PyPI package publishing on release

### Code Quality
- **Formatting**: Automated code formatting (gofmt, black)
- **Linting**: Static analysis and vetting
- **Dependency Management**: Automated dependency updates

## 📚 Documentation

- **README**: Comprehensive setup and usage guide
- **Examples**: Working code samples for both Go and Python
- **API Documentation**: Complete SDK reference
- **Architecture Diagrams**: System design and data flow

## 🔗 Links

- **GitHub Repository**: https://github.com/AmpyFin/ampy-observability
- **Go SDK**: `github.com/ampyfin/ampy-observability/go/ampyobs`
- **Python Package**: https://pypi.org/project/ampyobs/
- **Documentation**: See README.md for complete setup guide

## 🏗 Architecture

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

## 🚀 Getting Started

1. **Start the observability stack**:
   ```bash
   cd deploy && docker compose up -d
   ```

2. **Install the SDK** (Go or Python):
   ```bash
   # Go
   go get github.com/ampyfin/ampy-observability/go/ampyobs
   
   # Python
   pip install ampyobs
   ```

3. **Run the examples**:
   ```bash
   # Go examples
   cd go/examples && go run .
   
   # Python examples
   python python/examples/producer.py
   ```

4. **View telemetry**:
   - Jaeger UI: http://localhost:16686
   - Prometheus: http://localhost:9090
   - Grafana: http://localhost:3000

## 📝 Breaking Changes

This is the initial release, so there are no breaking changes.

## 🐛 Known Issues

None at this time.

## 🔮 What's Next

Future releases will include:
- Additional metric types and dimensions
- Enhanced dashboard templates
- Performance optimizations
- Extended language support

---

**Full Changelog**: This is the initial release. See the repository for complete development history.

**Contributors**: AmpyFin Team

**Support**: For issues and questions, please open an issue on GitHub.
