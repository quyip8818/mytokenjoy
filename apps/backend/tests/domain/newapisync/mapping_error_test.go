//go:build testhook

package newapisync_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/newapisync/platformkey"
	"github.com/tokenjoy/backend/internal/domain/newapisync/syncdeps"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

type errMappings struct {
	err error
}

func (e errMappings) GetMappingByPlatformKeyID(context.Context, uuid.UUID) (*store.PlatformKeyMapping, error) {
	return nil, e.err
}
func (e errMappings) GetMappingByKeyHash(context.Context, string) (*store.PlatformKeyMapping, error) {
	return nil, e.err
}
func (e errMappings) FindMappingByNewAPIKeyID(context.Context, int64) (*store.PlatformKeyMapping, error) {
	return nil, e.err
}
func (e errMappings) ListMappingsByNewAPIKeyIDs(context.Context, []int64) ([]store.PlatformKeyMapping, error) {
	return nil, e.err
}
func (e errMappings) ListMappingsByMemberID(context.Context, uuid.UUID) ([]store.PlatformKeyMapping, error) {
	return nil, e.err
}
func (e errMappings) ListMappingsByDepartmentID(context.Context, uuid.UUID) ([]store.PlatformKeyMapping, error) {
	return nil, e.err
}
func (e errMappings) ListMappingsByProjectID(context.Context, uuid.UUID) ([]store.PlatformKeyMapping, error) {
	return nil, e.err
}
func (e errMappings) ListMappingsByPlatformKeyIDs(context.Context, []uuid.UUID) ([]store.PlatformKeyMapping, error) {
	return nil, e.err
}
func (e errMappings) ListActiveMappingsByCompany(context.Context, uuid.UUID) ([]store.PlatformKeyMapping, error) {
	return nil, e.err
}
func (e errMappings) UpsertMapping(context.Context, store.PlatformKeyMapping) error { return e.err }
func (e errMappings) UpdateMappingSync(context.Context, uuid.UUID, int64, string, time.Time) error {
	return e.err
}

func TestDisablePlatformKey_MappingLookupError(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	want := errors.New("mapping db down")
	d := syncdeps.Deps{
		Cfg:      cfg,
		Store:    st,
		Client:   newapi.NewAdminPortAdapter(&mock.StubAdminClient{}),
		Mappings: errMappings{err: want},
	}
	err := platformkey.DisablePlatformKey(testutil.Ctx(), d, uuid.MustParse("00000000-0000-7000-0000-00000000bb01"))
	if !errors.Is(err, want) {
		t.Fatalf("expected mapping error, got %v", err)
	}
}

func TestSyncRevokePlatformKey_MappingLookupError(t *testing.T) {
	t.Parallel()
	want := fmt.Errorf("mapping db down")
	d := syncdeps.Deps{
		Cfg:      config.Config{NewAPIConfig: config.NewAPIConfig{NewAPIEnabled: true}},
		Client:   newapi.NewAdminPortAdapter(&mock.StubAdminClient{}),
		Mappings: errMappings{err: want},
	}
	err := platformkey.SyncRevokePlatformKey(context.Background(), d, uuid.MustParse("00000000-0000-7000-0000-00000000bb01"))
	if !errors.Is(err, want) {
		t.Fatalf("expected mapping error, got %v", err)
	}
}

func TestSyncRevokePlatformKey_MissingMappingNoop(t *testing.T) {
	t.Parallel()
	d := syncdeps.Deps{
		Cfg:      config.Config{NewAPIConfig: config.NewAPIConfig{NewAPIEnabled: true}},
		Client:   newapi.NewAdminPortAdapter(&mock.StubAdminClient{}),
		Mappings: errMappings{err: nil},
	}
	if err := platformkey.SyncRevokePlatformKey(context.Background(), d, uuid.MustParse("00000000-0000-7000-0000-00000000bb02")); err != nil {
		t.Fatalf("expected noop, got %v", err)
	}
}
