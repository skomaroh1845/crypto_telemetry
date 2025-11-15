package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/skomaroh1845/crypto_telemetry/notifier-service/internal/config"
	"github.com/skomaroh1845/crypto_telemetry/notifier-service/internal/handlers"
	"github.com/skomaroh1845/crypto_telemetry/notifier-service/internal/orchestrator"
	"github.com/skomaroh1845/crypto_telemetry/notifier-service/internal/server"
	"github.com/skomaroh1845/crypto_telemetry/notifier-service/internal/services"
	"github.com/skomaroh1845/crypto_telemetry/notifier-service/internal/telegram"
	"github.com/skomaroh1845/crypto_telemetry/notifier-service/internal/telemetry"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	slog.SetDefault(logger)

	// Load configuration from environment variables
	cfg := config.Load()

	// Validate required configuration
	if cfg.TelegramToken == "" {
		slog.Error("TELEGRAM_BOT_TOKEN environment variable is required")
	}

	// Initialize telemetry (tracing and metrics)
	TracerProvider, MeterProvider, err := telemetry.InitTelemetry(cfg)
	if err != nil {
		slog.Error("Failed to initialize telemetry", "error", err)
	}
	defer func() {
		if err := TracerProvider.Shutdown(context.Background()); err != nil {
			slog.Error("Error shutting down tracer provider", "error", err)
		}
		if err := MeterProvider.Shutdown(context.Background()); err != nil {
			slog.Error("Error shutting down meter provider", "error", err)
		}
	}()

	// Initialize metrics
	metrics, err := telemetry.NewMetrics()
	if err != nil {
		slog.Error("Failed to initialize metrics", "error", err)
	}

	// Initialize Telegram bot
	bot, err := telegram.NewBot(cfg)
	if err != nil {
		slog.Error("Failed to initialize Telegram bot", "error", err)
	}

	// Initialize clients for other services
	marketDataClient := services.NewMarketDataService(cfg.MarketDataServiceURL, cfg.HTTPTimeout)
	decisionClient := services.NewDecisionService(cfg.DecisionServiceURL, cfg.HTTPTimeout)
	telegramService := services.NewTelegramService(cfg.TelegramToken, cfg.HTTPTimeout)

	// Initialize workflow orchestrator
	workflowOrchestrator := orchestrator.NewWorkflowOrchestrator(
		marketDataClient,
		decisionClient,
		telegramService,
		metrics,
	)

	// Initialize Telegram poller

	telegramBot, err := telegram.NewBot(cfg)
	if err != nil {
		slog.Error("Failed to initialize Telegram bot", "error", err)
	}

	telegramPoller := telegram.NewPoller(
		telegramBot,
		workflowOrchestrator,
		metrics,
	)

	if telegramPoller == nil {
		slog.Error("Failed to initialize telegramPoller", "error", err)
	}

	// Initialize handlers
	telegramHandler := handlers.NewTelegramHandler(bot, workflowOrchestrator, metrics)

	// Create and setup HTTP server
	srv := server.New(cfg)
	srv.SetupRoutes(telegramHandler, metrics)

	// Start server in a goroutine
	go func() {
		if err := srv.Start(); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	go func() {
		telegramPoller.Start(context.Background())
	}()

	slog.Info("Starting Notifier Service", "port", cfg.ServicePort, "enviroment", cfg.Environment)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down Notifier Service...")

	// Даем время для завершения текущих операций
	//ctx
	_, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// TODO: Graceful shutdown сервера и бота

	slog.Info("Notifier Service stopped")
}
