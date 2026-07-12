//go:build testhook

package gatewayfix

import (
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/newapiunits"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	orgfix "github.com/tokenjoy/backend/tests/testutil/org"
)

const defaultGatewayFullKey = "sk-test-gateway-key"

func normalizeGatewayOpts(opts GatewayScenarioOpts) GatewayScenarioOpts {
	if opts.CompanyID == 0 {
		opts.CompanyID = contract.DefaultCompanyID
	}
	if opts.DepartmentID == "" {
		opts.DepartmentID = contract.IDDept3
	}
	if opts.NewAPIWalletUserID == 0 {
		opts.NewAPIWalletUserID = 99
	}
	if opts.RemainQuota == 0 {
		opts.RemainQuota = 10000
	}
	if opts.CompanyStatus == "" {
		opts.CompanyStatus = store.CompanyStatusActive
	}
	return opts
}

func applyGatewayCompanyState(t *testing.T, cfg config.Config, st store.Store, opts GatewayScenarioOpts) {
	t.Helper()
	ctx := testutil.CtxForCompany(opts.CompanyID)

	walletPoint := 100000.0
	if opts.WalletBalancePoint != nil {
		walletPoint = *opts.WalletBalancePoint
	}
	if err := st.Company().SetWalletRemain(ctx, opts.CompanyID, walletPoint, nil); err != nil {
		t.Fatal(err)
	}
	if err := st.Company().UpdateNewAPIWalletUserID(ctx, opts.CompanyID, opts.NewAPIWalletUserID); err != nil {
		t.Fatal(err)
	}
	if opts.CompanyStatus != store.CompanyStatusActive {
		if err := st.Company().UpdateStatus(ctx, opts.CompanyID, opts.CompanyStatus); err != nil {
			t.Fatal(err)
		}
	}
}

func applyGatewayBudgetState(t *testing.T, cfg config.Config, st store.Store, opts GatewayScenarioOpts) {
	t.Helper()
	ctx := testutil.CtxForCompany(opts.CompanyID)

	tree, err := common.LoadBudgetTree(ctx, st.Org().Nodes())
	if err != nil {
		t.Fatal(err)
	}
	if !setBudgetOnTree(tree, opts.DepartmentID, opts.Budget, 0) {
		t.Fatalf("department %s not found in budget tree", opts.DepartmentID)
	}
	if err := orgfix.PersistBudgetTree(ctx, st, tree); err != nil {
		t.Fatal(err)
	}

	periodKey := pkgbudget.OpenSnapshotKey(pkgbudget.PeriodMonthly, cfg.Clock()).String()
	if err := st.BudgetConsumed().SetConsumed(ctx, store.AxisKindOrgNode, opts.DepartmentID, periodKey, opts.Consumed); err != nil {
		t.Fatal(err)
	}
}

type gatewayKeySetup struct {
	fullKey       string
	platformKeyID string
	memberID      string
}

func applyGatewayKeyMapping(t *testing.T, st store.Store, opts GatewayScenarioOpts) gatewayKeySetup {
	t.Helper()
	ctx := testutil.CtxForCompany(opts.CompanyID)

	members, err := st.Org().Members(ctx)
	if err != nil {
		t.Fatal(err)
	}
	memberID := contract.IDMember1
	if len(members) > 0 {
		memberID = members[0].ID
	}

	setup := gatewayKeySetup{fullKey: defaultGatewayFullKey, platformKeyID: contract.IDPlatformKey1}
	keys, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) == 0 {
		m := memberID
		keys = []types.PlatformKey{{
			ID:        "plk-gateway-test",
			Name:      "Gateway Test Key",
			KeyPrefix: "sk-test",
			FullKey:   &setup.fullKey,
			MemberID:  &m,
			Status:    "active",
			CreatedAt: "2026-06-19",
		}}
		setup.platformKeyID = keys[0].ID
	} else {
		found := false
		for i := range keys {
			if keys[i].ID == contract.IDPlatformKey1 {
				keys[i].FullKey = &setup.fullKey
				keys[i].Status = "active"
				found = true
			}
		}
		if !found {
			keys[0].FullKey = &setup.fullKey
			keys[0].Status = "active"
			setup.platformKeyID = keys[0].ID
		}
	}
	if err := st.Keys().SetPlatformKeys(ctx, keys); err != nil {
		t.Fatal(err)
	}

	for i := range keys {
		if keys[i].ID == setup.platformKeyID && keys[i].MemberID != nil {
			memberID = *keys[i].MemberID
			break
		}
	}
	for i := range members {
		if members[i].ID == memberID {
			members[i].DepartmentID = opts.DepartmentID
			break
		}
	}
	if err := st.Org().SetMembers(ctx, members); err != nil {
		t.Fatal(err)
	}
	setup.memberID = memberID

	tokenID := int64(42)
	remain := opts.RemainQuota
	if err := st.PlatformKeyMappings().UpsertMapping(ctx, store.PlatformKeyMapping{
		CompanyID:            opts.CompanyID,
		PlatformKeyID:        setup.platformKeyID,
		NewAPIKeyID:          &tokenID,
		MemberID:             testutil.StrPtr(memberID),
		DepartmentID:         opts.DepartmentID,
		SyncStatus:           store.MappingSyncStatusSynced,
		NewAPIGroup:          newapiunits.NewAPIGroupForDepartment(opts.DepartmentID),
		NewAPIKeyRemainQuota: &remain,
	}); err != nil {
		t.Fatal(err)
	}
	return setup
}

func ConfigureGatewayStore(t *testing.T, cfg config.Config, st store.Store, opts GatewayScenarioOpts) string {
	t.Helper()
	opts = normalizeGatewayOpts(opts)
	applyGatewayCompanyState(t, cfg, st, opts)
	applyGatewayBudgetState(t, cfg, st, opts)
	setup := applyGatewayKeyMapping(t, st, opts)
	return setup.fullKey
}
