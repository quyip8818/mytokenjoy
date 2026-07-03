package app

import (
	"log/slog"

	"github.com/tokenjoy/backend/internal/config"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
)

func wireIngestService(cfg config.Config, i infra, logger *slog.Logger) *domainusage.IngestService {
	return domainusage.NewIngestService(cfg, i.store, i.notifier, logger)
}

func wireReader(i infra) domainusage.Reader {
	return domainusage.NewReader(i.store)
}

func wireReadModel(i infra) domainusage.ReadModel {
	return wireReader(i)
}
