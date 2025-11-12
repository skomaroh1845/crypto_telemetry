package main

import (
	"encoding/json"
	"net/http"
)

type DecisionResponse struct {
	Decision string `json:"decision"`
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

func decisionHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
	resp := DecisionResponse{Decision: "buy"}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
