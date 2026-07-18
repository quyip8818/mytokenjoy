//go:build testhook

package provision_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/postgres"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
	newapisynctf "github.com/tokenjoy/backend/tests/testutil/newapisync"
)

func TestBootstrapDemoWalletUserCreatesWallet(t *testing.T) {
	t.Parallel()
	var createdUsername string
	var nextTokenID int64 = 500
	stub := &mock.StubAdminClient{
		User: newapi.User{ID: 501},
		CreateUserFn: func(_ context.Context, req newapi.CreateUserRequest) (newapi.User, error) {
			createdUsername = req.Username
			return newapi.User{ID: 501}, nil
		},
		CreateTokenFn: func(_ context.Context, _ newapi.CreateTokenRequest) (newapi.Token, error) {
			nextTokenID++
			return newapi.Token{ID: nextTokenID, Key: fmt.Sprintf("sk-bootstrap-%d", nextTokenID), RemainQuota: 1000}, nil
		},
		GetTokenFn: func(_ context.Context, tokenID int64) (newapi.Token, error) {
			return newapi.Token{ID: tokenID, Key: fmt.Sprintf("sk-bootstrap-%d", tokenID), RemainQuota: 1000}, nil
		},
	}
	sync, st := newapisynctf.NewLocalTestService(t, stub)
	ctx := testutil.CtxForCompany(contract.LocalCompanyID)
	if err := sync.Bootstrap(ctx, contract.LocalCompanyID); err != nil {
		t.Fatal(err)
	}
	if stub.CreateUserCalls != 1 {
		t.Fatalf("expected one CreateUser call, got %d", stub.CreateUserCalls)
	}
	expectedUsername := company.WalletUsername(contract.LocalCompanyID)
	if createdUsername != expectedUsername {
		t.Fatalf("expected username %q, got %q", expectedUsername, createdUsername)
	}
	co, err := st.Company().GetByID(ctx, contract.LocalCompanyID)
	if err != nil {
		t.Fatal(err)
	}
	if co.NewAPIWalletUserID == nil || *co.NewAPIWalletUserID != 501 {
		t.Fatalf("expected wallet user 501, got %v", co.NewAPIWalletUserID)
	}
}

func TestBootstrapSyncsActiveSeedKey(t *testing.T) {
	t.Parallel()
	var nextTokenID int64 = 800
	stub := &mock.StubAdminClient{
		CreateTokenFn: func(_ context.Context, _ newapi.CreateTokenRequest) (newapi.Token, error) {
			nextTokenID++
			return newapi.Token{ID: nextTokenID, Key: fmt.Sprintf("sk-bootstrap-%d", nextTokenID), RemainQuota: 1000}, nil
		},
		GetTokenFn: func(_ context.Context, tokenID int64) (newapi.Token, error) {
			return newapi.Token{ID: tokenID, Key: fmt.Sprintf("sk-bootstrap-%d", tokenID), RemainQuota: 1000}, nil
		},
	}
	sync, st := newapisynctf.NewLocalTestService(t, stub)
	ctx := testutil.CtxForCompany(contract.LocalCompanyID)
	if err := sync.Bootstrap(ctx, contract.LocalCompanyID); err != nil {
		t.Fatal(err)
	}
	mapping, err := st.PlatformKeyMappings().GetMappingByPlatformKeyID(ctx, contract.IDPlatformKey1)
	if err != nil {
		t.Fatal(err)
	}
	if mapping == nil || mapping.SyncStatus != store.MappingSyncStatusSynced {
		t.Fatalf("expected synced mapping for seed key, got %+v", mapping)
	}
	hash, ok, err := st.Keys().PlatformKeyHashByID(ctx, contract.IDPlatformKey1)
	if err != nil || !ok || hash == store.HashPlatformKey("pending:"+contract.IDPlatformKey1.String()) {
		t.Fatalf("expected non-pending key hash, ok=%v err=%v hash=%s", ok, err, hash)
	}
}

