package org_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestGetFieldMappingsSeedsDefaults(t *testing.T) {
	svc := newTestService(t)
	mappings, err := svc.GetFieldMappings(testutil.Ctx(), string(types.PlatformFeishu))
	if err != nil {
		t.Fatal(err)
	}
	if len(mappings) != 6 {
		t.Fatalf("expected 6 mappings, got %d", len(mappings))
	}
}

func TestSaveAndTestFieldMappings(t *testing.T) {
	svc := newTestService(t)
	ctx := testutil.Ctx()
	custom := []types.FieldMapping{
		{SourceField: "name", SourceLabel: "Name", TargetField: "name", Required: true},
	}
	if err := svc.SaveFieldMappings(ctx, types.FieldMappingConfig{
		Platform: types.PlatformFeishu,
		Mappings: custom,
	}); err != nil {
		t.Fatal(err)
	}
	mappings, err := svc.GetFieldMappings(ctx, string(types.PlatformFeishu))
	if err != nil {
		t.Fatal(err)
	}
	if len(mappings) != 1 || mappings[0].SourceField != "name" {
		t.Fatalf("unexpected mappings %+v", mappings)
	}
	result, err := svc.TestFieldMapping(ctx, string(types.PlatformFeishu), "张三")
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success || result.Preview["姓名"] != "张三" {
		t.Fatalf("unexpected test result %+v", result)
	}
}
