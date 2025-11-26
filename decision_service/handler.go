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
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type DecisionResponse struct {
	Decision string `json:"decision"`
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

func decisionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tracer := otel.Tracer("decision-service")

	// Start a span for the entire decision process
	ctx, span := tracer.Start(ctx, "decision-process",
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
	tracer := otel.Tracer("decision-service")
	ctx, span := tracer.Start(ctx, "health-check")
	defer span.End()

	dataServiceURL := "http://localhost:8080"
	if url := os.Getenv("DATA_SERVICE_URL"); url != "" {
		dataServiceURL = url
	}

	resp, err := http.Get(dataServiceURL + "/health")
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Health check failed")
		return fmt.Errorf("failed to connect to data service: %w", err)
	}
	defer resp.Body.Close()

	span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("data service health check failed with status %d", resp.StatusCode)
		span.RecordError(err)
		span.SetStatus(codes.Error, "Health check failed")
		return err
	}

	span.SetStatus(codes.Ok, "Health check passed")
	return nil
}

func getMarketData(ctx context.Context, symbol string) (ai.MarketData, error) {
	tracer := otel.Tracer("decision-service")
	ctx, span := tracer.Start(ctx, "get-market-data",
		trace.WithAttributes(attribute.String("symbol", symbol)),
	)
	defer span.End()

	var marketData ai.MarketData

	dataServiceURL := "http://localhost:8080"
	if url := os.Getenv("DATA_SERVICE_URL"); url != "" {
		dataServiceURL = url
	}

	url := fmt.Sprintf("%s/price?symbol=%s", dataServiceURL, symbol)
	resp, err := http.Get(url)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Market data fetch failed")
		return marketData, fmt.Errorf("failed to fetch market data: %w", err)
	}
	defer resp.Body.Close()

	span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("data service returned status %d", resp.StatusCode)
		span.RecordError(err)
		span.SetStatus(codes.Error, "Market data fetch failed")
		return marketData, err
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
		span.SetStatus(codes.Error, "Market data parse failed")
		return marketData, fmt.Errorf("failed to parse exchange response: %w", err)
	}

	if exchangeResp.Status != "success" {
		err := fmt.Errorf("exchange API returned status: %s", exchangeResp.Status)
		span.RecordError(err)
		span.SetStatus(codes.Error, "Market data fetch failed")
		return marketData, err
	}

	if len(exchangeResp.Symbols) == 0 {
		err := fmt.Errorf("no symbol data found for %s", symbol)
		span.RecordError(err)
		span.SetStatus(codes.Error, "Market data fetch failed")
		return marketData, err
	}

	symbolData := exchangeResp.Symbols[0]
	price, err := strconv.ParseFloat(symbolData.Last, 64)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Market data parse failed")
		return marketData, fmt.Errorf("failed to parse price: %w", err)
	}

	marketData = ai.MarketData{
		Price:     price,
		Volume:    0,
		Timestamp: parseTimestamp(symbolData.Date),
	}

	span.SetAttributes(
		attribute.Float64("market.price", price),
		attribute.String("market.timestamp", marketData.Timestamp.String()),
	)
	span.SetStatus(codes.Ok, "Market data fetched successfully")
	return marketData, nil
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
