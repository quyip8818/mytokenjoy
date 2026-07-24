package bootstrap

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/store"
)

func insertCompanies(ctx context.Context, exec TableWriter, appCfg config.Config, bsCfg Config, tokenJoyID, companyID uuid.UUID) error {
	// TokenJoy internal company (always testing type).
	if _, err := exec.Exec(ctx, `
		INSERT INTO companies (id, name, type, status)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO NOTHING
	`, tokenJoyID, "TokenJoy", store.CompanyTypeTesting, store.CompanyStatusActive); err != nil {
		return fmt.Errorf("insert tokenjoy company: %w", err)
	}

	// Local/selfhosted company.
	name := resolveCompanyName(appCfg, bsCfg)
	companyType := store.CompanyTypeSelfhosted
	if appCfg.SupportSaas {
		companyType = store.CompanyTypeDemo
	}
	if _, err := exec.Exec(ctx, `
		INSERT INTO companies (id, name, type, status)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO NOTHING
	`, companyID, name, companyType, store.CompanyStatusActive); err != nil {
		return fmt.Errorf("insert company: %w", err)
	}
	return nil
}

// resolveCompanyName prefers appCfg.ResolvedCompanyName (env var), falls back to bootstrap yaml.
func resolveCompanyName(appCfg config.Config, bsCfg Config) string {
	if n := appCfg.ResolvedCompanyName(); strings.TrimSpace(n) != "" {
		return n
	}
	return bsCfg.Company.Name
}
