package main

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/preloved-services/auth-service/config"
	"github.com/preloved-services/auth-service/internal/db"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// ── 1. Logger (set up first so every step below can log) ──
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})

	// config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("❌ Failed to load config — check your .env file")
	}
	log.Info().Err(err).Msg("Failed to load config see.env file")
	
	// seeting gin to release mode when on production
	if cfg.AppEnv == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	// db connextion
	log.Info().Msg("Connecting to posgre & redis")
	connections,err :=db.Connect(cfg.PostgresDSN,cfg.RedisAddr,cfg.RedisPassword)
	if err != nil {
		log.Fatal().Err(err).Msg("❌ Failed to connect to database — check your .env file")
	}
	defer connections.Close()
	
	log.Info().Msg("Successfully connected to database & redis")
	
	

}

// requestLogger logs every HTTP request with method, path, status, and duration.
// Shows in your terminal when running locally.
func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		log.Info().
			Str("method", c.Request.Method).
			Str("path", path).
			Int("status", c.Writer.Status()).
			Dur("duration", time.Since(start)).
			Str("ip", c.ClientIP()).
			Msg("request")

	}
}
