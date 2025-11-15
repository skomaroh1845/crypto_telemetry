// Create ai/groq.go
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

type GroqClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

type GroqRequest struct {
	Model    string        `json:"model"`
	Messages []GroqMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type GroqMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type GroqResponse struct {
	Choices []GroqChoice `json:"choices"`
}

type GroqChoice struct {
	Message GroqMessage `json:"message"`
}

func NewGroqClient() *GroqClient {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		// Groq sometimes allows requests without API key for testing
		fmt.Println("Warning: GROQ_API_KEY environment variable not set")
	}

	return &GroqClient{
		apiKey:  apiKey,
		baseURL: "https://api.groq.com/openai/v1",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *GroqClient) GetDecision(data MarketData) (DecisionResponse, error) {
	prompt := c.createTradingPrompt(data)

	groqReq := GroqRequest{
		Model: "llama3-8b-8192", // Free model, very fast
		Messages: []GroqMessage{
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

	response, err := c.makeRequest(groqReq)
	if err != nil {
		return DecisionResponse{}, fmt.Errorf("failed to call Groq API: %w", err)
	}

	decision, err := c.parseDecision(response)
	if err != nil {
		return DecisionResponse{}, fmt.Errorf("failed to parse AI decision: %w", err)
	}

	return DecisionResponse{Decision: decision}, nil
}

func (c *GroqClient) createTradingPrompt(data MarketData) string {
	return fmt.Sprintf(`
Analyze this market data and provide a trading decision:

Current Price: $%.2f
Volume: %.2f
Timestamp: %s

Based on this data, should I buy, sell, or hold? Respond with only one word: buy, sell, or hold.
    `, data.Price, data.Volume, data.Timestamp.Format(time.RFC3339))
}

func (c *GroqClient) makeRequest(req GroqRequest) (*GroqResponse, error) {
	requestBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var groqResp GroqResponse
	if err := json.NewDecoder(resp.Body).Decode(&groqResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &groqResp, nil
}

func (c *GroqClient) parseDecision(response *GroqResponse) (string, error) {
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

func (c *GroqClient) normalizeDecision(decision string) string {
	decision = strings.TrimSpace(decision)
	decision = strings.ToLower(decision)
	decision = strings.TrimRight(decision, ".,!?")
	return decision
}
