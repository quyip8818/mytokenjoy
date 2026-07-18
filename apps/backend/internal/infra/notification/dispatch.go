package notification

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	domainnotification "github.com/tokenjoy/backend/internal/domain/notification"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/infra/jobs"
)

// Dispatch is the main entry point for the notification pipeline.
func (s *Service) Dispatch(ctx context.Context, event domainnotification.Event) error {
	// 1. Render the message
	msg := s.renderer.Render(event)

	// 2. Load user preferences
	prefs := s.loadPreferences(ctx, event.RecipientID)

	// 2a. Check quiet hours — only critical notifications bypass
	if event.ResolvedPriority() != domainnotification.PriorityCritical && IsInQuietHours(prefs.QuietHours) {
		s.logger.Debug("notification suppressed by quiet hours",
			"event", event.EventType,
			"recipient", event.RecipientID)
		return nil
	}

	// 3. Resolve target channels (preferences ∩ configured ∩ fallback chain ∩ rate limits)
	channels := s.resolveChannels(event, prefs)

	if len(channels) == 0 {
		s.logger.Debug("no channels resolved for notification",
			"event", event.EventType,
			"recipient", event.RecipientID)
		return nil
	}

	// 4. Deliver to each resolved channel
	var lastErr error
	for _, ch := range channels {
		if err := ch.Send(ctx, event.RecipientID, msg); err != nil {
			s.logger.Warn("channel delivery failed",
				"channel", ch.Name(),
				"event", event.EventType,
				"recipient", event.RecipientID,
				"error", err)
			s.recordFailure(ctx, ch.Name(), event, msg, err)
			lastErr = err
			continue
		}
		s.logger.Debug("notification delivered",
			"channel", ch.Name(),
			"event", event.EventType,
			"recipient", event.RecipientID)
	}

	// Log channel always fires for audit trail (separate from user-facing channels)
	logCh, ok := s.registry.Get(domainnotification.ChannelLog)
	if ok {
		_ = logCh.Send(ctx, event.RecipientID, msg)
	}

	return lastErr
}

// DispatchAsync enqueues notification delivery as background jobs via RiverQueue.
// Falls back to synchronous Dispatch if no enqueuer is configured.
func (s *Service) DispatchAsync(ctx context.Context, event domainnotification.Event) error {
	if s.enqueuer == nil {
		return s.Dispatch(ctx, event)
	}

	// Load preferences for quiet hours and channel resolution
	prefs := s.loadPreferences(ctx, event.RecipientID)

	// Check quiet hours — only critical notifications bypass
	if event.ResolvedPriority() != domainnotification.PriorityCritical && IsInQuietHours(prefs.QuietHours) {
		s.logger.Debug("notification suppressed by quiet hours (async)",
			"event", event.EventType,
			"recipient", event.RecipientID)
		return nil
	}

	// Render the message
	msg := s.renderer.Render(event)

	// Resolve channels
	channels := s.resolveChannels(event, prefs)

	if len(channels) == 0 {
		return nil
	}

	payload, _ := json.Marshal(msg.Payload)

	// Enqueue a delivery job per channel
	for _, ch := range channels {
		args := jobs.NotificationDeliveryArgs{
			CompanyID:   event.CompanyID,
			Channel:     ch.Name(),
			RecipientID: event.RecipientID,
			EventType:   event.EventType,
			Title:       msg.Title,
			Body:        msg.Body,
			Payload:     payload,
			Priority:    event.ResolvedPriority(),
		}
		if err := jobs.InsertNotificationDelivery(ctx, s.enqueuer, args); err != nil {
			s.logger.Warn("failed to enqueue notification delivery",
				"channel", ch.Name(),
				"event", event.EventType,
				"error", err)
			// Fall back to synchronous delivery for this channel
			_ = ch.Send(ctx, event.RecipientID, msg)
		}
	}

	// Log channel fires immediately (audit trail, not user-facing)
	logCh, ok := s.registry.Get(domainnotification.ChannelLog)
	if ok {
		_ = logCh.Send(ctx, event.RecipientID, msg)
	}

	return nil
}

// loadPreferences loads user notification preferences, falling back to defaults on error.
func (s *Service) loadPreferences(ctx context.Context, userID uuid.UUID) domainnotification.UserPreferences {
	entries, err := s.store.NotificationPreference().Get(ctx, userID)
	if err != nil {
		s.logger.Debug("failed to load notification preferences, using defaults",
			"user", userID, "error", err)
		return defaultPreferences(userID)
	}

	prefs := domainnotification.UserPreferences{
		UserID:      userID,
		Preferences: make([]domainnotification.PreferenceEntry, len(entries)),
	}
	for i, e := range entries {
		prefs.Preferences[i] = domainnotification.PreferenceEntry{
			Category: e.Category,
			Channel:  e.Channel,
			Enabled:  e.Enabled,
		}
	}
	return prefs
}

func defaultPreferences(userID uuid.UUID) domainnotification.UserPreferences {
	return domainnotification.UserPreferences{UserID: userID}
}

// resolveChannels determines which channels to deliver to based on:
// 1. The event's priority fallback chain
// 2. User preferences (category × channel)
// 3. Channel configuration status (IsConfigured)
// 4. Rate limits (per user per channel)
func (s *Service) resolveChannels(event domainnotification.Event, prefs domainnotification.UserPreferences) []Channel {
	if prefs.GlobalMute {
		return nil
	}

	category := event.ResolvedCategory()
	priority := event.ResolvedPriority()
	fallbackChain := domainnotification.PriorityFallbackChain(priority)

	var result []Channel
	for _, chName := range fallbackChain {
		if chName == domainnotification.ChannelLog {
			continue
		}
		if !prefs.IsChannelEnabled(category, chName) {
			continue
		}

		ch, ok := s.registry.Get(chName)
		if !ok || !ch.IsConfigured() {
			continue
		}

		// Apply rate limit for SMS and Email channels
		rateLimitKey := event.RecipientID.String() + ":" + chName
		switch chName {
		case domainnotification.ChannelSMS:
			if !s.smsRateLimiter.Allow(rateLimitKey) {
				s.logger.Debug("sms rate limited", "recipient", event.RecipientID)
				continue
			}
		case domainnotification.ChannelEmail:
			if !s.emailRateLimiter.Allow(rateLimitKey) {
				s.logger.Debug("email rate limited", "recipient", event.RecipientID)
				continue
			}
		}

		result = append(result, ch)
	}

	// For critical priority with no channels resolved, force in_app as last resort
	if len(result) == 0 && priority == domainnotification.PriorityCritical {
		if ch, ok := s.registry.Get(domainnotification.ChannelInApp); ok && ch.IsConfigured() {
			result = append(result, ch)
		}
	}

	return result
}

// recordFailure logs a delivery failure to the notification_log.
func (s *Service) recordFailure(ctx context.Context, channelName string, event domainnotification.Event, msg domainnotification.RenderedMessage, deliveryErr error) {
	payload, _ := json.Marshal(msg.Payload)
	entry := types.NotificationLogEntry{
		ID:        uuid.Must(uuid.NewV7()),
		Channel:   channelName,
		EventType: event.EventType,
		UserID:    event.RecipientID,
		Title:     msg.Title,
		Body:      msg.Body,
		Payload:   payload,
		Status:    types.NotificationStatusFailed,
		Error:     deliveryErr.Error(),
	}
	_ = s.store.Notification().Append(ctx, entry)
}
