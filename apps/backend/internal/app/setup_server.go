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
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tokenjoy/backend/internal/config"
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

		// 2. Register company: call SaaS or generate locally
		companyID, err := registerCompany(ctx, cfg, req, idempotencyKey, logger)
		if err != nil {
			logger.Error("register company", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to register company: " + err.Error()})
			return
		}

		// 3. Persist setup state + admin user in a single transaction.
		// If any step fails the whole thing rolls back — no half-initialized state.
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

		// 4. Create admin user (users table only — member/roles created by bootstrap)
		if err := createAdminUserTx(ctx, tx, req); err != nil {
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

// registerCompany calls SaaS platform or generates UUID locally depending on config.
func registerCompany(ctx context.Context, cfg config.Config, req setupInitRequest, idempotencyKey string, logger *slog.Logger) (uuid.UUID, error) {
	if cfg.SetupOfflineMode || strings.TrimSpace(cfg.SaasPlatformURL) == "" {
		// Offline mode: generate UUID v7 locally
		id := uuid.Must(uuid.NewV7())
		logger.Info("offline mode: generated local company ID", "companyId", id)
		return id, nil
	}

	// Call SaaS: POST /api/platform/register-local
	body := map[string]string{
		"name":           strings.TrimSpace(req.CompanyName),
		"industry":       req.Industry,
		"size":           req.Size,
		"idempotencyKey": idempotencyKey,
	}
	payload, _ := json.Marshal(body)

	url := strings.TrimRight(cfg.SaasPlatformURL, "/") + "/api/platform/register-local"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return uuid.Nil, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Registration-Secret", cfg.SaasRegistrationSecret)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return uuid.Nil, fmt.Errorf("call SaaS register-local: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return uuid.Nil, fmt.Errorf("SaaS register-local returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		CompanyID string `json:"companyId"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return uuid.Nil, fmt.Errorf("parse SaaS response: %w", err)
	}
	id, err := uuid.Parse(result.CompanyID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("parse company ID from SaaS: %w", err)
	}
	return id, nil
}

// createAdminUserTx inserts a user row with bcrypt-hashed password within a transaction.
// Only writes the users table — member/role creation happens during bootstrap.
// ponytail: setup 只在空库首次运行，不会有 email 冲突。直接 INSERT。
func createAdminUserTx(ctx context.Context, tx pgx.Tx, req setupInitRequest) error {
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
	now := time.Now().UTC()

	_, err = tx.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, name, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 'active', $5, $5)
	`, userID, email, string(hash), name, now)
	if err != nil {
		return fmt.Errorf("insert admin user: %w", err)
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
