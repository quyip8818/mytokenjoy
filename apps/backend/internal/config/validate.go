package config

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
)

func (c Config) validate() error {
	if err := c.validateCore(); err != nil {
		return err
	}
	if err := c.validateDeploy(); err != nil {
		return err
	}
	if err := c.validateNewAPI(); err != nil {
		return err
	}
	return nil
}

func (c Config) validateCore() error {
	if strings.TrimSpace(c.DatabaseURL) == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.TokenJoyCompanyID <= 0 || c.LocalCompanyID <= 0 {
		return fmt.Errorf("TOKENJOY_COMPANY_ID and LOCAL_COMPANY_ID must be positive")
	}
	if c.TokenJoyCompanyID == c.LocalCompanyID {
		return fmt.Errorf("TOKENJOY_COMPANY_ID and LOCAL_COMPANY_ID must differ")
	}
	if c.TokenJoyCompanyID >= 1000000 || c.LocalCompanyID >= 1000000 {
		return fmt.Errorf("TOKENJOY_COMPANY_ID and LOCAL_COMPANY_ID must be < 1000000")
	}
	if !c.SupportSaas && strings.TrimSpace(c.CompanyName) == "" {
		return fmt.Errorf("COMPANY_NAME is required when SUPPORT_SAAS=false")
	}
	if strings.TrimSpace(c.SessionSecret) == "" {
		return fmt.Errorf("SESSION_SECRET is required")
	}
	if err := validateDataSourceCredentialKey(c.DataSourceCredentialKey); err != nil {
		return err
	}
	if c.SupportSaas && strings.TrimSpace(c.PlatformSessionSecret) == "" {
		return fmt.Errorf("PLATFORM_SESSION_SECRET is required when SUPPORT_SAAS=true")
	}
	if c.LogSchemaIsolated && !c.IngestEnabled() {
		return fmt.Errorf("log schema isolation requires LOG_DATABASE_URL")
	}
	return nil
}

func (c Config) validateDeploy() error {
	if err := validateBootstrapMode(c.BootstrapMode); err != nil {
		return err
	}
	if err := validateDeployEnv(c.DeployEnv); err != nil {
		return err
	}
	if err := validateClockAnchor(c.ClockAnchor); err != nil {
		return err
	}
	if c.IsProductionDeploy() {
		return c.validateProductionContract()
	}
	return nil
}

func (c Config) validateNewAPI() error {
	if c.NewAPIGatewayEnabled && !c.NewAPIEnabled {
		return fmt.Errorf("NEW_API_ENABLED must be true when NEWAPI_GATEWAY_ENABLED=true")
	}
	if c.IngestEnabled() && strings.TrimSpace(c.NewAPIWebhookSecret) == "" {
		return fmt.Errorf("NEW_API_WEBHOOK_SECRET is required when LOG_DATABASE_URL is set")
	}
	if !c.NewAPIEnabled {
		return nil
	}
	if strings.TrimSpace(c.NewAPIBaseURL) == "" {
		return fmt.Errorf("NEW_API_BASE_URL is required when NEW_API_ENABLED=true")
	}
	if err := validateNewAPIBaseURL(c.NewAPIBaseURL); err != nil {
		return err
	}
	if strings.TrimSpace(c.NewAPIAdminToken) == "" {
		return fmt.Errorf("NEW_API_ADMIN_TOKEN is required when NEW_API_ENABLED=true")
	}
	return nil
}

func validateDataSourceCredentialKey(raw string) error {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return fmt.Errorf("DATA_SOURCE_CREDENTIAL_KEY is required")
	}
	if decoded, err := base64.StdEncoding.DecodeString(trimmed); err == nil && len(decoded) == 32 {
		return nil
	}
	if decoded, err := hex.DecodeString(trimmed); err == nil && len(decoded) == 32 {
		return nil
	}
	return fmt.Errorf("DATA_SOURCE_CREDENTIAL_KEY is invalid")
}

func validateNewAPIBaseURL(raw string) error {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return fmt.Errorf("NEW_API_BASE_URL is invalid: %w", err)
	}
	if parsed.Path != "" && parsed.Path != "/" {
		return fmt.Errorf("NEW_API_BASE_URL must not include a path")
	}
	return nil
}
