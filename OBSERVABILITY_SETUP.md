# Observability Stack - MVP Setup Complete

## 🎯 Overview

A complete observability stack has been set up with:
- ✅ **OTEL Collector** - Central telemetry data receiver
- ✅ **Prometheus** - Metrics storage and querying
- ✅ **Grafana Tempo** - Distributed tracing backend
- ✅ **Grafana Loki** - Log aggregation
- ✅ **Grafana** - Unified visualization dashboard
- ✅ **Mock Service** - Generates test telemetry data

## 🚀 Quick Start

### 1. Start the Stack

```bash
cd /home/khamitov02/crypto_telemetry
docker-compose up -d
```

Or use the management script:

```bash
cd /home/khamitov02/crypto_telemetry/observability
./manage.sh start
```

### 2. Access Services

| Service | URL | Credentials |
|---------|-----|-------------|
| **Grafana** | http://localhost:3000 | admin/admin |
| **Prometheus** | http://localhost:9090 | - |
| **Tempo** | http://localhost:3200 | - |
| **Loki** | http://localhost:3100 | - |

### 3. View Telemetry in Grafana

1. Open http://localhost:3000
2. Login with `admin`/`admin`
3. Go to **Explore** (compass icon)
4. Select datasource:
   - **Prometheus** for metrics
   - **Tempo** for traces  
   - **Loki** for logs

## 📊 Data Flow

```
┌──────────────────────┐
│   Your Services      │
│   Mock Service       │
└──────────┬───────────┘
           │ OTLP Protocol
           │ (HTTP: 4318, gRPC: 4317)
           ▼
┌──────────────────────┐
│   OTEL Collector     │ ◄── Processes & Routes Data
└──┬──────┬────────┬───┘
   │      │        │
   │      │        └─────────┐
   │      │                  │
   ▼      ▼                  ▼
┌─────┐ ┌──────┐        ┌──────┐
│Prome│ │Tempo │        │ Loki │
│theus│ │      │        │      │
└──┬──┘ └───┬──┘        └───┬──┘
   │        │               │
   └────────┴───────────────┘
            │
            ▼
    ┌──────────────┐
    │   Grafana    │ ◄── Unified Dashboard
    └──────────────┘
```

## 📁 File Structure

```
crypto_telemetry/
├── docker-compose.yml                    # Main compose file with all services
├── observability/
│   ├── README.md                        # Detailed documentation
│   ├── manage.sh                        # Management script
│   ├── otel-collector/
│   │   └── otel-collector-config.yaml  # Collector configuration
│   ├── prometheus/
│   │   └── prometheus.yml              # Prometheus configuration
│   ├── tempo/
│   │   └── tempo.yaml                  # Tempo configuration
│   ├── loki/
│   │   └── loki.yaml                   # Loki configuration
│   └── grafana/
│       └── provisioning/
│           ├── datasources/
│           │   └── datasources.yaml    # Auto-configured datasources
│           └── dashboards/
│               └── dashboards.yaml     # Dashboard provisioning
└── mock_telemetry_service/
    ├── Dockerfile
    ├── requirements.txt
    └── main.py                          # Mock service generating telemetry
```

## 🔧 Configuration Highlights

### OTEL Collector
- **Receives**: OTLP (gRPC: 4317, HTTP: 4318)
- **Exports to**:
  - Prometheus (via Prometheus exporter on :8889)
  - Tempo (via OTLP)
  - Loki (via Loki exporter)
- **Processors**: Batch, Memory limiter, Resource enrichment

### Prometheus
- **Scrape interval**: 15s
- **Retention**: 2 hours
- **Scrapes**:
  - OTEL Collector metrics (:8889)
  - OTEL Collector internal (:8888)
  - Self-monitoring
  - Tempo, Loki metrics

### Tempo
- **Protocol**: OTLP (gRPC & HTTP)
- **Storage**: Local filesystem
- **Retention**: 2 hours
- **Features**: Service graphs, span metrics, metrics generation

### Loki
- **Schema**: TSDB v13
- **Storage**: Local filesystem
- **Retention**: 2 hours
- **Compaction**: Enabled (10m interval)

### Grafana
- **Pre-configured datasources**: Prometheus, Tempo, Loki
- **Features**:
  - Trace-to-logs correlation
  - Trace-to-metrics correlation
  - Service maps
  - Node graphs
  - TraceQL editor enabled

## 🧪 Mock Service

The mock service generates realistic telemetry data:

### Traces
- Nested spans (parent-child relationships)
- Crypto trading operations (fetch_price, fetch_orderbook, etc.)
- Multiple exchanges (binance, coinbase, kraken, bybit, okx)
- Complex workflows with multiple steps

### Metrics
- **crypto.requests.total** (counter) - Total requests by symbol/exchange
- **crypto.price.current** (gauge) - Current cryptocurrency prices
- **crypto.request.duration** (histogram) - Request latency distribution

### Logs
- Structured logging with consistent fields
- Different log levels (INFO, ERROR, DEBUG)
- Contextual information (symbol, exchange, price, etc.)
- Trace correlation via trace IDs

### Error Simulation
- 10% random error rate
- Various error types (rate limits, timeouts, API errors)
- Proper error attributes in spans

## 📝 Sample Queries

### Prometheus (Metrics)

```promql
# Request rate by symbol
rate(crypto_requests_total[5m])

# 95th percentile latency
histogram_quantile(0.95, rate(crypto_request_duration_bucket[5m]))

# Error rate percentage
100 * sum(rate(crypto_requests_total{status="error"}[5m])) / sum(rate(crypto_requests_total[5m]))

# Requests by exchange
sum by (exchange) (rate(crypto_requests_total[5m]))
```

