package modelcatalog_test

import (
	"context"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/adminport"
	"github.com/tokenjoy/backend/internal/domain/models"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/support/modelcatalog"
	"github.com/tokenjoy/backend/internal/support/simulate"
)

// --- Minimal mock models repo (only what CreateModel/ListModels/UpdateModel need) ---

type inMemModelsRepo struct {
	mu     sync.Mutex
	models []types.ModelInfo
}

func (r *inMemModelsRepo) Models(_ context.Context) ([]types.ModelInfo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]types.ModelInfo, len(r.models))
	copy(out, r.models)
	return out, nil
}
func (r *inMemModelsRepo) ModelByType(_ context.Context, modelType string) (*types.ModelInfo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.models {
		if r.models[i].Type == modelType {
			m := r.models[i]
			return &m, nil
		}
	}
	return nil, nil
}
func (r *inMemModelsRepo) ModelByProviderType(_ context.Context, provider, modelType string) (*types.ModelInfo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.models {
		if r.models[i].Provider == provider && r.models[i].Type == modelType {
			m := r.models[i]
			return &m, nil
		}
	}
	return nil, nil
}
func (r *inMemModelsRepo) GlobalModelByProviderType(context.Context, string, string) (*types.ModelInfo, error) {
	return nil, nil
}
func (r *inMemModelsRepo) ModelByID(_ context.Context, id uuid.UUID) (*types.ModelInfo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.models {
		if r.models[i].ID == id {
			m := r.models[i]
			return &m, nil
		}
	}
	return nil, nil
}
func (r *inMemModelsRepo) ModelByIDs(context.Context, []int64) ([]types.ModelInfo, error) {
	return nil, nil
}
func (r *inMemModelsRepo) InsertModel(_ context.Context, m types.ModelInfo) (types.ModelInfo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	m.ID = uuid.Must(uuid.NewV7())
	r.models = append(r.models, m)
	return m, nil
}
func (r *inMemModelsRepo) UpdateModel(_ context.Context, m types.ModelInfo) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.models {
		if r.models[i].ID == m.ID {
			r.models[i] = m
			return nil
		}
	}
	return nil
}
func (r *inMemModelsRepo) SyncFromPlatform(context.Context, uuid.UUID, []types.ModelInfo) error {
	return nil
}
func (r *inMemModelsRepo) Allowlist() store.ModelAllowlistRepository { return nil }

// modelsStore satisfies models.Store (narrow interface: Models() + Org()).
type modelsStore struct {
	repo *inMemModelsRepo
}

func (s *modelsStore) Models() store.ModelsRepository { return s.repo }
func (s *modelsStore) Org() store.OrgRepository       { return nil } // not needed for pricing tests

// --- Tracking admin client that records UpsertModelRatio calls ---

type pricingTracker struct {
	adminport.Port // embed nil interface — only override what we use
	mu             sync.Mutex
	upsertCalls    []ratioCall
	pricingData    []adminport.ModelPricing
}

type ratioCall struct {
	ModelType       string
	InputPrice      float64
	OutputPrice     float64
	CacheInputPrice float64
}

func (c *pricingTracker) UpsertModelRatio(_ context.Context, modelType string, inputPrice, outputPrice, cacheInputPrice float64) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.upsertCalls = append(c.upsertCalls, ratioCall{modelType, inputPrice, outputPrice, cacheInputPrice})
	// Also persist so ListModelPricing reflects the write (simulates real NewAPI behavior).
	mr, cr, ca := modelcatalog.RatioFromPrice(inputPrice, outputPrice, cacheInputPrice)
	for i := range c.pricingData {
		if c.pricingData[i].ModelName == modelType {
			c.pricingData[i].ModelRatio = mr
			c.pricingData[i].CompletionRatio = cr
			c.pricingData[i].CacheRatio = ca
			return nil
		}
	}
	c.pricingData = append(c.pricingData, adminport.ModelPricing{ModelName: modelType, ModelRatio: mr, CompletionRatio: cr, CacheRatio: ca})
	return nil
}

func (c *pricingTracker) ListModelPricing(_ context.Context) ([]adminport.ModelPricing, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]adminport.ModelPricing, len(c.pricingData))
	copy(out, c.pricingData)
	return out, nil
}

func (c *pricingTracker) RebuildAbilities(context.Context) error { return nil }

// --- Helper ---

func newTestSvc(client *pricingTracker, seedModels ...types.ModelInfo) models.Service {
	st := &modelsStore{repo: &inMemModelsRepo{models: seedModels}}
	cfg := config.Config{}
	cfg.TokenJoyCompanyID = uuid.MustParse("00000000-0000-7000-8000-000000000001")
	return models.NewService(cfg, st, client, nil, simulate.NewDelayer(false))
}

// --- Integration tests: verify TJ model write → NewAPI price update ---

