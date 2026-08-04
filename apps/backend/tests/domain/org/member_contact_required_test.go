package org_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestCreateMemberRejectsNoContact(t *testing.T) {
	t.Parallel()
	svc := newTestOrgService(t)
	_, err := svc.CreateMember(testutil.Ctx(), types.CreateMemberInput{
		User:   types.CreateMemberUserInput{Name: "No Contact"},
		Member: types.CreateMemberData{DepartmentID: contract.IDDept5},
	}, uuid.Nil)
	de := asDomainError(t, err)
	if de.Status != domain.StatusBadRequest {
		t.Fatalf("expected 400, got %d", de.Status)
	}
}

func TestCreateMemberAcceptsPhoneOnly(t *testing.T) {
	t.Parallel()
	svc := newTestOrgService(t)
	_, err := svc.CreateMember(testutil.Ctx(), types.CreateMemberInput{
		User:   types.CreateMemberUserInput{Name: "Phone Only", Phone: "13700009999"},
		Member: types.CreateMemberData{DepartmentID: contract.IDDept5},
	}, uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateMemberAcceptsEmailOnly(t *testing.T) {
	t.Parallel()
	svc := newTestOrgService(t)
	_, err := svc.CreateMember(testutil.Ctx(), types.CreateMemberInput{
		User:   types.CreateMemberUserInput{Name: "Email Only", Email: "emailonly@example.com"},
		Member: types.CreateMemberData{DepartmentID: contract.IDDept5},
	}, uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestBatchImportRejectsNoContact(t *testing.T) {
	t.Parallel()
	svc := newTestOrgService(t)
	result, err := svc.BatchImport(testutil.Ctx(), []types.BatchImportRow{
		{Name: "Has Contact", Phone: "13800001111", DepartmentName: "测试组"},
		{Name: "No Contact", DepartmentName: "测试组"},
	}, uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Imported != 1 {
		t.Fatalf("expected 1 imported, got %d", result.Imported)
	}
	if len(result.Failures) != 1 {
		t.Fatalf("expected 1 failure, got %d", len(result.Failures))
	}
	if result.Failures[0].Row != 2 {
		t.Fatalf("expected failure on row 2, got row %d", result.Failures[0].Row)
	}
}
