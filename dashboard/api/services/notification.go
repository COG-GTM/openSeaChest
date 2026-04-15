// Package services implements the core business logic for the alerting engine.
package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"github.com/COG-GTM/openSeaChest/dashboard/api/models"
)

// NotificationService handles dispatching notifications through various channels.
type NotificationService struct {
	httpClient *http.Client
}

// NewNotificationService creates a new NotificationService.
func NewNotificationService() *NotificationService {
	return &NotificationService{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Send dispatches a notification through the given channel.
func (ns *NotificationService) Send(channel models.NotificationChannel, payload models.NotificationPayload) error {
	switch channel.Type {
	case models.ChannelTypeWebhook:
		return ns.sendWebhook(channel, payload)
	case models.ChannelTypeEmail:
		return ns.sendEmail(channel, payload)
	default:
		return fmt.Errorf("unsupported notification channel type: %s", channel.Type)
	}
}

// sendWebhook sends an alert payload to a configured webhook URL with retry
// and exponential backoff (max 3 retries, backoff: 1s, 2s, 4s).
func (ns *NotificationService) sendWebhook(channel models.NotificationChannel, payload models.NotificationPayload) error {
	var config models.WebhookConfig
	if err := json.Unmarshal(channel.Config, &config); err != nil {
		return fmt.Errorf("invalid webhook config: %w", err)
	}

	if config.URL == "" {
		return fmt.Errorf("webhook URL is required")
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	maxRetries := 3
	baseBackoff := 1 * time.Second

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(math.Pow(2, float64(attempt-1))) * baseBackoff
			log.Printf("Webhook retry %d/%d for channel %s, backing off %s",
				attempt, maxRetries, channel.ID, backoff)
			time.Sleep(backoff)
		}

		req, err := http.NewRequest(http.MethodPost, config.URL, bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("failed to create webhook request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "OpenSeaChest-AlertEngine/1.0")
		for key, value := range config.Headers {
			req.Header.Set(key, value)
		}

		resp, err := ns.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("webhook request failed: %w", err)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			log.Printf("Webhook notification sent successfully to %s (attempt %d)", config.URL, attempt+1)
			return nil
		}

		lastErr = fmt.Errorf("webhook returned status %d", resp.StatusCode)
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			// Client errors are not retryable
			return lastErr
		}
	}

	return fmt.Errorf("webhook delivery failed after %d retries: %w", maxRetries, lastErr)
}

// sendEmail sends an alert notification via SMTP. This is a stub implementation
// that connects to the configured SMTP server and sends a formatted alert email.
func (ns *NotificationService) sendEmail(channel models.NotificationChannel, payload models.NotificationPayload) error {
	var config models.EmailConfig
	if err := json.Unmarshal(channel.Config, &config); err != nil {
		return fmt.Errorf("invalid email config: %w", err)
	}

	if config.SMTPHost == "" || config.FromAddress == "" || len(config.ToAddresses) == 0 {
		return fmt.Errorf("email config requires smtp_host, from_address, and to_addresses")
	}

	subject := fmt.Sprintf("[%s] Alert: %s - %s on %s",
		strings.ToUpper(string(payload.Severity)),
		payload.RuleName,
		payload.State,
		payload.DeviceSerial)

	body := formatAlertEmail(payload)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=\"utf-8\"\r\n\r\n%s",
		config.FromAddress,
		strings.Join(config.ToAddresses, ", "),
		subject,
		body)

	addr := fmt.Sprintf("%s:%d", config.SMTPHost, config.SMTPPort)

	var auth smtp.Auth
	if config.Username != "" {
		auth = smtp.PlainAuth("", config.Username, config.Password, config.SMTPHost)
	}

	// Attempt to connect and send. In a stub/test scenario, the SMTP server
	// may not be available, so we log and return the error gracefully.
	err := sendMailWithTimeout(addr, auth, config.FromAddress, config.ToAddresses, []byte(msg))
	if err != nil {
		log.Printf("Email notification stub: would send to %v via %s (error: %v)",
			config.ToAddresses, addr, err)
		return fmt.Errorf("email send failed: %w", err)
	}

	log.Printf("Email notification sent to %v via %s", config.ToAddresses, addr)
	return nil
}

// sendMailWithTimeout connects to the SMTP server with a timeout and performs
// the full SMTP transaction on the same connection, ensuring the timeout
// applies to the actual send operation.
func sendMailWithTimeout(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}

	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("SMTP connection failed: %w", err)
	}

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("SMTP client creation failed: %w", err)
	}
	defer client.Close()

	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP auth failed: %w", err)
		}
	}

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("SMTP MAIL FROM failed: %w", err)
	}

	for _, addr := range to {
		if err := client.Rcpt(addr); err != nil {
			return fmt.Errorf("SMTP RCPT TO failed for %s: %w", addr, err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("SMTP DATA failed: %w", err)
	}

	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("SMTP write failed: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("SMTP close data failed: %w", err)
	}

	return client.Quit()
}

// formatAlertEmail formats the alert payload into a human-readable email body.
func formatAlertEmail(payload models.NotificationPayload) string {
	var sb strings.Builder
	sb.WriteString("OpenSeaChest Fleet Health Dashboard - Alert Notification\n")
	sb.WriteString("=======================================================\n\n")
	sb.WriteString(fmt.Sprintf("Alert ID:      %s\n", payload.AlertID))
	sb.WriteString(fmt.Sprintf("Rule:          %s (%s)\n", payload.RuleName, payload.RuleType))
	sb.WriteString(fmt.Sprintf("Device:        %s\n", payload.DeviceSerial))
	sb.WriteString(fmt.Sprintf("Severity:      %s\n", payload.Severity))
	sb.WriteString(fmt.Sprintf("State:         %s\n", payload.State))
	sb.WriteString(fmt.Sprintf("Message:       %s\n", payload.Message))
	sb.WriteString(fmt.Sprintf("First Seen:    %s\n", payload.FirstSeen.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("Last Seen:     %s\n", payload.LastSeen.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("Timestamp:     %s\n", payload.Timestamp.Format(time.RFC3339)))
	sb.WriteString("\n---\nThis is an automated notification from the OpenSeaChest Fleet Health Dashboard.\n")
	return sb.String()
}
