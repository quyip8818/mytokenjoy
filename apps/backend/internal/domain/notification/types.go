// Package notification defines domain types for the notification module.
package notification

import

// --- Channels ---
"github.com/google/uuid"

const (
	ChannelEmail   = "email"
	ChannelSMS     = "sms"
	ChannelInApp   = "in_app"
	ChannelLog     = "log"
	ChannelWebhook = "webhook"
)

// AllChannels returns all supported delivery channels.
func AllChannels() []string {
	return []string{ChannelEmail, ChannelSMS, ChannelInApp, ChannelLog, ChannelWebhook}
}

// UserFacingChannels returns channels that users can configure in preferences.
func UserFacingChannels() []string {
	return []string{ChannelEmail, ChannelSMS, ChannelInApp}
}

// --- Priority ---

const (
	PriorityCritical = "critical"
	PriorityHigh     = "high"
	PriorityNormal   = "normal"
	PriorityLow      = "low"
)

// PriorityFallbackChain returns the channel fallback order for a given priority.
func PriorityFallbackChain(priority string) []string {
	switch priority {
	case PriorityCritical:
		return []string{ChannelSMS, ChannelEmail, ChannelInApp}
	case PriorityHigh:
		return []string{ChannelEmail, ChannelInApp}
	case PriorityNormal:
		return []string{ChannelInApp}
	case PriorityLow:
		return []string{ChannelInApp}
	default:
		return []string{ChannelInApp}
	}
}

// --- Categories ---

const (
	CategoryBudgetAlert       = "budget_alert"
	CategoryKeyExpiration     = "key_expiration"
	CategoryUsageReport       = "usage_report"
	CategorySecurityEvent     = "security_event"
	CategorySystemMaintenance = "system_maintenance"
	CategoryOverrun           = "overrun"
)

// AllCategories returns all notification categories.
func AllCategories() []string {
	return []string{
		CategoryBudgetAlert,
		CategoryKeyExpiration,
		CategoryUsageReport,
		CategorySecurityEvent,
		CategorySystemMaintenance,
		CategoryOverrun,
	}
}

// CategoryDefaultChannels returns the default enabled channels for a category.
func CategoryDefaultChannels(category string) []string {
	switch category {
	case CategoryBudgetAlert:
		return []string{ChannelEmail, ChannelInApp}
	case CategoryKeyExpiration:
		return []string{ChannelEmail, ChannelInApp}
	case CategoryUsageReport:
		return []string{ChannelEmail, ChannelInApp}
	case CategorySecurityEvent:
		return []string{ChannelEmail, ChannelSMS, ChannelInApp}
	case CategorySystemMaintenance:
		return []string{ChannelInApp}
	case CategoryOverrun:
		return []string{ChannelEmail, ChannelInApp}
	default:
		return []string{ChannelInApp}
	}
}

// --- Event Types ---

const (
	EventSyncThresholdExceeded      = "sync_threshold_exceeded"
	EventOverrunBlocked             = "overrun_blocked"
	EventOverdraftExpanded          = "overdraft_expanded"
	EventBudgetAlertReached         = "budget_alert_reached"
	EventKeyExpired                 = "key_expired"
	EventKeyExpiringSoon            = "key_expiring_soon"
	EventUsageWeeklyReport          = "usage_weekly_report"
	EventSecurityLoginNewDevice     = "security_login_new_device"
	EventSystemMaintenanceScheduled = "system_maintenance_scheduled"
)

// EventCategory maps an event type to its category.
func EventCategory(eventType string) string {
	switch eventType {
	case EventSyncThresholdExceeded, EventOverrunBlocked, EventOverdraftExpanded, EventBudgetAlertReached:
		return CategoryBudgetAlert
	case EventKeyExpired, EventKeyExpiringSoon:
		return CategoryKeyExpiration
	case EventUsageWeeklyReport:
		return CategoryUsageReport
	case EventSecurityLoginNewDevice:
		return CategorySecurityEvent
	case EventSystemMaintenanceScheduled:
		return CategorySystemMaintenance
	default:
		return CategoryBudgetAlert // safe default
	}
}

// --- Status ---

const (
	StatusPending = "pending"
	StatusSent    = "sent"
	StatusFailed  = "failed"
	StatusRead    = "read"
)

// --- NotificationEvent (trigger payload, channel-agnostic) ---

type Event struct {
	EventType   string
	RecipientID uuid.UUID
	CompanyID   uuid.UUID
	Payload     map[string]any
	Metadata    EventMetadata
}

type EventMetadata struct {
	DeduplicationKey string
	Priority         string
	GroupKey         string
	Category         string
}

// ResolvedCategory returns the category from metadata, or infers it from event type.
func (e Event) ResolvedCategory() string {
	if e.Metadata.Category != "" {
		return e.Metadata.Category
	}
	return EventCategory(e.EventType)
}

// ResolvedPriority returns the priority from metadata, or defaults to normal.
func (e Event) ResolvedPriority() string {
	if e.Metadata.Priority != "" {
		return e.Metadata.Priority
	}
	return PriorityNormal
}

// --- RenderedMessage ---

type RenderedMessage struct {
	Title   string
	Body    string
	Payload map[string]any
}

// --- User Preference ---

type PreferenceEntry struct {
	Category string
	Channel  string
	Enabled  bool
}

type UserPreferences struct {
	UserID      uuid.UUID
	Preferences []PreferenceEntry
	GlobalMute  bool
	QuietHours  *QuietHours
}

type QuietHours struct {
	Start    string // HH:mm
	End      string // HH:mm
	Timezone string
}

// IsChannelEnabled checks if a specific channel is enabled for a category.
// Returns the default if no explicit preference is set.
func (p UserPreferences) IsChannelEnabled(category, channel string) bool {
	if p.GlobalMute {
		return false
	}
	for _, entry := range p.Preferences {
		if entry.Category == category && entry.Channel == channel {
			return entry.Enabled
		}
	}
	// Default: check if channel is in the category's default channels
	for _, ch := range CategoryDefaultChannels(category) {
		if ch == channel {
			return true
		}
	}
	return false
}
