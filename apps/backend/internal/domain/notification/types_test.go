package notification

import "testing"

func TestEventResolvedCategory(t *testing.T) {
	t.Parallel()

	// Explicit category in metadata takes precedence
	e := Event{
		EventType: EventBudgetAlertReached,
		Metadata:  EventMetadata{Category: CategorySecurityEvent},
	}
	if got := e.ResolvedCategory(); got != CategorySecurityEvent {
		t.Fatalf("expected %s, got %s", CategorySecurityEvent, got)
	}

	// Falls back to EventCategory mapping
	e2 := Event{EventType: EventKeyExpired}
	if got := e2.ResolvedCategory(); got != CategoryKeyExpiration {
		t.Fatalf("expected %s, got %s", CategoryKeyExpiration, got)
	}
}

func TestEventResolvedPriority(t *testing.T) {
	t.Parallel()

	e := Event{Metadata: EventMetadata{Priority: PriorityCritical}}
	if got := e.ResolvedPriority(); got != PriorityCritical {
		t.Fatalf("expected critical, got %s", got)
	}

	e2 := Event{}
	if got := e2.ResolvedPriority(); got != PriorityNormal {
		t.Fatalf("expected normal default, got %s", got)
	}
}

func TestUserPreferencesIsChannelEnabled(t *testing.T) {
	t.Parallel()

	prefs := UserPreferences{
		UserID: "u1",
		Preferences: []PreferenceEntry{
			{Category: CategoryBudgetAlert, Channel: ChannelEmail, Enabled: false},
			{Category: CategoryBudgetAlert, Channel: ChannelInApp, Enabled: true},
		},
	}

	// Explicit disabled
	if prefs.IsChannelEnabled(CategoryBudgetAlert, ChannelEmail) {
		t.Fatal("email should be disabled for budget_alert")
	}

	// Explicit enabled
	if !prefs.IsChannelEnabled(CategoryBudgetAlert, ChannelInApp) {
		t.Fatal("in_app should be enabled for budget_alert")
	}

	// No explicit entry → use category defaults
	// security_event defaults include sms
	if !prefs.IsChannelEnabled(CategorySecurityEvent, ChannelSMS) {
		t.Fatal("sms should default to enabled for security_event")
	}

	// system_maintenance doesn't default to email
	if prefs.IsChannelEnabled(CategorySystemMaintenance, ChannelEmail) {
		t.Fatal("email should not be enabled by default for system_maintenance")
	}
}

func TestUserPreferencesGlobalMute(t *testing.T) {
	t.Parallel()

	prefs := UserPreferences{
		UserID:     "u1",
		GlobalMute: true,
	}

	if prefs.IsChannelEnabled(CategoryBudgetAlert, ChannelInApp) {
		t.Fatal("global mute should disable all channels")
	}
}

func TestPriorityFallbackChain(t *testing.T) {
	t.Parallel()

	critical := PriorityFallbackChain(PriorityCritical)
	if len(critical) != 3 || critical[0] != ChannelSMS {
		t.Fatalf("critical chain should start with sms, got %v", critical)
	}

	high := PriorityFallbackChain(PriorityHigh)
	if len(high) != 2 || high[0] != ChannelEmail {
		t.Fatalf("high chain should start with email, got %v", high)
	}

	normal := PriorityFallbackChain(PriorityNormal)
	if len(normal) != 1 || normal[0] != ChannelInApp {
		t.Fatalf("normal chain should be [in_app], got %v", normal)
	}
}

func TestEventCategoryMapping(t *testing.T) {
	t.Parallel()

	cases := []struct {
		eventType string
		expected  string
	}{
		{EventBudgetAlertReached, CategoryBudgetAlert},
		{EventOverrunBlocked, CategoryBudgetAlert},
		{EventKeyExpired, CategoryKeyExpiration},
		{EventKeyExpiringSoon, CategoryKeyExpiration},
		{EventUsageWeeklyReport, CategoryUsageReport},
		{EventSecurityLoginNewDevice, CategorySecurityEvent},
		{EventSystemMaintenanceScheduled, CategorySystemMaintenance},
	}

	for _, tc := range cases {
		got := EventCategory(tc.eventType)
		if got != tc.expected {
			t.Errorf("EventCategory(%q) = %q, want %q", tc.eventType, got, tc.expected)
		}
	}
}
