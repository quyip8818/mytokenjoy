package budget

import (
	"context"
	"encoding/json"
)

type OverrunProcessor interface {
	ProcessOverrunPayload(ctx context.Context, raw json.RawMessage) error
}
