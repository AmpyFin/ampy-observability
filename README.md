# ampy-observability — Logging • Metrics • Tracing SDKs (Open Source)

> **Purpose:** Deliver **uniform, low‑friction observability** across AmpyFin’s Go and Python services:
> - **Logs:** structured JSON with consistent fields (e.g., `run_id`, `universe_id`, `as_of`), correlated to traces and ampy‑bus envelopes.
> - **Metrics:** standard counters/gauges/histograms via Prometheus & OpenTelemetry.
> - **Tracing:** distributed tracing with W3C context propagation (`traceparent`) across `ampy-bus` hops.
>
> **Artifacts:** Language SDKs (Go, Python) + a **reference docker‑compose observability stack** (collector, metrics, tracing, logs) for local/dev and CI smoke tests.
>
> This README is **LLM‑ready**: it specifies *what to build and how it should behave*—no code or repo layout. Shapes/fields and expected behavior are defined for deterministic implementation.

---

## 1) Mission & Success Criteria

### Mission
Create a **single, consistent observability layer** so every AmpyFin module (ingestion, ML, OMS, broker connectors, backtester) emits **correlated logs, metrics, and traces** with minimal app code, enabling:
- Faster debugging & incident response
- Deterministic audits of trading sessions
- Capacity planning and SLO management
- Transparent OSS adoption paths

### Success looks like
- Every log line is **JSON**, redacted, includes a minimal **context core** (service, env, run_id, trace/span IDs) and **domain context** (symbol, mic, client_order_id) when applicable.
- All services export **OTLP** (OpenTelemetry) and Prometheus‑compatible metrics with standardized names and **bounded label contracts**.
- Traces stitch across processes and **ampy‑bus** boundaries (publish→route→consume) with stable sampling and low overhead.
- Dashboards & alerts are ready‑to‑use; **p95/p99** latencies and error budgets are visible per domain and per service.
- Observability overhead stays within budget (**<3% CPU**, **<50MB** RAM per service typical; tunable).

---

## 2) Problems This Solves

- **Inconsistent logs**: Different structures → slow debugging and broken correlation.
- **Missing context**: No `trace_id`/`run_id` makes incident RCA hard.
- **Metric cardinality explosions** from ad‑hoc labels increase costs and reduce fidelity.
- **Fragmented tracing**: No end‑to‑end view across `ampy-bus` topics and services.
- **Vendor lock‑in risk**: Non‑standard APIs hinder portability; OTel keeps us neutral.

---

## 3) Scope (What `ampy-observability` Covers)

- **SDKs (Go, Python)** for: structured logs, metrics, and tracing with consistent defaults & context propagation.
- **Correlation rules** tying **logs ⇄ metrics ⇄ traces ⇄ ampy‑bus envelopes** (message_id, schema_fqdn, topic, partition_key).
- **Resource attributes** (service metadata) and **semantic conventions** for domain events (bars, ticks, signals, orders, fills, positions, news, fx).
- **Sampling & performance guidance**, backpressure/error handling for exporters.
- **Reference docker‑compose stack** for local/dev: OTel Collector (gateway), Prometheus, tracing backend, log aggregator, Grafana.

**Non‑goals:** No app business logic; no repository layout here; no vendor‑specific lock‑in.

---

## 4) Global Conventions

### 4.1 Resource Attributes (attach to every telemetry signal)
- `service.name` — logical service (e.g., `ampy-oms`, `broker-alpaca`, `yfinance-go`, `ampy-model-server`)
- `service.version` — app/version (e.g., git SHA or semver)
- `deployment.environment` — `dev` | `paper` | `prod`
- `service.instance.id` — runtime instance identity
- `cloud.region` — e.g., `us-east-1`
- `ampy.bus.cluster` — logical bus plane name
- `ampy.schema_version` — deployed `ampy-proto` bundle version

### 4.2 Context Core (present in **every log**; propagated to spans & metrics as labels only where appropriate)
- `trace_id`, `span_id` (W3C)
- `run_id` (pipeline run/session)
- `as_of` (logical time for domain state)
- `universe_id` (if applicable)
- `message_id` (if associated with a bus envelope)
- `client_order_id` (if order‑related)
- `symbol`, `mic` (if instrument‑related)

### 4.3 PII & Secrets
- **PII forbidden**. **Secrets redacted** in logs & attributes. Emit fingerprints/hashes only if required and non‑reversible.

### 4.4 Time & Units
- All timestamps **UTC** ISO‑8601 with **ns** precision where available.
- Metrics use **base units** (ms for durations, bytes for sizes, bp for rates when applicable).

