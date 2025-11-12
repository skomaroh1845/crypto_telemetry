# Observability Stack MVP

This directory contains the complete observability setup for the crypto telemetry system using OpenTelemetry, Prometheus, Tempo, Loki, and Grafana.

## Architecture

```
┌─────────────────┐
│  Applications   │
│  (Services)     │
└────────┬────────┘
         │ OTLP (HTTP/gRPC)
         ▼
┌─────────────────┐
│ OTEL Collector  │ ◄── Receives all telemetry data
└────┬───┬────┬───┘
     │   │    │
     │   │    └──────────┐
     │   │               │
     ▼   ▼               ▼
┌────────┐  ┌──────┐  ┌──────┐
│Promethe│  │Tempo │  │ Loki │
│  us    │  │      │  │      │
│(Metrics)  │(Trace)  │(Logs)│
└────┬───┘  └──┬───┘  └──┬───┘
     │         │          │
     └─────────┴──────────┘
               │
               ▼
         ┌──────────┐
         │ Grafana  │ ◄── Unified visualization
         └──────────┘
```

## Components

### 1. **OTEL Collector** (Port 4317, 4318, 8888, 8889)
- Receives telemetry data via OTLP protocol (gRPC: 4317, HTTP: 4318)
- Routes data to appropriate backends:
  - **Traces** → Tempo
  - **Metrics** → Prometheus (via Prometheus exporter)
  - **Logs** → Loki
- Exposes internal metrics on port 8888 and 8889

### 2. **Prometheus** (Port 9090)
- Time-series database for metrics
- Scrapes metrics from:
  - OTEL Collector (port 8889)
  - OTEL Collector internal metrics (port 8888)
  - Self-monitoring
  - Tempo, Loki
- **Retention**: 2 hours (configurable)

### 3. **Grafana Tempo** (Port 3200)
- Distributed tracing backend
- Receives traces from OTEL Collector via OTLP
- Stores traces with 2-hour retention
- Generates service graphs and span metrics

### 4. **Grafana Loki** (Port 3100)
- Log aggregation system
- Receives logs from OTEL Collector
- **Retention**: 2 hours
- Supports LogQL for querying

### 5. **Grafana** (Port 3000)
- Unified observability dashboard
- Pre-configured datasources:
  - Prometheus (default)
  - Tempo
  - Loki
- Correlates traces, metrics, and logs
- **Credentials**: admin/admin

### 6. **Mock Telemetry Service**
- Generates synthetic telemetry data for testing
- Simulates cryptocurrency trading operations
- Produces:
  - Traces with nested spans
  - Metrics (counters, histograms, gauges)
  - Structured logs
  - Occasional errors for testing

## Quick Start

### 1. Start the Stack

```bash
docker-compose up -d
```

### 2. Verify Services

```bash
# Check all services are running
docker-compose ps

# Check OTEL Collector logs
docker-compose logs -f otel-collector

# Check mock service is generating data
docker-compose logs -f mock-telemetry-service
```

### 3. Access UIs

- **Grafana**: http://localhost:3000 (admin/admin)
- **Prometheus**: http://localhost:9090
- **Tempo**: http://localhost:3200
- **Loki**: http://localhost:3100

### 4. View Data in Grafana

1. Open Grafana at http://localhost:3000
2. Login with admin/admin
3. Navigate to **Explore**
4. Select different datasources:
   - **Prometheus**: View metrics
   - **Tempo**: View traces
   - **Loki**: View logs

## Configuration Files

### OTEL Collector (`otel-collector/otel-collector-config.yaml`)
- **Receivers**: OTLP (gRPC + HTTP)
- **Processors**: Batch, Memory limiter, Resource
- **Exporters**: Prometheus, Tempo (OTLP), Loki

### Prometheus (`prometheus/prometheus.yml`)
- Scrape configs for all services
- 2-hour retention
- 15s scrape interval

### Tempo (`tempo/tempo.yaml`)
- OTLP receivers (gRPC + HTTP)
- Local storage backend
- Metrics generator enabled
- 2-hour block retention

