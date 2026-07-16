//go:build testhook

package budgetfix

import (
	"log/slog"
	"os"
	"testing"

	"github.com/tokenjoy/backend/internal/adapter"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/domain/newapisync"
	"github.com/tokenjoy/backend/internal/domain/newapisync/policy"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/infra/notification"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

func NewOverrunService(
	t *testing.T,
	cfg config.Config,
	st store.Store,
	stub *mock.StubAdminClient,
	notifier types.Notifier,
) *budget.OverrunService {
	t.Helper()
	newAPISync := newapisync.New(cfg, st, newapi.NewAdminPortAdapter(stub), policy.NewChannelPolicy(cfg), adapter.NewNewAPISyncEnqueuer(jobs.NoopEnqueuer{}))
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	if notifier == nil {
		notifier = notification.NewService(cfg, st, logger)
	}
	return budget.NewOverrunService(cfg, st, newAPISync, notifier, logger)
}
