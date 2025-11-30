package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Handler struct {
	exchangeClient *ExchangeClient
	tracer         trace.Tracer
	metrics        *Metrics
}

func NewHandler(exchangeClient *ExchangeClient, metrics *Metrics) *Handler {
	return &Handler{
		exchangeClient: exchangeClient,
		tracer:         otel.Tracer("data-service"),
		metrics:        metrics,
	}
}

func (h *Handler) GetPrice(w http.ResponseWriter, r *http.Request) {

	start := time.Now()

	// Извлекаем context из HTTP request
	// Контекст уже содержит родительский спан от middleware или от входящего запроса
	ctx := r.Context()

	// Создаем явный спан для обработки запроса GetPrice
	// Этот спан будет дочерним от спана в контексте (если он есть)
	// или корневым спаном, если родительского спана нет
	ctx, span := h.tracer.Start(ctx, "get_price_handler",
		trace.WithAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.url", r.URL.String()),
			attribute.String("http.route", "/price"),
		),
	)
	defer span.End()

	// Get symbol from query parameter, default to BTC if not provided
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		symbol = "BTC"
	}

	// Добавляем атрибут symbol в спан
	span.SetAttributes(attribute.String("request.symbol", symbol))

	// Fetch price data from exchange API
	exchange_resp, err := h.exchangeClient.GetPrice(ctx, symbol)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, fmt.Sprintf("Failed to fetch price: %v", err))
		log.Printf("Error fetching price for symbol %s: %v", symbol, err)
		http.Error(w, fmt.Sprintf("Failed to fetch price: %v", err), http.StatusInternalServerError)

		h.metrics.RecordRequest(ctx, r.URL.Path, time.Since(start).Seconds(), true)
		return
	}

	// Return full response from exchange API
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(exchange_resp); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to encode response")
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)

		h.metrics.RecordRequest(ctx, r.URL.Path, time.Since(start).Seconds(), true)
		return
	}

	// Успешное выполнение
	span.SetAttributes(
		attribute.Int("response.symbols_count", len(exchange_resp.Symbols)),
	)
	span.SetStatus(codes.Ok, "success")

	h.metrics.RecordRequest(ctx, r.URL.Path, time.Since(start).Seconds(), false)
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	// start := time.Now()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

	// h.metrics.RecordRequest(r.Context(), r.URL.Path, time.Since(start).Seconds(), false)
}
