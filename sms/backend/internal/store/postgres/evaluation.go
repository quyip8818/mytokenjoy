package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"sms/backend/internal/domain/evaluation"
	"sms/backend/internal/domain/types"
	"sms/backend/internal/store"
)

func (s *Store) ListEvaluations(ctx context.Context, f evaluation.ListFilter) (*types.PagedResult[types.Evaluation], error) {
	var conditions []string
	var args []any
	idx := 1

	if f.SupplierID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("e.supplier_id = $%d", idx))
		args = append(args, f.SupplierID)
		idx++
	}
	if f.Period != "" {
		conditions = append(conditions, fmt.Sprintf("e.period = $%d", idx))
		args = append(args, f.Period)
		idx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	err := s.pool.QueryRow(ctx, "SELECT count(*) FROM evaluations e "+where, args...).Scan(&total)
	if err != nil {
		return nil, err
	}

	offset := (f.Page - 1) * f.PageSize
	query := fmt.Sprintf(
		`SELECT e.id, e.supplier_id, e.evaluator_id, e.period,
		        e.quality, e.performance, e.price, e.service, e.compliance,
		        e.total_score, e.grade, e.comment, e.created_at,
		        sup.name as supplier_name, u.real_name as evaluator_name
		 FROM evaluations e
		 LEFT JOIN suppliers sup ON sup.id = e.supplier_id
		 LEFT JOIN users u ON u.id = e.evaluator_id
		 %s ORDER BY e.created_at DESC LIMIT $%d OFFSET $%d`,
		where, idx, idx+1,
	)
	args = append(args, f.PageSize, offset)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []types.Evaluation
	for rows.Next() {
		var e types.Evaluation
		if err := rows.Scan(&e.ID, &e.SupplierID, &e.EvaluatorID, &e.Period,
			&e.Quality, &e.Performance, &e.Price, &e.Service, &e.Compliance,
			&e.TotalScore, &e.Grade, &e.Comment, &e.CreatedAt,
			&e.SupplierName, &e.EvaluatorName); err != nil {
			return nil, err
		}
		items = append(items, e)
	}
	if items == nil {
		items = []types.Evaluation{}
	}
	return &types.PagedResult[types.Evaluation]{Items: items, Total: total, Page: f.Page, PageSize: f.PageSize}, nil
}

func (s *Store) GetEvaluation(ctx context.Context, id uuid.UUID) (*types.Evaluation, error) {
	var e types.Evaluation
	err := s.pool.QueryRow(ctx,
		`SELECT e.id, e.supplier_id, e.evaluator_id, e.period,
		        e.quality, e.performance, e.price, e.service, e.compliance,
		        e.total_score, e.grade, e.comment, e.created_at,
		        sup.name, u.real_name
		 FROM evaluations e
		 LEFT JOIN suppliers sup ON sup.id = e.supplier_id
		 LEFT JOIN users u ON u.id = e.evaluator_id
		 WHERE e.id = $1`, id,
	).Scan(&e.ID, &e.SupplierID, &e.EvaluatorID, &e.Period,
		&e.Quality, &e.Performance, &e.Price, &e.Service, &e.Compliance,
		&e.TotalScore, &e.Grade, &e.Comment, &e.CreatedAt,
		&e.SupplierName, &e.EvaluatorName)
	return &e, store.WrapNotFound(err)
}

func (s *Store) CreateEvaluation(ctx context.Context, e *types.Evaluation) error {
	e.ID = newID()
	return s.pool.QueryRow(ctx,
		`INSERT INTO evaluations (id, supplier_id, evaluator_id, period, quality, performance, price, service, compliance, total_score, grade, comment)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		 RETURNING created_at`,
		e.ID, e.SupplierID, e.EvaluatorID, e.Period, e.Quality, e.Performance,
		e.Price, e.Service, e.Compliance, e.TotalScore, e.Grade, e.Comment,
	).Scan(&e.CreatedAt)
}

func (s *Store) UpdateEvaluation(ctx context.Context, id uuid.UUID, e *types.Evaluation) error {
	ct, err := s.pool.Exec(ctx,
		`UPDATE evaluations SET supplier_id=$1, period=$2, quality=$3, performance=$4,
		 price=$5, service=$6, compliance=$7, total_score=$8, grade=$9, comment=$10
		 WHERE id=$11`,
		e.SupplierID, e.Period, e.Quality, e.Performance,
		e.Price, e.Service, e.Compliance, e.TotalScore, e.Grade, e.Comment, id,
	)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return types.ErrNotFound
	}
	return nil
}

func (s *Store) DeleteEvaluation(ctx context.Context, id uuid.UUID) error {
	ct, err := s.pool.Exec(ctx, `DELETE FROM evaluations WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return types.ErrNotFound
	}
	return nil
}

func (s *Store) GetWeights(ctx context.Context) ([]types.EvaluationWeight, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, dimension, weight FROM evaluation_weights ORDER BY dimension`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []types.EvaluationWeight
	for rows.Next() {
		var w types.EvaluationWeight
		if err := rows.Scan(&w.ID, &w.Dimension, &w.Weight); err != nil {
			return nil, err
		}
		items = append(items, w)
	}
	if items == nil {
		items = []types.EvaluationWeight{}
	}
	return items, nil
}

func (s *Store) UpdateWeights(ctx context.Context, weights []types.EvaluationWeight) error {
	for _, w := range weights {
		_, err := s.pool.Exec(ctx,
			`UPDATE evaluation_weights SET weight = $1 WHERE dimension = $2`,
			w.Weight, w.Dimension,
		)
		if err != nil {
			return err
		}
	}
	return nil
}
