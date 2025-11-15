package config

import (
	"log/slog"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServicePort          string
	TelegramToken        string
	OTLPCollectorURL     string
	Environment          string
	MarketDataServiceURL string
	DecisionServiceURL   string
	WebhookURL           string
	HTTPTimeout          int
	PollInterval         time.Duration // Polling interval for Telegram
}

func Load() *Config {
	slog.Info("TELEGRAM_BOT_TOKEN", "TELEGRAM_BOT_TOKEN", getEnv("TELEGRAM_BOT_TOKEN", ""))

	return &Config{
		ServicePort:          getEnv("SERVICE_PORT", "8080"),
		TelegramToken:        getEnv("TELEGRAM_BOT_TOKEN", ""),
		OTLPCollectorURL:     getEnv("OTEL_COLLECTOR_URL", "otel-collector:4317"),
		Environment:          getEnv("ENVIRONMENT", "development"),
		MarketDataServiceURL: getEnv("MARKET_DATA_SERVICE_URL", "http://market-data-service:8081"),
		DecisionServiceURL:   getEnv("DECISION_SERVICE_URL", "http://decision-service:8082"),
		WebhookURL:           getEnv("WEBHOOK_URL", ""),
		HTTPTimeout:          getEnvAsInt("HTTP_TIMEOUT", 10),
		PollInterval:         getEnvAsDuration("POLL_INTERVAL", 2*time.Second),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// Helper function to get environment variable as duration
func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
