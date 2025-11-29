package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type ExchangeClient struct {
	baseURL string
	apiKey  string
	client  *http.Client
	tracer  trace.Tracer
}

type ExchangeResponse struct {
	Status  string       `json:"status"`
	Symbols []SymbolData `json:"symbols"`
}

type SymbolData struct {
	Symbol                string `json:"symbol"`
	Last                  string `json:"last"`
	LastBTC               string `json:"last_btc"`
	Lowest                string `json:"lowest"`
	Highest               string `json:"highest"`
	Date                  string `json:"date"`
	DailyChangePercentage string `json:"daily_change_percentage"`
	SourceExchange        string `json:"source_exchange"`
}

func NewExchangeClient(baseURL, apiKey string) *ExchangeClient {
	return &ExchangeClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		tracer: otel.Tracer("data-service"),
	}
}

func (c *ExchangeClient) GetPrice(ctx context.Context, symbol string) (*ExchangeResponse, error) {

	// Создаем вложенный span для операции "exchange_api_call"
	ctx, span := c.tracer.Start(ctx, "exchange_api_call",
		trace.WithAttributes(
			attribute.String("symbol", symbol),
			attribute.String("api.url", c.baseURL),
		),
	)
	defer span.End() // Завершаем span при выходе из функции

	if symbol == "" {
		span.RecordError(fmt.Errorf("symbol parameter is required"))
		span.SetStatus(codes.Error, "symbol parameter is required")
		return nil, fmt.Errorf("symbol parameter is required")
	}

	url := fmt.Sprintf("%s?symbol=%s", c.baseURL, symbol)

	// Создаем HTTP request с context
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create request")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("accept", "*/*")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	// Инжектируем trace context в заголовки запроса
	propagator := otel.GetTextMapPropagator()
	propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))

	// Выполняем запрос
	resp, err := c.client.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to execute request")
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Извлекаем trace context из заголовков ответа
	respHeaders := propagation.HeaderCarrier(resp.Header)
	remoteCtx := propagator.Extract(ctx, respHeaders)

	// Связываем полученный trace context со спаном
	remoteSpan := trace.SpanFromContext(remoteCtx)
	if remoteSpan.SpanContext().IsValid() {
		// Добавляем информацию о remote span в атрибуты
		span.SetAttributes(
			attribute.String("remote.trace_id", remoteSpan.SpanContext().TraceID().String()),
			attribute.String("remote.span_id", remoteSpan.SpanContext().SpanID().String()),
		)
		// Добавляем как event с информацией о remote span
		span.AddEvent("exchange_api_response_received",
			trace.WithAttributes(
				attribute.String("remote.trace_id", remoteSpan.SpanContext().TraceID().String()),
				attribute.String("remote.span_id", remoteSpan.SpanContext().SpanID().String()),
			),
		)
	}

	// Добавляем атрибуты в span
	span.SetAttributes(
		attribute.Int("http.status_code", resp.StatusCode),
		attribute.String("http.method", "GET"),
	)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		err := fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
		span.RecordError(err)
		span.SetStatus(codes.Error, "API returned error status")
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to read response body")
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var exchange_resp ExchangeResponse
	if err := json.Unmarshal(body, &exchange_resp); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to parse response")
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if exchange_resp.Status != "success" {
		err := fmt.Errorf("API returned status: %s", exchange_resp.Status)
		span.RecordError(err)
		span.SetStatus(codes.Error, "API returned non-success status")
		return nil, err
	}

	if len(exchange_resp.Symbols) == 0 {
		err := fmt.Errorf("no symbols in response")
		span.RecordError(err)
		span.SetStatus(codes.Error, "no symbols in response")
		return nil, err
	}

	// Успешное выполнение
	span.SetStatus(codes.Ok, "success")
	return &exchange_resp, nil
}
