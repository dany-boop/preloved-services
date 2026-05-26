// Package types defines shared structs used across all services.
// Import this to avoid duplicating the same types in every service.

package types

import "time"

// ──────────────────────────────────────────────
// Standard API Response wrappers
// All endpoints return one of these
// ──────────────────────────────────────────────

// Response is the standard success response envelope.
// Example: {"success": true, "data": {...}, "message": "OK"}
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse is the standard error envelope.
// Example: {"success": false, "error": "invalid credentials", "code": "AUTH_INVALID"}
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"` // machine-readable error code
}

// PaginatedResponse wraps list responses with pagination info.
type PaginatedResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Meta    Pagination  `json:"meta"`
}

// Pagination holds page info returned with list endpoints.
type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// ──────────────────────────────────────────────
// Shared domain types
// ──────────────────────────────────────────────

// UserRole defines access levels in the system.
type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
	RoleAgent UserRole = "agent" // AI Agent (Anti Gravity internal)
)

// UserStatus tracks account state.
type UserStatus string

const (
	StatusActive    UserStatus = "active"
	StatusInactive  UserStatus = "inactive"
	StatusSuspended UserStatus = "suspended"
	StatusPending   UserStatus = "pending" // email not verified yet
)

// BaseModel is embedded in all DB models for common fields.
type BaseModel struct {
	ID        string    `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// ──────────────────────────────────────────────
// RabbitMQ message payloads
// Typed messages sent between services
// ──────────────────────────────────────────────

// EmailMessage is published to the email.queue
type EmailMessage struct {
	To       string            `json:"to"`
	Template string            `json:"template"` // "welcome" | "reset_password" | "verify_email"
	Data     map[string]string `json:"data"`     // template variables
}

// NotificationMessage is published to the notification.queue
type NotificationMessage struct {
	UserID  string `json:"user_id"`
	Title   string `json:"title"`
	Body    string `json:"body"`
	Type    string `json:"type"` // "push" | "in_app" | "sms"
	Payload string `json:"payload,omitempty"` // JSON string for extra data
}

// AITaskMessage is published to the ai.tasks.queue for async AI jobs
type AITaskMessage struct {
	TaskID  string                 `json:"task_id"`
	Type    string                 `json:"type"`   // "recommendation" | "embedding" | "summarize"
	Payload map[string]interface{} `json:"payload"`
}
