package config

import (
	"os"
	"strconv"
)

type Config struct {
	ServicePort          string
	TelegramToken        string
	OTLPCollectorURL     string
	Environment          string
	MarketDataServiceURL string
	DecisionServiceURL   string
}

func Load() *Config {
	return &Config{
		ServicePort:          getEnv("SERVICE_PORT", "8080"),
		TelegramToken:        getEnv("TELEGRAM_BOT_TOKEN", ""),
		OTLPCollectorURL:     getEnv("OTEL_COLLECTOR_URL", "otel-collector:4317"),
		Environment:          getEnv("ENVIRONMENT", "development"),
		MarketDataServiceURL: getEnv("MARKET_DATA_SERVICE_URL", "http://market-data-service:8081"),
		DecisionServiceURL:   getEnv("DECISION_SERVICE_URL", "http://decision-service:8082"),
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