### Loki (Logs)

```logql
# All logs from mock service
{service_name="mock-telemetry-service"}

# Only errors
{service_name="mock-telemetry-service"} |= "ERROR"

# Logs for specific crypto symbol
{service_name="mock-telemetry-service"} | json | symbol="BTC"

# Logs with high prices
{service_name="mock-telemetry-service"} | json | price > 40000

# Count logs by level
sum by (level) (count_over_time({service_name="mock-telemetry-service"}[5m]))
```

### Tempo (Traces)

```traceql
# All traces for Bitcoin
{.crypto.symbol = "BTC"}

# Slow traces (over 200ms)
{duration > 200ms}

# Failed traces
{.error = true}

# Specific operation type
{name = "process_trading_signal"}

# Traces from specific exchange
{.crypto.exchange = "binance"}

# Complex query: slow BTC trades with errors
{.crypto.symbol = "BTC" && duration > 200ms && .error = true}
```

## 🔍 Troubleshooting

### Check Service Health

```bash
# Using the management script
./observability/manage.sh status

# Check all containers
docker-compose ps

# View logs
docker-compose logs -f otel-collector
docker-compose logs -f mock-telemetry-service
```

### Verify Data Flow

```bash
# Test telemetry pipeline
./observability/manage.sh test

# Check OTEL Collector is receiving data
curl http://localhost:8888/metrics | grep otelcol_receiver_accepted

# Check Prometheus has data
curl "http://localhost:9090/api/v1/query?query=up"

# Check Loki has logs
curl -G -s "http://localhost:3100/loki/api/v1/query" --data-urlencode 'query={service_name="mock-telemetry-service"}' | jq
```

### Common Issues

1. **No data in Grafana**
   - Check OTEL Collector logs: `docker-compose logs otel-collector`
   - Verify mock service is running: `docker-compose ps mock-telemetry-service`
   - Check Prometheus targets: http://localhost:9090/targets

2. **Port already in use**
   - Modify port mappings in `docker-compose.yml`
   - Or stop conflicting services

3. **Services not starting**
   - Check Docker resources (memory, disk)
   - Review service logs: `docker-compose logs <service>`
   - Try: `docker-compose down -v && docker-compose up -d`

## 🔄 Management Commands

```bash
cd observability

# Start stack
./manage.sh start

# Stop stack
./manage.sh stop

# Restart stack
./manage.sh restart

# Show status
./manage.sh status

# Show logs (all or specific service)
./manage.sh logs
./manage.sh logs otel-collector

# Show URLs
./manage.sh urls

# Test pipeline
./manage.sh test

# Query metrics
./manage.sh metrics

# Clean all data (removes volumes)
./manage.sh clean
```

## 🔌 Integrating Your Services

To send telemetry from your services:

### 1. Set Environment Variables

```yaml
environment:
  - OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4318
  - OTEL_SERVICE_NAME=your-service-name
```

### 2. Add Dependency

```yaml
depends_on:
  - otel-collector
```

### 3. Use OpenTelemetry SDK

See `mock_telemetry_service/main.py` for a complete example.

## 📊 Grafana Setup Tips

### Creating Dashboards

1. Go to **Dashboards** → **New** → **New Dashboard**
2. Add panels for:
   - Request rate (Prometheus)
   - Error rate (Prometheus)
   - Latency distribution (Prometheus)
   - Recent traces (Tempo)
   - Log stream (Loki)

### Correlating Data

- Click on a trace ID in logs → Opens in Tempo
- Click on "Logs for this span" in Tempo → Opens in Loki
- Click on metrics in Tempo → Opens in Prometheus

### Pre-built Queries

The datasources are configured with correlations:
- Traces link to logs automatically
- Traces link to metrics automatically
- Service maps show request flow

## 🎯 Next Steps

1. **Customize Retention**: Adjust retention periods in configs (currently 2h)
2. **Add Alerting**: Configure alert rules in Prometheus
3. **Create Dashboards**: Build custom Grafana dashboards
4. **Instrument Services**: Add OpenTelemetry to your actual services
5. **Add Sampling**: Configure trace sampling for production
6. **Security**: Add authentication, TLS, API keys for production

## 📚 References

- [OpenTelemetry Docs](https://opentelemetry.io/docs/)
- [OTEL Collector](https://opentelemetry.io/docs/collector/)
- [Prometheus](https://prometheus.io/docs/)
- [Tempo](https://grafana.com/docs/tempo/)
- [Loki](https://grafana.com/docs/loki/)
- [Grafana](https://grafana.com/docs/grafana/)
- [OpenTelemetry Semantic Conventions](https://opentelemetry.io/docs/specs/semconv/)
- [Prometheus Naming Conventions](https://prometheus.io/docs/practices/naming/)

## ✅ What's Working

- ✅ OTEL Collector receiving OTLP data
- ✅ Prometheus scraping metrics from OTEL Collector
- ✅ Tempo receiving and storing traces
- ✅ Loki receiving and storing logs
- ✅ Grafana visualizing all data sources
- ✅ Mock service generating test data
- ✅ Trace-to-logs correlation
- ✅ Trace-to-metrics correlation
- ✅ 2-hour retention on all systems
- ✅ Self-monitoring of observability stack

---

**Need help?** Check the detailed documentation in `observability/README.md`

