package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"sms/backend/internal/domain/model"
	"sms/backend/internal/domain/types"
	"sms/backend/internal/store"
)

func (s *Store) ListModels(ctx context.Context, f model.ListFilter) (*types.PagedResult[types.AiModel], error) {
	var conditions []string
	var args []any
	idx := 1

	if f.Keyword != "" {
		conditions = append(conditions, fmt.Sprintf("m.model_name ILIKE $%d", idx))
		args = append(args, "%"+f.Keyword+"%")
		idx++
	}
	if f.SupplierID != nil && *f.SupplierID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("m.supplier_id = $%d", idx))
		args = append(args, f.SupplierID)
		idx++
	}
	if f.ModelType != "" {
		conditions = append(conditions, fmt.Sprintf("m.model_type = $%d", idx))
		args = append(args, f.ModelType)
		idx++
	}
	if f.Status != "" {
		conditions = append(conditions, fmt.Sprintf("m.status = $%d", idx))
		args = append(args, f.Status)
		idx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	err := s.pool.QueryRow(ctx, "SELECT count(*) FROM models m "+where, args...).Scan(&total)
	if err != nil {
		return nil, err
	}

	offset := (f.Page - 1) * f.PageSize
	query := fmt.Sprintf(
		`SELECT m.id, m.supplier_id, m.model_name, m.model_id, m.model_type, m.context_length,
		        m.cost_input, m.cost_output, m.input_price, m.output_price, m.discount,
		        m.status, m.source, m.description, m.created_at, m.updated_at,
		        sup.name as supplier_name, ch.name as channel_name
		 FROM models m
		 LEFT JOIN suppliers sup ON sup.id = m.supplier_id
		 LEFT JOIN channels ch ON ch.id = m.channel_id
		 %s ORDER BY m.created_at DESC LIMIT $%d OFFSET $%d`,
		where, idx, idx+1,
	)
	args = append(args, f.PageSize, offset)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []types.AiModel
	for rows.Next() {
		var m types.AiModel
		if err := rows.Scan(&m.ID, &m.SupplierID, &m.ModelName, &m.ModelID, &m.ModelType, &m.ContextLength,
			&m.CostInput, &m.CostOutput, &m.InputPrice, &m.OutputPrice, &m.Discount,
			&m.Status, &m.Source, &m.Description, &m.CreatedAt, &m.UpdatedAt,
			&m.SupplierName, &m.ChannelName); err != nil {
			return nil, err
		}
		items = append(items, m)
	}
	if items == nil {
		items = []types.AiModel{}
	}
	return &types.PagedResult[types.AiModel]{Items: items, Total: total, Page: f.Page, PageSize: f.PageSize}, nil
}

func (s *Store) GetModel(ctx context.Context, id uuid.UUID) (*types.AiModel, error) {
	var m types.AiModel
	err := s.pool.QueryRow(ctx,
		`SELECT m.id, m.supplier_id, m.model_name, m.model_id, m.model_type, m.context_length,
		        m.cost_input, m.cost_output, m.input_price, m.output_price, m.discount,
		        m.status, m.source, m.description, m.created_at, m.updated_at,
		        sup.name as supplier_name, ch.name as channel_name
		 FROM models m
		 LEFT JOIN suppliers sup ON sup.id = m.supplier_id
		 LEFT JOIN channels ch ON ch.id = m.channel_id
		 WHERE m.id = $1`, id,
	).Scan(&m.ID, &m.SupplierID, &m.ModelName, &m.ModelID, &m.ModelType, &m.ContextLength,
		&m.CostInput, &m.CostOutput, &m.InputPrice, &m.OutputPrice, &m.Discount,
		&m.Status, &m.Source, &m.Description, &m.CreatedAt, &m.UpdatedAt,
		&m.SupplierName, &m.ChannelName)
	return &m, store.WrapNotFound(err)
}

func (s *Store) CreateModel(ctx context.Context, m *types.AiModel) error {
	m.ID = newID()
	err := s.pool.QueryRow(ctx,
		`INSERT INTO models (id, supplier_id, model_name, model_id, model_type, context_length, input_price, output_price, discount, status, description)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 RETURNING created_at, updated_at`,
		m.ID, m.SupplierID, m.ModelName, m.ModelID, m.ModelType, m.ContextLength, m.InputPrice, m.OutputPrice, m.Discount, m.Status, m.Description,
	).Scan(&m.CreatedAt, &m.UpdatedAt)
	return store.WrapConflict(err)
}

func (s *Store) UpdateModel(ctx context.Context, id uuid.UUID, m *types.AiModel) error {
	ct, err := s.pool.Exec(ctx,
		`UPDATE models SET supplier_id=$1, model_name=$2, model_id=$3, model_type=$4, context_length=$5,
		 input_price=$6, output_price=$7, discount=$8, status=$9, description=$10
		 WHERE id=$11`,
		m.SupplierID, m.ModelName, m.ModelID, m.ModelType, m.ContextLength, m.InputPrice, m.OutputPrice, m.Discount, m.Status, m.Description, id,
	)
	if err != nil {
		return store.WrapConflict(err)
	}
	if ct.RowsAffected() == 0 {
		return types.ErrNotFound
	}
	return nil
}

func (s *Store) DeleteModel(ctx context.Context, id uuid.UUID) error {
	ct, err := s.pool.Exec(ctx, `DELETE FROM models WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return types.ErrNotFound
	}
	return nil
}

func (s *Store) ListModelsWithModelID(ctx context.Context) ([]types.AiModel, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, supplier_id, model_name, model_id, model_type, context_length,
		        input_price, output_price, discount, status, description, created_at, updated_at
		 FROM models WHERE model_id IS NOT NULL AND model_id != ''`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []types.AiModel
	for rows.Next() {
		var m types.AiModel
		if err := rows.Scan(&m.ID, &m.SupplierID, &m.ModelName, &m.ModelID, &m.ModelType, &m.ContextLength,
			&m.InputPrice, &m.OutputPrice, &m.Discount, &m.Status, &m.Description,
			&m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, m)
	}
	if items == nil {
		items = []types.AiModel{}
	}
	return items, nil
}
