# 🎉 Observability Stack Setup - COMPLETE

## ✅ What Was Created

### 1. Docker Compose Configuration
- **File**: `docker-compose.yml` (updated)
- **Services Added**:
  - `otel-collector` - OpenTelemetry Collector v0.91.0
  - `prometheus` - Prometheus v2.48.1
  - `tempo` - Grafana Tempo v2.3.1
  - `loki` - Grafana Loki v2.9.3
  - `grafana` - Grafana v10.2.3
  - `mock-telemetry-service` - Test data generator

### 2. Configuration Files

#### OTEL Collector
```
observability/otel-collector/otel-collector-config.yaml
```
- OTLP receivers (gRPC + HTTP)
- Batch, memory limiter, resource processors
- Exports to Prometheus, Tempo, Loki

#### Prometheus
```
observability/prometheus/prometheus.yml
```
- Scrapes OTEL Collector, Tempo, Loki
- 2-hour retention
- 15s scrape interval

#### Tempo
```
observability/tempo/tempo.yaml
```
- OTLP receivers
- Local storage
- 2-hour retention
- Metrics generator enabled

#### Loki
```
observability/loki/loki.yaml
```
- TSDB schema v13
- 2-hour retention
- Compaction enabled

#### Grafana
```
observability/grafana/provisioning/datasources/datasources.yaml
observability/grafana/provisioning/dashboards/dashboards.yaml
```
- Pre-configured Prometheus, Tempo, Loki datasources
- Trace-to-logs correlation
- Trace-to-metrics correlation

### 3. Mock Telemetry Service
```
mock_telemetry_service/
├── Dockerfile
├── requirements.txt
└── main.py
```
- Generates realistic crypto trading telemetry
- Traces with nested spans
- Metrics (counters, gauges, histograms)
- Structured logs with trace correlation
- 10% error rate for testing

### 4. Documentation
```
OBSERVABILITY_SETUP.md         - Quick start guide
observability/README.md         - Detailed documentation
observability/ARCHITECTURE.md   - Architecture deep-dive
observability/SETUP_SUMMARY.md  - This file
```

### 5. Management Tools
```
start-observability.sh          - Quick start script
observability/manage.sh         - Full management CLI
observability/.gitignore        - Git ignore file
```

## 🚀 Quick Start

### Option 1: Quick Start Script
```bash
cd /home/khamitov02/crypto_telemetry
./start-observability.sh
```

### Option 2: Docker Compose
```bash
cd /home/khamitov02/crypto_telemetry
docker-compose up -d
```

### Option 3: Management CLI
```bash
cd /home/khamitov02/crypto_telemetry/observability
./manage.sh start
```

## 🌐 Access Points

| Service | URL | Purpose |
|---------|-----|---------|
| **Grafana** | http://localhost:3000 | Dashboards & visualization |
| **Prometheus** | http://localhost:9090 | Metrics queries |
| **Tempo** | http://localhost:3200 | Traces API |
| **Loki** | http://localhost:3100 | Logs API |
| **OTEL Collector** | http://localhost:4318 | OTLP HTTP endpoint |
| **OTEL Collector** | http://localhost:4317 | OTLP gRPC endpoint |
| **OTEL Metrics** | http://localhost:8888/metrics | Collector internal metrics |

**Grafana Credentials**: `admin` / `admin`

## 📊 Data Flow

```
Your Service → OTEL Collector → Prometheus (metrics)
                             → Tempo (traces)
                             → Loki (logs)
                                       ↓
                              All feed into Grafana
```

## 🔍 Verification Steps

### 1. Check Services Are Running
```bash
docker-compose ps
# All services should show "Up" status
```

### 2. Check Mock Service Logs
```bash
docker-compose logs -f mock-telemetry-service
# Should see logs about crypto operations
```

### 3. Check OTEL Collector
```bash
curl http://localhost:8888/metrics | grep otelcol_receiver_accepted
# Should show increasing counter values
```

### 4. Check Prometheus
```bash
curl -s "http://localhost:9090/api/v1/query?query=up" | jq '.data.result[].metric.job'
# Should show: prometheus, otel-collector, tempo, loki
```

### 5. Check Grafana
```bash
curl -s http://localhost:3000/api/health | jq
# Should show: {"database": "ok", "version": "..."}
```

## 📈 View Data in Grafana

1. **Open Grafana**: http://localhost:3000
2. **Login**: admin / admin
3. **Navigate to Explore** (compass icon in sidebar)
4. **Try these queries**:

### Metrics (Prometheus)
```promql
# Request rate
rate(crypto_requests_total[5m])

# 95th percentile latency
histogram_quantile(0.95, rate(crypto_request_duration_bucket[5m]))

# Error rate
rate(crypto_requests_total{status="error"}[5m])
```

### Traces (Tempo)
```traceql
# All traces
{}

# Bitcoin trades
{.crypto.symbol = "BTC"}

# Slow traces
{duration > 200ms}

# Error traces
{.error = true}
```

### Logs (Loki)
```logql
# All logs
{service_name="mock-telemetry-service"}

# Errors only
{service_name="mock-telemetry-service"} |= "ERROR"

# Bitcoin logs
{service_name="mock-telemetry-service"} | json | symbol="BTC"
```

## 🛠️ Common Operations

### View Logs
```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f otel-collector
docker-compose logs -f mock-telemetry-service
```

### Restart Services
```bash
docker-compose restart otel-collector
# or
./observability/manage.sh restart
```

