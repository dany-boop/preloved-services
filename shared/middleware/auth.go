// Package middleware provides reusable Gin middleware for all services.
//
// USAGE in any service:
//   router := gin.Default()
//   router.Use(middleware.CORS(cfg.AllowedOrigins))
//
//   protected := router.Group("/api/v1")
//   protected.Use(middleware.Auth(cfg.JWTSecret))
//   protected.GET("/profile", handler.GetProfile)

package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/antigravity/shared/jwt"
	"github.com/antigravity/shared/types"
	"github.com/google/uuid"
)

// ──────────────────────────────────────────────
// Auth Middleware
// ──────────────────────────────────────────────

// Auth validates the Bearer JWT token on protected routes.
// On success: sets "user_id", "email", "role" in the Gin context.
// On failure: returns 401 and stops the request.
func Auth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, types.ErrorResponse{
				Error: "authorization header required",
				Code:  "AUTH_MISSING",
			})
			c.Abort()
			return
		}

		// Must be "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, types.ErrorResponse{
				Error: "invalid authorization format, use: Bearer <token>",
				Code:  "AUTH_FORMAT",
			})
			c.Abort()
			return
		}

		// Validate the JWT
		claims, err := jwt.ValidateToken(parts[1], jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, types.ErrorResponse{
				Error: "invalid or expired token",
				Code:  "AUTH_INVALID",
			})
			c.Abort()
			return
		}

		// Store user info in context for handlers to use
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// RequireRole checks that the authenticated user has the required role.
// Always use AFTER Auth() middleware.
//
// Example: route.Use(middleware.Auth(secret), middleware.RequireRole("admin"))
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, types.ErrorResponse{
				Error: "role not found in context",
				Code:  "AUTH_NO_ROLE",
			})
			c.Abort()
			return
		}

		for _, r := range roles {
			if userRole == r {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, types.ErrorResponse{
			Error: "insufficient permissions",
			Code:  "AUTH_FORBIDDEN",
		})
		c.Abort()
	}
}

// ──────────────────────────────────────────────
// CORS Middleware
// ──────────────────────────────────────────────

// CORS sets headers to allow cross-origin requests.
// allowedOrigins: slice of allowed origins e.g. ["http://localhost:3000"]
func CORS(allowedOrigins []string) gin.HandlerFunc {
	originsMap := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		originsMap[o] = true
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if originsMap[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Correlation-ID")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// ──────────────────────────────────────────────
// Correlation ID Middleware
// ──────────────────────────────────────────────

// CorrelationID ensures every request has a correlation ID.
// If KrakenD passes one via header, use it. Otherwise generate a new one.
// This ID is propagated to all downstream services for tracing.
func CorrelationID() gin.HandlerFunc {
	return func(c *gin.Context) {
		correlationID := c.GetHeader("X-Correlation-ID")
		if correlationID == "" {
			correlationID = uuid.New().String()
		}

		// Make it available to handlers
		c.Set("correlation_id", correlationID)

		// Echo it back in the response header
		c.Header("X-Correlation-ID", correlationID)

		c.Next()
	}
}
