package store

import "encoding/json"

// RiverJobView is a test-only shape for assertions against river_job rows.
type RiverJobView struct {
	ID        string
	Kind      string
	Payload   json.RawMessage
	Status    string
	LastError *string
}
