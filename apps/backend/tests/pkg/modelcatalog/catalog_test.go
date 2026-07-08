package modelcatalog_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/modelcatalog"
)

func TestDedupeEffectiveTenantOverridesGlobal(t *testing.T) {
	t.Parallel()
	items := []types.ModelInfo{
		{ModelID: 1, Provider: "openai", Type: "gpt-4o", CompanyID: 1},
		{ModelID: 2, Provider: "openai", Type: "gpt-4o", CompanyID: 2},
	}
	out := modelcatalog.DedupeEffective(items)
	if len(out) != 1 || out[0].ModelID != 2 {
		t.Fatalf("expected tenant override, got %+v", out)
	}
}

func TestDedupeEffectiveDifferentProviderSameType(t *testing.T) {
	t.Parallel()
	items := []types.ModelInfo{
		{ModelID: 1, Provider: "openai", Type: "gpt-4o", Enabled: true},
		{ModelID: 2, Provider: types.ProviderCustom, Type: "gpt-4o", Enabled: true},
	}
	out := modelcatalog.DedupeEffective(items)
	if len(out) != 2 {
		t.Fatalf("expected both models, got %d", len(out))
	}
}

func TestFilterEnabledIDs(t *testing.T) {
	t.Parallel()
	catalog := []types.ModelInfo{
		{ModelID: 1, Type: "a", Enabled: true},
		{ModelID: 2, Type: "b", Enabled: false},
		{ModelID: 3, Type: "c", Enabled: true},
	}
	out := modelcatalog.FilterEnabledIDs(catalog, []int64{1, 2, 3, 99})
	if len(out) != 2 || out[0] != 1 || out[1] != 3 {
		t.Fatalf("unexpected %v", out)
	}
}

func TestIsCallTypeAllowed(t *testing.T) {
	t.Parallel()
	catalog := []types.ModelInfo{
		{ModelID: 1, Provider: "openai", Type: "gpt-4o", Enabled: true},
		{ModelID: 2, Provider: types.ProviderCustom, Type: "gpt-4o", Enabled: true},
		{ModelID: 3, Provider: "openai", Type: "claude", Enabled: false},
	}
	if !modelcatalog.IsCallTypeAllowed(catalog, []int64{2}, "gpt-4o") {
		t.Fatal("expected allowed")
	}
	if modelcatalog.IsCallTypeAllowed(catalog, []int64{3}, "claude") {
		t.Fatal("expected disabled model blocked")
	}
	if modelcatalog.IsCallTypeAllowed(catalog, []int64{1}, "claude") {
		t.Fatal("expected not in allowlist")
	}
}

func TestResolveIDForCallTypePrefersCustom(t *testing.T) {
	t.Parallel()
	catalog := []types.ModelInfo{
		{ModelID: 1, Provider: "openai", Type: "gpt-4o", Enabled: true},
		{ModelID: 2, Provider: types.ProviderCustom, Type: "gpt-4o", Enabled: true},
	}
	id, ok := modelcatalog.ResolveIDForCallType(catalog, []int64{1, 2}, "gpt-4o")
	if !ok || id == nil || *id != 2 {
		t.Fatalf("expected custom model id 2, got %v ok=%v", id, ok)
	}
}

func TestValidateWritableIDs(t *testing.T) {
	t.Parallel()
	catalog := []types.ModelInfo{
		{ModelID: 1, Enabled: true},
		{ModelID: 2, Enabled: false},
	}
	if err := modelcatalog.ValidateWritableIDs(catalog, []int64{1}); err != nil {
		t.Fatal(err)
	}
	if err := modelcatalog.ValidateWritableIDs(catalog, []int64{2}); err != modelcatalog.ErrModelDisabled {
		t.Fatalf("expected disabled error, got %v", err)
	}
	if err := modelcatalog.ValidateWritableIDs(catalog, []int64{9}); err != modelcatalog.ErrUnknownModelID {
		t.Fatalf("expected unknown error, got %v", err)
	}
}

func TestCallTypesForIDs(t *testing.T) {
	t.Parallel()
	catalog := []types.ModelInfo{
		{ModelID: 1, Type: "gpt-4o", Enabled: true},
		{ModelID: 2, Type: "gpt-4o", Enabled: true},
		{ModelID: 3, Type: "claude", Enabled: false},
	}
	out := modelcatalog.CallTypesForIDs(catalog, []int64{1, 2, 3})
	if len(out) != 1 || out[0] != "gpt-4o" {
		t.Fatalf("unexpected %v", out)
	}
}
