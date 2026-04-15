// Package services implements the core business logic for the alerting engine.
package services

import (
	"fmt"
	"sync"
	"time"

	"github.com/COG-GTM/openSeaChest/dashboard/api/models"
	"github.com/google/uuid"
)

// Store provides an in-memory data store for alert rules, alerts, alert history,
// and notification channels. In production this would be backed by PostgreSQL.
type Store struct {
	mu       sync.RWMutex
	rules    map[string]*models.AlertRule
	alerts   map[string]*models.Alert
	history  []models.AlertHistory
	channels map[string]*models.NotificationChannel

	// deduplication tracking: key is "rule_id:device_serial"
	activeAlertIndex map[string]string // dedup key -> alert ID

	// temperature debounce tracking: key is "rule_id:device_serial"
	tempConsecutive map[string]int
}

// NewStore creates a new in-memory store.
func NewStore() *Store {
	return &Store{
		rules:            make(map[string]*models.AlertRule),
		alerts:           make(map[string]*models.Alert),
		history:          make([]models.AlertHistory, 0),
		channels:         make(map[string]*models.NotificationChannel),
		activeAlertIndex: make(map[string]string),
		tempConsecutive:  make(map[string]int),
	}
}

// --- Alert Rule CRUD ---

// CreateRule creates a new alert rule and returns it.
func (s *Store) CreateRule(req models.CreateRuleRequest) (*models.AlertRule, error) {
	if !models.IsValidRuleType(req.RuleType) {
		return nil, fmt.Errorf("invalid rule type: %s", req.RuleType)
	}
	if req.Name == "" {
		return nil, fmt.Errorf("rule name is required")
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	now := time.Now().UTC()
	rule := &models.AlertRule{
		ID:                   uuid.New().String(),
		Name:                 req.Name,
		RuleType:             req.RuleType,
		Config:               req.Config,
		Severity:             req.Severity,
		Enabled:              enabled,
		NotificationChannels: req.NotificationChannels,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	s.mu.Lock()
	s.rules[rule.ID] = rule
	s.mu.Unlock()

	return rule, nil
}

// GetRule returns a rule by ID.
func (s *Store) GetRule(id string) (*models.AlertRule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rule, ok := s.rules[id]
	if !ok {
		return nil, fmt.Errorf("rule not found: %s", id)
	}
	return rule, nil
}

// ListRules returns all alert rules.
func (s *Store) ListRules() []models.AlertRule {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rules := make([]models.AlertRule, 0, len(s.rules))
	for _, r := range s.rules {
		rules = append(rules, *r)
	}
	return rules
}

// UpdateRule updates an existing alert rule.
func (s *Store) UpdateRule(id string, req models.UpdateRuleRequest) (*models.AlertRule, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rule, ok := s.rules[id]
	if !ok {
		return nil, fmt.Errorf("rule not found: %s", id)
	}

	if req.Name != nil {
		rule.Name = *req.Name
	}
	if req.Config != nil {
		rule.Config = *req.Config
	}
	if req.Severity != nil {
		rule.Severity = *req.Severity
	}
	if req.Enabled != nil {
		rule.Enabled = *req.Enabled
	}
	if req.NotificationChannels != nil {
		rule.NotificationChannels = req.NotificationChannels
	}
	rule.UpdatedAt = time.Now().UTC()

	return rule, nil
}

// DeleteRule deletes an alert rule by ID.
func (s *Store) DeleteRule(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.rules[id]; !ok {
		return fmt.Errorf("rule not found: %s", id)
	}
	delete(s.rules, id)
	return nil
}

// --- Alert Operations ---

// dedupKey returns the deduplication key for a rule+device combination.
func dedupKey(ruleID, deviceSerial string) string {
	return ruleID + ":" + deviceSerial
}

// GetActiveAlertForDevice returns the active alert for a given rule and device, if any.
func (s *Store) GetActiveAlertForDevice(ruleID, deviceSerial string) *models.Alert {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := dedupKey(ruleID, deviceSerial)
	alertID, ok := s.activeAlertIndex[key]
	if !ok {
		return nil
	}
	alert, ok := s.alerts[alertID]
	if !ok || !alert.IsActive() {
		return nil
	}
	return alert
}

// CreateOrDeduplicateAlert creates a new alert or updates last_seen if a duplicate exists.
// Returns the alert and a boolean indicating whether it was newly created.
func (s *Store) CreateOrDeduplicateAlert(ruleID, deviceSerial, message string) (*models.Alert, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := dedupKey(ruleID, deviceSerial)
	now := time.Now().UTC()

	// Check for existing active alert (deduplication)
	if alertID, ok := s.activeAlertIndex[key]; ok {
		if existing, ok := s.alerts[alertID]; ok && existing.IsActive() {
			existing.LastSeen = now
			if message != "" {
				existing.Message = message
			}
			return existing, false, nil
		}
	}

	// Create new alert
	alert := &models.Alert{
		ID:           uuid.New().String(),
		RuleID:       ruleID,
		DeviceSerial: deviceSerial,
		State:        models.AlertStateFiring,
		Message:      message,
		FirstSeen:    now,
		LastSeen:     now,
	}

	s.alerts[alert.ID] = alert
	s.activeAlertIndex[key] = alert.ID

	// Record history
	s.history = append(s.history, models.AlertHistory{
		ID:        uuid.New().String(),
		AlertID:   alert.ID,
		OldState:  models.AlertStateOK,
		NewState:  models.AlertStateFiring,
		ChangedAt: now,
		ChangedBy: "system",
	})

	return alert, true, nil
}

// AcknowledgeAlert transitions an alert from FIRING to ACKNOWLEDGED.
func (s *Store) AcknowledgeAlert(alertID, acknowledgedBy string) (*models.Alert, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	alert, ok := s.alerts[alertID]
	if !ok {
		return nil, fmt.Errorf("alert not found: %s", alertID)
	}

	if !alert.State.ValidTransition(models.AlertStateAcknowledged) {
		return nil, fmt.Errorf("cannot acknowledge alert in state %s", alert.State)
	}

	oldState := alert.State
	now := time.Now().UTC()
	alert.State = models.AlertStateAcknowledged
	alert.AcknowledgedAt = &now
	alert.AcknowledgedBy = acknowledgedBy

	s.history = append(s.history, models.AlertHistory{
		ID:        uuid.New().String(),
		AlertID:   alertID,
		OldState:  oldState,
		NewState:  models.AlertStateAcknowledged,
		ChangedAt: now,
		ChangedBy: acknowledgedBy,
	})

	return alert, nil
}

// ResolveAlert transitions an alert to RESOLVED.
func (s *Store) ResolveAlert(alertID, resolvedBy string) (*models.Alert, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	alert, ok := s.alerts[alertID]
	if !ok {
		return nil, fmt.Errorf("alert not found: %s", alertID)
	}

	if !alert.State.ValidTransition(models.AlertStateResolved) {
		return nil, fmt.Errorf("cannot resolve alert in state %s", alert.State)
	}

	oldState := alert.State
	now := time.Now().UTC()
	alert.State = models.AlertStateResolved
	alert.ResolvedAt = &now
	alert.ResolvedBy = resolvedBy

	// Remove from active index
	key := dedupKey(alert.RuleID, alert.DeviceSerial)
	delete(s.activeAlertIndex, key)

	s.history = append(s.history, models.AlertHistory{
		ID:        uuid.New().String(),
		AlertID:   alertID,
		OldState:  oldState,
		NewState:  models.AlertStateResolved,
		ChangedAt: now,
		ChangedBy: resolvedBy,
	})

	return alert, nil
}

// ListActiveAlerts returns all alerts in FIRING or ACKNOWLEDGED state.
func (s *Store) ListActiveAlerts() []models.Alert {
	s.mu.RLock()
	defer s.mu.RUnlock()

	active := make([]models.Alert, 0)
	for _, alert := range s.alerts {
		if alert.IsActive() {
			active = append(active, *alert)
		}
	}
	return active
}

// ListAlertHistory returns paginated alert history.
func (s *Store) ListAlertHistory(page, pageSize int) ([]models.Alert, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	all := make([]models.Alert, 0, len(s.alerts))
	for _, alert := range s.alerts {
		all = append(all, *alert)
	}

	total := len(all)
	start := (page - 1) * pageSize
	if start >= total {
		return []models.Alert{}, total
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	return all[start:end], total
}

// GetAlert returns an alert by ID.
func (s *Store) GetAlert(id string) (*models.Alert, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	alert, ok := s.alerts[id]
	if !ok {
		return nil, fmt.Errorf("alert not found: %s", id)
	}
	return alert, nil
}

// --- Temperature Debounce ---

// IncrementTempCount increments the consecutive temperature violation count
// for a given rule+device and returns the new count.
func (s *Store) IncrementTempCount(ruleID, deviceSerial string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := dedupKey(ruleID, deviceSerial)
	s.tempConsecutive[key]++
	return s.tempConsecutive[key]
}

// ResetTempCount resets the consecutive temperature violation count.
func (s *Store) ResetTempCount(ruleID, deviceSerial string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := dedupKey(ruleID, deviceSerial)
	delete(s.tempConsecutive, key)
}

// --- Notification Channels ---

// CreateChannel creates a new notification channel.
func (s *Store) CreateChannel(ch *models.NotificationChannel) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.channels[ch.ID] = ch
}

// GetChannel returns a notification channel by ID.
func (s *Store) GetChannel(id string) (*models.NotificationChannel, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ch, ok := s.channels[id]
	if !ok {
		return nil, fmt.Errorf("notification channel not found: %s", id)
	}
	return ch, nil
}

// ListChannels returns all notification channels.
func (s *Store) ListChannels() []models.NotificationChannel {
	s.mu.RLock()
	defer s.mu.RUnlock()

	chs := make([]models.NotificationChannel, 0, len(s.channels))
	for _, ch := range s.channels {
		chs = append(chs, *ch)
	}
	return chs
}
