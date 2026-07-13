package newapisync_test

import (
	"context"
	"testing"

	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/pkg/newapiunits"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

func TestSyncUpsertProviderKeyEnsuresPrivateGroup(t *testing.T) {
	t.Parallel()
	var ensuredGroup string
	stub := &mock.StubAdminClient{
		EnsureGroupFn: func(_ context.Context, group, _ string) error {
			ensuredGroup = group
			return nil
		},
		UpsertChannelFn: func(_ context.Context, req newapi.UpsertChannelRequest) (newapi.Channel, error) {
			if req.Group != ensuredGroup {
				t.Fatalf("expected channel group %q, got %q", ensuredGroup, req.Group)
			}
			return newapi.Channel{ID: 7}, nil
		},
	}
	sync, st := newSyncWithStubAndCfg(t, stub)
	ctx := testutil.Ctx()
	keys, err := st.Keys().ProviderKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) == 0 {
		t.Fatal("expected seeded provider keys")
	}
	pk := keys[0]
	pk.SecretKey = "sk-provider"
	keys[0] = pk
	if err := st.Keys().SetProviderKeys(ctx, keys); err != nil {
		t.Fatal(err)
	}
	if err := sync.SyncUpsertProviderKey(ctx, pk.ID); err != nil {
		t.Fatal(err)
	}
	want := newapiunits.NewAPIGroupForDepartment("dept-3")
	if ensuredGroup != want {
		t.Fatalf("expected ensured group %q, got %q", want, ensuredGroup)
	}
	if stub.EnsureGroupCalls != 1 || stub.UpsertChannelCalls != 1 {
		t.Fatalf("expected one EnsureGroup and one UpsertChannel, got ensure=%d upsert=%d", stub.EnsureGroupCalls, stub.UpsertChannelCalls)
	}
}

func TestSyncUpsertProviderKeyEnsuresSaaSGroup(t *testing.T) {
	t.Parallel()
	stub := &mock.StubAdminClient{
		EnsureGroupFn: func(_ context.Context, group, _ string) error {
			if group != "platform_shared" {
				t.Fatalf("expected platform_shared, got %q", group)
			}
			return nil
		},
		UpsertChannelFn: func(_ context.Context, req newapi.UpsertChannelRequest) (newapi.Channel, error) {
			if req.Group != "platform_shared" {
				t.Fatalf("expected platform_shared group, got %q", req.Group)
			}
			return newapi.Channel{ID: 8}, nil
		},
	}
	sync, st := newSyncWithStubAndCfg(t, stub, testutil.WithSupportSaas(true))
	ctx := testutil.Ctx()
	keys, err := st.Keys().ProviderKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	pk := keys[0]
	pk.SecretKey = "sk-provider"
	keys[0] = pk
	if err := st.Keys().SetProviderKeys(ctx, keys); err != nil {
		t.Fatal(err)
	}
	if err := sync.SyncUpsertProviderKey(ctx, pk.ID); err != nil {
		t.Fatal(err)
	}
}
