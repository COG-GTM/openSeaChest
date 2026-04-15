// Package models defines the data structures for the alerting engine.
package models

import (
	"encoding/json"
	"time"
)

// ChannelType represents the type of notification channel.
type ChannelType string

const (
	ChannelTypeWebhook ChannelType = "webhook"
	ChannelTypeEmail   ChannelType = "email"
)

// NotificationChannel defines a configured notification delivery channel.
type NotificationChannel struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Type      ChannelType     `json:"channel_type"`
	Config    json.RawMessage `json:"config"`
	Enabled   bool            `json:"enabled"`
	CreatedAt time.Time       `json:"created_at"`
}

// WebhookConfig is the configuration for webhook notification channels.
type WebhookConfig struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
	Secret  string            `json:"secret,omitempty"`
}

// EmailConfig is the configuration for email notification channels.
type EmailConfig struct {
	SMTPHost     string   `json:"smtp_host"`
	SMTPPort     int      `json:"smtp_port"`
	FromAddress  string   `json:"from_address"`
	ToAddresses  []string `json:"to_addresses"`
	Username     string   `json:"username,omitempty"`
	Password     string   `json:"password,omitempty"`
	UseTLS       bool     `json:"use_tls"`
}

// NotificationPayload is the payload sent through notification channels.
type NotificationPayload struct {
	AlertID      string     `json:"alert_id"`
	RuleName     string     `json:"rule_name"`
	RuleType     RuleType   `json:"rule_type"`
	DeviceSerial string     `json:"device_serial"`
	Severity     Severity   `json:"severity"`
	State        AlertState `json:"state"`
	Message      string     `json:"message"`
	FirstSeen    time.Time  `json:"first_seen"`
	LastSeen     time.Time  `json:"last_seen"`
	Timestamp    time.Time  `json:"timestamp"`
}
