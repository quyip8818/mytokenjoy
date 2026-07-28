//go:build testhook

package smssync_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/riverqueue/river"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/infra/river/periodic"
	"github.com/tokenjoy/backend/internal/integration/sms"
	"github.com/tokenjoy/backend/internal/worker/smssync"
)

// ---------------------------------------------------------------------------
// Test doubles
// ---------------------------------------------------------------------------

type stubFetcher struct {
	catalog *sms.Catalog
	called  atomic.Int32
}

func (s *stubFetcher) FetchCatalog(_ context.Context) (*sms.Catalog, error) {
	s.called.Add(1)
	return s.catalog, nil
}

type stubTarget struct {
	called atomic.Int32
}

func (s *stubTarget) UpsertChannel(_ context.Context, _ sms.CatalogChannel) error {
	s.called.Add(1)
	return nil
}
func (s *stubTarget) UpsertModelRatio(_ context.Context, _ string, _, _ float64) error {
	s.called.Add(1)
	return nil
}
func (s *stubTarget) UpsertModel(_ context.Context, _ sms.CatalogModel) error {
	s.called.Add(1)
	return nil
}
func (s *stubTarget) RebuildAbilities(_ context.Context) error { return nil }
func (s *stubTarget) DisableStaleModels(_ context.Context, _ []string) (int, error) {
	return 0, nil
}

// ---------------------------------------------------------------------------
// 1. Job args registration — SMSSyncArgs implements river.JobArgs
// ---------------------------------------------------------------------------

func TestSMSSyncArgs_Kind(t *testing.T) {
	t.Parallel()

	args := jobs.SMSSyncArgs{}
	kind := args.Kind()
	if kind == "" {
		t.Fatal("SMSSyncArgs.Kind() must return a non-empty string")
	}
	if kind != jobs.KindSMSSync {
		t.Fatalf("SMSSyncArgs.Kind() = %q, want %q", kind, jobs.KindSMSSync)
	}
}

func TestSMSSyncArgs_InsertOpts(t *testing.T) {
	t.Parallel()

	args := jobs.SMSSyncArgs{}
	opts := args.InsertOpts()

	// Should use the low-priority queue (non-critical periodic work)
	if opts.Queue != config.RiverQueueLow {
		t.Fatalf("InsertOpts.Queue = %q, want %q", opts.Queue, config.RiverQueueLow)
	}
}

// Compile-time check: SMSSyncArgs satisfies river.JobArgs interface.
var _ river.JobArgs = jobs.SMSSyncArgs{}

// ---------------------------------------------------------------------------
// 2. River worker execution — Work() delegates to smssync.Worker.Execute
// ---------------------------------------------------------------------------

func TestSMSSyncRiverWorker_Work_CallsExecute(t *testing.T) {
	t.Parallel()

	fetcher := &stubFetcher{catalog: &sms.Catalog{
		Models: []sms.CatalogModel{
			{ModelID: "gpt-4o", DisplayName: "GPT-4o", InputPrice: 5, OutputPrice: 15},
		},
	}}
	target := &stubTarget{}

	w := smssync.New(fetcher, target)

	// Simulate what the River worker's Work() method should do: call Execute.
	err := w.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute returned unexpected error: %v", err)
	}
	if fetcher.called.Load() != 1 {
		t.Fatalf("expected fetcher called once, got %d", fetcher.called.Load())
	}
	// UpsertModelRatio + UpsertModel = 2 target calls for 1 model
	if target.called.Load() < 2 {
		t.Fatalf("expected at least 2 target calls, got %d", target.called.Load())
	}
}

// ---------------------------------------------------------------------------
// 3. Periodic registration — SMS sync periodic job gated on config
// ---------------------------------------------------------------------------

func TestBuildSMSSyncPeriodicJobs_Enabled(t *testing.T) {
	t.Parallel()

	cfg := config.Config{}
	cfg.RiverEnabled = true
	cfg.RiverPeriodicEnabled = true
	cfg.SMSSyncEnabled = true
	cfg.SMSSyncIntervalSec = 300

	periodicJobs := periodic.BuildSMSSyncJobs(cfg)
	if len(periodicJobs) == 0 {
		t.Fatal("expected at least one periodic job when SMS sync is enabled")
	}
}

func TestBuildSMSSyncPeriodicJobs_DisabledWhenRiverOff(t *testing.T) {
	t.Parallel()

	cfg := config.Config{}
	cfg.RiverEnabled = false
	cfg.RiverPeriodicEnabled = true
	cfg.SMSSyncEnabled = true

	periodicJobs := periodic.BuildSMSSyncJobs(cfg)
	if len(periodicJobs) != 0 {
		t.Fatalf("expected no periodic jobs when River is disabled, got %d", len(periodicJobs))
	}
}

func TestBuildSMSSyncPeriodicJobs_DisabledWhenPeriodicOff(t *testing.T) {
	t.Parallel()

	cfg := config.Config{}
	cfg.RiverEnabled = true
	cfg.RiverPeriodicEnabled = false
	cfg.SMSSyncEnabled = true

	periodicJobs := periodic.BuildSMSSyncJobs(cfg)
	if len(periodicJobs) != 0 {
		t.Fatalf("expected no periodic jobs when periodic is disabled, got %d", len(periodicJobs))
	}
}

func TestBuildSMSSyncPeriodicJobs_DisabledWhenSMSOff(t *testing.T) {
	t.Parallel()

	cfg := config.Config{}
	cfg.RiverEnabled = true
	cfg.RiverPeriodicEnabled = true
	cfg.SMSSyncEnabled = false

	periodicJobs := periodic.BuildSMSSyncJobs(cfg)
	if len(periodicJobs) != 0 {
		t.Fatalf("expected no periodic jobs when SMS sync is disabled, got %d", len(periodicJobs))
	}
}

// ---------------------------------------------------------------------------
// 4. Graceful fallback — ticker still works when River is off
// ---------------------------------------------------------------------------

func TestSMSSyncWorker_TickerFallback(t *testing.T) {
	t.Parallel()

	fetcher := &stubFetcher{catalog: &sms.Catalog{
		Models: []sms.CatalogModel{
			{ModelID: "test-model", DisplayName: "Test", InputPrice: 1, OutputPrice: 2},
		},
	}}
	target := &stubTarget{}

	// Use a very short interval to verify the ticker fires
	w := smssync.NewWithInterval(fetcher, target, 50*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// Run blocks until ctx is cancelled — runs in goroutine like the fallback path
	done := make(chan struct{})
	go func() {
		w.Run(ctx)
		close(done)
	}()

	<-done

	// Should have been called at least twice: once immediately + at least one ticker fire
	calls := fetcher.called.Load()
	if calls < 2 {
		t.Fatalf("expected fetcher called at least 2 times (immediate + ticker), got %d", calls)
	}
}