### Loki (`loki/loki.yaml`)
- TSDB schema (v13)
- Filesystem storage
- 2-hour retention
- Compaction enabled

### Grafana Datasources (`grafana/provisioning/datasources/datasources.yaml`)
- Auto-provisioned datasources
- Trace-to-logs correlation
- Trace-to-metrics correlation
- Service map and node graph enabled

## Useful Queries

### Prometheus (Metrics)

```promql
# Request rate by symbol
rate(crypto_requests_total[5m])

# Request duration p95
histogram_quantile(0.95, rate(crypto_request_duration_bucket[5m]))

# Error rate
rate(crypto_requests_total{status="error"}[5m])
```

### Loki (Logs)

```logql
# All logs from mock service
{service_name="mock-telemetry-service"}

# Only errors
{service_name="mock-telemetry-service"} |= "error"

# Logs for specific symbol
{service_name="mock-telemetry-service"} | json | symbol="BTC"
```

### Tempo (Traces)

```traceql
# All traces for a specific symbol
{.crypto.symbol = "BTC"}

# Slow traces (>200ms)
{duration > 200ms}

# Error traces
{.error = true}

# Specific operation
{name = "process_trading_signal"}
```

## Monitoring the Observability Stack

The stack monitors itself:

- **OTEL Collector**: Metrics at http://localhost:8888/metrics
- **Prometheus**: Self-monitoring at http://localhost:9090/targets
- **Tempo**: Metrics exposed to Prometheus
- **Loki**: Metrics exposed to Prometheus

## Data Retention

All components are configured with **2-hour retention** as requested:

- Prometheus: `--storage.tsdb.retention.time=2h`
- Tempo: `compaction.block_retention: 2h`
- Loki: `retention_period: 2h`

## Troubleshooting

### No data in Grafana?

1. Check OTEL Collector is receiving data:
   ```bash
   docker-compose logs otel-collector | grep -i "traces\|metrics\|logs"
   ```

2. Verify mock service is running:
   ```bash
   docker-compose logs mock-telemetry-service
   ```

3. Check Prometheus targets:
   - Open http://localhost:9090/targets
   - All targets should be "UP"

### Services not starting?

```bash
# Check logs for specific service
docker-compose logs <service-name>

# Restart services
docker-compose restart

# Complete restart
docker-compose down -v
docker-compose up -d
```

### Port conflicts?

If ports are already in use, modify the port mappings in `docker-compose.yml`.

## Integrating Your Services

To send telemetry from your services to this stack:

### Environment Variables

```yaml
environment:
  - OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4318  # HTTP
  # or
  - OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4317  # gRPC
  - OTEL_SERVICE_NAME=your-service-name
```

### Example with Python

```python
from opentelemetry import trace
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor

# Setup
provider = TracerProvider()
processor = BatchSpanProcessor(
    OTLPSpanExporter(endpoint="http://otel-collector:4318/v1/traces")
)
provider.add_span_processor(processor)
trace.set_tracer_provider(provider)

# Use
tracer = trace.get_tracer(__name__)
with tracer.start_as_current_span("operation"):
    # Your code here
    pass
```

## References

- [OpenTelemetry Collector](https://opentelemetry.io/docs/collector/)
- [Prometheus](https://prometheus.io/docs/)
- [Grafana Tempo](https://grafana.com/docs/tempo/)
- [Grafana Loki](https://grafana.com/docs/loki/)
- [OpenTelemetry Semantic Conventions](https://opentelemetry.io/docs/specs/semconv/)
- [Prometheus Naming Best Practices](https://prometheus.io/docs/practices/naming/)

## Network Architecture

All services are connected via `crypto_telemetry_network` bridge network, allowing them to communicate using service names as hostnames.

## Volume Persistence

Data is persisted in named Docker volumes:
- `prometheus_data`: Prometheus TSDB
- `tempo_data`: Tempo traces
- `loki_data`: Loki logs
- `grafana_data`: Grafana dashboards and settings

To reset all data:
```bash
docker-compose down -v
```

