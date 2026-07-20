package bootstrap

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func insertModels(ctx context.Context, exec TableWriter, companyID uuid.UUID, cfg Config) error {
	for _, m := range cfg.Models {
		modelID := deterministicModelID(companyID, m.CallType)
		if _, err := exec.Exec(ctx, `
			INSERT INTO models (id, company_id, call_type, name, provider, input_ratio, output_ratio, enabled)
			VALUES ($1, $2, $3, $4, $5, $6, $7, TRUE)
			ON CONFLICT (company_id, call_type) DO NOTHING
		`, modelID, companyID, m.CallType, m.Name, m.Provider, m.InputRatio, m.OutputRatio); err != nil {
			return fmt.Errorf("insert model %s: %w", m.CallType, err)
		}
	}
	return nil
}

func deterministicModelID(companyID uuid.UUID, callType string) uuid.UUID {
	return uuid.NewSHA1(companyID, []byte("model:"+callType))
}
