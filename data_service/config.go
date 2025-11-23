package main

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Port                 string
	APIKey               string
	ExchangeAPIURL       string
	OTELExporterEndpoint string
}

func LoadConfig() (*Config, error) {
	api_key := os.Getenv("EXCHANGE_API_KEY")
	if api_key == "" {
		return nil, fmt.Errorf("EXCHANGE_API_KEY environment variable is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	otel_endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if otel_endpoint == "" {
		otel_endpoint = "otel-collector:4318"
	}
	// Remove http:// or https:// prefix if present
	// otlptracehttp/otlpmetrichttp WithEndpoint expects host:port format
	for strings.HasPrefix(otel_endpoint, "http://") {
		otel_endpoint = strings.TrimPrefix(otel_endpoint, "http://")
	}
	for strings.HasPrefix(otel_endpoint, "https://") {
		otel_endpoint = strings.TrimPrefix(otel_endpoint, "https://")
	}
	// Remove trailing path if present
	if idx := strings.Index(otel_endpoint, "/"); idx != -1 {
		otel_endpoint = otel_endpoint[:idx]
	}
	// Ensure we have clean host:port format
	otel_endpoint = strings.TrimSpace(otel_endpoint)

	exchange_api_url := os.Getenv("EXCHANGE_API_URL")
	if exchange_api_url == "" {
		exchange_api_url = "https://api.freecryptoapi.com/v1/getData"
	}

	return &Config{
		Port:                 port,
		APIKey:               api_key,
		ExchangeAPIURL:       exchange_api_url,
		OTELExporterEndpoint: otel_endpoint,
	}, nil
}
