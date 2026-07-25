package evaluation

import (
	"context"
	"fmt"
	"math"
	"sync/atomic"

	"github.com/google/uuid"
	"sms/backend/internal/domain/types"
)

type Store interface {
	ListEvaluations(ctx context.Context, f ListFilter) (*types.PagedResult[types.Evaluation], error)
	GetEvaluation(ctx context.Context, id uuid.UUID) (*types.Evaluation, error)
	CreateEvaluation(ctx context.Context, e *types.Evaluation) error
	UpdateEvaluation(ctx context.Context, id uuid.UUID, e *types.Evaluation) error
	DeleteEvaluation(ctx context.Context, id uuid.UUID) error
	GetWeights(ctx context.Context) ([]types.EvaluationWeight, error)
	UpdateWeights(ctx context.Context, weights []types.EvaluationWeight) error
}

type ListFilter struct {
	Page       int
	PageSize   int
	SupplierID uuid.UUID
	Period     string
}

type CreateInput struct {
	SupplierID  uuid.UUID `json:"supplierId"`
	Period      string    `json:"period"`
	Quality     int       `json:"quality"`
	Performance int       `json:"performance"`
	Price       int       `json:"price"`
	Service     int       `json:"service"`
	Compliance  int       `json:"compliance"`
	Comment     *string   `json:"comment"`
}

type UpdateInput = CreateInput

type Service struct {
	store   Store
	weights atomic.Pointer[map[string]int]
}

func NewService(store Store) *Service {
	s := &Service{store: store}
	s.refreshWeights(context.Background())
	return s
}

func (s *Service) refreshWeights(ctx context.Context) {
	weights, _ := s.store.GetWeights(ctx)
	m := make(map[string]int, len(weights))
	for _, w := range weights {
		m[w.Dimension] = w.Weight
	}
	s.weights.Store(&m)
}

func (s *Service) CalcScore(quality, performance, price, service, compliance int) (float64, string) {
	w := *s.weights.Load()
	total := float64(quality)*float64(w["quality"])/100 +
		float64(performance)*float64(w["performance"])/100 +
		float64(price)*float64(w["price"])/100 +
		float64(service)*float64(w["service"])/100 +
		float64(compliance)*float64(w["compliance"])/100
	score := math.Round(total*20*100) / 100
	return score, toGrade(score)
}

func toGrade(score float64) string {
	switch {
	case score >= 90:
		return "A"
	case score >= 80:
		return "B"
	case score >= 60:
		return "C"
	default:
		return "D"
	}
}

func (s *Service) List(ctx context.Context, f ListFilter) (*types.PagedResult[types.Evaluation], error) {
	if f.PageSize > 100 {
		f.PageSize = 100
	}
	return s.store.ListEvaluations(ctx, f)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*types.Evaluation, error) {
	return s.store.GetEvaluation(ctx, id)
}

func (s *Service) Create(ctx context.Context, evaluatorID uuid.UUID, input CreateInput) (*types.Evaluation, error) {
	if input.SupplierID == uuid.Nil || input.Period == "" {
		return nil, fmt.Errorf("%w: 供应商和评估周期不能为空", types.ErrValidation)
	}
	score, grade := s.CalcScore(input.Quality, input.Performance, input.Price, input.Service, input.Compliance)
	e := &types.Evaluation{
		SupplierID:  input.SupplierID,
		EvaluatorID: evaluatorID,
		Period:      input.Period,
		Quality:     input.Quality,
		Performance: input.Performance,
		Price:       input.Price,
		Service:     input.Service,
		Compliance:  input.Compliance,
		TotalScore:  score,
		Grade:       grade,
		Comment:     input.Comment,
	}
	if err := s.store.CreateEvaluation(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateInput) (*types.Evaluation, error) {
	score, grade := s.CalcScore(input.Quality, input.Performance, input.Price, input.Service, input.Compliance)
	e := &types.Evaluation{
		SupplierID:  input.SupplierID,
		Period:      input.Period,
		Quality:     input.Quality,
		Performance: input.Performance,
		Price:       input.Price,
		Service:     input.Service,
		Compliance:  input.Compliance,
		TotalScore:  score,
		Grade:       grade,
		Comment:     input.Comment,
	}
	if err := s.store.UpdateEvaluation(ctx, id, e); err != nil {
		return nil, err
	}
	e.ID = id
	return e, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.store.DeleteEvaluation(ctx, id)
}

func (s *Service) GetWeights(ctx context.Context) ([]types.EvaluationWeight, error) {
	return s.store.GetWeights(ctx)
}

func (s *Service) UpdateWeights(ctx context.Context, weights []types.EvaluationWeight) error {
	total := 0
	for _, w := range weights {
		total += w.Weight
	}
	if total != 100 {
		return fmt.Errorf("%w: 权重合计必须为 100", types.ErrValidation)
	}
	if err := s.store.UpdateWeights(ctx, weights); err != nil {
		return err
	}
	s.refreshWeights(ctx)
	return nil
}
