package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/skomaroh1845/crypto_telemetry/notifier-service/internal/http"
	"github.com/skomaroh1845/crypto_telemetry/notifier-service/internal/models"
)

// DecisionService handles HTTP communication with Decision Service
type DecisionService struct {
	baseURL string
	client  *http.Client
	tracer  trace.Tracer
}

// NewDecisionService creates a new Decision service client
func NewDecisionService(baseURL string, timeout int) *DecisionService {
	return &DecisionService{
		baseURL: strings.TrimSuffix(baseURL, "/decision_service:8081"),
		client:  http.NewClient(timeout),
		tracer:  otel.Tracer("decision-service"),
	}
}

// GetDecision fetches trading decision via HTTP using symbol query parameter
func (s *DecisionService) GetDecision(ctx context.Context, symbol string) (*models.DecisionResponse, error) {
	ctx, span := s.tracer.Start(ctx, "DecisionService.GetDecision")
	defer span.End()

	decisionReq := models.DecisionRequest{
		Symbol: symbol,
	}

	jsonData, err := json.Marshal(decisionReq)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Construct URL with symbol query parameter
	//decisionURL := fmt.Sprintf("%s/decision", s.baseURL)
	decisionURL := fmt.Sprintf("http://decision_service:8081/decision?symbol=%s", symbol)

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	var response models.DecisionResponse
	err = s.client.Post(ctx, decisionURL, headers, bytes.NewBuffer(jsonData), &response)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get decision: %w", err)
	}

	span.SetAttributes(
		attribute.String("symbol", symbol),
		attribute.String("decision", response.Decision),
	)

	return &response, nil
}

// HealthCheck checks if the decision service is healthy
func (s *DecisionService) HealthCheck(ctx context.Context) error {
	ctx, span := s.tracer.Start(ctx, "DecisionService.HealthCheck")
	defer span.End()

	healthURL := fmt.Sprintf("%s/health", s.baseURL)
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	var healthResponse models.HealthResponse
	err := s.client.Get(ctx, healthURL, headers, &healthResponse)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("decision service health check failed: %w", err)
	}

	if healthResponse.Status != "healthy" {
		return fmt.Errorf("decision service is not healthy: %s", healthResponse.Status)
	}

	span.SetAttributes(attribute.String("health.status", healthResponse.Status))
	return nil
}
