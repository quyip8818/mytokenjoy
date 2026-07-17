package modelcatalog_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/modelcatalog"
)

var (
	catModel1 = uuid.MustParse("00000000-0000-7000-0000-0000000000b1")
	catModel2 = uuid.MustParse("00000000-0000-7000-0000-0000000000b2")
	catModel3 = uuid.MustParse("00000000-0000-7000-0000-0000000000b3")
	catModel4 = uuid.MustParse("00000000-0000-7000-0000-0000000000b4")
	catModel9 = uuid.MustParse("00000000-0000-7000-0000-0000000000b9")

	catCompany1 = uuid.MustParse("00000000-0000-7000-0000-000000000001")
	catCompany2 = uuid.MustParse("00000000-0000-7000-0000-000000000002")
)

func TestDedupeEffectiveTenantOverridesGlobal(t *testing.T) {
	t.Parallel()
	items := []types.ModelInfo{
		{ID: catModel1, Provider: "openai", Type: "gpt-4o", CompanyID: catCompany1},
		{ID: catModel2, Provider: "openai", Type: "gpt-4o", CompanyID: catCompany2},
	}
	out := modelcatalog.DedupeEffective(items)
	if len(out) != 1 || out[0].ID != catModel2 {
		t.Fatalf("expected tenant override, got %+v", out)
	}
}

func TestDedupeEffectiveDifferentProviderSameType(t *testing.T) {
	t.Parallel()
	items := []types.ModelInfo{
		{ID: catModel1, Provider: "openai", Type: "gpt-4o", Enabled: true},
		{ID: catModel2, Provider: types.ProviderCustom, Type: "gpt-4o", Enabled: true},
	}
	out := modelcatalog.DedupeEffective(items)
	if len(out) != 2 {
		t.Fatalf("expected both models, got %d", len(out))
	}
}

func TestFilterEnabledIDs(t *testing.T) {
	t.Parallel()
	catalog := []types.ModelInfo{
		{ID: catModel1, Type: "a", Enabled: true},
		{ID: catModel2, Type: "b", Enabled: false},
		{ID: catModel3, Type: "c", Enabled: true},
	}
	out := modelcatalog.FilterEnabledIDs(catalog, []uuid.UUID{catModel1, catModel2, catModel3, catModel9})
	if len(out) != 2 || out[0] != catModel1 || out[1] != catModel3 {
		t.Fatalf("unexpected %v", out)
	}
}

func TestIsCallTypeAllowed(t *testing.T) {
	t.Parallel()
	catalog := []types.ModelInfo{
		{ID: catModel1, Provider: "openai", Type: "gpt-4o", Enabled: true},
		{ID: catModel2, Provider: types.ProviderCustom, Type: "gpt-4o", Enabled: true},
		{ID: catModel3, Provider: "openai", Type: "claude", Enabled: false},
	}
	if !modelcatalog.IsCallTypeAllowed(catalog, []uuid.UUID{catModel2}, "gpt-4o") {
		t.Fatal("expected allowed")
	}
	if modelcatalog.IsCallTypeAllowed(catalog, []uuid.UUID{catModel3}, "claude") {
		t.Fatal("expected disabled model blocked")
	}
	if modelcatalog.IsCallTypeAllowed(catalog, []uuid.UUID{catModel1}, "claude") {
		t.Fatal("expected not in allowlist")
	}
}

func TestResolveIDForCallTypePrefersCustom(t *testing.T) {
	t.Parallel()
	catalog := []types.ModelInfo{
		{ID: catModel1, Provider: "openai", Type: "gpt-4o", Enabled: true},
		{ID: catModel2, Provider: types.ProviderCustom, Type: "gpt-4o", Enabled: true},
	}
	id, ok := modelcatalog.ResolveIDForCallType(catalog, []uuid.UUID{catModel1, catModel2}, "gpt-4o")
	if !ok || id == nil || *id != catModel2 {
		t.Fatalf("expected custom model id catModel2, got %v ok=%v", id, ok)
	}
}

func TestValidateWritableIDs(t *testing.T) {
	t.Parallel()
	catalog := []types.ModelInfo{
		{ID: catModel1, Enabled: true},
		{ID: catModel2, Enabled: false},
	}
	if err := modelcatalog.ValidateWritableIDs(catalog, []uuid.UUID{catModel1}); err != nil {
		t.Fatal(err)
	}
	if err := modelcatalog.ValidateWritableIDs(catalog, []uuid.UUID{catModel2}); err != modelcatalog.ErrModelDisabled {
		t.Fatalf("expected disabled error, got %v", err)
	}
	if err := modelcatalog.ValidateWritableIDs(catalog, []uuid.UUID{catModel9}); err != modelcatalog.ErrUnknownModelID {
		t.Fatalf("expected unknown error, got %v", err)
	}
}

func TestCallTypesForIDs(t *testing.T) {
	t.Parallel()
	catalog := []types.ModelInfo{
		{ID: catModel1, Type: "gpt-4o", Enabled: true},
		{ID: catModel2, Type: "gpt-4o", Enabled: true},
		{ID: catModel3, Type: "claude", Enabled: false},
	}
	out := modelcatalog.CallTypesForIDs(catalog, []uuid.UUID{catModel1, catModel2, catModel3})
	if len(out) != 1 || out[0] != "gpt-4o" {
		t.Fatalf("unexpected %v", out)
	}
}
