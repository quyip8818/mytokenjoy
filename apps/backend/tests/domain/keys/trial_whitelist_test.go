//go:build testhook

package keys_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/ctxcompany"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestCreatePlatformKeyTrialRejectsRealModel(t *testing.T) {
	t.Parallel()
	svc, _, _ := newKeysServiceWithNewAPI(t)

	ctx := domaincompany.WithContext(testutil.Ctx(), ctxcompany.Info{
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
	testutil.AssertDomainStatus(t, err, domain.StatusUnprocessable)
}

func TestCreatePlatformKeyDemoRejectsRealModel(t *testing.T) {
	t.Parallel()
	svc, _, _ := newKeysServiceWithNewAPI(t)

	ctx := domaincompany.WithContext(testutil.Ctx(), ctxcompany.Info{
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
	testutil.AssertDomainStatus(t, err, domain.StatusUnprocessable)
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

func TestCreatePlatformKeyTrialRejectsMixedWhitelist(t *testing.T) {
	t.Parallel()
	svc, _, _ := newKeysServiceWithNewAPI(t)

	ctx := domaincompany.WithContext(testutil.Ctx(), ctxcompany.Info{
		CompanyID: contract.DefaultCompanyID,
		Type:      store.CompanyTypeTrial,
		Status:    store.CompanyStatusActive,
	})

	memberID := contract.IDMember1
	_, err := svc.CreatePlatformKey(ctx, types.CreatePlatformKeyInput{
		Name:           "trial-mixed",
		Scope:          types.PlatformKeyScopeMember,
		MemberID:       &memberID,
		Budget:         1000,
		ModelWhitelist: []uuid.UUID{contract.IDModelTest, contract.IDModel1}, // mix of test + real
	})
	testutil.AssertDomainStatus(t, err, domain.StatusUnprocessable)
}

func TestCreatePlatformKeyTrialRejectsUnknownModel(t *testing.T) {
	t.Parallel()
	svc, _, _ := newKeysServiceWithNewAPI(t)

	ctx := domaincompany.WithContext(testutil.Ctx(), ctxcompany.Info{
		CompanyID: contract.DefaultCompanyID,
		Type:      store.CompanyTypeTrial,
		Status:    store.CompanyStatusActive,
	})

	memberID := contract.IDMember1
	_, err := svc.CreatePlatformKey(ctx, types.CreatePlatformKeyInput{
		Name:           "trial-unknown",
		Scope:          types.PlatformKeyScopeMember,
		MemberID:       &memberID,
		Budget:         1000,
		ModelWhitelist: []uuid.UUID{uuid.MustParse("00000000-0000-7000-0000-ffffffffffff")},
	})
	testutil.AssertDomainStatus(t, err, domain.StatusUnprocessable)
}
