// Package models defines the data structures for the alerting engine.
package models

import (
	"time"
)

// AlertState represents the current state of an alert.
type AlertState string

const (
	AlertStateOK           AlertState = "ok"
	AlertStateFiring       AlertState = "firing"
	AlertStateAcknowledged AlertState = "acknowledged"
	AlertStateResolved     AlertState = "resolved"
)

// ValidTransition checks whether a state transition is valid according to the
// alert state machine: OK -> FIRING -> ACKNOWLEDGED -> RESOLVED.
// Additionally, FIRING -> RESOLVED is allowed (condition clears without ack).
func (s AlertState) ValidTransition(next AlertState) bool {
	switch s {
	case AlertStateOK:
		return next == AlertStateFiring
	case AlertStateFiring:
		return next == AlertStateAcknowledged || next == AlertStateResolved
	case AlertStateAcknowledged:
		return next == AlertStateResolved
	case AlertStateResolved:
		return next == AlertStateFiring // can re-fire after resolution
	default:
		return false
	}
}

// Alert represents an active or historical alert instance.
type Alert struct {
	ID             string     `json:"id"`
	RuleID         string     `json:"rule_id"`
	DeviceSerial   string     `json:"device_serial"`
	State          AlertState `json:"state"`
	Message        string     `json:"message"`
	FirstSeen      time.Time  `json:"first_seen"`
	LastSeen       time.Time  `json:"last_seen"`
	AcknowledgedAt *time.Time `json:"acknowledged_at,omitempty"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty"`
	AcknowledgedBy string     `json:"acknowledged_by,omitempty"`
	ResolvedBy     string     `json:"resolved_by,omitempty"`
}

// IsActive returns true if the alert is in a FIRING or ACKNOWLEDGED state.
func (a *Alert) IsActive() bool {
	return a.State == AlertStateFiring || a.State == AlertStateAcknowledged
}

// AlertHistory records a state transition for an alert.
type AlertHistory struct {
	ID        string     `json:"id"`
	AlertID   string     `json:"alert_id"`
	OldState  AlertState `json:"old_state"`
	NewState  AlertState `json:"new_state"`
	ChangedAt time.Time  `json:"changed_at"`
	ChangedBy string     `json:"changed_by"`
}

// AlertListResponse is the paginated response for alert queries.
type AlertListResponse struct {
	Alerts     []Alert `json:"alerts"`
	TotalCount int     `json:"total_count"`
	Page       int     `json:"page"`
	PageSize   int     `json:"page_size"`
}

// AcknowledgeRequest is the request body for acknowledging an alert.
type AcknowledgeRequest struct {
	AcknowledgedBy string `json:"acknowledged_by"`
}

// ResolveRequest is the request body for manually resolving an alert.
type ResolveRequest struct {
	ResolvedBy string `json:"resolved_by"`
}
