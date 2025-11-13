package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/skomaroh1845/crypto_telemetry/notifier-service/internal/config"
	"github.com/skomaroh1845/crypto_telemetry/notifier-service/internal/telemetry"
)

func main() {
	// create basic structured logger

	// TODO: add output source
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// load config
	cfg := config.Load()

	// OTLP init
	tp, err := telemetry.InitTracer(cfg)
	if err != nil {
		slog.Error("Failed to initialize telemetry: %v", err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			slog.Error("Error shutting down tracer provider: %v", err)
		}
	}()

	// TODO: init Telegram bot
	// TODO: init HTTP server
	// TODO: init handler

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
