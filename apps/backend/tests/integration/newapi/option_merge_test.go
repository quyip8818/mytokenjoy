//go:build integration

package newapi_test

import (
	"encoding/json"
	"testing"

	"github.com/tokenjoy/backend/internal/integration/newapi"
)

func TestMergeGroupOptionAddsUserUsableGroups(t *testing.T) {
	t.Parallel()
	raw := `{"default":"Default","vip":"VIP"}`
	merged, skip, err := newapi.MergeGroupOption(raw, "UserUsableGroups", "dept-dept-3", "后端组")
	if err != nil || skip {
		t.Fatalf("unexpected skip=%v err=%v", skip, err)
	}
	var data map[string]string
	if err := json.Unmarshal([]byte(merged), &data); err != nil {
		t.Fatal(err)
	}
	if data["dept-dept-3"] != "后端组" {
		t.Fatalf("unexpected data: %+v", data)
	}
}

func TestMergeGroupOptionSkipsExisting(t *testing.T) {
	t.Parallel()
	raw := `{"dept-dept-3":"后端组"}`
	_, skip, err := newapi.MergeGroupOption(raw, "UserUsableGroups", "dept-dept-3", "后端组")
	if err != nil || !skip {
		t.Fatalf("expected skip, got skip=%v err=%v", skip, err)
	}
}
