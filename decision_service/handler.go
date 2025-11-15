package main

import (
	"encoding/json"
	"net/http"

	"github.com/skomaroh1845/crypto_telemetry/decision_service/ai"
)

type DecisionResponse struct {
	Decision string `json:"decision"`
}

type DecisionRequest struct {
	//TODO dublicate MarketData?
	MarketData ai.MarketData `json:"market_data"`
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

	var req DecisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	aiClient := ai.NewGroqClient()
	decision, err := aiClient.GetDecision(req.MarketData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(decision)
}
