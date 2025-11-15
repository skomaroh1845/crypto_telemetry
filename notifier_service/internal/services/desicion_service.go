package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/skomaroh1845/crypto_telemetry/notifier-service/internal/http"
	"github.com/skomaroh1845/crypto_telemetry/notifier-service/internal/models"
)

// DecisionService handles HTTP communication with Decision Service
type DecisionService struct {
	baseURL string
	client  *http.Client
}

// NewDecisionService creates a new Decision service client
func NewDecisionService(baseURL string, timeout int) *DecisionService {
	return &DecisionService{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		client:  http.NewClient(timeout),
	}
}

// GetDecision fetches trading decision via HTTP
func (s *DecisionService) GetDecision(ctx context.Context, marketData *models.MarketDataResponse) (*models.DecisionResponse, error) {
	// Prepare request payload
	decisionReq := models.DecisionRequest{
		Crypto:    marketData.Crypto,
		Price:     marketData.Price,
		Currency:  marketData.Currency,
		Timestamp: marketData.Timestamp,
	}

	jsonData, err := json.Marshal(decisionReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal decision request: %w", err)
	}

	// Construct URL for decision service
	url := fmt.Sprintf("%s/api/v1/decide", s.baseURL)

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	var response models.DecisionResponse
	err = s.client.Post(ctx, url, headers, bytes.NewBuffer(jsonData), &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get decision: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("decision service error: %s", response.Error)
	}

	return &response, nil
}
