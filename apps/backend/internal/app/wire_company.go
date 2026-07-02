package app

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/config"
	domainbilling "github.com/tokenjoy/backend/internal/domain/billing"
	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/store"
)

func wireCompany(cfg config.Config, i infra) domaincompany.Service {
	return domaincompany.NewService(cfg, i.store, i.adminClient)
}

func wireBilling(cfg config.Config, i infra) domainbilling.Service {
	rebalanceEnqueue := func(ctx context.Context, companyID int64) error {
		return i.store.Relay().EnqueueRebalance(ctx, store.RebalanceAxisCompany, fmt.Sprintf("%d", companyID))
	}
	return domainbilling.NewService(cfg, i.store, i.adminClient, i.wallet, rebalanceEnqueue)
}
