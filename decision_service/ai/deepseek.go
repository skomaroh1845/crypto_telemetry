package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type DeepSeekClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

type DeepSeekRequest struct {
	Model    string            `json:"model"`
	Messages []DeepSeekMessage `json:"messages"`
	Stream   bool              `json:"stream"`
}

type DeepSeekMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type DeepSeekResponse struct {
	Choices []DeepSeekChoice `json:"choices"`
}

type DeepSeekChoice struct {
	Message DeepSeekMessage `json:"message"`
}

func NewDeepSeekClient() *DeepSeekClient {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		fmt.Println("Warning: DEEPSEEK_API_KEY environment variable not set")
	}

	baseURL := os.Getenv("DEEPSEEK_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.deepseek.com/v1"
	}

	return &DeepSeekClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second, //TODO
		},
	}
}

func (c *DeepSeekClient) GetDecision(data MarketData) (DecisionResponse, error) {
	prompt := c.createTradingPrompt(data)

	deepSeekReq := DeepSeekRequest{
		Model: "deepseek-chat",
		Messages: []DeepSeekMessage{
			{
				Role:    "system",
				Content: "You are a professional trading analyst. Analyze the given market data and provide ONLY a single word decision: 'buy', 'sell', or 'hold'. Do not provide any explanations or additional text.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Stream: false,
	}

	fmt.Println("Hello there!")
	response, err := c.makeRequest(deepSeekReq)
	if err != nil {
		return DecisionResponse{}, fmt.Errorf("failed to call DeepSeek API: %w", err)
	}

	decision, err := c.parseDecision(response)
	if err != nil {
		return DecisionResponse{}, fmt.Errorf("failed to parse AI decision: %w", err)
	}

	return DecisionResponse{Decision: decision}, nil
}

func (c *DeepSeekClient) createTradingPrompt(data MarketData) string {
	return fmt.Sprintf(`
Analyze this market data and provide a trading decision:

Current Price: $%.2f
Volume: %.2f
Timestamp: %s

Based on this data, should I buy, sell, or hold? Respond with only one word: buy, sell, or hold.
    `, data.Price, data.Volume, data.Timestamp.Format(time.RFC3339))
}

func (c *DeepSeekClient) makeRequest(req DeepSeekRequest) (*DeepSeekResponse, error) {
	requestBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	fmt.Println("I am in 109")
	httpReq, err := http.NewRequest("POST", c.baseURL+"/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	fmt.Println("I am in 114")
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()
	fmt.Println("I am in 123")
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Println(string(body))
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	fmt.Println("I am in 128")
	var deepSeekResp DeepSeekResponse
	if err := json.NewDecoder(resp.Body).Decode(&deepSeekResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	fmt.Println("I am in 133")
	return &deepSeekResp, nil
}

func (c *DeepSeekClient) parseDecision(response *DeepSeekResponse) (string, error) {
	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	decision := response.Choices[0].Message.Content
	decision = c.normalizeDecision(decision)

	switch decision {
	case "buy", "sell", "hold":
		return decision, nil
	default:
		return "", fmt.Errorf("invalid decision received: %s", decision)
	}
}

func (c *DeepSeekClient) normalizeDecision(decision string) string {
	decision = strings.TrimSpace(decision)
	decision = strings.ToLower(decision)

	decision = strings.TrimRight(decision, ".,!?")

	return decision
}
