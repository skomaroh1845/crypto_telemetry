package models

// MarketDataResponse response from Market Data Service
type MarketDataResponse struct {
	Crypto    string  `json:"crypto"`
	Price     float64 `json:"price"`
	Currency  string  `json:"currency"`
	Timestamp int64   `json:"timestamp"`
	Exchange  string  `json:"exchange"`
	Success   bool    `json:"success"`
	Error     string  `json:"error,omitempty"`
}

// DecisionRequest request to Decision Service
type DecisionRequest struct {
	Crypto    string  `json:"crypto"`
	Price     float64 `json:"price"`
	Currency  string  `json:"currency"`
	Timestamp int64   `json:"timestamp"`
}

// DecisionResponse response from Decision Service
type DecisionResponse struct {
	Decision   string  `json:"decision"`
	Reason     string  `json:"reason"`
	Confidence float64 `json:"confidence"`
	Timestamp  int64   `json:"timestamp"`
	Success    bool    `json:"success"`
	Error      string  `json:"error,omitempty"`
}

// TelegramMessage represents a message sent to Telegram API
type TelegramMessage struct {
	ChatID    int64  `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

// TelegramResponse represents response from Telegram API
type TelegramResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description,omitempty"`
}

// ServiceHealth represents health check response
type ServiceHealth struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Timestamp int64  `json:"timestamp"`
}