func TestCreateModelPushesPrice(t *testing.T) {
	t.Parallel()
	client := &pricingTracker{}
	svc := newTestSvc(client)

	_, err := svc.CreateModel(context.Background(), types.CreateModelInput{
		Type:        "test-model",
		BaseURL:     "http://test",
		InputPrice:  2.5,
		OutputPrice: 10.0,
	})
	if err != nil {
		t.Fatal(err)
	}

	client.mu.Lock()
	defer client.mu.Unlock()
	if len(client.upsertCalls) != 1 {
		t.Fatalf("expected 1 UpsertModelRatio call, got %d", len(client.upsertCalls))
	}
	c := client.upsertCalls[0]
	if c.ModelType != "test-model" || c.InputPrice != 2.5 || c.OutputPrice != 10.0 {
		t.Errorf("got (%q, %f, %f), want (test-model, 2.5, 10.0)", c.ModelType, c.InputPrice, c.OutputPrice)
	}
}

func TestUpdateModelPushesPrice(t *testing.T) {
	t.Parallel()
	client := &pricingTracker{}
	svc := newTestSvc(client)

	created, err := svc.CreateModel(context.Background(), types.CreateModelInput{
		Type:        "update-model",
		BaseURL:     "http://test",
		InputPrice:  1.0,
		OutputPrice: 3.0,
	})
	if err != nil {
		t.Fatal(err)
	}

	newInput := 5.0
	newOutput := 15.0
	_, err = svc.UpdateModel(context.Background(), created.ID, types.UpdateModelInput{
		InputPrice:  &newInput,
		OutputPrice: &newOutput,
	})
	if err != nil {
		t.Fatal(err)
	}

	client.mu.Lock()
	defer client.mu.Unlock()
	if len(client.upsertCalls) != 2 {
		t.Fatalf("expected 2 calls (create + update), got %d", len(client.upsertCalls))
	}
	c := client.upsertCalls[1]
	if c.InputPrice != 5.0 || c.OutputPrice != 15.0 {
		t.Errorf("update push got (%f, %f), want (5.0, 15.0)", c.InputPrice, c.OutputPrice)
	}
}

func TestListModelsWithPricingMerges(t *testing.T) {
	t.Parallel()
	client := &pricingTracker{
		pricingData: []adminport.ModelPricing{
			{ModelName: "gpt-4o", ModelRatio: 1.25, CompletionRatio: 4.0},     // input=2.5, output=10.0
			{ModelName: "claude-3.5", ModelRatio: 0.75, CompletionRatio: 5.0}, // input=1.5, output=7.5
		},
	}
	seed := []types.ModelInfo{
		{ID: uuid.Must(uuid.NewV7()), Type: "gpt-4o", Name: "GPT-4o", Provider: "openai", Deprecated: false, Capabilities: []string{"chat"}, Source: "platform"},
		{ID: uuid.Must(uuid.NewV7()), Type: "claude-3.5", Name: "Claude", Provider: "anthropic", Deprecated: false, Capabilities: []string{"chat"}, Source: "platform"},
	}
	svc := newTestSvc(client, seed...)

	result, err := svc.ListModelsWithPricing(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 models, got %d", len(result))
	}

	for _, m := range result {
		switch m.Type {
		case "gpt-4o":
			if m.InputPrice != 2.5 || m.OutputPrice != 10.0 {
				t.Errorf("gpt-4o: got (%f, %f), want (2.5, 10.0)", m.InputPrice, m.OutputPrice)
			}
		case "claude-3.5":
			if m.InputPrice != 1.5 || m.OutputPrice != 7.5 {
				t.Errorf("claude: got (%f, %f), want (1.5, 7.5)", m.InputPrice, m.OutputPrice)
			}
		}
	}
}

func TestCreateThenListShowsPrice(t *testing.T) {
	t.Parallel()
	client := &pricingTracker{}
	svc := newTestSvc(client)

	_, err := svc.CreateModel(context.Background(), types.CreateModelInput{
		Type:        "e2e-model",
		BaseURL:     "http://test",
		InputPrice:  3.0,
		OutputPrice: 9.0,
	})
	if err != nil {
		t.Fatal(err)
	}

	// After create, ListModelsWithPricing should return the model with prices from NewAPI.
	result, err := svc.ListModelsWithPricing(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1, got %d", len(result))
	}
	if result[0].InputPrice != 3.0 || result[0].OutputPrice != 9.0 {
		t.Errorf("roundtrip: got (%f, %f), want (3.0, 9.0)", result[0].InputPrice, result[0].OutputPrice)
	}
}

func TestCreateModelZeroPriceSkipsPush(t *testing.T) {
	t.Parallel()
	client := &pricingTracker{}
	svc := newTestSvc(client)

	_, err := svc.CreateModel(context.Background(), types.CreateModelInput{
		Type:    "free-model",
		BaseURL: "http://test",
		// No price → should NOT push to NewAPI
	})
	if err != nil {
		t.Fatal(err)
	}

	client.mu.Lock()
	defer client.mu.Unlock()
	if len(client.upsertCalls) != 0 {
		t.Fatalf("expected 0 calls for zero-price, got %d", len(client.upsertCalls))
	}
}
