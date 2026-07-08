package org_test

import (
	"strings"
	"testing"

	"github.com/tokenjoy/backend/tests/testutil"
)

func TestCreateRoleRejectsDuplicateName(t *testing.T) {
	t.Parallel()
	svc := newTestOrgService(t)
	ctx := testutil.Ctx()

	_, err := svc.CreateRole(ctx, "TestRole", []string{"p-1"})
	if err != nil {
		t.Fatalf("first CreateRole failed: %v", err)
	}

	_, err = svc.CreateRole(ctx, "TestRole", []string{"p-2"})
	if err == nil {
		t.Fatal("expected error for duplicate role name")
	}
	de := asDomainError(t, err)
	if !strings.Contains(de.Message, "already exists") {
		t.Fatalf("expected 'already exists' error, got: %s", de.Message)
	}
}

func TestCreateRoleRejectsEmptyName(t *testing.T) {
	t.Parallel()
	svc := newTestOrgService(t)
	ctx := testutil.Ctx()

	_, err := svc.CreateRole(ctx, "  ", []string{"p-1"})
	if err == nil {
		t.Fatal("expected error for whitespace-only role name")
	}
	de := asDomainError(t, err)
	if !strings.Contains(de.Message, "empty") {
		t.Fatalf("expected 'empty' error, got: %s", de.Message)
	}
}

func TestCreateRoleSucceedsWithUniqueName(t *testing.T) {
	t.Parallel()
	svc := newTestOrgService(t)
	ctx := testutil.Ctx()

	role, err := svc.CreateRole(ctx, "UniqueRole", []string{"p-1"})
	if err != nil {
		t.Fatalf("CreateRole failed: %v", err)
	}
	if role.Name != "UniqueRole" {
		t.Fatalf("expected name 'UniqueRole', got '%s'", role.Name)
	}
}
