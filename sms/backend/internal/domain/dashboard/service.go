package dashboard

import (
	"context"

	"sms/backend/internal/domain/types"
)

type Store interface {
	DashboardCards(ctx context.Context) (*types.DashboardCards, error)
	DashboardCharts(ctx context.Context) (*types.DashboardCharts, error)
	ExpiringContracts(ctx context.Context, days int) ([]types.ExpiringContract, error)
	RecentOrders(ctx context.Context, limit int) ([]types.PurchaseOrder, error)
}

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) Cards(ctx context.Context) (*types.DashboardCards, error) {
	return s.store.DashboardCards(ctx)
}

func (s *Service) Charts(ctx context.Context) (*types.DashboardCharts, error) {
	return s.store.DashboardCharts(ctx)
}

func (s *Service) Expiring(ctx context.Context) ([]types.ExpiringContract, error) {
	return s.store.ExpiringContracts(ctx, 30)
}

func (s *Service) RecentOrders(ctx context.Context) ([]types.PurchaseOrder, error) {
	return s.store.RecentOrders(ctx, 5)
}
