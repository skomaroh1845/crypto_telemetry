# Observability Stack Architecture

## System Overview

This document describes the architecture of the observability stack for the crypto telemetry system.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Application Layer                        │
│                                                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       │
│  │ Data Service │  │Decision Svc  │  │ Mock Service │       │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘       │
└─────────┼──────────────────┼──────────────────┼──────────────┘
          │                  │                  │
          │ OTLP HTTP/gRPC   │                  │
          │ (4317/4318)      │                  │
          └──────────────────┴──────────────────┘
                             │
┌────────────────────────────┼─────────────────────────────────┐
│                  Telemetry Collection Layer                   │
│                            │                                  │
│                  ┌─────────▼──────────┐                       │
│                  │  OTEL Collector    │                       │
│                  │  ┌──────────────┐  │                       │
│                  │  │  Receivers   │  │ ← OTLP gRPC (4317)   │
│                  │  │  (OTLP)      │  │ ← OTLP HTTP (4318)   │
│                  │  └──────┬───────┘  │                       │
│                  │         │          │                       │
│                  │  ┌──────▼───────┐  │                       │
│                  │  │ Processors   │  │                       │
│                  │  │ • Batch      │  │                       │
│                  │  │ • Memory Lim │  │                       │
│                  │  │ • Resource   │  │                       │
│                  │  └──────┬───────┘  │                       │
│                  │         │          │                       │
│                  │  ┌──────▼───────┐  │                       │
│                  │  │  Exporters   │  │                       │
│                  │  │ • Prometheus │  │ → :8889              │
│                  │  │ • OTLP/Tempo │  │ → tempo:4317         │
│                  │  │ • Loki       │  │ → loki:3100          │
│                  │  └──────────────┘  │                       │
│                  └────────────────────┘                       │
│                  Metrics: :8888/metrics                       │
└───────────────────────────┬──────────────────────────────────┘
                            │
                            │ Processed Data
                            │
         ┌──────────────────┼──────────────────┐
         │                  │                  │
         ▼                  ▼                  ▼
┌────────────────┐  ┌────────────────┐  ┌────────────────┐
│   Prometheus   │  │     Tempo      │  │      Loki      │
│                │  │                │  │                │
│  Time-Series   │  │  Distributed   │  │  Log           │
│  Database      │  │  Tracing       │  │  Aggregation   │
│                │  │                │  │                │
│  • Metrics     │  │  • Traces      │  │  • Logs        │
│  • 2h retain   │  │  • 2h retain   │  │  • 2h retain   │
│  • 15s scrape  │  │  • OTLP        │  │  • TSDB        │
│                │  │  • Service Map │  │  • Compaction  │
└────────┬───────┘  └────────┬───────┘  └────────┬───────┘
         │                   │                   │
         │      Port 9090    │    Port 3200      │  Port 3100
         │                   │                   │
         └───────────────────┴───────────────────┘
                             │
                             │ Query APIs
                             │