### Stop Stack
```bash
docker-compose stop
# or
./observability/manage.sh stop
```

### Clean All Data
```bash
docker-compose down -v
# WARNING: This removes all data!
```

### Test Pipeline
```bash
./observability/manage.sh test
```

## 🎯 Key Features Implemented

### ✅ OpenTelemetry Collector
- [x] OTLP HTTP receiver (port 4318)
- [x] OTLP gRPC receiver (port 4317)
- [x] Batch processor
- [x] Memory limiter
- [x] Prometheus exporter
- [x] Tempo (OTLP) exporter
- [x] Loki exporter
- [x] Internal telemetry

### ✅ Prometheus
- [x] Scrapes OTEL Collector metrics
- [x] 2-hour retention
- [x] Self-monitoring
- [x] Monitors all backends

### ✅ Tempo
- [x] OTLP receivers
- [x] Local storage
- [x] 2-hour retention
- [x] Metrics generator
- [x] Service graphs

### ✅ Loki
- [x] HTTP push endpoint
- [x] TSDB storage
- [x] 2-hour retention
- [x] Compaction enabled

### ✅ Grafana
- [x] Pre-configured datasources
- [x] Trace-to-logs correlation
- [x] Trace-to-metrics correlation
- [x] TraceQL editor enabled
- [x] Service maps
- [x] Node graphs

### ✅ Mock Service
- [x] Generates traces
- [x] Generates metrics
- [x] Generates logs
- [x] Realistic crypto scenarios
- [x] Error simulation
- [x] Multiple exchanges/symbols

## 📦 Docker Volumes

Persistent data is stored in named volumes:
- `prometheus_data` - Metrics TSDB
- `tempo_data` - Trace storage
- `loki_data` - Log storage
- `grafana_data` - Dashboards, users, settings

## 🔧 Customization

### Adjust Retention
Edit the respective config files:
- **Prometheus**: `prometheus.yml` → `retention.time: 2h`
- **Tempo**: `tempo.yaml` → `block_retention: 2h`
- **Loki**: `loki.yaml` → `retention_period: 2h`

### Change Ports
Edit `docker-compose.yml` port mappings:
```yaml
ports:
  - "3000:3000"  # host:container
```

### Add More Services
Copy the pattern from `data_service`:
```yaml
environment:
  - OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4318
  - OTEL_SERVICE_NAME=my-service-name
depends_on:
  - otel-collector
```

## 🐛 Troubleshooting

### Services Won't Start
```bash
# Check logs
docker-compose logs <service-name>

# Check Docker resources
docker stats

# Restart everything
docker-compose down
docker-compose up -d
```

### No Data in Grafana
```bash
# 1. Check mock service is running
docker-compose ps mock-telemetry-service

# 2. Check OTEL Collector
docker-compose logs otel-collector | grep -i "traces\|metrics\|logs"

# 3. Check Prometheus targets
curl http://localhost:9090/targets
# All should be "UP"

# 4. Run test
./observability/manage.sh test
```

### Port Conflicts
If ports are in use:
1. Stop conflicting services
2. Or modify ports in `docker-compose.yml`

### Performance Issues
- Increase Docker memory allocation
- Reduce mock service frequency (edit `main.py`)
- Adjust OTEL Collector batch settings

## 📚 Next Steps

### 1. Instrument Your Services
Add OpenTelemetry to your actual services:
```python
# Python example
from opentelemetry import trace
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter
# ... see mock_telemetry_service/main.py for full example
```

### 2. Create Dashboards
- Go to Grafana → Dashboards → New Dashboard
- Add panels for your metrics
- Save and share

### 3. Set Up Alerts
- Prometheus → Alerts
- Define alert rules
- Configure notification channels

### 4. Production Hardening
- Add TLS/SSL
- Enable authentication
- Increase retention
- Use external storage (S3, GCS)
- Set up backups

## 🎓 Learning Resources

### Sample Queries to Try

**Prometheus**:
```promql
# Top 5 symbols by request count
topk(5, sum by (symbol) (rate(crypto_requests_total[5m])))

# Average latency by exchange
avg by (exchange) (rate(crypto_request_duration_sum[5m]) / rate(crypto_request_duration_count[5m]))
```

**Loki**:
```logql
# Count errors by symbol
sum by (symbol) (count_over_time({service_name="mock-telemetry-service"} |= "ERROR" [5m]))

# Parse JSON and filter
{service_name="mock-telemetry-service"} | json | price > 2000
```

**Tempo**:
```traceql
# Find slow Bitcoin operations
{.crypto.symbol = "BTC" && duration > 300ms}

# Operations from a specific exchange
{.crypto.exchange = "binance"}
```

## ✨ Summary

You now have a fully functional observability stack with:
- ✅ Centralized telemetry collection (OTEL Collector)
- ✅ Metrics storage and querying (Prometheus)
- ✅ Distributed tracing (Tempo)
- ✅ Log aggregation (Loki)
- ✅ Unified visualization (Grafana)
- ✅ Test data generation (Mock Service)
- ✅ Complete documentation
- ✅ Management tools

**Everything is connected and ready to use!**

## 🤝 Support

For detailed information, see:
- **Quick Reference**: `OBSERVABILITY_SETUP.md`
- **Detailed Guide**: `observability/README.md`
- **Architecture**: `observability/ARCHITECTURE.md`

For management commands:
```bash
./observability/manage.sh help
```

---

**Happy Observing! 📊🔍📈**

