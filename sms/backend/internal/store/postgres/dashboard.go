package postgres

import (
	"context"

	"sms/backend/internal/domain/types"
)

func (s *Store) DashboardCards(ctx context.Context) (*types.DashboardCards, error) {
	var cards types.DashboardCards
	err := s.pool.QueryRow(ctx, `
		SELECT
			(SELECT count(*) FROM suppliers),
			(SELECT count(*) FROM suppliers WHERE status = 'active'),
			(SELECT count(*) FROM contracts WHERE status = 'active')
	`).Scan(&cards.SupplierTotal, &cards.ActiveSuppliers, &cards.ActiveContracts)
	return &cards, err
}

func (s *Store) DashboardCharts(ctx context.Context) (*types.DashboardCharts, error) {
	charts := &types.DashboardCharts{}

	rows, err := s.pool.Query(ctx,
		`SELECT grade, count(*) FROM evaluations GROUP BY grade ORDER BY grade`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var lc types.LabelCount
		if err := rows.Scan(&lc.Label, &lc.Count); err != nil {
			return nil, err
		}
		charts.GradeDistribution = append(charts.GradeDistribution, lc)
	}
	if charts.GradeDistribution == nil {
		charts.GradeDistribution = []types.LabelCount{}
	}

	return charts, nil
}

func (s *Store) ExpiringContracts(ctx context.Context, days int) ([]types.ExpiringContract, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT c.id, c.title, c.contract_no, c.end_date::text, sup.name
		 FROM contracts c JOIN suppliers sup ON sup.id = c.supplier_id
		 WHERE c.status = 'active' AND c.end_date <= CURRENT_DATE + $1
		 ORDER BY c.end_date`, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []types.ExpiringContract
	for rows.Next() {
		var ec types.ExpiringContract
		if err := rows.Scan(&ec.ID, &ec.Title, &ec.ContractNo, &ec.EndDate, &ec.SupplierName); err != nil {
			return nil, err
		}
		items = append(items, ec)
	}
	if items == nil {
		items = []types.ExpiringContract{}
	}
	return items, nil
}