┌────────────────────────────▼─────────────────────────────────┐
│                   Visualization Layer                         │
│                                                               │
│                    ┌──────────────┐                           │
│                    │   Grafana    │                           │
│                    │              │                           │
│                    │  Port 3000   │                           │
│                    │              │                           │
│  ┌─────────────────┴──────────────┴────────────────┐         │
│  │         Pre-configured Datasources               │         │
│  │  • Prometheus (default)                          │         │
│  │  • Tempo (with trace-to-logs correlation)        │         │
│  │  • Loki (with trace correlation)                 │         │
│  └──────────────────────────────────────────────────┘         │
│                                                               │
│  Features:                                                    │
│  • Unified dashboards                                         │
│  • Trace → Logs → Metrics correlation                         │
│  • Service maps and node graphs                               │
│  • TraceQL editor                                             │
└───────────────────────────────────────────────────────────────┘
```

## Component Details

### 1. Application Layer

**Services:**
- Data Service (existing)
- Decision Service (planned)
- Mock Telemetry Service (testing)

**Telemetry:**
- Instrumented with OpenTelemetry SDK
- Sends OTLP over HTTP (port 4318) or gRPC (port 4317)
- Generates traces, metrics, and logs

### 2. OTEL Collector

**Role:** Central telemetry hub

**Receivers:**
- `otlp/grpc` - Port 4317
- `otlp/http` - Port 4318

**Processors:**
- `batch` - Batches telemetry data (10s timeout, 1024 batch size)
- `memory_limiter` - Prevents OOM (512 MiB limit)
- `resource` - Enriches resource attributes

**Exporters:**
- `prometheus` - Exposes metrics on port 8889 for Prometheus to scrape
- `otlp/tempo` - Forwards traces to Tempo via OTLP
- `loki` - Pushes logs to Loki HTTP endpoint
- `logging` - Debug output to console

**Internal Telemetry:**
- Port 8888 - Prometheus metrics about collector itself
- Detailed metrics and logging enabled

### 3. Prometheus

**Role:** Time-series metrics database

**Configuration:**
- Scrape interval: 15 seconds
- Retention: 2 hours
- Storage: TSDB with named volume

**Scrape Targets:**
- `otel-collector:8889` - Application metrics from OTEL Collector
- `otel-collector:8888` - OTEL Collector internal metrics
- `tempo:3200` - Tempo internal metrics
- `loki:3100` - Loki internal metrics
- `prometheus:9090` - Self-monitoring

**Features:**
- PromQL query language
- Built-in alerting (Alertmanager integration ready)
- Web UI on port 9090

### 4. Grafana Tempo

**Role:** Distributed tracing backend

**Configuration:**
- Protocol: OTLP (gRPC + HTTP)
- Storage: Local filesystem
- Retention: 2 hours
- Block duration: 5 minutes

**Features:**
- TraceQL query language
- Service graphs (topology visualization)
- Span metrics generation
- Metrics generator → Prometheus integration
- Node graph support

**Ports:**
- 3200 - HTTP API
- 4317 - OTLP gRPC (internal)
- 4318 - OTLP HTTP (internal)

### 5. Grafana Loki

**Role:** Log aggregation system

**Configuration:**
- Schema: TSDB v13
- Storage: Local filesystem
- Retention: 2 hours
- Compaction: Every 10 minutes

**Features:**
- LogQL query language
- Label-based indexing
- Integration with Promtail (optional)
- Multi-tenancy support (disabled for simplicity)

**Ports:**
- 3100 - HTTP API for push/query

### 6. Grafana

**Role:** Unified visualization and dashboarding

**Pre-configured Datasources:**

1. **Prometheus** (default)
   - URL: http://prometheus:9090
   - 15s refresh interval
   - Supports metric queries and alerting

2. **Tempo**
   - URL: http://tempo:3200
   - Trace-to-logs integration
   - Trace-to-metrics integration
   - Service map enabled
   - Node graph enabled

3. **Loki**
   - URL: http://loki:3100
   - Max 1000 lines per query
   - Derived fields for trace correlation

**Features:**
- Explore view for ad-hoc queries
- Dashboard creation and sharing
- TraceQL editor (feature flag enabled)
- Correlations between all data types
- User management (admin/admin)

**Port:** 3000

## Data Flow Sequences

### Trace Flow

```
Application
    │
    │ 1. Generate span
    │    with context
    │
    ▼
OTEL SDK
    │
    │ 2. Export via OTLP
    │    (HTTP/gRPC)
    │
    ▼
OTEL Collector
    │
    │ 3. Batch process
    │    Add attributes
    │
    ▼
Tempo (OTLP)
    │
    │ 4. Store trace
    │    Index by trace_id
    │
    ▼
Grafana
    │
    └─ 5. Query with TraceQL
       Visualize service map
       Jump to logs
```

### Metrics Flow

```
Application
    │
    │ 1. Record metric
    │    (counter/gauge/histogram)
    │
    ▼
OTEL SDK
    │
    │ 2. Aggregate & export
    │    via OTLP
    │
    ▼
OTEL Collector
    │
    │ 3. Batch process
    │    Expose as Prometheus
    │    endpoint (:8889)
    │
    ▼
Prometheus
    │
    │ 4. Scrape endpoint
    │    Store in TSDB
    │
    ▼
Grafana
    │
    └─ 5. Query with PromQL
       Create dashboards
       Set up alerts
```

### Logs Flow

```
Application
    │
    │ 1. Emit log with
    │    structured fields
    │
    ▼
OTEL SDK
    │
    │ 2. Export via OTLP
    │    Include trace context
    │
    ▼
OTEL Collector
    │
    │ 3. Batch process
    │    Push to Loki
    │
    ▼
