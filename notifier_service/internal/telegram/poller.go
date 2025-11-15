package telegram

import (
	"context"
	"log"
	"log/slog"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/skomaroh1845/crypto_telemetry/notifier-service/internal/orchestrator"
	"github.com/skomaroh1845/crypto_telemetry/notifier-service/internal/telemetry"
)

// Poller handles Telegram message polling using the bot API
type Poller struct {
	bot          *Bot
	orchestrator *orchestrator.WorkflowOrchestrator
	metrics      *telemetry.Metrics
	tracer       trace.Tracer
	isRunning    bool
}

// NewPoller creates a new Telegram poller
func NewPoller(
	bot *Bot,
	orchestrator *orchestrator.WorkflowOrchestrator,
	metrics *telemetry.Metrics,
) *Poller {
	return &Poller{
		bot:          bot,
		orchestrator: orchestrator,
		metrics:      metrics,
		tracer:       otel.Tracer("telegram-poller"),
		isRunning:    false,
	}
}

// Start begins polling for Telegram messages
func (p *Poller) Start(ctx context.Context) {
	p.isRunning = true

	slog.Info("Starting Telegram poller", "username", p.bot.api.Self.UserName)
	slog.Info("Bot is ready to receive messages...")

	// Get updates channel
	updates := p.bot.GetUpdatesChannel()

	for p.isRunning {
		select {
		case <-ctx.Done():
			p.isRunning = false
			log.Println("Telegram poller stopped via context")
			return
		case update := <-updates:
			p.processUpdate(ctx, update)
		}
	}
}

// Stop gracefully stops the poller
func (p *Poller) Stop() {
	p.isRunning = false
	p.bot.Stop()
	log.Println("Telegram poller stopped")
}

func (p *Poller) handleStart(ctx context.Context, update *tgbotapi.Update) error {
	ctx, span := p.tracer.Start(ctx, "TelegramPoller.handleStart")

	slog.Info("handling start command from user")
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Добро пожаловать в бот по рекомендации криптовалют")

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("рекомендации"),
			tgbotapi.NewKeyboardButton("помощь"),
		))

	msg.ReplyMarkup = keyboard

	if err := p.bot.SendMessage(ctx, update.Message.Chat.ID, &msg); err != nil {
		span.RecordError(err)
		log.Printf("Failed to send initial message: %v", err)
		return err
	}

	return nil
}

func (p *Poller) handleHelp(ctx context.Context, update *tgbotapi.Update) error {
	ctx, span := p.tracer.Start(ctx, "TelegramPoller.handleHelp")

	slog.Info("handling help command from user")
	msg := tgbotapi.NewMessage(update.Message.Chat.ID,
		"\\help - помошь\n\\start - старт бота\n\\advice - рекомендации к покупке",
	)

	if err := p.bot.SendMessage(ctx, update.Message.Chat.ID, &msg); err != nil {
		span.RecordError(err)
		log.Printf("Failed to send initial message: %v", err)
		return err
	}

	return nil
}

func (p *Poller) handleAdvice(ctx context.Context, update tgbotapi.Update) error {
	p.processUserRequest(ctx, update.Message.Chat.ID, update.Message.Text, update.Message.Chat.UserName)
	return nil
}

// processUpdate processes a single Telegram update
func (p *Poller) processUpdate(ctx context.Context, update tgbotapi.Update) {
	ctx, span := p.tracer.Start(ctx, "TelegramPoller.ProcessUpdate")
	defer span.End()

	// Skip if no message or text
	// update.Message.Text == ""
	if update.Message == nil {
		return
	}

	span.SetAttributes(
		attribute.Int64("telegram.chat_id", update.Message.Chat.ID),
		attribute.String("telegram.username", update.SentFrom().UserName),
		attribute.String("user.message", update.Message.Text),
		attribute.Int("update.id", update.UpdateID),
	)

	slog.Info("Received message", "username", update.SentFrom().UserName, "messageText", update.Message.Text, "command", update.Message.Command())

	if update.Message.Command() == "start" || update.Message.Text == "старт" {
		err := p.handleStart(ctx, &update)
		if err != nil {
			span.RecordError(err)
		}

		p.metrics.RequestsCounter.Add(ctx, 1)
		return
	} else if update.Message.Command() == "help" || update.Message.Text == "помощь" {
		err := p.handleHelp(ctx, &update)
		if err != nil {
			span.RecordError(err)
		}

		p.metrics.RequestsCounter.Add(ctx, 1)
		return
	} else if update.Message.Command() == "advice" || update.Message.Text == "рекомендации" {
		go func() {
			err := p.handleAdvice(ctx, update)
			if err != nil {
				span.RecordError(err)
			}

			p.metrics.RequestsCounter.Add(ctx, 1)
		}()

		return
	}

	// Process the user request in a goroutine to avoid blocking the poller
	//go p.processUserRequest(ctx, chatID, messageText, username)
}

// processUserRequest processes the user request asynchronously
func (p *Poller) processUserRequest(ctx context.Context, chatID int64, messageText string, username string) {
	ctx, span := p.tracer.Start(ctx, "TelegramPoller.ProcessUserRequest")
	defer span.End()

	// Process the user request
	if err := p.orchestrator.ProcessUserRequest(ctx, chatID, messageText); err != nil {
		span.RecordError(err)
		log.Printf("Failed to process user request from @%s: %v", username, err)

		// Send error message to user
		//errorMsg := "❌ Sorry, I encountered an error while processing your request. Please try again later."
		//if sendErr := p.bot.SendMessage(ctx, chatID, errorMsg); sendErr != nil {
		//	log.Printf("Failed to send error message: %v", sendErr)
		//}
	}
}

// isCryptoAnalysisCommand checks if the message is a crypto analysis command
func (p *Poller) isCryptoAnalysisCommand(message string) bool {
	triggers := []string{
		"крипт", "crypto", "биткоин", "bitcoin", "эфириум", "ethereum",
		"цена", "price", "анализ", "analysis", "что делать", "what should",
		"купить", "buy", "продать", "sell", "btc", "eth", "ada", "sol",
		"cardano", "solana", "investment", "инвестиц", "совет", "advice",
		"should i", "what about", "recommend", "recommendation",
	}

	message = strings.ToLower(message)
	for _, trigger := range triggers {
		if strings.Contains(message, trigger) {
			return true
		}
	}
	return false
}
