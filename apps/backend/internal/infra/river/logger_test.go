package riverinfra

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestQuietLoggerDropsProducerHeartbeat(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	base := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	logger := quietLogger(base)

	logger.Info("producer: Producer job counts",
		slog.Uint64("num_completed_jobs", 3),
		slog.Int("num_jobs_running", 0),
		slog.Int("num_jobs_stuck", 0),
		slog.String("queue", "default"),
	)
	logger.Info("River client started", slog.String("client_id", "abc"))
	logger.Info("producer: Producer job counts",
		slog.Uint64("num_completed_jobs", 3),
		slog.Int("num_jobs_running", 0),
		slog.Int("num_jobs_stuck", 2),
		slog.String("queue", "default"),
	)

	out := buf.String()
	if strings.Count(out, "Producer job counts") != 1 {
		t.Fatalf("expected only stuck heartbeat kept, got:\n%s", out)
	}
	if !strings.Contains(out, "River client started") {
		t.Fatalf("expected real info log kept, got:\n%s", out)
	}
	if !strings.Contains(out, `"num_jobs_stuck":2`) {
		t.Fatalf("expected stuck heartbeat kept, got:\n%s", out)
	}
}
