package org_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/tests/testutil"
)

// TestBatchImportNestedDepartmentPath tests importing with path notation like "技术部/前端组"
func TestBatchImportNestedDepartmentPath(t *testing.T) {
	t.Parallel()
	svc := newTestOrgService(t)
	result, err := svc.BatchImport(testutil.Ctx(), []types.BatchImportRow{
		{Name: "路径导入", Phone: "13900000001", DepartmentName: "技术部/前端组", EmployeeId: "P001"},
	}, uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Imported != 1 {
		t.Fatalf("expected 1 imported, got %d (failures: %v)", result.Imported, result.Failures)
	}
}

// TestBatchImportNestedDepartmentPathDeep tests 3-level path like "总公司/技术部/后端组"
func TestBatchImportNestedDepartmentPathDeep(t *testing.T) {
	t.Parallel()
	svc := newTestOrgService(t)
	result, err := svc.BatchImport(testutil.Ctx(), []types.BatchImportRow{
		{Name: "深路径导入", Phone: "13900000002", DepartmentName: "总公司/技术部/后端组", EmployeeId: "P002"},
	}, uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Imported != 1 {
		t.Fatalf("expected 1 imported, got %d (failures: %v)", result.Imported, result.Failures)
	}
}

// TestBatchImportAutoCreateDepartment tests that unknown departments are auto-created under root
func TestBatchImportAutoCreateDepartment(t *testing.T) {
	t.Parallel()
	svc := newTestOrgService(t)
	result, err := svc.BatchImport(testutil.Ctx(), []types.BatchImportRow{
		{Name: "新部门成员", Phone: "13900000003", DepartmentName: "财务部", EmployeeId: "P003"},
	}, uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Imported != 1 {
		t.Fatalf("expected 1 imported, got %d (failures: %v)", result.Imported, result.Failures)
	}
	// Verify the department was actually created
	depts, err := svc.GetDepartmentTree(testutil.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, d := range depts[0].Children {
		if d.Name == "财务部" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected auto-created department 财务部 to exist")
	}
}

// TestBatchImportAutoCreateNestedDepartment tests that "新部门/子部门" auto-creates the chain
func TestBatchImportAutoCreateNestedDepartment(t *testing.T) {
	t.Parallel()
	svc := newTestOrgService(t)
	result, err := svc.BatchImport(testutil.Ctx(), []types.BatchImportRow{
		{Name: "嵌套新建", Phone: "13900000004", DepartmentName: "研发中心/AI组", EmployeeId: "P004"},
	}, uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Imported != 1 {
		t.Fatalf("expected 1 imported, got %d (failures: %v)", result.Imported, result.Failures)
	}
	// Verify both levels created
	depts, err := svc.GetDepartmentTree(testutil.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	var parent *types.Department
	for i := range depts[0].Children {
		if depts[0].Children[i].Name == "研发中心" {
			parent = &depts[0].Children[i]
			break
		}
	}
	if parent == nil {
		t.Fatal("expected auto-created department 研发中心 to exist")
	}
	found := false
	for _, child := range parent.Children {
		if child.Name == "AI组" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected auto-created sub-department AI组 under 研发中心")
	}
}

// TestBatchImportExistingParentNewChild auto-creates only the leaf under existing parent
func TestBatchImportExistingParentNewChild(t *testing.T) {
	t.Parallel()
	svc := newTestOrgService(t)
	result, err := svc.BatchImport(testutil.Ctx(), []types.BatchImportRow{
		{Name: "新子部门成员", Phone: "13900000005", DepartmentName: "技术部/DevOps组", EmployeeId: "P005"},
	}, uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Imported != 1 {
		t.Fatalf("expected 1 imported, got %d (failures: %v)", result.Imported, result.Failures)
	}
}

// TestBatchImportMultiRowMixed tests the real scenario: existing depts + new dept in one batch
func TestBatchImportMultiRowMixed(t *testing.T) {
	t.Parallel()
	svc := newTestOrgService(t)
	result, err := svc.BatchImport(testutil.Ctx(), []types.BatchImportRow{
		{Name: "张三", Phone: "13800138000", Email: "a@example.com", DepartmentName: "技术部", EmployeeId: "EMP001"},
		{Name: "李散", Phone: "13800138001", Email: "b@example.com", DepartmentName: "市场部", EmployeeId: "EMP002"},
		{Name: "陈单", Phone: "13800138002", Email: "c@example.com", DepartmentName: "客服部", EmployeeId: "EMP003"},
	}, uuid.Nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Imported != 3 {
		t.Fatalf("expected 3 imported, got %d (failures: %v)", result.Imported, result.Failures)
	}
	if len(result.Failures) != 0 {
		t.Fatalf("expected 0 failures, got %v", result.Failures)
	}
}

// TestBatchImportDuplicateEmailReportsPerRow tests that same email across rows
// reports a friendly per-row error instead of a generic "保存失败"
func TestBatchImportDuplicateEmailReportsPerRow(t *testing.T) {
	t.Parallel()
	svc := newTestOrgService(t)
	result, err := svc.BatchImport(testutil.Ctx(), []types.BatchImportRow{
		{Name: "张三", Phone: "13800138000", Email: "same@example.com", DepartmentName: "技术部", EmployeeId: "EMP001"},
		{Name: "李散", Phone: "13800138001", Email: "same@example.com", DepartmentName: "市场部", EmployeeId: "EMP002"},
		{Name: "陈单", Phone: "13800138002", Email: "same@example.com", DepartmentName: "客服部", EmployeeId: "EMP003"},
	}, uuid.Nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// First row should succeed, 2nd and 3rd should fail with duplicate user error
	if result.Imported != 1 {
		t.Fatalf("expected 1 imported, got %d", result.Imported)
	}
	if len(result.Failures) != 2 {
		t.Fatalf("expected 2 failures, got %d: %v", len(result.Failures), result.Failures)
	}
	for _, f := range result.Failures {
		if !strings.Contains(f.Reason, "已存在") {
			t.Errorf("expected duplicate user message, got: %s", f.Reason)
		}
	}
}
