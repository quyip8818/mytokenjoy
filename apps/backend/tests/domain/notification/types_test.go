package notification_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/notification"
)

func TestEventResolvedCategory(t *testing.T) {
	t.Parallel()

	e := notification.Event{
		EventType: notification.EventBudgetAlertReached,
		Metadata:  notification.EventMetadata{Category: notification.CategorySecurityEvent},
	}
	if got := e.ResolvedCategory(); got != notification.CategorySecurityEvent {
		t.Fatalf("expected %s, got %s", notification.CategorySecurityEvent, got)
	}

	e2 := notification.Event{EventType: notification.EventKeyExpired}
	if got := e2.ResolvedCategory(); got != notification.CategoryKeyExpiration {
		t.Fatalf("expected %s, got %s", notification.CategoryKeyExpiration, got)
	}
}

func TestEventResolvedPriority(t *testing.T) {
	t.Parallel()

	e := notification.Event{Metadata: notification.EventMetadata{Priority: notification.PriorityCritical}}
	if got := e.ResolvedPriority(); got != notification.PriorityCritical {
		t.Fatalf("expected critical, got %s", got)
	}

	e2 := notification.Event{}
	if got := e2.ResolvedPriority(); got != notification.PriorityNormal {
		t.Fatalf("expected normal default, got %s", got)
	}
}

func TestUserPreferencesIsChannelEnabled(t *testing.T) {
	t.Parallel()

	prefs := notification.UserPreferences{
		UserID: uuid.MustParse("00000000-0000-7000-0000-000000000001"),
		Preferences: []notification.PreferenceEntry{
			{Category: notification.CategoryBudgetAlert, Channel: notification.ChannelEmail, Enabled: false},
			{Category: notification.CategoryBudgetAlert, Channel: notification.ChannelInApp, Enabled: true},
		},
	}

	if prefs.IsChannelEnabled(notification.CategoryBudgetAlert, notification.ChannelEmail) {
		t.Fatal("email should be disabled for budget_alert")
	}

	if !prefs.IsChannelEnabled(notification.CategoryBudgetAlert, notification.ChannelInApp) {
		t.Fatal("in_app should be enabled for budget_alert")
	}

	if !prefs.IsChannelEnabled(notification.CategorySecurityEvent, notification.ChannelSMS) {
		t.Fatal("sms should default to enabled for security_event")
	}

	if prefs.IsChannelEnabled(notification.CategorySystemMaintenance, notification.ChannelEmail) {
		t.Fatal("email should not be enabled by default for system_maintenance")
	}
}

func TestUserPreferencesGlobalMute(t *testing.T) {
	t.Parallel()

	prefs := notification.UserPreferences{
		UserID:     uuid.MustParse("00000000-0000-7000-0000-000000000001"),
		GlobalMute: true,
	}

	if prefs.IsChannelEnabled(notification.CategoryBudgetAlert, notification.ChannelInApp) {
		t.Fatal("global mute should disable all channels")
	}
}

func TestPriorityFallbackChain(t *testing.T) {
	t.Parallel()

	critical := notification.PriorityFallbackChain(notification.PriorityCritical)
	if len(critical) != 3 || critical[0] != notification.ChannelSMS {
		t.Fatalf("critical chain should start with sms, got %v", critical)
	}

	high := notification.PriorityFallbackChain(notification.PriorityHigh)
	if len(high) != 2 || high[0] != notification.ChannelEmail {
		t.Fatalf("high chain should start with email, got %v", high)
	}

	normal := notification.PriorityFallbackChain(notification.PriorityNormal)
	if len(normal) != 1 || normal[0] != notification.ChannelInApp {
		t.Fatalf("normal chain should be [in_app], got %v", normal)
	}
}

func TestEventCategoryMapping(t *testing.T) {
	t.Parallel()

	cases := []struct {
		eventType string
		expected  string
	}{
		{notification.EventBudgetAlertReached, notification.CategoryBudgetAlert},
		{notification.EventOverrunBlocked, notification.CategoryBudgetAlert},
		{notification.EventKeyExpired, notification.CategoryKeyExpiration},
		{notification.EventKeyExpiringSoon, notification.CategoryKeyExpiration},
		{notification.EventUsageWeeklyReport, notification.CategoryUsageReport},
		{notification.EventSecurityLoginNewDevice, notification.CategorySecurityEvent},
		{notification.EventSystemMaintenanceScheduled, notification.CategorySystemMaintenance},
	}

	for _, tc := range cases {
		got := notification.EventCategory(tc.eventType)
		if got != tc.expected {
			t.Errorf("EventCategory(%q) = %q, want %q", tc.eventType, got, tc.expected)
		}
	}
}
