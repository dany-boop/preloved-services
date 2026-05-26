// Package logger provides a structured logger with correlation ID support.
// Used by ALL services — import this instead of using log.Println directly.
//
// USAGE:
//   logger.Init("debug")
//   log := logger.WithCorrelationID("req-123")
//   log.Info().Str("user_id", "456").Msg("user logged in")

package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Init sets up the global logger.
// level: "debug" | "info" | "warn" | "error"
func Init(level string) {
	// Pretty print in development, JSON in production
	if os.Getenv("APP_ENV") == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		})
	}

	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(lvl)

	// Add service name to every log line
	zerolog.DefaultContextLogger = &log.Logger
}

// WithCorrelationID returns a logger that includes the correlation ID.
// Every HTTP request should have one — generated at the gateway.
func WithCorrelationID(correlationID string) zerolog.Logger {
	return log.With().Str("correlation_id", correlationID).Logger()
}

// WithService returns a logger tagged with the service name.
// Call once at startup: logger.WithService("auth-service")
func WithService(serviceName string) zerolog.Logger {
	return log.With().Str("service", serviceName).Logger()
}
