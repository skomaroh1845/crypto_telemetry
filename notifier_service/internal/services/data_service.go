package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/skomaroh1845/crypto_telemetry/notifier-service/internal/http"
	"github.com/skomaroh1845/crypto_telemetry/notifier-service/internal/models"
)

// MarketDataService handles HTTP communication with Market Data Service
type MarketDataService struct {
	baseURL string
	client  *http.Client
}

// NewMarketDataService creates a new Market Data service client
func NewMarketDataService(baseURL string, timeout int) *MarketDataService {
	return &MarketDataService{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		client:  http.NewClient(timeout),
	}
}

// GetCurrentPrice fetches current cryptocurrency price via HTTP
func (s *MarketDataService) GetCurrentPrice(ctx context.Context, crypto string) (*models.MarketDataResponse, error) {
	// Construct URL for market data service
	url := fmt.Sprintf("%s/api/v1/price/%s", s.baseURL, strings.ToLower(crypto))

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	var response models.MarketDataResponse
	err := s.client.Get(ctx, url, headers, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get market data: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("market data service error: %s", response.Error)
	}

	return &response, nil
}
