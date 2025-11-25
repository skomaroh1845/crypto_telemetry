package telegram

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/skomaroh1845/crypto_telemetry/notifier-service/internal/models"
	"github.com/skomaroh1845/crypto_telemetry/notifier-service/internal/services"
	"github.com/skomaroh1845/crypto_telemetry/notifier-service/internal/telemetry"
)

// WorkflowOrchestrator coordinates the workflow with Decision Service only
type WorkflowOrchestrator struct {
	decisionService *services.DecisionService
	telegramBot     *Bot
	metrics         *telemetry.Metrics
	tracer          trace.Tracer
}

// NewWorkflowOrchestrator creates a new workflow orchestrator
func NewWorkflowOrchestrator(
	decisionService *services.DecisionService,
	telegramBot *Bot,
	metrics *telemetry.Metrics,
) *WorkflowOrchestrator {
	return &WorkflowOrchestrator{
		decisionService: decisionService,
		telegramBot:     telegramBot,
		metrics:         metrics,
		tracer:          otel.Tracer("workflow-orchestrator"),
	}
}

// ProcessUserRequest processes a user request by calling Decision Service only
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

	msg := tgbotapi.NewMessage(chatID, "🔍 Analyzing cryptocurrency market...")

	// Send initial processing message
	if err := o.telegramBot.SendMessage(ctx, chatID, &msg); err != nil {
		span.RecordError(err)
		o.metrics.ErrorsCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("error.type", "telegram_initial")))
		return fmt.Errorf("failed to send initial message: %w", err)
	}

	// 1. Extract cryptocurrency symbol from message
	symbol := o.extractCryptoSymbol(message)
	span.SetAttributes(attribute.String("crypto.symbol", symbol))

	// 2. Get decision from Decision Service via HTTP
	decision, err := o.decisionService.GetDecision(ctx, symbol)
	if err != nil {
		span.RecordError(err)
		o.metrics.ErrorsCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("error.type", "decision_service")))

		// Send error message to user
		msg := tgbotapi.NewMessage(chatID, "❌ Failed to analyze cryptocurrency. Please try again later.")

		if sendErr := o.telegramBot.SendMessage(ctx, chatID, &msg); sendErr != nil {
			log.Printf("Failed to send error message: %v", sendErr)
		}
		return fmt.Errorf("failed to get decision: %w", err)
	}

	// 3. Format and send result

	msg = tgbotapi.NewMessage(chatID, o.formatDecisionMessage(decision))

	if err := o.telegramBot.SendMessage(ctx, chatID, &msg); err != nil {
		span.RecordError(err)
		o.metrics.ErrorsCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("error.type", "telegram_final")))
		return fmt.Errorf("failed to send decision message: %w", err)
	}

	// Record metrics
	duration := time.Since(startTime).Seconds()
	o.metrics.ProcessingDuration.Record(ctx, duration)
	o.metrics.MessagesSentCounter.Add(ctx, 2) // Initial + final message

	span.SetAttributes(
		attribute.String("decision", decision.Decision),
		attribute.Float64("processing.duration_seconds", duration),
	)

	return nil
}

// extractCryptoSymbol extracts cryptocurrency symbol from user message
func (o *WorkflowOrchestrator) extractCryptoSymbol(message string) string {
	message = strings.ToLower(message)

	symbols := map[string]string{
		"bitcoin":  "BTC",
		"btc":      "BTC",
		"ethereum": "ETH",
		"eth":      "ETH",
		"cardano":  "ADA",
		"ada":      "ADA",
		"solana":   "SOL",
		"sol":      "SOL",
		"ripple":   "XRP",
		"xrp":      "XRP",
		"dogecoin": "DOGE",
		"doge":     "DOGE",
		"litecoin": "LTC",
		"ltc":      "LTC",
	}

	for keyword, symbol := range symbols {
		if strings.Contains(message, keyword) {
			return symbol
		}
	}

	// Default to Bitcoin
	return "BTC"
}

// formatDecisionMessage formats the decision into a user-friendly message
func (o *WorkflowOrchestrator) formatDecisionMessage(decision *models.DecisionResponse) string {
	var emoji string
	switch strings.ToLower(decision.Decision) {
	case "buy", "покупать":
		emoji = "🟢 BUY"
	case "sell", "продавать":
		emoji = "🔴 SELL"
	case "hold", "держать":
		emoji = "🟡 HOLD"
	default:
		emoji = "⚪ " + strings.ToUpper(decision.Decision)
	}

	return fmt.Sprintf(
		"%s\n\n",
		emoji,
	)
}
