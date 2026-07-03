package org_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/datasource"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
)

func importedDept(id, extID, name string) types.Department {
	ext := extID
	src := types.DeptSourceImported
	return types.Department{ID: id, Name: name, ExternalID: &ext, Source: &src}
}

func manualDept(id, name string) types.Department {
	src := types.DeptSourceManual
	return types.Department{ID: id, Name: name, Source: &src}
}

func importedMember(id, extID, name, email, phone string) types.Member {
	ext := extID
	return types.Member{
		ID: id, Name: name, Email: email, Phone: phone,
		ExternalID: &ext, Source: types.MemberSourceImported,
	}
}

func TestBuildSyncDiffRemoteAdditions(t *testing.T) {
	diff := pkgorg.BuildSyncDiff(
		nil, nil,
		[]datasource.RemoteDepartment{{ExternalID: "d1", Name: "Eng"}},
		[]datasource.RemoteMember{{ExternalID: "u1", Name: "Alice", Email: "a@x.com"}},
	)
	if len(diff.AddDepartments) != 1 || diff.AddDepartments[0].ExternalID != "d1" {
		t.Fatalf("unexpected add departments: %+v", diff.AddDepartments)
	}
	if len(diff.AddMembers) != 1 || diff.AddMembers[0].ExternalID != "u1" {
		t.Fatalf("unexpected add members: %+v", diff.AddMembers)
	}
}

func TestBuildSyncDiffRemoteRemovals(t *testing.T) {
	diff := pkgorg.BuildSyncDiff(
		[]types.Department{importedDept("dept-1", "d1", "Eng")},
		[]types.Member{importedMember("m-1", "u1", "Alice", "a@x.com", "138")},
		nil, nil,
	)
	if len(diff.RemoveDepartments) != 1 || diff.RemoveDepartments[0].ID != "dept-1" {
		t.Fatalf("unexpected remove departments: %+v", diff.RemoveDepartments)
	}
	if len(diff.RemoveMembers) != 1 || diff.RemoveMembers[0].ID != "m-1" {
		t.Fatalf("unexpected remove members: %+v", diff.RemoveMembers)
	}
}

func TestBuildSyncDiffRenames(t *testing.T) {
	diff := pkgorg.BuildSyncDiff(
		[]types.Department{importedDept("dept-1", "d1", "Eng")},
		[]types.Member{importedMember("m-1", "u1", "Alice", "a@x.com", "138")},
		[]datasource.RemoteDepartment{{ExternalID: "d1", Name: "Engineering"}},
		[]datasource.RemoteMember{{ExternalID: "u1", Name: "Alice", Email: "alice@x.com", Mobile: "139"}},
	)
	if len(diff.UpdateDepartments) != 1 || diff.UpdateDepartments[0].Name != "Engineering" {
		t.Fatalf("unexpected update departments: %+v", diff.UpdateDepartments)
	}
	if len(diff.UpdateMembers) != 1 || diff.UpdateMembers[0].Email != "alice@x.com" {
		t.Fatalf("unexpected update members: %+v", diff.UpdateMembers)
	}
}

func TestBuildSyncDiffSkipsManualSources(t *testing.T) {
	ext := "d-manual"
	diff := pkgorg.BuildSyncDiff(
		[]types.Department{manualDept("dept-manual", "Local"), importedDept("dept-1", "d1", "Eng")},
		[]types.Member{
			{ID: "m-manual", Name: "Bob", Source: types.MemberSourceManual, ExternalID: &ext},
			importedMember("m-1", "u1", "Alice", "a@x.com", "138"),
		},
		nil,
		[]datasource.RemoteMember{{ExternalID: "u-manual", Name: "Ghost"}},
	)
	if len(diff.RemoveDepartments) != 1 || diff.RemoveDepartments[0].ID != "dept-1" {
		t.Fatalf("manual dept should be skipped, got removes: %+v", diff.RemoveDepartments)
	}
	if len(diff.RemoveMembers) != 1 || diff.RemoveMembers[0].ID != "m-1" {
		t.Fatalf("manual member should be skipped, got removes: %+v", diff.RemoveMembers)
	}
}