---

## 5) Logs (JSON) — Shapes & Fields

**Goal:** human‑searchable and machine‑joinable (by IDs); consistent across languages.

**Levels:** `debug` | `info` | `warn` | `error` | `fatal`

**Base fields (always present):**
```json
{
  "ts": "2025-09-05T19:31:05.123456Z",
  "level": "info",
  "service": "ampy-oms",
  "env": "prod",
  "service_version": "2.3.1@abc123",
  "trace_id": "4b5b3f2a0f9d4e3db4c8a1f0e3a7c812",
  "span_id": "b6a7f1c9d2e34a5b",
  "run_id": "live_trading_44",
  "as_of": "2025-09-05T19:31:05Z",
  "message": "Placed limit order",
  "event": "order.submit"
}
```

**Domain context (attach when relevant):**
```json
{
  "client_order_id": "co_20250905_001",
  "account_id": "ALPACA-LIVE-01",
  "symbol": "AAPL",
  "mic": "XNAS",
  "limit_price": {"scaled": 1919900, "scale": 4},
  "quantity": {"scaled": 1000000, "scale": 2},
  "tif": "DAY"
}
```

**Bus correlation (when tied to a message):**
```json
{
  "ampy_bus": {
    "topic": "ampy/prod/orders/v1/requests",
    "schema_fqdn": "ampy.orders.v1.OrderRequest",
    "message_id": "018f5e32-9b2a-7cde-9333-4f1ab2a49e77",
    "partition_key": "co_20250905_001"
  }
}
```

**Error example (with redaction):**
```json
{
  "ts": "2025-09-05T19:31:06.010Z",
  "level": "error",
  "service": "broker-alpaca",
  "env": "prod",
  "trace_id": "a0c1b2d3e4f5061728394a5b6c7d8e9f",
  "run_id": "live_trading_44",
  "message": "Order rejected by broker",
  "event": "order.reject",
  "client_order_id": "co_20250905_001",
  "reject_reason": "risk_check",
  "http_status": 429,
  "retry_in": "250ms",
  "secret_ref": "aws-sm://ALPACA_SECRET#value",
  "broker_payload_redacted": "***"
}
```

**Cardinality rules for logs**
- Keep **field sets stable**; avoid arbitrary maps with unbounded keys.
- Use **domain enums** for `event` (e.g., `bars.ingest`, `ticks.trade`, `signals.emit`, `orders.submit`, `fills.update`, `positions.snapshot`).

---

## 6) Metrics — Names, Labels, Buckets

**Backends:** Apps export OTLP metrics; Prometheus scrapes via collector or native endpoints.

**Naming pattern:** `ampy.<subsystem>.<metric>` (dots inside) **or** OTel semantic names—be consistent.

**Global labels (low‑cardinality):**
- `service`, `env`, `region`, `version`
- `source`, `producer` (where applicable)
- `broker` (for execution path)
- `domain` (bars|ticks|signals|orders|fills|positions|news|fx)
- `outcome` (ok|retry|dlq|reject), `reason` (enum)

**Histograms & buckets (guidance):**
- **Latency** (ms): `[1, 2, 5, 10, 20, 50, 100, 200, 500, 1000, 2000]`
- **Batch size** (bytes): `[1024, 4096, 16384, 65536, 262144, 1048576, 4194304]`
- **Records per batch**: `[1, 5, 10, 20, 50, 100, 250, 500, 1000]`

**Core metric set (illustrative):**
```
ampy.bus.produced_total{topic,service,env}
ampy.bus.consumed_total{topic,service,env}
ampy.bus.delivery_latency_ms{topic}            # histogram
ampy.bus.consumer_lag{topic,consumer}          # gauge
ampy.bus.dlq_total{topic,reason}               # counter

ampy.oms.order_submit_total{broker,env,outcome}
ampy.oms.order_latency_ms{broker}              # histogram (submit→ack)
ampy.oms.rejections_total{broker,reason}

ampy.ingest.throughput_msgs_per_s{source,domain}
ampy.ingest.decode_fail_total{source,reason}

ampy.ml.inference_latency_ms{model_id}         # histogram
ampy.ml.signals_emitted_total{model_id,domain}
ampy.ml.errors_total{model_id,reason}

ampy.fx.rate_staleness_ms{pair}                # gauge
ampy.fx.provider_fail_total{provider,reason}

ampy.runtime.gc_pause_ms{service}              # histogram
ampy.runtime.mem_bytes{service}                # gauge
```

**Cardinality guardrails**
- Labels **must** be bounded enumerations. Do **not** use `client_order_id` or `symbol` as labels; put them in logs or trace attributes.

