package main

import (
	"fmt"
	"os"
)

type Config struct {
	Port          string
	APIKey        string
	ExchangeAPIURL string
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
		otel_endpoint = "localhost:4317"
	}

	exchange_api_url := "https://api.freecryptoapi.com/v1/getData"

	return &Config{
		Port:                port,
		APIKey:              api_key,
		ExchangeAPIURL:      exchange_api_url,
		OTELExporterEndpoint: otel_endpoint,
	}, nil
}

