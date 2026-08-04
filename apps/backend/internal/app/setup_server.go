package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/grants"
	"github.com/tokenjoy/backend/seed/bootstrap"
	"github.com/tokenjoy/backend/seed/contract"
	"golang.org/x/crypto/bcrypt"
)

// SetupResult is returned by RunSetupServer after the user completes initialization.
type SetupResult struct {
	CompanyID uuid.UUID
}

// RunSetupServer starts a temporary HTTP server that serves the setup UI and API.
// It blocks until the user completes setup, then returns the new company ID.
// The server serves the frontend static files and exposes /api/setup/* endpoints.
func RunSetupServer(ctx context.Context, pool *pgxpool.Pool, cfg config.Config, logger *slog.Logger) (SetupResult, error) {
	logger.Info("=== SETUP MODE ===")
	logger.Info("System not initialized. Please open the browser and complete setup.")

	resultCh := make(chan SetupResult, 1)
	errCh := make(chan error, 1)

	r := chi.NewRouter()
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)

	// Setup API endpoints
	r.Get("/api/setup/status", handleSetupStatus())
	r.Post("/api/setup/init", handleSetupInit(pool, cfg, resultCh, errCh, logger))

	// Serve frontend static files (the setup page is built into the frontend bundle)
	if cfg.StaticDir != "" {
		fileServer(r, cfg.StaticDir)
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Start server in background
	go func() {
		logger.Info("setup server starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("setup server: %w", err)
		}
	}()

	// Wait for setup completion or context cancellation
	select {
	case result := <-resultCh:
		// Graceful shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
		logger.Info("setup complete, server shutting down", "companyId", result.CompanyID)
		return result, nil
	case err := <-errCh:
		_ = srv.Close()
		return SetupResult{}, err
	case <-ctx.Done():
		_ = srv.Close()
		return SetupResult{}, ctx.Err()
	}
}

// fileServer serves static files with SPA fallback (index.html for non-file paths).
func fileServer(r chi.Router, dir string) {
	fs := http.Dir(dir)
	fileHandler := http.FileServer(fs)
	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		// Try to serve the file; if not found, serve index.html (SPA routing)
		path := r.URL.Path
		if f, err := fs.Open(path); err == nil {
			f.Close()
			fileHandler.ServeHTTP(w, r)
		} else {
			// SPA fallback
			r.URL.Path = "/"
			fileHandler.ServeHTTP(w, r)
		}
	})
}

func handleSetupStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"ready": true,
		})
	}
}

type setupInitRequest struct {
	CompanyName   string `json:"companyName"`
	Industry      string `json:"industry"`
	Size          string `json:"size"`
	AdminEmail    string `json:"adminEmail"`
	AdminPassword string `json:"adminPassword"`
	AdminName     string `json:"adminName"`
}

