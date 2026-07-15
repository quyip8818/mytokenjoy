package usage

import (
	"encoding/json"
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

func TestShouldEnqueueOverrun(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		summaries     []store.CombinedKeySummary
		platformKeyID string
		want          bool
	}{
		{
			name:          "empty platform key ID → skip",
			summaries:     []store.CombinedKeySummary{{PlatformKeyID: "pk-1", Remain: 0}},
			platformKeyID: "",
			want:          false,
		},
		{
			name:          "nil summaries (Unconstrained) → skip",
			summaries:     nil,
			platformKeyID: "pk-1",
			want:          false,
		},
		{
			name:          "remain > 0 → skip",
			summaries:     []store.CombinedKeySummary{{PlatformKeyID: "pk-1", Remain: 10}},
			platformKeyID: "pk-1",
			want:          false,
		},
		{
			name:          "remain == 0 → enqueue",
			summaries:     []store.CombinedKeySummary{{PlatformKeyID: "pk-1", Remain: 0}},
			platformKeyID: "pk-1",
			want:          true,
		},
		{
			name:          "remain < 0 → enqueue",
			summaries:     []store.CombinedKeySummary{{PlatformKeyID: "pk-1", Remain: -5}},
			platformKeyID: "pk-1",
			want:          true,
		},
		{
			name:          "key not in summaries (Unknown) → enqueue for safety",
			summaries:     []store.CombinedKeySummary{{PlatformKeyID: "pk-other", Remain: 50}},
			platformKeyID: "pk-1",
			want:          true,
		},
		{
			name:          "empty summaries slice (Unknown) → enqueue for safety",
			summaries:     []store.CombinedKeySummary{},
			platformKeyID: "pk-1",
			want:          true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ShouldEnqueueOverrun(tc.summaries, tc.platformKeyID)
			if got != tc.want {
				t.Errorf("ShouldEnqueueOverrun() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestOverrunPayloadFromEffects(t *testing.T) {
	t.Parallel()

	t.Run("nil effects → nil", func(t *testing.T) {
		if got := OverrunPayloadFromEffects(nil); got != nil {
			t.Errorf("expected nil, got %s", got)
		}
	})

	t.Run("empty payload → nil", func(t *testing.T) {
		effects := &IngestEffects{OverrunPayload: nil}
		if got := OverrunPayloadFromEffects(effects); got != nil {
			t.Errorf("expected nil, got %s", got)
		}
	})

	t.Run("valid payload → returned", func(t *testing.T) {
		payload := json.RawMessage(`{"departmentId":"dept-1"}`)
		effects := &IngestEffects{OverrunPayload: payload}
		got := OverrunPayloadFromEffects(effects)
		if string(got) != `{"departmentId":"dept-1"}` {
			t.Errorf("got %s", got)
		}
	})
}

func TestOverrunPayloadFromEntry(t *testing.T) {
	t.Parallel()
	memberID := "m-1"
	projectID := "proj-1"
	entry := types.UsageLedgerEntry{
		DepartmentID:  "dept-1",
		PlatformKeyID: "pk-1",
		MemberID:      &memberID,
		ProjectID:     &projectID,
	}
	raw := overrunPayloadFromEntry(entry, "2026-07")
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatal(err)
	}
	if payload["departmentId"] != "dept-1" {
		t.Errorf("departmentId = %v", payload["departmentId"])
	}
	if payload["platformKeyId"] != "pk-1" {
		t.Errorf("platformKeyId = %v", payload["platformKeyId"])
	}
	if payload["memberId"] != "m-1" {
		t.Errorf("memberId = %v", payload["memberId"])
	}
	if payload["projectId"] != "proj-1" {
		t.Errorf("projectId = %v", payload["projectId"])
	}
	if payload["periodKey"] != "2026-07" {
		t.Errorf("periodKey = %v", payload["periodKey"])
	}
}