---

## 7) Tracing — Spans, Links, Sampling

**Goal:** end‑to‑end visibility (ingestion → features → models → signals → OMS → broker → fills → positions), stitched across **ampy‑bus**.

**Propagation:** W3C `traceparent` / `tracestate` in `ampy-bus` headers. Consumers **continue** the trace; fan‑out uses **span links**.

**Key spans (names & attributes):**
- `ingest.fetch` — attrs: `source`, `dataset`, `symbol`, `mic`, `count`
- `ingest.publish` — attrs: `topic`, `message_id`, `schema_fqdn`, `records`
- `features.compute` — attrs: `window`, `universe_id`, `as_of`
- `model.infer` — attrs: `model_id`, `batch`, `latency_ms`, `horizon`
- `signals.emit` — attrs: `model_id`, `symbol`, `score`
- `oms.submit` — attrs: `broker`, `client_order_id`, `side`, `notional_usd`
- `broker.ack` — attrs: `broker_order_id`, `status`
- `fills.update` — attrs: `fill_qty`, `price`, `account_id`
- `positions.snapshot` — attrs: `account_id`, `symbol`, `pnl_unrealized`
- `bus.route` — attrs: `topic`, `partition_key`, `lag_ms`

**Span link example (fan‑out):**
```
Parent: model.infer (trace_id=...)
Child:  signals.emit (continues trace)
Bus publish span: bus.publish (links to signals.emit)
Downstream consumer: oms.submit (continues trace from bus headers)
```

**Sampling policy:**
- Default **parent‑based, probabilistic** (e.g., 0.25).
- **Always‑sample** for safety‑critical events (order rejects, DLQ).
- **Tail‑sampling** (in collector) for high‑latency or error outliers.

**Performance budget:**
- SDK overhead: **<5 µs** per log/metric call typical; spans batched.
- Exporters backpressure; **drop** (with counters) when unsafe; **never block** OMS hot paths.

---

## 8) Reference docker‑compose Stack (Conceptual)

> No YAML here—this defines the **components and contracts** your compose file must satisfy.

**Components**
- **OpenTelemetry Collector (Gateway)**: Receives OTLP (gRPC/HTTP); batching, retries, attributes/redaction, tail‑sampling; fans out to backends.
- **Prometheus**: Scrapes OTel metrics endpoint or receives via collector.
- **Tracing backend**: Jaeger or Tempo (OTLP ingestion).
- **Log aggregator**: Loki or Elasticsearch‑equivalent; JSON log ingestion via collector or sidecar.
- **Grafana**: Unified dashboards over metrics/traces/logs.

**Collector pipelines (logical)**
- **receivers**: `otlp` (metrics/traces/logs), `prometheus` (optional), `loki` (optional)
- **processors**: `batch`, `memory_limiter`, `attributes` (resource injection), `redaction`, `tail_sampling`
- **exporters**: `otlphttp`/`otlpgprc` to tracing backend, `prometheusremotewrite` to TSDB, `loki` for logs

**Dev defaults**
- Single‑host footprint; minimal retention (logs 1–3 days, traces 1–3 days, metrics 7–14 days).
- Example dashboards & alert rules preloaded (see §9).

**Ports (suggested)**
- Collector OTLP gRPC: `4317`
- Collector OTLP HTTP: `4318`
- Prometheus UI: `9090`
- Grafana UI: `3000`
- Tracing backend UI (Jaeger): `16686`
- Loki API: `3100`

---

## 9) Dashboards & Alerts (Examples)

**Dashboards**
- **Trading Path (prod)**: panels for `oms.order_latency_ms p95/p99`, `rejections_total by reason`, `bus.delivery_latency_ms by topic`, `fills vs orders time‑series`, `positions snapshot lag`, `model inference latency`, `news ingest freshness`.
- **Ingestion Health**: throughput, decode failures, DLQ counts, consumer lag by topic.
- **Resource**: GC pauses, memory, CPU, file descriptors.

**Alerts (semantic, not code)**
- `oms.order_rejects` rate > threshold for 5m (labels: broker=alpaca, reason=risk_check).
- `bus.delivery_latency_ms` p99 > 150ms for bars or > 50ms for orders.
- `bus.dlq_total` increases over baseline (per topic).
- `ml.inference_latency_ms` p95 > SLO for 10m.
- `fx.rate_staleness_ms` > allowed max for 30s (pair‑scoped).
- **Watchdog**: no `positions.snapshot` for N minutes during trading hours.

---

## 10) Security & Compliance