Loki
    │
    │ 4. Index by labels
    │    Store log lines
    │
    ▼
Grafana
    │
    └─ 5. Query with LogQL
       Filter and search
       Jump to traces
```

## Network Architecture

All services run in the `crypto_telemetry_network` Docker bridge network.

**Service Discovery:**
- Services reference each other by container name
- Docker's internal DNS resolves names to IPs
- No external service discovery needed

**Port Mapping:**
- Only UI/API ports are exposed to host
- Internal ports remain within Docker network
- Security through network isolation

## Storage and Persistence

**Volumes:**
- `prometheus_data` - Prometheus TSDB
- `tempo_data` - Tempo blocks and WAL
- `loki_data` - Loki chunks and index
- `grafana_data` - Grafana dashboards and settings

**Retention Strategy:**
- All systems: 2 hours retention
- Suitable for development/testing
- Increase for production use

**Data Cleanup:**
```bash
docker-compose down -v  # Removes all volumes
```

## Scaling Considerations

### Current Setup (Development)
- Single instance of each component
- Local storage
- No replication
- 2-hour retention

### Production Recommendations
- OTEL Collector: Multiple instances behind load balancer
- Prometheus: Federation or Thanos for long-term storage
- Tempo: S3/GCS backend with multiple queriers
- Loki: Object storage with multiple ingesters
- Grafana: HA mode with shared database

## Monitoring the Monitors

The observability stack monitors itself:

1. **OTEL Collector**
   - Internal metrics: http://localhost:8888/metrics
   - Scraped by Prometheus

2. **Prometheus**
   - Self-scraping enabled
   - Targets view: http://localhost:9090/targets

3. **Tempo**
   - Metrics endpoint scraped by Prometheus
   - Health: http://localhost:3200/ready

4. **Loki**
   - Metrics endpoint scraped by Prometheus
   - Health: http://localhost:3100/ready

5. **Grafana**
   - Health endpoint: http://localhost:3000/api/health

## Security Considerations

### Current Setup (Development)
- ❌ No authentication between services
- ❌ No TLS encryption
- ❌ No API keys
- ✅ Network isolation via Docker

### Production Requirements
- ✅ TLS for all inter-service communication
- ✅ Authentication tokens/API keys
- ✅ mTLS for OTLP
- ✅ Grafana authentication (OAuth, LDAP)
- ✅ Network policies/firewall rules
- ✅ Secret management (Vault, k8s secrets)

## OpenTelemetry Semantic Conventions

The mock service follows OTel semantic conventions:

**Resource Attributes:**
- `service.name` - Service identifier
- `service.version` - Version string
- `deployment.environment` - Environment name

**Span Attributes:**
- `crypto.symbol` - Cryptocurrency symbol
- `crypto.exchange` - Exchange name
- `operation.type` - Operation type
- `error` - Error flag
- `http.status_code` - HTTP status

**Metric Names:**
- `crypto.requests.total` - Following OTel naming
- `crypto.price.current` - Descriptive names
- `crypto.request.duration` - With units

## Troubleshooting Guide

### Data Not Flowing

1. **Check OTEL Collector**
   ```bash
   docker-compose logs otel-collector | grep -i error
   curl http://localhost:8888/metrics | grep receiver_accepted
   ```

2. **Check Exporters**
   ```bash
   curl http://localhost:8888/metrics | grep exporter_sent
   ```

3. **Check Backends**
   ```bash
   # Prometheus
   curl http://localhost:9090/-/healthy
   
   # Tempo
   curl http://localhost:3200/ready
   
   # Loki
   curl http://localhost:3100/ready
   ```

### Performance Issues

- Increase OTEL Collector memory limit
- Adjust batch processor settings
- Reduce scrape frequency
- Increase retention periods

### High Cardinality

- Review metric labels
- Use histograms instead of gauges where appropriate
- Implement sampling for traces
- Use label drop processors in OTEL Collector

## References

- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
- [OTEL Collector Configuration](https://opentelemetry.io/docs/collector/configuration/)
- [Prometheus Best Practices](https://prometheus.io/docs/practices/)
- [Tempo Documentation](https://grafana.com/docs/tempo/)
- [Loki Documentation](https://grafana.com/docs/loki/)
- [Grafana Correlations](https://grafana.com/docs/grafana/latest/administration/correlations/)

