package auth

import (
	"context"
)

const (
	DefaultContextKey = "default"
)

// AuditLogger interface for audit logging
type AuditLogger interface {
	LogEvent(ctx context.Context, event AuditEvent) error
}

// AuditEvent represents an authentication audit event
type AuditEvent struct {
	EventType string            `json:"event_type"`
	UserID    string            `json:"user_id"`
	ClientIP  string            `json:"client_ip"`
	UserAgent string            `json:"user_agent"`
	Success   bool              `json:"success"`
	Error     string            `json:"error,omitempty"`
	Timestamp string            `json:"timestamp"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// Option function type for configuring authentication services
type Option func(interface{}) error
