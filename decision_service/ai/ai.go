package ai

import "decision-service/handler"

type AIClient interface {
	GetDecision(data handler.MarketData) (handler.DecisionResponse, error)
}
