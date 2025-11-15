package handlers

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"

	"github.com/skomaroh1845/crypto_telemetry/notifier-service/internal/orchestrator"
	"github.com/skomaroh1845/crypto_telemetry/notifier-service/internal/telegram"
	"github.com/skomaroh1845/crypto_telemetry/notifier-service/internal/telemetry"
)

// BaseHandler provides common functionality for all handlers
type BaseHandler struct {
	tracer  trace.Tracer
	metrics *telemetry.Metrics
}

// NewBaseHandler creates a new base handler
func NewBaseHandler(metrics *telemetry.Metrics) *BaseHandler {
	return &BaseHandler{
		metrics: metrics,
	}
}

// TelegramHandler handles Telegram webhook requests
type TelegramHandler struct {
	*BaseHandler
	bot          *telegram.Bot
	orchestrator *orchestrator.WorkflowOrchestrator
}

// NewTelegramHandler creates a new Telegram handler
func NewTelegramHandler(bot *telegram.Bot, orchestrator *orchestrator.WorkflowOrchestrator, metrics *telemetry.Metrics) *TelegramHandler {
	return &TelegramHandler{
		BaseHandler:  NewBaseHandler(metrics),
		bot:          bot,
		orchestrator: orchestrator,
	}
}

// HandleWebhook handles incoming Telegram webhook messages
func (h *TelegramHandler) HandleWebhook(c *gin.Context) {
	// TODO: Implement webhook handling
	c.JSON(200, gin.H{
		"status":  "ok",
		"message": "Telegram webhook endpoint is working",
	})
}
