package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ExchangeClient struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

type ExchangeResponse struct {
	Status  string `json:"status"`
	Symbols []SymbolData `json:"symbols"`
}

type SymbolData struct {
	Symbol                string `json:"symbol"`
	Last                  string `json:"last"`
	LastBTC               string `json:"last_btc"`
	Lowest                string `json:"lowest"`
	Highest               string `json:"highest"`
	Date                  string `json:"date"`
	DailyChangePercentage string `json:"daily_change_percentage"`
	SourceExchange        string `json:"source_exchange"`
}

func NewExchangeClient(baseURL, apiKey string) *ExchangeClient {
	return &ExchangeClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *ExchangeClient) GetPrice(symbol string) (*ExchangeResponse, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol parameter is required")
	}
	
	url := fmt.Sprintf("%s?symbol=%s", c.baseURL, symbol)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("accept", "*/*")
	req.Header.Set("Authorization", "Bearer " + c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var exchange_resp ExchangeResponse
	if err := json.Unmarshal(body, &exchange_resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if exchange_resp.Status != "success" {
		return nil, fmt.Errorf("API returned status: %s", exchange_resp.Status)
	}

	if len(exchange_resp.Symbols) == 0 {
		return nil, fmt.Errorf("no symbols in response")
	}

	return &exchange_resp, nil
}

