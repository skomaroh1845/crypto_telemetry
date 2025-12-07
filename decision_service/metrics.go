package main

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	serviceName = "decision_service"
)

type Metrics struct {
	requestCount      metric.Int64Counter
	requestErrorCount metric.Int64Counter
	requestDuration   metric.Float64Histogram
}

func NewMetrics() (*Metrics, error) {
	meter := otel.Meter(serviceName)

	// Счетчик общего количества запросов
	requestCount, err := meter.Int64Counter(
		serviceName+"_request_count_total",
		metric.WithDescription("Total number of HTTP requests"),
	)
	if err != nil {
		return nil, err
	}

	// Счетчик количества ошибочных запросов
	requestErrorCount, err := meter.Int64Counter(
		serviceName+"_request_error_count",
		metric.WithDescription("Total number of HTTP request errors"),
	)
	if err != nil {
		return nil, err
	}

	// Гистограмма времени выполнения запроса
	requestDuration, err := meter.Float64Histogram(
		serviceName+"_request_duration_sec",
		metric.WithDescription("HTTP request duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	return &Metrics{
		requestCount:      requestCount,
		requestErrorCount: requestErrorCount,
		requestDuration:   requestDuration,
	}, nil
}

func (m *Metrics) RecordRequest(ctx context.Context, route string, duration float64, isError bool) {
	attrs := []attribute.KeyValue{
		attribute.String("route", route),
	}

	// Инкрементируем счетчик запросов
	m.requestCount.Add(ctx, 1, metric.WithAttributes(attrs...))

	// Если была ошибка, инкрементируем счетчик ошибок
	if isError {
		m.requestErrorCount.Add(ctx, 1, metric.WithAttributes(attrs...))
	}

	// Записываем время выполнения
	m.requestDuration.Record(ctx, duration, metric.WithAttributes(attrs...))
}