func TestBootstrapRepairsPendingHashOnSyncedMapping(t *testing.T) {
	t.Parallel()
	tokenID := int64(66)
	const bearer = "sk-repair-hash"
	var nextTokenID int64 = 700
	stub := &mock.StubAdminClient{
		GetTokenKeyFn: func(_ context.Context, id int64) (string, error) {
			if id != tokenID {
				t.Fatalf("unexpected token id %d", id)
			}
			return bearer, nil
		},
		CreateTokenFn: func(_ context.Context, _ newapi.CreateTokenRequest) (newapi.Token, error) {
			nextTokenID++
			return newapi.Token{ID: nextTokenID, Key: fmt.Sprintf("sk-other-%d", nextTokenID), RemainQuota: 1000}, nil
		},
	}
	sync, st := newapisynctf.NewLocalTestService(t, stub)
	ctx := testutil.CtxForCompany(contract.LocalCompanyID)
	if err := st.PlatformKeyMappings().UpsertMapping(ctx, store.PlatformKeyMapping{
		CompanyID:     contract.LocalCompanyID,
		PlatformKeyID: contract.IDPlatformKey1,
		DepartmentID:  contract.IDDept3,
		NewAPIGroup:   "dept-dept-3",
		SyncStatus:    store.MappingSyncStatusSynced,
		NewAPIKeyID:   &tokenID,
	}); err != nil {
		t.Fatal(err)
	}
	keys, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for i := range keys {
		if keys[i].ID != contract.IDPlatformKey1 {
			continue
		}
		keys[i].FullKey = nil
		keys[i].KeyPrefix = "pending..."
		if err := st.Keys().SetPlatformKeys(ctx, keys); err != nil {
			t.Fatal(err)
		}
		break
	}
	if err := sync.Bootstrap(ctx, contract.LocalCompanyID); err != nil {
		t.Fatal(err)
	}
	hash, ok, err := st.Keys().PlatformKeyHashByID(ctx, contract.IDPlatformKey1)
	if err != nil || !ok || hash != store.HashPlatformKey(bearer) {
		t.Fatalf("expected repaired hash, ok=%v err=%v hash=%s", ok, err, hash)
	}
}

func TestReconcileMissingTokenRecreatesSync(t *testing.T) {
	t.Parallel()
	staleID := int64(99)
	var nextTokenID int64 = 900
	stub := &mock.StubAdminClient{
		CreateTokenFn: func(_ context.Context, _ newapi.CreateTokenRequest) (newapi.Token, error) {
			nextTokenID++
			return newapi.Token{ID: nextTokenID, Key: fmt.Sprintf("sk-recreated-%d", nextTokenID), RemainQuota: 1000}, nil
		},
		GetTokenFn: func(_ context.Context, id int64) (newapi.Token, error) {
			if id == staleID {
				return newapi.Token{}, errors.New("token not found")
			}
			return newapi.Token{ID: id, Key: fmt.Sprintf("sk-recreated-%d", id), RemainQuota: 1000}, nil
		},
	}
	sync, st := newapisynctf.NewLocalTestService(t, stub)
	ctx := testutil.CtxForCompany(contract.LocalCompanyID)
	if err := st.PlatformKeyMappings().UpsertMapping(ctx, store.PlatformKeyMapping{
		CompanyID:     contract.LocalCompanyID,
		PlatformKeyID: contract.IDPlatformKey1,
		DepartmentID:  contract.IDDept3,
		NewAPIGroup:   "dept-dept-3",
		SyncStatus:    store.MappingSyncStatusSynced,
		NewAPIKeyID:   &staleID,
	}); err != nil {
		t.Fatal(err)
	}
	if err := sync.Bootstrap(ctx, contract.LocalCompanyID); err != nil {
		t.Fatal(err)
	}
	mapping, err := st.PlatformKeyMappings().GetMappingByPlatformKeyID(ctx, contract.IDPlatformKey1)
	if err != nil {
		t.Fatal(err)
	}
	if mapping == nil || mapping.SyncStatus != store.MappingSyncStatusSynced {
		t.Fatalf("expected synced mapping after reconcile, got %+v", mapping)
	}
	if mapping.NewAPIKeyID == nil || *mapping.NewAPIKeyID == staleID {
		t.Fatalf("expected replaced newapi key id, got %v", mapping.NewAPIKeyID)
	}
	hash, ok, err := st.Keys().PlatformKeyHashByID(ctx, contract.IDPlatformKey1)
	if err != nil || !ok || hash == store.HashPlatformKey("pending:"+contract.IDPlatformKey1.String()) {
		t.Fatalf("expected non-pending key hash, ok=%v err=%v hash=%s", ok, err, hash)
	}
}

