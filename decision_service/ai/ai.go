package ai

import "time"

type MarketData struct {
	Price     float64   `json:"price"`
	Volume    float64   `json:"volume"`
	Timestamp time.Time `json:"timestamp"`
}

type DecisionResponse struct {
	Decision string `json:"decision"`
}

type AIClient interface {
	GetDecision(data MarketData) (DecisionResponse, error)
}
