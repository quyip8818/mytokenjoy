package workers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/riverqueue/river"
	"github.com/tokenjoy/backend/internal/domain/company"
	domainnotification "github.com/tokenjoy/backend/internal/domain/notification"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/infra/notification"
)

// NotificationDeliveryWorker handles async notification delivery through a specific channel.
type NotificationDeliveryWorker struct {
	river.WorkerDefaults[jobs.NotificationDeliveryArgs]
	registry *notification.Registry
}

func NewNotificationDeliveryWorker(registry *notification.Registry) *NotificationDeliveryWorker {
	return &NotificationDeliveryWorker{registry: registry}
}

func (w *NotificationDeliveryWorker) Work(ctx context.Context, job *river.Job[jobs.NotificationDeliveryArgs]) error {
	args := job.Args

	// Set company context for store operations
	ctx = company.WithDefaultCompany(ctx, args.CompanyID)

	ch, ok := w.registry.Get(args.Channel)
	if !ok {
		return river.JobCancel(fmt.Errorf("channel %q not registered", args.Channel))
	}
	if !ch.IsConfigured() {
		return river.JobCancel(fmt.Errorf("channel %q not configured", args.Channel))
	}

	var payload map[string]any
	if len(args.Payload) > 0 {
		_ = json.Unmarshal(args.Payload, &payload)
	}
	if payload == nil {
		payload = make(map[string]any)
	}
	payload["eventType"] = args.EventType

	msg := domainnotification.RenderedMessage{
		Title:   args.Title,
		Body:    args.Body,
		Payload: payload,
	}

	return ch.Send(ctx, args.RecipientID.String(), msg)
}
