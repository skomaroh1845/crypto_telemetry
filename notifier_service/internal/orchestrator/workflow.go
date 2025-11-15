package orchestrator

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/skomaroh1845/crypto_telemetry/notifier-service/internal/models"
	"github.com/skomaroh1845/crypto_telemetry/notifier-service/internal/services"
	"github.com/skomaroh1845/crypto_telemetry/notifier-service/internal/telemetry"
)

// WorkflowOrchestrator coordinates the workflow between services via HTTP
type WorkflowOrchestrator struct {
	marketDataService *services.MarketDataService
	decisionService   *services.DecisionService
	telegramService   *services.TelegramService
	metrics           *telemetry.Metrics
	tracer            trace.Tracer
}

// NewWorkflowOrchestrator creates a new workflow orchestrator
func NewWorkflowOrchestrator(
	marketDataService *services.MarketDataService,
	decisionService *services.DecisionService,
	telegramService *services.TelegramService,
	metrics *telemetry.Metrics,
) *WorkflowOrchestrator {
	return &WorkflowOrchestrator{
		marketDataService: marketDataService,
		decisionService:   decisionService,
		telegramService:   telegramService,
		metrics:           metrics,
		tracer:            otel.Tracer("workflow-orchestrator"),
	}
}

// ProcessUserRequest processes a user request through the entire workflow via HTTP calls
func (o *WorkflowOrchestrator) ProcessUserRequest(ctx context.Context, chatID int64, message string) error {
	ctx, span := o.tracer.Start(ctx, "WorkflowOrchestrator.ProcessUserRequest")
	defer span.End()

	startTime := time.Now()

	// Increment requests counter
	o.metrics.RequestsCounter.Add(ctx, 1)

	span.SetAttributes(
		attribute.Int64("telegram.chat_id", chatID),
		attribute.String("user.message", message),
	)

	// Send initial processing message
	if err := o.telegramService.SendMessage(ctx, chatID, "🔍 Analyzing cryptocurrency market..."); err != nil {
		span.RecordError(err)
		o.metrics.ErrorsCounter.Add(ctx, 1)
		return fmt.Errorf("failed to send initial message: %w", err)
	}

	// 1. Extract cryptocurrency from message
	crypto := o.extractCryptoFromMessage(message)
	span.SetAttributes(attribute.String("crypto.selected", crypto))

	// 2. Get current price via HTTP
	marketData, err := o.marketDataService.GetCurrentPrice(ctx, crypto)
	if err != nil {
		span.RecordError(err)
		o.metrics.ErrorsCounter.Add(ctx, 1)
		o.telegramService.SendMessage(ctx, chatID, "❌ Failed to fetch market data. Please try again later.")
		return fmt.Errorf("failed to get market data: %w", err)
	}

	// 3. Get decision via HTTP
	decision, err := o.decisionService.GetDecision(ctx, marketData)
	if err != nil {
		span.RecordError(err)
		o.metrics.ErrorsCounter.Add(ctx, 1)
		o.telegramService.SendMessage(ctx, chatID, "❌ Failed to analyze market. Please try again later.")
		return fmt.Errorf("failed to get decision: %w", err)
	}

	// 4. Format and send result
	responseText := o.formatDecisionMessage(marketData, decision)
	if err := o.telegramService.SendMessage(ctx, chatID, responseText); err != nil {
		span.RecordError(err)
		o.metrics.ErrorsCounter.Add(ctx, 1)
		return fmt.Errorf("failed to send decision message: %w", err)
	}

	// Record metrics
	duration := time.Since(startTime).Seconds()
	o.metrics.ProcessingDuration.Record(ctx, duration)
	o.metrics.MessagesSentCounter.Add(ctx, 2) // Initial + final message

	span.SetAttributes(
		attribute.String("crypto", marketData.Crypto),
		attribute.Float64("price", marketData.Price),
		attribute.String("decision", decision.Decision),
		attribute.Float64("confidence", decision.Confidence),
		attribute.Float64("processing.duration_seconds", duration),
	)

	return nil
}

// extractCryptoFromMessage extracts cryptocurrency symbol from user message
func (o *WorkflowOrchestrator) extractCryptoFromMessage(message string) string {
	message = strings.ToLower(message)

	cryptos := map[string]string{
		"bitcoin":  "BTC",
		"btc":      "BTC",
		"ethereum": "ETH",
		"eth":      "ETH",
		"cardano":  "ADA",
		"ada":      "ADA",
		"solana":   "SOL",
		"sol":      "SOL",
	}

	for keyword, symbol := range cryptos {
		if strings.Contains(message, keyword) {
			return symbol
		}
	}

	// Default to Bitcoin
	return "BTC"
}

// formatDecisionMessage formats the decision into a user-friendly message
func (o *WorkflowOrchestrator) formatDecisionMessage(marketData *models.MarketDataResponse, decision *models.DecisionResponse) string {
	var emoji string
	switch strings.ToLower(decision.Decision) {
	case "buy", "покупать":
		emoji = "🟢"
	case "sell", "продавать":
		emoji = "🔴"
	case "hold", "держать":
		emoji = "🟡"
	default:
		emoji = "⚪"
	}

	return fmt.Sprintf(
		"%s *Analysis Result* (%s)\n\n"+
			"💵 Current Price: $%.2f\n"+
			"📊 Decision: %s\n"+
			"🎯 Confidence: %.0f%%\n"+
			"📝 Reasoning: %s",
		emoji,
		marketData.Crypto,
		marketData.Price,
		decision.Decision,
		decision.Confidence*100,
		decision.Reason,
	)
}
