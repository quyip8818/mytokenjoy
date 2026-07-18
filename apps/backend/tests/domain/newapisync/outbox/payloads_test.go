//go:build testhook

package outbox_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
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

	companyID1 := uuid.MustParse("00000000-0000-7000-0000-000000000001")
	companyID3 := uuid.MustParse("00000000-0000-7000-0000-000000000003")
	companyID4 := uuid.MustParse("00000000-0000-7000-0000-000000000004")
	companyID5 := uuid.MustParse("00000000-0000-7000-0000-000000000005")
	platformKeyID := uuid.MustParse("00000000-0000-7000-0000-00000000f001")
	deptID := uuid.MustParse("00000000-0000-7000-0000-000000000d03")

	t.Run("create key", func(t *testing.T) {
		t.Parallel()
		var wire struct {
			CompanyID     string `json:"companyId"`
			PlatformKeyID string `json:"platformKeyId"`
		}
		assertOutboxWireJSON(t, outbox.CreateKeyOutboxPayload{
			CompanyID: companyID1, PlatformKeyID: platformKeyID,
		}, &wire)
		if wire.CompanyID != companyID1.String() || wire.PlatformKeyID != platformKeyID.String() {
			t.Fatalf("wire mismatch: %+v", wire)
		}
	})

	t.Run("upsert channel", func(t *testing.T) {
		t.Parallel()
		providerKeyID := uuid.MustParse("00000000-0000-7000-0000-00000000a001")
		var wire struct {
			CompanyID     string `json:"companyId"`
			ProviderKeyID string `json:"providerKeyId"`
		}
		assertOutboxWireJSON(t, outbox.UpsertChannelOutboxPayload{
			CompanyID: companyID3, ProviderKeyID: providerKeyID,
		}, &wire)
		if wire.CompanyID != companyID3.String() || wire.ProviderKeyID != providerKeyID.String() {
			t.Fatalf("wire mismatch: %+v", wire)
		}
	})

	t.Run("update model limits", func(t *testing.T) {
		t.Parallel()
		var wire struct {
			CompanyID    string `json:"companyId"`
			DepartmentID string `json:"departmentId"`
		}
		assertOutboxWireJSON(t, outbox.UpdateModelLimitsOutboxPayload{
			CompanyID: companyID4, DepartmentID: deptID,
		}, &wire)
		if wire.CompanyID != companyID4.String() || wire.DepartmentID != deptID.String() {
			t.Fatalf("wire mismatch: %+v", wire)
		}
	})

	t.Run("rebalance axis", func(t *testing.T) {
		t.Parallel()
		axisID := uuid.MustParse("00000000-0000-7000-0000-000000000d03")
		var wire struct {
			CompanyID string `json:"companyId"`
			AxisKind  string `json:"axisKind"`
			AxisID    string `json:"axisId"`
		}
		assertOutboxWireJSON(t, outbox.RebalanceAxisOutboxPayload{
			CompanyID: companyID5, AxisKind: "department", AxisID: axisID,
		}, &wire)
		if wire.CompanyID != companyID5.String() || wire.AxisKind != "department" || wire.AxisID != axisID.String() {
			t.Fatalf("wire mismatch: %+v", wire)
		}
	})
}
