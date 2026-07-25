package evaluation_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"sms/backend/internal/domain/evaluation"
	"sms/backend/internal/domain/types"
)

// --- mock store ---

type mockStore struct {
	evals   []types.Evaluation
	weights []types.EvaluationWeight
}

func newMockStore() *mockStore {
	return &mockStore{
		weights: []types.EvaluationWeight{
			{Dimension: "quality", Weight: 20},
			{Dimension: "performance", Weight: 20},
			{Dimension: "price", Weight: 20},
			{Dimension: "service", Weight: 20},
			{Dimension: "compliance", Weight: 20},
		},
	}
}

func (m *mockStore) ListEvaluations(_ context.Context, f evaluation.ListFilter) (*types.PagedResult[types.Evaluation], error) {
	return &types.PagedResult[types.Evaluation]{
		Items: m.evals, Total: len(m.evals), Page: f.Page, PageSize: f.PageSize,
	}, nil
}

func (m *mockStore) GetEvaluation(_ context.Context, id uuid.UUID) (*types.Evaluation, error) {
	for i := range m.evals {
		if m.evals[i].ID == id {
			return &m.evals[i], nil
		}
	}
	return nil, types.ErrNotFound
}

func (m *mockStore) CreateEvaluation(_ context.Context, e *types.Evaluation) error {
	e.ID = uuid.Must(uuid.NewV7())
	m.evals = append(m.evals, *e)
	return nil
}

func (m *mockStore) UpdateEvaluation(_ context.Context, id uuid.UUID, e *types.Evaluation) error {
	for i := range m.evals {
		if m.evals[i].ID == id {
			e.ID = id
			m.evals[i] = *e
			return nil
		}
	}
	return types.ErrNotFound
}

func (m *mockStore) DeleteEvaluation(_ context.Context, id uuid.UUID) error {
	for i := range m.evals {
		if m.evals[i].ID == id {
			m.evals = append(m.evals[:i], m.evals[i+1:]...)
			return nil
		}
	}
	return types.ErrNotFound
}

func (m *mockStore) GetWeights(_ context.Context) ([]types.EvaluationWeight, error) {
	return m.weights, nil
}

func (m *mockStore) UpdateWeights(_ context.Context, weights []types.EvaluationWeight) error {
	m.weights = weights
	return nil
}

// --- tests ---

var testEvaluatorID = uuid.Must(uuid.NewV7())
var testSupplierID = uuid.Must(uuid.NewV7())

func newService() *evaluation.Service {
	return evaluation.NewService(newMockStore())
}

func TestCalcScore_AllFives(t *testing.T) {
	t.Parallel()
	svc := newService()
	// 各维度均为5，权重各20，加权和=5，*20=100
	score, grade := svc.CalcScore(5, 5, 5, 5, 5)
	if score != 100 {
		t.Fatalf("expected 100, got %v", score)
	}
	if grade != "A" {
		t.Fatalf("expected grade A, got %s", grade)
	}
}

func TestCalcScore_AllOnes(t *testing.T) {
	t.Parallel()
	svc := newService()
	// 各维度均为1，加权和=1，*20=20
	score, grade := svc.CalcScore(1, 1, 1, 1, 1)
	if score != 20 {
		t.Fatalf("expected 20, got %v", score)
	}
	if grade != "D" {
		t.Fatalf("expected grade D, got %s", grade)
	}
}

func TestCreate_Success(t *testing.T) {
	t.Parallel()
	svc := newService()
	e, err := svc.Create(context.Background(), testEvaluatorID, evaluation.CreateInput{
		SupplierID: testSupplierID, Period: "2024-Q1",
		Quality: 4, Performance: 4, Price: 4, Service: 4, Compliance: 4,
	})
	if err != nil {
		t.Fatal(err)
	}
	if e.ID == uuid.Nil {
		t.Fatal("expected non-nil ID")
	}
	if e.Grade == "" {
		t.Fatal("expected grade to be calculated")
	}
}

func TestCreate_ValidationError(t *testing.T) {
	t.Parallel()
	svc := newService()
	_, err := svc.Create(context.Background(), testEvaluatorID, evaluation.CreateInput{})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestUpdateWeights_MustSum100(t *testing.T) {
	t.Parallel()
	svc := newService()
	err := svc.UpdateWeights(context.Background(), []types.EvaluationWeight{
		{Dimension: "quality", Weight: 50},
		{Dimension: "performance", Weight: 10},
	})
	if err == nil {
		t.Fatal("expected error: weights must sum 100")
	}
}

func TestUpdateWeights_Success(t *testing.T) {
	t.Parallel()
	svc := newService()
	err := svc.UpdateWeights(context.Background(), []types.EvaluationWeight{
		{Dimension: "quality", Weight: 30},
		{Dimension: "performance", Weight: 25},
		{Dimension: "price", Weight: 20},
		{Dimension: "service", Weight: 15},
		{Dimension: "compliance", Weight: 10},
	})
	if err != nil {
		t.Fatal(err)
	}
}