func handleSetupInit(
	pool *pgxpool.Pool,
	cfg config.Config,
	resultCh chan<- SetupResult,
	errCh chan<- error,
	logger *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req setupInitRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}

		// Validate required fields
		if strings.TrimSpace(req.CompanyName) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "companyName is required"})
			return
		}
		if strings.TrimSpace(req.AdminEmail) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "adminEmail is required"})
			return
		}
		if len(req.AdminPassword) < 8 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "adminPassword must be at least 8 characters"})
			return
		}

		ctx := r.Context()

		// 1. Generate idempotency key and write to system_settings
		idempotencyKey := uuid.Must(uuid.NewV7()).String()
		if err := setSystemSetting(ctx, pool, "setup_idempotency_key", idempotencyKey); err != nil {
			logger.Error("write idempotency key", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		}

		// 2. Register company: call SaaS (creates company + user/member + wallet + token)
		reg, err := registerCompany(ctx, cfg, req, idempotencyKey, logger)
		if err != nil {
			logger.Error("register company", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to register company: " + err.Error()})
			return
		}
		companyID := reg.CompanyID

		// 2b. Sync currencies from SaaS (companies.billing_currency FK requires it).
		if err := syncCurrenciesFromSaaS(ctx, pool, cfg); err != nil {
			logger.Error("sync currencies from SaaS", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		}

		// 2c. Apply bootstrap data (permissions, roles, company, org root, etc.).
		bootstrapCfg := cfg
		bootstrapCfg.CompanyID = companyID
		bootstrapCfg.CompanyName = strings.TrimSpace(req.CompanyName)
		bsCfg, err := bootstrap.LoadConfig(os.Getenv("BOOTSTRAP_CONFIG_PATH"))
		if err != nil {
			logger.Error("load bootstrap config", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		}
		if err := bootstrap.ApplyBootstrap(ctx, pool, bootstrapCfg, bsCfg); err != nil {
			logger.Error("apply bootstrap", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		}

		// 3. Persist setup state in a single transaction.
		// Admin user/member already created on SaaS; Local only stores settings + local admin user for login.
		tx, err := pool.Begin(ctx)
		if err != nil {
			logger.Error("begin setup tx", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		}
		defer tx.Rollback(ctx)

		if err := setSystemSettingTx(ctx, tx, "setup_company_id", companyID.String()); err != nil {
			logger.Error("write setup_company_id", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		}
		if err := setSystemSettingTx(ctx, tx, "setup_company_name", strings.TrimSpace(req.CompanyName)); err != nil {
			logger.Error("write setup_company_name", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		}
		if err := setSystemSettingTx(ctx, tx, "setup_admin_email", strings.TrimSpace(req.AdminEmail)); err != nil {
			logger.Error("write setup_admin_email", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		}
		if reg.SyncToken != "" {
			if err := setSystemSettingTx(ctx, tx, "catalog_sync_token", reg.SyncToken); err != nil {
				logger.Error("write catalog_sync_token", "error", err)
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
				return
			}
		}
		if reg.PlatformKey != "" {
			if err := setSystemSettingTx(ctx, tx, "saas_platform_key", reg.PlatformKey); err != nil {
				logger.Error("write saas_platform_key", "error", err)
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
				return
			}
		}
		if reg.WalletUserID > 0 {
			if err := setSystemSettingTx(ctx, tx, "saas_wallet_user_id", fmt.Sprintf("%d", reg.WalletUserID)); err != nil {
				logger.Error("write saas_wallet_user_id", "error", err)
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
				return
			}
		}

		// Create local admin user (for Local login — SaaS has its own copy).
		if err := createAdminUserTx(ctx, tx, req, companyID); err != nil {
			logger.Error("create admin user", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		}

		if err := tx.Commit(ctx); err != nil {
			logger.Error("commit setup tx", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		}

		// 5. Return success
		writeJSON(w, http.StatusOK, map[string]any{
			"companyId": companyID.String(),
			"status":    "ok",
		})

		// Signal completion — the setup server will shut down
		resultCh <- SetupResult{
			CompanyID: companyID,
		}
	}
}

// registerResult holds the full response from SaaS register-local.
type registerResult struct {
	CompanyID    uuid.UUID
	WalletUserID int64
	PlatformKey  string
	SyncToken    string
}

// registerCompany calls SaaS platform to register a selfhosted company.
// SaaS creates company + admin user/member + NewAPI wallet user + unlimited token.
func registerCompany(ctx context.Context, cfg config.Config, req setupInitRequest, idempotencyKey string, logger *slog.Logger) (registerResult, error) {
	if strings.TrimSpace(cfg.SaasPlatformURL) == "" {
		return registerResult{}, fmt.Errorf("SAAS_PLATFORM_URL is required for setup (no offline mode)")
	}

	// Call SaaS: POST /api/platform/register-local
	body := map[string]string{
		"name":           strings.TrimSpace(req.CompanyName),
		"industry":       req.Industry,
		"size":           req.Size,
		"adminEmail":     strings.TrimSpace(req.AdminEmail),
		"adminPassword":  req.AdminPassword,
		"adminName":      strings.TrimSpace(req.AdminName),
		"idempotencyKey": idempotencyKey,
	}
	payload, _ := json.Marshal(body)

	url := strings.TrimRight(cfg.SaasPlatformURL, "/") + "/api/platform/register-local"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return registerResult{}, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Registration-Secret", cfg.SaasRegistrationSecret)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return registerResult{}, fmt.Errorf("cannot reach SaaS platform at %s: %w", cfg.SaasPlatformURL, err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return registerResult{}, fmt.Errorf("SaaS register-local returned %d: %s", resp.StatusCode, string(respBody))
	}

	var parsed struct {
		CompanyID    string `json:"companyId"`
		WalletUserID int64  `json:"walletUserId"`
		PlatformKey  string `json:"platformKey"`
		SyncToken    string `json:"syncToken"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return registerResult{}, fmt.Errorf("parse SaaS response: %w", err)
	}
	id, err := uuid.Parse(parsed.CompanyID)
	if err != nil {
		return registerResult{}, fmt.Errorf("parse company ID from SaaS: %w", err)
	}
	return registerResult{
		CompanyID:    id,
		WalletUserID: parsed.WalletUserID,
		PlatformKey:  parsed.PlatformKey,
		SyncToken:    parsed.SyncToken,
	}, nil
}

// createAdminUserTx creates the full admin identity in local DB:
// user row + member row + super_admin role assignment.
// ponytail: setup 只在空库首次运行，不会有冲突。直接 INSERT。
// Prerequisites: bootstrap must have run first (currencies, companies, roles, org_nodes exist).
func createAdminUserTx(ctx context.Context, tx pgx.Tx, req setupInitRequest, companyID uuid.UUID) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	email := strings.TrimSpace(req.AdminEmail)
	name := strings.TrimSpace(req.AdminName)
	if name == "" {
		name = email
	}
	userID := uuid.Must(uuid.NewV7())
	memberID := uuid.Must(uuid.NewV7())
	rootDeptID := contract.IDDept1
	now := time.Now().UTC()

	// 1. User (auth identity).
	if _, err := tx.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, name, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 'active', $5, $5)
	`, userID, email, string(hash), name, now); err != nil {
		return fmt.Errorf("insert admin user: %w", err)
	}

	// 2. Member (company-scoped identity).
	if _, err := tx.Exec(ctx, `
		INSERT INTO members (id, company_id, user_id, alias, department_id, status, source, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, 'active', 'setup', $6, $6)
	`, memberID, companyID, userID, name, rootDeptID, now); err != nil {
		return fmt.Errorf("insert admin member: %w", err)
	}

	// 3. Super admin role.
	if _, err := tx.Exec(ctx, `
		INSERT INTO member_roles (company_id, member_id, role_id)
		VALUES ($1, $2, $3)
	`, companyID, memberID, grants.IDSuperAdmin); err != nil {
		return fmt.Errorf("insert admin member_role: %w", err)
	}

	return nil
}

// setSystemSettingTx upserts a key-value pair into system_settings within a transaction.
func setSystemSettingTx(ctx context.Context, tx pgx.Tx, key, value string) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO system_settings (key, value) VALUES ($1, $2)
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value
	`, key, value)
	return err
}

// setSystemSetting upserts a key-value pair into system_settings (non-transactional).
func setSystemSetting(ctx context.Context, pool *pgxpool.Pool, key, value string) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO system_settings (key, value) VALUES ($1, $2)
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value
	`, key, value)
	return err
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// syncCurrenciesFromSaaS fetches currencies from SaaS platform and writes them to local DB.
// ponytail: currencies endpoint is public (no auth). Simple GET + INSERT.
func syncCurrenciesFromSaaS(ctx context.Context, pool *pgxpool.Pool, cfg config.Config) error {
	url := strings.TrimRight(cfg.SaasPlatformURL, "/") + "/api/platform/sync/catalog/currencies"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("fetch currencies: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("currencies endpoint returned %d", resp.StatusCode)
	}

	var parsed struct {
		Data []struct {
			Code         string `json:"code"`
			QuotaPerUnit int64  `json:"quotaPerUnit"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return fmt.Errorf("decode currencies: %w", err)
	}

	for _, c := range parsed.Data {
		if _, err := pool.Exec(ctx, `
			INSERT INTO currencies (currency, quota_per_unit, enabled)
			VALUES ($1, $2, TRUE)
			ON CONFLICT (currency) DO UPDATE SET quota_per_unit = EXCLUDED.quota_per_unit, enabled = TRUE
		`, c.Code, c.QuotaPerUnit); err != nil {
			return fmt.Errorf("insert currency %s: %w", c.Code, err)
		}
	}
	return nil
}
