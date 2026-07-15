package notification

import (
	"fmt"

	domainnotification "github.com/tokenjoy/backend/internal/domain/notification"
)

// Renderer converts a notification Event into a RenderedMessage for channel delivery.
type Renderer struct{}

func NewRenderer() *Renderer {
	return &Renderer{}
}

// Render produces a RenderedMessage from a domain event.
// It extracts title/body from payload if present, or generates defaults.
func (r *Renderer) Render(event domainnotification.Event) domainnotification.RenderedMessage {
	title := extractString(event.Payload, "title")
	body := extractString(event.Payload, "body")

	if title == "" {
		title = defaultTitle(event.EventType)
	}
	if body == "" {
		body = extractString(event.Payload, "message")
	}

	// Include eventType in the payload for downstream channels
	enrichedPayload := make(map[string]any, len(event.Payload)+1)
	for k, v := range event.Payload {
		enrichedPayload[k] = v
	}
	enrichedPayload["eventType"] = event.EventType

	return domainnotification.RenderedMessage{
		Title:   title,
		Body:    body,
		Payload: enrichedPayload,
	}
}

func extractString(payload map[string]any, key string) string {
	if v, ok := payload[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func defaultTitle(eventType string) string {
	switch eventType {
	case domainnotification.EventBudgetAlertReached:
		return "Budget Alert"
	case domainnotification.EventOverrunBlocked:
		return "Overrun Blocked"
	case domainnotification.EventOverdraftExpanded:
		return "Overdraft Expanded"
	case domainnotification.EventSyncThresholdExceeded:
		return "Sync Threshold Exceeded"
	case domainnotification.EventKeyExpired:
		return "Key Expired"
	case domainnotification.EventKeyExpiringSoon:
		return "Key Expiring Soon"
	case domainnotification.EventUsageWeeklyReport:
		return "Weekly Usage Report"
	case domainnotification.EventSecurityLoginNewDevice:
		return "New Device Login"
	case domainnotification.EventSystemMaintenanceScheduled:
		return "System Maintenance"
	default:
		return fmt.Sprintf("Notification: %s", eventType)
	}
}