- **TLS everywhere**; mTLS between apps and collector.
- **ACLs** on collector endpoints; **rate limits** to prevent abuse.
- **Redaction** processor removes secrets/PII from attributes/logs.
- **Retention policy**: logs (prod) 7 days, traces 7–14 days, metrics 30–90 days (tunable; align with compliance).
- **Multitenancy** by `deployment.environment` & `service.name`; no cross‑env reads by default.

---

## 11) Interop with Ampy Stack

- **ampy-proto**: Decimal/money & identity fields (symbol, mic) appear in logs as structured objects; do **not** stringify precision away.
- **ampy-bus**: Envelope headers (`message_id`, `schema_fqdn`, `topic`, `partition_key`, `run_id`, `traceparent`) are mirrored into logs/trace attrs.
- **ampy-config**: Emits `config.*` metrics; config changes/secret rotations appear as structured events and traces; redaction enforced.

---

## 12) Validation & Testing (What “Good” Looks Like)

- **Golden telemetry**: Sample logs/metrics/traces for each domain path (bars→signals→orders→fills→positions).
- **Correlation tests**: Given a `client_order_id`, pivot from a log line → trace → metrics panels.
- **Sampling tests**: Verify parent‑based sampling keeps causality; tail‑sampling catches high‑latency/error outliers.
- **Cardinality tests**: Ensure labels remain bounded under stress (no symbol/client_order_id in labels).
- **Backpressure tests**: Simulate exporter outage; confirm SDK drops with counters and does not block hot paths.
- **Security tests**: Inject secret values; verify redaction in logs/attrs.

---

## 13) Acceptance Criteria (Definition of Done for `ampy-observability` v1)

- [ ] **Go & Python SDKs** expose uniform logging, metrics, tracing with the **resource & context core** and domain conventions in this doc. 
- [ ] **OTLP exporters** work in dev (compose) and staging; TLS/mTLS options present.
- [ ] **Docker‑compose stack** runs locally with collector, metrics, tracing, logs, and Grafana dashboards.
- [ ] **Correlation** across logs/metrics/traces works end‑to‑end for at least: bars→signals→orders→fills→positions.
- [ ] **Dashboards & alerts** are defined (semantic content) and load cleanly; p95/p99 SLOs visible.
- [ ] **Performance budgets** met; backpressure & redaction verified; no PII or secrets in sinks.
- [ ] **Golden telemetry** captured for regression; CI smoke test ensures basic pipelines ingest.

---

## 14) End‑to‑End Narrative (Example)

1) `yfinance-go` ingests AAPL 1m bars. It logs `bars.ingest` with `symbol=XNAS.AAPL`, emits `ampy.bus.produced_total`, and creates a `bus.publish` span (linking to downstream).  
2) `ampy-features` consumes bars, logs `features.compute`, records computation latency histogram, and creates a `model.infer` span with `model_id=hyper@…`.  
3) `ampy-model-server` logs `signals.emit` and publishes to bus with `traceparent`.  
4) `ampy-oms` consumes signals, logs `order.submit` (JSON), emits metrics for order counts/latency, and creates an `oms.submit` span.  
5) `broker-alpaca` logs `order.reject` or `broker.ack`; fills update positions.  
6) Grafana shows p95 OMS latency degrading; alert fires; traces pinpoint bus delivery lag → collector shows exporter throttling → rollback applied; SLO returns to green.

---

## 15) Architecture (Conceptual)

```
+-------------------+     OTLP gRPC/HTTP     +--------------------+
|  AmpyFin Services |  --------------------> |  OTel Collector     |
|  (Go/Python SDKs) |                        |  (gateway/pipelines)|
+---------+---------+                        +----------+---------+
          |                                           / | \
          | Logs (JSON) via OTLP or sidecar          /  |  \
          v                                         v   v   v
   +------+--------+                        +--------+ +------+ +---------+
   |  Log Backend  |  <----- Loki/OTLP ---- | Traces | |  TSDB| | Grafana |
   | (Loki/ES‑like)|                        | (Jaeger| |(Prom)| |  Dash   |
   +---------------+                        |  Tempo) | +------+ +---------+
                                            +--------+
```

**Key contracts**
- OTel **resource attributes** (§4.1) are mandatory on all signals.
- **Context core** (§4.2) appears on logs; associated subset appears on traces/metrics where appropriate.
- **Backpressure & drop** on exporters is counted and observable.
- **Redaction** is enforced in collector and SDKs.

---

## 16) Configuration Model

Configuration is surfaced via **env vars** and/or **`ampy-config`** keys. Env overrides take precedence.

**Environment variables (suggested):**

