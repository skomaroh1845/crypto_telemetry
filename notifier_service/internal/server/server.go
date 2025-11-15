package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/skomaroh1845/crypto_telemetry/notifier-service/internal/config"
	"github.com/skomaroh1845/crypto_telemetry/notifier-service/internal/handlers"
	"github.com/skomaroh1845/crypto_telemetry/notifier-service/internal/telemetry"
)

// Server represents the HTTP server
type Server struct {
	router *gin.Engine
	server *http.Server
	config *config.Config
}

// New creates a new HTTP server
func New(cfg *config.Config) *Server {
	// Set Gin mode based on environment
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	return &Server{
		router: router,
		config: cfg,
	}
}

// SetupRoutes configures all the routes for the server
func (s *Server) SetupRoutes(telegramHandler *handlers.TelegramHandler, metrics *telemetry.Metrics) {
	// Add OpenTelemetry middleware to all routes
	s.router.Use(otelgin.Middleware("notifier-service"))

	// Health check endpoint
	s.router.GET("/health", s.healthHandler)
	s.router.GET("/ready", s.readyHandler)

	// API v1 routes
	v1 := s.router.Group("/api/v1")
	{
		v1.POST("/notification", s.notificationHandler)
	}

	// Telegram webhook route
	s.router.POST("/webhook/telegram", telegramHandler.HandleWebhook)

	// Metrics endpoint for Prometheus
	//s.router.GET("/metrics", gin.WrapH(telemetry.PrometheusHandler()))
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.server = &http.Server{
		Addr:         ":" + s.config.ServicePort,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Starting HTTP server on port %s", s.config.ServicePort)
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Shutting down HTTP server...")
	return s.server.Shutdown(ctx)
}

// healthHandler handles health check requests
func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "notifier-service",
		"timestamp": time.Now().UTC(),
	})
}

// readyHandler handles readiness check requests
func (s *Server) readyHandler(c *gin.Context) {
	// TODO: Add readiness checks (database connections, etc.)
	c.JSON(http.StatusOK, gin.H{
		"status":    "ready",
		"service":   "notifier-service",
		"timestamp": time.Now().UTC(),
	})
}

// notificationHandler handles incoming notifications from other services
func (s *Server) notificationHandler(c *gin.Context) {
	// TODO: Implement notification handling logic
	c.JSON(http.StatusOK, gin.H{
		"status":  "received",
		"message": "Notification endpoint is working",
	})
}
