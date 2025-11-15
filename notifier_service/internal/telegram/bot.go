package telegram

import (
	"context"
	"fmt"
	"log"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/skomaroh1845/crypto_telemetry/notifier-service/internal/config"
)

// Bot wraps the telegram-bot-api with OpenTelemetry instrumentation
type Bot struct {
	api    *tgbotapi.BotAPI
	config *config.Config
	tracer trace.Tracer
}

// NewBot creates a new Telegram bot instance
func NewBot(cfg *config.Config) (*Bot, error) {
	// Initialize the bot with the token
	api, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		slog.Error("Failed to create bot", "token", cfg.TelegramToken)

		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	bot := &Bot{
		api:    api,
		config: cfg,
		tracer: otel.Tracer("telegram-bot"),
	}

	log.Printf("Authorized on account %s", bot.api.Self.UserName)
	return bot, nil
}

// SendMessage sends a message to a Telegram user with OpenTelemetry instrumentation
func (b *Bot) SendMessage(ctx context.Context, chatID int64, msg *tgbotapi.MessageConfig) error {
	ctx, span := b.tracer.Start(ctx, "Telegram.SendMessage")
	defer span.End()

	// Send the message
	_, err := b.api.Send(*msg)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to send message: %w", err)
	}

	span.SetAttributes(
		attribute.Int64("telegram.chat_id", chatID),
		attribute.String("telegram.message_type", "outgoing"),
	)

	return nil
}

// GetUpdatesChannel returns a channel for receiving updates (for polling)
func (b *Bot) GetUpdatesChannel() tgbotapi.UpdatesChannel {
	// Configure update polling
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30 // Timeout in seconds for long polling

	return b.api.GetUpdatesChan(u)
}

// Stop stops the bot and cleans up resources
func (b *Bot) Stop() {
	b.api.StopReceivingUpdates()
}