| Variable | Meaning | Default |
| --- | --- | --- |
| `OBS_SERVICE_NAME` | Overrides `service.name` | inferred |
| `OBS_SERVICE_VERSION` | Overrides `service.version` | inferred |
| `OBS_ENV` | `dev|paper|prod` | `dev` |
| `OBS_REGION` | Cloud/region label | empty |
| `OBS_OTLP_ENDPOINT` | `https://collector:4318` or `grpc://collector:4317` | `http://collector:4318` |
| `OBS_OTLP_INSECURE` | `true/false` | `true` (dev) |
| `OBS_SAMPLER` | `parent,traceidratio,always_on,always_off` | `parent` |
| `OBS_SAMPLER_RATIO` | `0.0..1.0` when ratio sampler | `0.25` |
| `OBS_EXPORT_TRACES` | `true/false` | `true` |
| `OBS_EXPORT_METRICS` | `true/false` | `true` |
| `OBS_EXPORT_LOGS` | `true/false` | `true` |
| `OBS_PROM_PORT` | If exposing native Prom endpoint | empty |
| `OBS_HEADERS` | `key1=val1,key2=val2` (for vendor APIs) | empty |
| `OBS_TLS_CA_FILE` | Path to CA for mTLS | empty |
| `OBS_TLS_CERT_FILE` | Client cert for mTLS | empty |
| `OBS_TLS_KEY_FILE` | Client key for mTLS | empty |
| `OBS_REDACTION_RULES` | Comma‑sep patterns for SDK redactor | sensible defaults |

**`ampy-config` keys (mirror):**
```
observability:
  otlp:
    endpoint: http://collector:4318
    insecure: true
    headers: {}
    tls:
      ca_file: null
      cert_file: null
      key_file: null
  sampler:
    type: parent
    ratio: 0.25
  export:
    traces: true
    metrics: true
    logs: true
  prometheus:
    port: null
  redaction:
    patterns:
      - "secret"
      - "password"
      - "token"
```

> **API keys?** None required for the default stack. If you opt into a vendor (Grafana Cloud, New Relic, Datadog, etc.), you’ll need their API token/headers and to set `OBS_HEADERS` accordingly.

---

## 17) Developer Usage (Examples without code)

**Instrument a new service**
1. Initialize the SDK with `service.name`, `service.version`, `env`.  
2. Emit structured **JSON logs** using the SDK’s logger; include `event` enums and **context core** attributes.  
3. Create **metrics** using predefined helpers; prefer bounded labels (`domain`, `outcome`, `reason`).  
4. Wrap operations in **spans** with semantic names (`ingest.fetch`, `oms.submit`, etc.), adding attributes from §7.  
5. Publish to/consume from **ampy‑bus** with **W3C trace context** in headers; copy envelope IDs into logs.  
6. Configure **OTLP endpoint** and **sampling** via env or `ampy-config`.  
7. Validate locally with the docker‑compose stack; check Grafana dashboards and Jaeger/Tempo traces.

**Operational playbook**
- Collector down? SDK should buffer, then **drop with counters**; alerts fire on `exporter_fail_total`.
- Cardinality spike? Reduce labels, adjust sampling, or tail‑sample outliers.
- Secret leak suspected? Enable aggressive **redaction** rules; verify via security tests (§12).

---

## 18) Versioning, Compatibility, and Packaging

- **SemVer** for the SDKs and compose stack manifests.
- **Language baselines**: Go ≥ 1.22; Python ≥ 3.10.
- **OTel compatibility**: Keep SDKs aligned with a pinned OTel version set per release tag.
- **Changelog** includes migration notes for metrics name/label changes and span attribute updates.

---

## 19) Roadmap

- **v1.1**: Runtime config reload; exemplars for high‑value spans; log‑to‑trace correlation UI links; auto‑dashboards per service.  
- **v1.2**: eBPF‑based runtime metrics (optional); structured log sampling; per‑topic lag SLOs.  
- **v1.3**: Vendor exporter modules (Grafana Cloud, Datadog, NR), canary sampling, adaptive sampling based on error rates.

---

## 20) Contributing

- Open issues with **use‑case**, **expected signal shapes**, and **cardinality impact** analysis.
- PRs must update **golden telemetry fixtures** and **docs** (dashboards/alerts) when signal shapes change.
- Run local compose stack and attach screenshots of dashboards/traces validating changes.

---

## 21) License

Open‑source under **Apache‑2.0** (same as other AmpyFin OSS modules).

---

*AmpyFin’s observability becomes cohesive, correlated, and actionable—unlocking quick root‑cause analysis and confident operations.*
