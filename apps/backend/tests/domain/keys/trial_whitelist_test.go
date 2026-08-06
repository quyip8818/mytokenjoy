//go:build testhook

package keys_test

import (
	"testing"

	"github.com/google/uuid"
	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/support/tenant"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

// Trial/demo accounts can now create keys with real models (subject to routing rules).
// The old restriction ("试用账户只能使用 test-model") has been removed per model-allocation-design.

func TestCreatePlatformKeyTrialAllowsRealModel(t *testing.T) {
	t.Parallel()
	svc, _, _ := newKeysServiceWithNewAPI(t)

	ctx := domaincompany.WithContext(testutil.Ctx(), tenant.Info{
		CompanyID: contract.DefaultCompanyID,
		Type:      store.CompanyTypeTrial,
		Status:    store.CompanyStatusActive,
	})

	memberID := contract.IDMember1
	_, err := svc.CreatePlatformKey(ctx, types.CreatePlatformKeyInput{
		Name:           "trial-real-model",
		Scope:          types.PlatformKeyScopeMember,
		MemberID:       &memberID,
		Budget:         1000,
		ModelWhitelist: []uuid.UUID{contract.IDModel1}, // deepseek-v4-pro — real model
	})
	if err != nil {
		t.Fatalf("expected trial key with real model to succeed, got %v", err)
	}
}

func TestCreatePlatformKeyDemoAllowsRealModel(t *testing.T) {
	t.Parallel()
	svc, _, _ := newKeysServiceWithNewAPI(t)

	ctx := domaincompany.WithContext(testutil.Ctx(), tenant.Info{
		CompanyID: contract.DefaultCompanyID,
		Type:      store.CompanyTypeDemo,
		Status:    store.CompanyStatusActive,
	})

	memberID := contract.IDMember1
	_, err := svc.CreatePlatformKey(ctx, types.CreatePlatformKeyInput{
		Name:           "demo-real-model",
		Scope:          types.PlatformKeyScopeMember,
		MemberID:       &memberID,
		Budget:         1000,
		ModelWhitelist: []uuid.UUID{contract.IDModel1},
	})
	if err != nil {
		t.Fatalf("expected demo key with real model to succeed, got %v", err)
	}
}

func TestCreatePlatformKeyStandardAllowsRealModel(t *testing.T) {
	t.Parallel()
	svc, _, _ := newKeysServiceWithNewAPI(t)

	// Standard company (default context Type is empty, treated as standard).
	memberID := contract.IDMember1
	_, err := svc.CreatePlatformKey(testutil.Ctx(), types.CreatePlatformKeyInput{
		Name:           "standard-real-model",
		Scope:          types.PlatformKeyScopeMember,
		MemberID:       &memberID,
		Budget:         1000,
		ModelWhitelist: []uuid.UUID{contract.IDModel1},
	})
	if err != nil {
		t.Fatalf("expected standard key with real model to succeed, got %v", err)
	}
}

func TestCreatePlatformKeyRejectsUnknownModel(t *testing.T) {
	t.Parallel()
	svc, _, _ := newKeysServiceWithNewAPI(t)

	memberID := contract.IDMember1
	_, err := svc.CreatePlatformKey(testutil.Ctx(), types.CreatePlatformKeyInput{
		Name:           "unknown-model",
		Scope:          types.PlatformKeyScopeMember,
		MemberID:       &memberID,
		Budget:         1000,
		ModelWhitelist: []uuid.UUID{uuid.MustParse("00000000-0000-7000-0000-ffffffffffff")},
	})
	if err == nil {
		t.Fatal("expected unknown model to be rejected")
	}
}
