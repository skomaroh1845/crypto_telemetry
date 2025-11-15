package telemetry

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

//
type Metrics struct {
	RequestsCounter     metric.Int64Counter
	ErrorsCounter       metric.Int64Counter
	ProcessingDuration  metric.Float64Histogram
	MessagesSentCounter metric.Int64Counter
}

func NewMetrics() (*Metrics, error) {
	meter := otel.Meter("notifier-service")

	requestsCounter, err := meter.Int64Counter("notifier_requests_total",
		metric.WithDescription("Total number of requests"),
	)
	if err != nil {
		return nil, err
	}

	errorsCounter, err := meter.Int64Counter("notifier_errors_total",
		metric.WithDescription("Total number of errors"),
	)
	if err != nil {
		return nil, err
	}

	processingDuration, err := meter.Float64Histogram("notifier_processing_duration_seconds",
		metric.WithDescription("Processing duration in seconds"),
	)
	if err != nil {
		return nil, err
	}

	messagesSentCounter, err := meter.Int64Counter("notifier_messages_sent_total",
		metric.WithDescription("Total number of messages sent to Telegram"),
	)
	if err != nil {
		return nil, err
	}

	return &Metrics{
		RequestsCounter:     requestsCounter,
		ErrorsCounter:       errorsCounter,
		ProcessingDuration:  processingDuration,
		MessagesSentCounter: messagesSentCounter,
	}, nil
}
