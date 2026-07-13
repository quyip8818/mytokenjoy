//go:build testhook

package outbox_test

import (
	"encoding/json"
	"testing"

	"github.com/tokenjoy/backend/internal/domain/newapisync/outbox"
)

func assertOutboxWireJSON(t *testing.T, payload any, wire any) {
	t.Helper()
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := json.Unmarshal(raw, wire); err != nil {
		t.Fatalf("unmarshal wire: %v json=%s", err, string(raw))
	}
}

func TestOutboxPayloadJSONRoundtrip(t *testing.T) {
	t.Parallel()

	t.Run("create key", func(t *testing.T) {
		t.Parallel()
		var wire struct {
			CompanyID     int64  `json:"companyId"`
			PlatformKeyID string `json:"platformKeyId"`
		}
		assertOutboxWireJSON(t, outbox.CreateKeyOutboxPayload{
			CompanyID: 1, PlatformKeyID: "plk-1",
		}, &wire)
		if wire.CompanyID != 1 || wire.PlatformKeyID != "plk-1" {
			t.Fatalf("wire mismatch: %+v", wire)
		}
	})

	t.Run("upsert channel", func(t *testing.T) {
		t.Parallel()
		var wire struct {
			CompanyID     int64  `json:"companyId"`
			ProviderKeyID string `json:"providerKeyId"`
		}
		assertOutboxWireJSON(t, outbox.UpsertChannelOutboxPayload{
			CompanyID: 3, ProviderKeyID: "pvk-1",
		}, &wire)
		if wire.CompanyID != 3 || wire.ProviderKeyID != "pvk-1" {
			t.Fatalf("wire mismatch: %+v", wire)
		}
	})

	t.Run("update model limits", func(t *testing.T) {
		t.Parallel()
		var wire struct {
			CompanyID    int64  `json:"companyId"`
			DepartmentID string `json:"departmentId"`
		}
		assertOutboxWireJSON(t, outbox.UpdateModelLimitsOutboxPayload{
			CompanyID: 4, DepartmentID: "dept-3",
		}, &wire)
		if wire.CompanyID != 4 || wire.DepartmentID != "dept-3" {
			t.Fatalf("wire mismatch: %+v", wire)
		}
	})

	t.Run("rebalance axis", func(t *testing.T) {
		t.Parallel()
		var wire struct {
			CompanyID int64  `json:"companyId"`
			AxisKind  string `json:"axisKind"`
			AxisID    string `json:"axisId"`
		}
		assertOutboxWireJSON(t, outbox.RebalanceAxisOutboxPayload{
			CompanyID: 5, AxisKind: "department", AxisID: "dept-3",
		}, &wire)
		if wire.CompanyID != 5 || wire.AxisKind != "department" || wire.AxisID != "dept-3" {
			t.Fatalf("wire mismatch: %+v", wire)
		}
	})
}
