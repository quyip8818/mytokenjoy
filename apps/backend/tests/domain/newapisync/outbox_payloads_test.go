package newapisync_test

import (
	"encoding/json"
	"testing"

	domainnewapisync "github.com/tokenjoy/backend/internal/domain/newapisync"
)

func assertJSONRoundtrip[T comparable](t *testing.T, payload T) {
	t.Helper()
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded T
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded != payload {
		t.Fatalf("roundtrip mismatch: got %#v want %#v", decoded, payload)
	}
}

func TestOutboxPayloadJSONRoundtrip(t *testing.T) {
	t.Parallel()

	t.Run("create key", func(t *testing.T) {
		t.Parallel()
		assertJSONRoundtrip(t, domainnewapisync.CreateKeyOutboxPayload{
			CompanyID: 1, PlatformKeyID: "plk-1",
		})
	})
	t.Run("update key", func(t *testing.T) {
		t.Parallel()
		assertJSONRoundtrip(t, domainnewapisync.UpdateKeyOutboxPayload{
			CompanyID: 2, PlatformKeyID: "plk-2",
		})
	})
	t.Run("upsert channel", func(t *testing.T) {
		t.Parallel()
		assertJSONRoundtrip(t, domainnewapisync.UpsertChannelOutboxPayload{
			CompanyID: 3, ProviderKeyID: "pvk-1",
		})
	})
	t.Run("update model limits", func(t *testing.T) {
		t.Parallel()
		assertJSONRoundtrip(t, domainnewapisync.UpdateModelLimitsOutboxPayload{
			CompanyID: 4, DepartmentID: "dept-3",
		})
	})
	t.Run("rebalance axis", func(t *testing.T) {
		t.Parallel()
		assertJSONRoundtrip(t, domainnewapisync.RebalanceAxisOutboxPayload{
			CompanyID: 5, AxisKind: "department", AxisID: "dept-3",
		})
	})
}
