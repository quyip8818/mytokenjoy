package notification

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/tokenjoy/backend/internal/config"
	domainnotification "github.com/tokenjoy/backend/internal/domain/notification"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/store"
)

// Service is the notification dispatcher. It resolves user preferences,
// selects channels, renders messages, and dispatches delivery.
type Service struct {
	cfg              config.Config
	store            store.Store
	logger           *slog.Logger
	registry         *Registry
	renderer         *Renderer
	hub              *SSEHub
	enqueuer         jobs.Enqueuer
	resolver         *RecipientResolver
	smsRateLimiter   *RateLimiter
	emailRateLimiter *RateLimiter
}

// NewService creates a notification service with the channel registry auto-populated from config.
func NewService(cfg config.Config, st store.Store, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}

	hub := NewSSEHub()
	registry := NewRegistry(logger)
	resolver := NewRecipientResolver(st)

	// Always register log and in_app channels
	registry.Register(NewLogChannel(logger))
	registry.Register(NewInAppChannel(st, logger, hub))

	// Conditionally register webhook channel
	registry.Register(NewWebhookChannel(cfg.NotifyWebhookURL, logger))

	// Conditionally register email channel (Resend)
	registry.Register(NewEmailChannel(EmailConfig{
		APIKey: cfg.ResendAPIKey,
		From:   cfg.ResendFrom,
	}, resolver, logger))

	// Conditionally register SMS channel (Aliyun)
	registry.Register(NewSMSChannel(SMSConfig{
		AccessKeyID:     cfg.AliyunSMSAccessKeyID,
		AccessKeySecret: cfg.AliyunSMSAccessKeySecret,
		SignName:        cfg.AliyunSMSSignName,
		TemplateCode:    cfg.AliyunSMSTemplateCode,
		Endpoint:        cfg.AliyunSMSEndpoint,
	}, resolver, logger))

	return &Service{
		cfg:              cfg,
		store:            st,
		logger:           logger,
		registry:         registry,
		renderer:         NewRenderer(),
		hub:              hub,
		enqueuer:         nil, // set via SetEnqueuer after river client is created
		resolver:         resolver,
		smsRateLimiter:   DefaultSMSRateLimiter(),
		emailRateLimiter: DefaultEmailRateLimiter(),
	}
}

// Hub returns the SSE hub for real-time notification push.
func (s *Service) Hub() *SSEHub { return s.hub }

// Registry returns the channel registry (for capabilities queries).
func (s *Service) Registry() *Registry { return s.registry }

// SetEnqueuer sets the job enqueuer for async delivery (called after river client init).
func (s *Service) SetEnqueuer(e jobs.Enqueuer) { s.enqueuer = e }

// SendDirect delivers a message directly to an address (phone/email) via the named channel,
// bypassing recipient resolution and user preferences. Used for verification codes, invites, etc.
func (s *Service) SendDirect(ctx context.Context, channel string, address string, msg domainnotification.RenderedMessage) error {
	ch, ok := s.registry.Get(channel)
	if !ok || !ch.IsConfigured() {
		s.logger.Warn("SendDirect: channel not available", "channel", channel)
		return fmt.Errorf("notification: channel %q not configured", channel)
	}
	direct, ok := ch.(DirectSender)
	if !ok {
		return fmt.Errorf("notification: channel %q does not support direct send", channel)
	}
	return direct.SendDirect(ctx, address, msg)
}

// --- types.Notifier interface (backward-compatible) ---

var _ types.Notifier = (*Service)(nil)

// Send implements types.Notifier for backward compatibility with existing callers.
// It converts the simple Notification into a domain Event and dispatches.
func (s *Service) Send(ctx context.Context, notification types.Notification) error {
	event := domainnotification.Event{
		EventType:   notification.EventType,
		RecipientID: notification.RecipientID,
		CompanyID:   store.CompanyID(ctx),
		Payload:     notification.Payload,
		Metadata: domainnotification.EventMetadata{
			Priority: domainnotification.PriorityNormal,
		},
	}
	return s.Dispatch(ctx, event)
}
