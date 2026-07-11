package usage_test

import (
	"errors"
	"testing"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
	newapisynctf "github.com/tokenjoy/backend/tests/testutil/newapisync"
	workerfix "github.com/tokenjoy/backend/tests/testutil/worker"
)

func TestIngestRejectsWhenDepartmentBudgetExceeded(t *testing.T) {
	t.Parallel()
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, RemainQuota: 1000}}
	_, st, _, ingest := workerfix.NewRunner(t, stub)
	ctx := testutil.Ctx()

	tree, err := common.LoadBudgetTree(ctx, st.Org().Nodes())
	if err != nil {
		t.Fatal(err)
	}
	node := pkgbudget.FindBudgetNode(tree, contract.IDDept3)
	if node == nil || node.Budget <= 0 {
		t.Fatalf("expected positive dept budget, got %+v", node)
	}
	testutil.SetDeptSnapshotConsumed(t, st, contract.IDDept3, node.Budget)

	newapisynctf.UpsertMapping(t, st, newapisynctf.DefaultMappingOpts())
	testutil.SeedConsumeLog(t, st, testutil.DefaultConsumeLog(91001, 99))

	err = ingest.IngestByLogID(ctx, 91001, types.SourceWebhook)
	if err == nil {
		t.Fatal("expected budget exceeded error")
	}
	var domainErr *domain.DomainError
	if !errors.As(err, &domainErr) || domainErr.Status != domain.StatusForbidden {
		t.Fatalf("expected forbidden domain error, got %v", err)
	}
}
