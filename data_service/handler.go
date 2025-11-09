package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Handler struct {
	exchangeClient *ExchangeClient
}

func NewHandler(exchangeClient *ExchangeClient) *Handler {
	return &Handler{
		exchangeClient: exchangeClient,
	}
}

func (h *Handler) GetPrice(w http.ResponseWriter, r *http.Request) {
	// Get symbol from query parameter, default to BTC if not provided
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		symbol = "BTC"
	}

	// Fetch price data from exchange API
	exchange_resp, err := h.exchangeClient.GetPrice(symbol)
	if err != nil {
		log.Printf("Error fetching price for symbol %s: %v", symbol, err)
		http.Error(w, fmt.Sprintf("Failed to fetch price: %v", err), http.StatusInternalServerError)
		return
	}

	// Return full response from exchange API
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(exchange_resp); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

