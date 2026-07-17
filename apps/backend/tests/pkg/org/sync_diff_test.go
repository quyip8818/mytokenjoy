package org_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/datasource"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
)

var (
	syncDept1   = uuid.MustParse("00000000-0000-7000-0000-000000000d01")
	syncMember1 = uuid.MustParse("00000000-0000-7000-0000-000000000e01")
	syncManual  = uuid.MustParse("00000000-0000-7000-0000-000000000d99")
	syncMManual = uuid.MustParse("00000000-0000-7000-0000-000000000e99")
)

func importedDept(id uuid.UUID, extID, name string) types.Department {
	ext := extID
	src := types.DeptSourceImported
	return types.Department{ID: id, Name: name, ExternalID: &ext, Source: &src}
}

func manualDept(id uuid.UUID, name string) types.Department {
	src := types.DeptSourceManual
	return types.Department{ID: id, Name: name, Source: &src}
}

func importedMember(id uuid.UUID, extID, name, email, phone string) types.Member {
	ext := extID
	return types.Member{
		ID: id, Name: name, Email: email, Phone: phone,
		ExternalID: &ext, Source: types.MemberSourceImported,
	}
}

func TestBuildSyncDiffRemoteAdditions(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	diff := pkgorg.BuildSyncDiff(
		[]types.Department{importedDept(syncDept1, "d1", "Eng")},
		[]types.Member{importedMember(syncMember1, "u1", "Alice", "a@x.com", "138")},
		nil, nil,
	)
	if len(diff.RemoveDepartments) != 1 || diff.RemoveDepartments[0].ID != syncDept1 {
		t.Fatalf("unexpected remove departments: %+v", diff.RemoveDepartments)
	}
	if len(diff.RemoveMembers) != 1 || diff.RemoveMembers[0].ID != syncMember1 {
		t.Fatalf("unexpected remove members: %+v", diff.RemoveMembers)
	}
}

func TestBuildSyncDiffRenames(t *testing.T) {
	t.Parallel()
	diff := pkgorg.BuildSyncDiff(
		[]types.Department{importedDept(syncDept1, "d1", "Eng")},
		[]types.Member{importedMember(syncMember1, "u1", "Alice", "a@x.com", "138")},
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
	t.Parallel()
	ext := "d-manual"
	diff := pkgorg.BuildSyncDiff(
		[]types.Department{manualDept(syncManual, "Local"), importedDept(syncDept1, "d1", "Eng")},
		[]types.Member{
			{ID: syncMManual, Name: "Bob", Source: types.MemberSourceManual, ExternalID: &ext},
			importedMember(syncMember1, "u1", "Alice", "a@x.com", "138"),
		},
		nil,
		[]datasource.RemoteMember{{ExternalID: "u-manual", Name: "Ghost"}},
	)
	if len(diff.RemoveDepartments) != 1 || diff.RemoveDepartments[0].ID != syncDept1 {
		t.Fatalf("manual dept should be skipped, got removes: %+v", diff.RemoveDepartments)
	}
	if len(diff.RemoveMembers) != 1 || diff.RemoveMembers[0].ID != syncMember1 {
		t.Fatalf("manual member should be skipped, got removes: %+v", diff.RemoveMembers)
	}
}
