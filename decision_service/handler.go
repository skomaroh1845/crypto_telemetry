package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/skomaroh1845/crypto_telemetry/decision_service/ai"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("decision-service")

type DecisionResponse struct {
	Decision string `json:"decision"`
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

func decisionHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "decision-process",
		trace.WithAttributes(attribute.String("handler", "decision")),
	)
	defer span.End()

	symbol := r.URL.Query().Get("symbol")
	span.SetAttributes(attribute.String("symbol", symbol))

	if err := checkDataServiceHealth(ctx); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Health check failed")
		http.Error(w, "Market data service is unavailable: "+err.Error(), http.StatusInternalServerError)
		return
	}

	market, err := getMarketData(ctx, symbol)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Market data fetch failed")
		http.Error(w, "Failed to fetch market data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	span.SetAttributes(
		attribute.Float64("price", market.Price),
		attribute.Float64("volume", market.Volume),
	)

	aiClient := ai.NewDaniilFrolovAI()
	decision, err := aiClient.GetDecision(ctx, market)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "AI decision failed")
		http.Error(w, "Failed to get AI decision:"+err.Error(), http.StatusInternalServerError)
		return
	}

	span.SetAttributes(attribute.String("final.decision", decision.Decision))
	span.SetStatus(codes.Ok, "Decision completed successfully")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(decision)
}

func checkDataServiceHealth(ctx context.Context) error {
	//tracer := otel.Tracer("data-service")
	ctx, span := tracer.Start(ctx, "data-service.health",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	client := http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}

	dataServiceURL := os.Getenv("DATA_SERVICE_URL")
	if dataServiceURL == "" {
		dataServiceURL = "http://localhost:8080"
	}

	req, _ := http.NewRequestWithContext(ctx, "GET", dataServiceURL+"/health", nil)

	resp, err := client.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Health check failed")
		return fmt.Errorf("failed to connect to data service: %w", err)
	}
	defer resp.Body.Close()

	span.SetAttributes(attribute.Int("http.status", resp.StatusCode))

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("data service returned %d", resp.StatusCode)
		span.RecordError(err)
		span.SetStatus(codes.Error, "Health check failed")
		return err
	}

	span.SetStatus(codes.Ok, "Health check passed")
	return nil
}

func getMarketData(ctx context.Context, symbol string) (ai.MarketData, error) {
	//tracer := otel.Tracer("data-service")
	ctx, span := tracer.Start(ctx, "data-service.get-price",
		trace.WithAttributes(attribute.String("symbol", symbol)),
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	//client := http.Client{
	//	Transport: otelhttp.NewTransport(http.DefaultTransport),
	//}

	client:=http.Client { Timeout: time.Duration(5) * time.Second }

	var result ai.MarketData

	dataServiceURL := os.Getenv("DATA_SERVICE_URL")
	if dataServiceURL == "" {
		dataServiceURL = "http://localhost:8080"
	}

	url := fmt.Sprintf("%s/price?symbol=%s", dataServiceURL, symbol)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)

	resp, err := client.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Market data fetch failed")
		return result, fmt.Errorf("failed to fetch market data: %w", err)
	}
	defer resp.Body.Close()

	span.SetAttributes(attribute.Int("http.status", resp.StatusCode))

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("data service returned %d", resp.StatusCode)
		span.RecordError(err)
		span.SetStatus(codes.Error, "Market data fetch failed")
		return result, err
	}

	var exchangeResp struct {
		Status  string `json:"status"`
		Symbols []struct {
			Symbol string `json:"symbol"`
			Last   string `json:"last"`
			Date   string `json:"date"`
		} `json:"symbols"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&exchangeResp); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Parse failed")
		return result, fmt.Errorf("failed to parse exchange response: %w", err)
	}

	if exchangeResp.Status != "success" {
		err := fmt.Errorf("exchange status: %s", exchangeResp.Status)
		span.RecordError(err)
		span.SetStatus(codes.Error, "API returned error")
		return result, err
	}

	if len(exchangeResp.Symbols) == 0 {
		err := fmt.Errorf("no symbol data for %s", symbol)
		span.RecordError(err)
		span.SetStatus(codes.Error, "No data")
		return result, err
	}

	s := exchangeResp.Symbols[0]
	price, err := strconv.ParseFloat(s.Last, 64)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Price parse failed")
		return result, fmt.Errorf("failed to parse price: %w", err)
	}

	result = ai.MarketData{
		Price:     price,
		Volume:    0,
		Timestamp: parseTimestamp(s.Date),
	}

	span.SetAttributes(
		attribute.Float64("market.price", price),
		attribute.String("market.timestamp", result.Timestamp.String()),
	)

	span.SetStatus(codes.Ok, "Market data OK")
	return result, nil
}

func parseTimestamp(timestamp string) time.Time {
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timestamp); err == nil {
			return t
		}
	}

	return time.Now()
}
