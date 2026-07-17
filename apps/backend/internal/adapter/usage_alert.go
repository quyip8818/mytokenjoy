package adapter

import (
	"context"

	"github.com/google/uuid"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	domainnotification "github.com/tokenjoy/backend/internal/domain/notification"
	"github.com/tokenjoy/backend/internal/infra/notification"
)

// budgetAlertPublisher adapts budget.AlertPublisher to infra/notification.Service.
type budgetAlertPublisher struct {
	svc *notification.Service
}

// NewBudgetAlertPublisher creates an AlertPublisher backed by the notification service.
func NewBudgetAlertPublisher(svc *notification.Service) domainbudget.AlertPublisher {
	if svc == nil {
		return domainbudget.NoopAlertPublisher
	}
	return &budgetAlertPublisher{svc: svc}
}

func (p *budgetAlertPublisher) PublishBudgetAlerts(ctx context.Context, alerts []domainbudget.BudgetAlertEvent) error {
	for _, alert := range alerts {
		recipientID, err := uuid.Parse(alert.RecipientID)
		if err != nil {
			continue
		}
		event := domainnotification.Event{
			EventType:   domainnotification.EventBudgetAlertReached,
			RecipientID: recipientID,
			CompanyID:   alert.CompanyID,
			Payload: map[string]any{
				"departmentId": alert.DepartmentID,
				"nodeName":     alert.NodeName,
				"ruleId":       alert.RuleID,
				"threshold":    alert.Threshold,
				"currentPct":   alert.CurrentPct,
				"consumed":     alert.Consumed,
				"budget":       alert.Budget,
				"periodKey":    alert.PeriodKey,
			},
			Metadata: domainnotification.EventMetadata{
				Priority:         domainnotification.PriorityNormal,
				Category:         domainnotification.CategoryBudgetAlert,
				DeduplicationKey: alert.DedupeKey,
			},
		}
		// DispatchAsync enqueues via River — failures are non-fatal.
		if err := p.svc.DispatchAsync(ctx, event); err != nil {
			continue
		}
	}
	return nil
}

var _ domainbudget.AlertPublisher = (*budgetAlertPublisher)(nil)
