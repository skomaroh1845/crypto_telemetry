#!/bin/bash

# Observability Stack Management Script

set -e

COMPOSE_FILE="../docker-compose.yml"
PROJECT_NAME="crypto_telemetry"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

print_header() {
    echo -e "${GREEN}================================${NC}"
    echo -e "${GREEN}$1${NC}"
    echo -e "${GREEN}================================${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ ${NC}$1"
}

print_success() {
    echo -e "${GREEN}✓ ${NC}$1"
}

print_error() {
    echo -e "${RED}✗ ${NC}$1"
}

start_stack() {
    print_header "Starting Observability Stack"
    
    cd ..
    docker-compose up -d otel-collector prometheus tempo loki grafana mock-telemetry-service
    
    print_success "All services started!"
    print_info "Waiting for services to be healthy..."
    sleep 5
    
    show_urls
}

stop_stack() {
    print_header "Stopping Observability Stack"
    
    cd ..
    docker-compose stop otel-collector prometheus tempo loki grafana mock-telemetry-service
    
    print_success "All services stopped!"
}

restart_stack() {
    print_header "Restarting Observability Stack"
    stop_stack
    sleep 2
    start_stack
}

show_logs() {
    print_header "Service Logs"
    
    if [ -z "$1" ]; then
        print_info "Showing logs for all observability services..."
        cd ..
        docker-compose logs -f otel-collector prometheus tempo loki grafana mock-telemetry-service
    else
        print_info "Showing logs for $1..."
        cd ..
        docker-compose logs -f "$1"
    fi
}

show_status() {
    print_header "Service Status"
    
    cd ..
    docker-compose ps otel-collector prometheus tempo loki grafana mock-telemetry-service
    
    echo ""
    print_info "Health checks:"
    
    # Check OTEL Collector
    if curl -s http://localhost:8888/metrics > /dev/null; then
        print_success "OTEL Collector is healthy"
    else
        print_error "OTEL Collector is not responding"
    fi
    
    # Check Prometheus
    if curl -s http://localhost:9090/-/healthy > /dev/null; then
        print_success "Prometheus is healthy"
    else
        print_error "Prometheus is not responding"
    fi
    
    # Check Tempo
    if curl -s http://localhost:3200/ready > /dev/null; then
        print_success "Tempo is healthy"
    else
        print_error "Tempo is not responding"
    fi
    
    # Check Loki
    if curl -s http://localhost:3100/ready > /dev/null; then
        print_success "Loki is healthy"
    else
        print_error "Loki is not responding"
    fi
    
    # Check Grafana
    if curl -s http://localhost:3000/api/health > /dev/null; then
        print_success "Grafana is healthy"
    else
        print_error "Grafana is not responding"
    fi
}

show_urls() {
    print_header "Service URLs"
    echo ""
    echo "  📊 Grafana:         http://localhost:3000 (admin/admin)"
    echo "  📈 Prometheus:      http://localhost:9090"
    echo "  🔍 Tempo:           http://localhost:3200"
    echo "  📝 Loki:            http://localhost:3100"
    echo "  🔄 OTEL Collector:  http://localhost:4318 (HTTP) / :4317 (gRPC)"
    echo "  📊 OTEL Metrics:    http://localhost:8888/metrics"
    echo ""
}

clean_data() {
    print_header "Cleaning Observability Data"
    
    read -p "This will remove all volumes and data. Are you sure? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        cd ..
        docker-compose down -v
        print_success "All data cleaned!"
    else
        print_info "Cancelled"
    fi
}

test_telemetry() {
    print_header "Testing Telemetry Pipeline"
    
    print_info "Checking if mock service is generating data..."
    cd ..
    docker-compose logs --tail=20 mock-telemetry-service
    
    echo ""
    print_info "Checking OTEL Collector metrics..."
    curl -s http://localhost:8888/metrics | grep -E "otelcol_receiver_accepted|otelcol_exporter_sent" | head -10
    
    echo ""
    print_info "Checking Prometheus targets..."
    curl -s http://localhost:9090/api/v1/targets | jq -r '.data.activeTargets[] | "\(.labels.job): \(.health)"'
}

query_metrics() {
    print_header "Sample Metrics Queries"
    
    print_info "Request rate by symbol:"
    curl -s "http://localhost:9090/api/v1/query?query=rate(crypto_requests_total[5m])" | jq -r '.data.result[] | "\(.metric.symbol): \(.value[1])"'
}

show_help() {
    echo "Observability Stack Management"
    echo ""
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  start       Start all observability services"
    echo "  stop        Stop all observability services"
    echo "  restart     Restart all observability services"
    echo "  status      Show service status and health"
    echo "  logs        Show logs (optionally specify service name)"
    echo "  urls        Show service URLs"
    echo "  clean       Remove all data and volumes"
    echo "  test        Test telemetry pipeline"
    echo "  metrics     Query sample metrics"
    echo "  help        Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 start"
    echo "  $0 logs otel-collector"
    echo "  $0 status"
}

# Main script logic
case "${1:-help}" in
    start)
        start_stack
        ;;
    stop)
        stop_stack
        ;;
    restart)
        restart_stack
        ;;
    status)
        show_status
        ;;
    logs)
        show_logs "$2"
        ;;
    urls)
        show_urls
        ;;
    clean)
        clean_data
        ;;
    test)
        test_telemetry
        ;;
    metrics)
        query_metrics
        ;;
    help|*)
        show_help
        ;;
esac

