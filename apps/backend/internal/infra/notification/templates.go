package notification

// templates.go — single source of truth for all notification template IDs.
// ponytail: hardcoded, no env vars. Add new event types here.

import domainnotification "github.com/tokenjoy/backend/internal/domain/notification"

// --- SMS (Aliyun) template codes ---
// Key "" is the default/fallback (verification code).

var smsTemplates = map[string]string{
	"":                  "SMS_337285106", // default (verification code)
	"verification_code": "SMS_337285106",
	"member_invite":     "SMS_510740074",
}

// SMSTemplateCode resolves an eventType to its Aliyun template code.
// Falls back to the default ("") template if no specific mapping exists.
func SMSTemplateCode(eventType string) string {
	if code := smsTemplates[eventType]; code != "" {
		return code
	}
	return smsTemplates[""]
}

// --- Email (Resend) template aliases ---

var emailTemplates = map[string]string{
	domainnotification.EventBudgetAlertReached:    "budget-alert",
	domainnotification.EventOverrunBlocked:        "overrun-blocked",
	domainnotification.EventOverdraftExpanded:     "overrun-blocked",
	domainnotification.EventSyncThresholdExceeded: "sync-threshold-exceeded",
	"verification_code":                           "verification-code",
	"company_invite":                              "company-invite",
	"member_invite":                               "company-invite",
}

// EmailTemplateID resolves eventType to the Resend template ID.
// Returns empty string if no template is configured for the event.
func EmailTemplateID(eventType string) string {
	return emailTemplates[eventType]
}
