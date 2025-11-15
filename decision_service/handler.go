package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/skomaroh1845/crypto_telemetry/decision_service/ai"
)

type DecisionResponse struct {
	Decision string `json:"decision"`
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

func decisionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	symbol := r.URL.Query().Get("symbol")
	if err := checkDataServiceHealth(); err != nil {
		http.Error(w, "Market data service is unavailable: "+err.Error(), http.StatusInternalServerError)
		return
	}

	market, err := getMarketData(symbol)
	if err != nil {
		http.Error(w, "Failed to fetch market data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	aiClient := ai.NewDaniilFrolovAI()
	decision, err := aiClient.GetDecision(market)
	if err != nil {
		http.Error(w, "Failed to get AI decision:"+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(decision)
}

func checkDataServiceHealth() error {
	dataServiceURL := "http://localhost:8080" //TODO take from config
	if url := os.Getenv("DATA_SERVICE_URL"); url != "" {
		dataServiceURL = url
	}

	resp, err := http.Get(dataServiceURL + "/health")
	if err != nil {
		return fmt.Errorf("failed to connect to data service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("data service health check failed with status %d", resp.StatusCode)
	}

	return nil
}

func getMarketData(symbol string) (ai.MarketData, error) {
	var marketData ai.MarketData

	dataServiceURL := "http://localhost:8080" //TODO take from config
	if url := os.Getenv("DATA_SERVICE_URL"); url != "" {
		dataServiceURL = url
	}

	url := fmt.Sprintf("%s/price?symbol=%s", dataServiceURL, symbol)
	resp, err := http.Get(url)
	if err != nil {
		return marketData, fmt.Errorf("failed to fetch market data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return marketData, fmt.Errorf("data service returned status %d", resp.StatusCode)
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
		return marketData, fmt.Errorf("failed to parse exchange response: %w", err)
	}

	if exchangeResp.Status != "success" {
		return marketData, fmt.Errorf("exchange API returned status: %s", exchangeResp.Status)
	}

	if len(exchangeResp.Symbols) == 0 {
		return marketData, fmt.Errorf("no symbol data found for %s", symbol)
	}

	symbolData := exchangeResp.Symbols[0]
	price, err := strconv.ParseFloat(symbolData.Last, 64)
	if err != nil {
		return marketData, fmt.Errorf("failed to parse price: %w", err)
	}

	marketData = ai.MarketData{
		Price:     price,
		Volume:    0,
		Timestamp: parseTimestamp(symbolData.Date),
	}

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
