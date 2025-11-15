package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/skomaroh1845/crypto_telemetry/notifier-service/internal/http"
	"github.com/skomaroh1845/crypto_telemetry/notifier-service/internal/models"
)

// TelegramService handles HTTP communication with Telegram API using polling
type TelegramService struct {
	botToken   string
	client     *http.Client
	tracer     trace.Tracer
	lastUpdate int // Track last processed update ID
}

// NewTelegramService creates a new Telegram service client
func NewTelegramService(botToken string, timeout int) *TelegramService {
	return &TelegramService{
		botToken:   botToken,
		client:     http.NewClient(timeout),
		tracer:     otel.Tracer("telegram-service"),
		lastUpdate: 0,
	}
}

// TelegramUpdate represents an incoming update from Telegram polling
type TelegramUpdate struct {
	UpdateID int              `json:"update_id"`
	Message  *TelegramMessage `json:"message"`
}

// TelegramMessage represents a message from Telegram
type TelegramMessage struct {
	MessageID int64        `json:"message_id"`
	From      TelegramUser `json:"from"`
	Chat      TelegramChat `json:"chat"`
	Text      string       `json:"text"`
	Date      int64        `json:"date"`
}

// TelegramUser represents a Telegram user
type TelegramUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
}

// TelegramChat represents a Telegram chat
type TelegramChat struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
}

// TelegramResponse represents response from Telegram API
type TelegramResponse struct {
	OK          bool             `json:"ok"`
	Result      []TelegramUpdate `json:"result"`
	Description string           `json:"description,omitempty"`
}

// SendMessage sends a message to Telegram user via HTTP
func (s *TelegramService) SendMessage(ctx context.Context, chatID int64, text string) error {
	ctx, span := s.tracer.Start(ctx, "Telegram.SendMessage")
	defer span.End()

	// Prepare message payload
	message := models.TelegramMessage{
		ChatID:    chatID,
		Text:      text,
		ParseMode: "Markdown",
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to marshal telegram message: %w", err)
	}

	// Construct Telegram API URL
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", s.botToken)

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	var response models.TelegramResponse
	err = s.client.Post(ctx, url, headers, bytes.NewBuffer(jsonData), &response)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to send telegram message: %w", err)
	}

	if !response.OK {
		span.RecordError(fmt.Errorf(response.Description))
		return fmt.Errorf("telegram API error: %s", response.Description)
	}

	span.SetAttributes(
		attribute.Int64("telegram.chat_id", chatID),
		attribute.String("telegram.message_type", "outgoing"),
	)

	return nil
}

// GetUpdates fetches new messages from Telegram using long polling
func (s *TelegramService) GetUpdates(ctx context.Context) ([]TelegramUpdate, error) {
	ctx, span := s.tracer.Start(ctx, "Telegram.GetUpdates")
	defer span.End()

	// Construct URL with offset for getting new updates only
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?timeout=30&offset=%d",
		s.botToken, s.lastUpdate+1)

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	var response TelegramResponse
	err := s.client.Get(ctx, url, headers, &response)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get telegram updates: %w", err)
	}

	if !response.OK {
		span.RecordError(fmt.Errorf(response.Description))
		return nil, fmt.Errorf("telegram API error: %s", response.Description)
	}

	// Update last processed update ID
	if len(response.Result) > 0 {
		s.lastUpdate = response.Result[len(response.Result)-1].UpdateID
	}

	span.SetAttributes(
		attribute.Int("updates.received", len(response.Result)),
		attribute.Int("last_update_id", s.lastUpdate),
	)

	return response.Result, nil
}

// GetBotInfo retrieves basic information about the bot
func (s *TelegramService) GetBotInfo(ctx context.Context) (*TelegramUser, error) {
	ctx, span := s.tracer.Start(ctx, "Telegram.GetBotInfo")
	defer span.End()

	url := fmt.Sprintf("https://api.telegram.org/bot%s/getMe", s.botToken)
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	var response struct {
		OK     bool         `json:"ok"`
		Result TelegramUser `json:"result"`
	}

	err := s.client.Get(ctx, url, headers, &response)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get bot info: %w", err)
	}

	if !response.OK {
		return nil, fmt.Errorf("failed to get bot info")
	}

	span.SetAttributes(
		attribute.String("bot.username", response.Result.Username),
		attribute.String("bot.first_name", response.Result.FirstName),
	)

	return &response.Result, nil
}
