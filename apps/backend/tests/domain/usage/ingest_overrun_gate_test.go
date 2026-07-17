package usage_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/store"
)

func TestShouldEnqueueOverrun(t *testing.T) {
	t.Parallel()
	pk1 := uuid.MustParse("00000000-0000-7000-0000-000000000f01")
	pkOther := uuid.MustParse("00000000-0000-7000-0000-000000000f02")
	tests := []struct {
		name          string
		summaries     []store.CombinedKeySummary
		platformKeyID uuid.UUID
		want          bool
	}{
		{
			name:          "empty platform key ID → skip",
			summaries:     []store.CombinedKeySummary{{PlatformKeyID: pk1, Remain: 0}},
			platformKeyID: uuid.Nil,
			want:          false,
		},
		{
			name:          "nil summaries (Unconstrained) → skip",
			summaries:     nil,
			platformKeyID: pk1,
			want:          false,
		},
		{
			name:          "remain > 0 → skip",
			summaries:     []store.CombinedKeySummary{{PlatformKeyID: pk1, Remain: 10}},
			platformKeyID: pk1,
			want:          false,
		},
		{
			name:          "remain == 0 → enqueue",
			summaries:     []store.CombinedKeySummary{{PlatformKeyID: pk1, Remain: 0}},
			platformKeyID: pk1,
			want:          true,
		},
		{
			name:          "remain < 0 → enqueue",
			summaries:     []store.CombinedKeySummary{{PlatformKeyID: pk1, Remain: -5}},
			platformKeyID: pk1,
			want:          true,
		},
		{
			name:          "key not in summaries (Unknown) → enqueue for safety",
			summaries:     []store.CombinedKeySummary{{PlatformKeyID: pkOther, Remain: 50}},
			platformKeyID: pk1,
			want:          true,
		},
		{
			name:          "empty summaries slice (Unknown) → enqueue for safety",
			summaries:     []store.CombinedKeySummary{},
			platformKeyID: pk1,
			want:          true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := usage.ShouldEnqueueOverrun(tc.summaries, tc.platformKeyID)
			if got != tc.want {
				t.Errorf("ShouldEnqueueOverrun() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestOverrunPayloadFromEffects(t *testing.T) {
	t.Parallel()

	t.Run("nil effects → nil", func(t *testing.T) {
		if got := usage.OverrunPayloadFromEffects(nil); got != nil {
			t.Errorf("expected nil, got %s", got)
		}
	})

	t.Run("empty payload → nil", func(t *testing.T) {
		effects := &usage.IngestEffects{OverrunPayload: nil}
		if got := usage.OverrunPayloadFromEffects(effects); got != nil {
			t.Errorf("expected nil, got %s", got)
		}
	})

	t.Run("valid payload → returned", func(t *testing.T) {
		payload := json.RawMessage(`{"departmentId":"dept-1"}`)
		effects := &usage.IngestEffects{OverrunPayload: payload}
		got := usage.OverrunPayloadFromEffects(effects)
		if string(got) != `{"departmentId":"dept-1"}` {
			t.Errorf("got %s", got)
		}
	})
}

func TestOverrunPayloadFromEntry(t *testing.T) {
	t.Parallel()
	memberID := uuid.MustParse("00000000-0000-7000-0000-000000000e01")
	projectID := uuid.MustParse("00000000-0000-7000-0000-000000000101")
	deptID := uuid.MustParse("00000000-0000-7000-0000-000000000d01")
	pkID := uuid.MustParse("00000000-0000-7000-0000-000000000f01")
	entry := types.UsageLedgerEntry{
		DepartmentID:  deptID,
		PlatformKeyID: pkID,
		MemberID:      &memberID,
		ProjectID:     &projectID,
	}
	raw := usage.OverrunPayloadFromEntry(entry, "2026-07")
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatal(err)
	}
	if payload["departmentId"] != deptID.String() {
		t.Errorf("departmentId = %v", payload["departmentId"])
	}
	if payload["platformKeyId"] != pkID.String() {
		t.Errorf("platformKeyId = %v", payload["platformKeyId"])
	}
	if payload["memberId"] != memberID.String() {
		t.Errorf("memberId = %v", payload["memberId"])
	}
	if payload["projectId"] != projectID.String() {
		t.Errorf("projectId = %v", payload["projectId"])
	}
	if payload["periodKey"] != "2026-07" {
		t.Errorf("periodKey = %v", payload["periodKey"])
	}
}