func TestBootstrapSkipsWhenAllReady(t *testing.T) {
	t.Parallel()
	var nextTokenID int64 = 800
	stub := &mock.StubAdminClient{
		CreateTokenFn: func(_ context.Context, _ newapi.CreateTokenRequest) (newapi.Token, error) {
			nextTokenID++
			return newapi.Token{ID: nextTokenID, Key: fmt.Sprintf("sk-bootstrap-%d", nextTokenID), RemainQuota: 1000}, nil
		},
		GetTokenFn: func(_ context.Context, tokenID int64) (newapi.Token, error) {
			return newapi.Token{ID: tokenID, Key: fmt.Sprintf("sk-bootstrap-%d", tokenID), Group: "dept-dept-3", RemainQuota: 1000}, nil
		},
	}
	sync, st := newapisynctf.NewLocalTestService(t, stub)
	ctx := testutil.CtxForCompany(contract.LocalCompanyID)
	if err := sync.Bootstrap(ctx, contract.LocalCompanyID); err != nil {
		t.Fatal(err)
	}
	callsAfterFirst := stub.CreateTokenCalls
	if err := sync.Bootstrap(ctx, contract.LocalCompanyID); err != nil {
		t.Fatal(err)
	}
	if stub.CreateTokenCalls != callsAfterFirst {
		t.Fatalf("expected no CreateToken calls on second bootstrap, before=%d after=%d", callsAfterFirst, stub.CreateTokenCalls)
	}
	_ = st
}

func TestBootstrapHealsZeroWalletUserID(t *testing.T) {
	t.Parallel()
	var nextTokenID int64 = 900
	stub := &mock.StubAdminClient{
		CreateUserFn: func(_ context.Context, _ newapi.CreateUserRequest) (newapi.User, error) {
			return newapi.User{ID: 777}, nil
		},
		CreateTokenFn: func(_ context.Context, _ newapi.CreateTokenRequest) (newapi.Token, error) {
			nextTokenID++
			return newapi.Token{ID: nextTokenID, Key: fmt.Sprintf("sk-bootstrap-%d", nextTokenID), RemainQuota: 1000}, nil
		},
		GetTokenFn: func(_ context.Context, tokenID int64) (newapi.Token, error) {
			return newapi.Token{ID: tokenID, Key: fmt.Sprintf("sk-bootstrap-%d", tokenID), RemainQuota: 1000}, nil
		},
	}
	sync, st := newapisynctf.NewLocalTestService(t, stub)
	ctx := testutil.CtxForCompany(contract.LocalCompanyID)
	if err := sync.Bootstrap(ctx, contract.LocalCompanyID); err != nil {
		t.Fatal(err)
	}
	pool := postgres.MainPool(st)
	if _, err := pool.Exec(ctx, `UPDATE companies SET newapi_wallet_user_id = 0 WHERE id = $1`, contract.LocalCompanyID); err != nil {
		t.Fatal(err)
	}
	createCallsBefore := stub.CreateUserCalls
	if err := sync.Bootstrap(ctx, contract.LocalCompanyID); err != nil {
		t.Fatal(err)
	}
	if stub.CreateUserCalls <= createCallsBefore {
		t.Fatalf("expected CreateUser on heal, before=%d after=%d", createCallsBefore, stub.CreateUserCalls)
	}
	co, err := st.Company().GetByID(ctx, contract.LocalCompanyID)
	if err != nil {
		t.Fatal(err)
	}
	if co.NewAPIWalletUserID == nil || *co.NewAPIWalletUserID != 777 {
		t.Fatalf("expected wallet user 777 after heal, got %v", co.NewAPIWalletUserID)
	}
}
