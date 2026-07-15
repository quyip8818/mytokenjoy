package jobs

import (
	"encoding/json"

	"github.com/riverqueue/river"
	"github.com/tokenjoy/backend/internal/config"
)

// NotificationDeliveryArgs represents a single-channel notification delivery job.
type NotificationDeliveryArgs struct {
	CompanyID   int64           `json:"company_id"`
	Channel     string          `json:"channel"`
	RecipientID string          `json:"recipient_id"`
	EventType   string          `json:"event_type"`
	Title       string          `json:"title"`
	Body        string          `json:"body"`
	Payload     json.RawMessage `json:"payload"`
	Priority    string          `json:"priority"`
}

func (NotificationDeliveryArgs) Kind() string { return KindNotificationDelivery }

func (NotificationDeliveryArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue:       config.RiverQueueDefault,
		MaxAttempts: 5,
	}
}
