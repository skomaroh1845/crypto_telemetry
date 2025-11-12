#!/bin/bash

# Quick Start Script for Observability Stack

echo "🚀 Starting Observability Stack..."
echo ""

# Start the observability services
docker-compose up -d otel-collector prometheus tempo loki grafana mock-telemetry-service

echo ""
echo "⏳ Waiting for services to start (15 seconds)..."
sleep 15

echo ""
echo "✓ Observability Stack is running!"
echo ""
echo "📊 Access Points:"
echo "  • Grafana:    http://localhost:3000 (admin/admin)"
echo "  • Prometheus: http://localhost:9090"
echo "  • Tempo:      http://localhost:3200"  
echo "  • Loki:       http://localhost:3100"
echo ""
echo "🔍 Useful Commands:"
echo "  • View logs:        docker-compose logs -f mock-telemetry-service"
echo "  • Check status:     docker-compose ps"
echo "  • Stop services:    docker-compose down"
echo "  • Management tool:  ./observability/manage.sh help"
echo ""
echo "📚 Documentation:"
echo "  • Setup Guide:      OBSERVABILITY_SETUP.md"
echo "  • Detailed Docs:    observability/README.md"
echo ""

