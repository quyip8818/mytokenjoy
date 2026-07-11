//go:build testhook

package saas

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/tokenjoy/backend/internal/app"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

const (
	PlatformBootstrapEmail    = "ops@tokenjoy.test"
	PlatformBootstrapPassword = "platform-pass-123"
)

func ApplyConfig(cfg *config.Config) {
	cfg.SupportSaas = true
	cfg.SimulateDelay = false
	cfg.PlatformBootstrapEmail = PlatformBootstrapEmail
	cfg.PlatformBootstrapPassword = PlatformBootstrapPassword
}

func Config(opts ...testutil.ConfigOption) config.Config {
	all := append([]testutil.ConfigOption{
		testutil.WithSupportSaas(true),
		testutil.WithPlatformBootstrap(PlatformBootstrapEmail, PlatformBootstrapPassword),
	}, opts...)
	return testutil.TestConfig(all...)
}

type NewAPIMock struct {
	Server     *httptest.Server
	mu         sync.Mutex
	quotas     map[int64]int64
	nextUserID int64
}

func StartNewAPIMock(t *testing.T) *NewAPIMock {
	t.Helper()
	m := &NewAPIMock{quotas: make(map[int64]int64), nextUserID: 200}
	m.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/user/":
			m.mu.Lock()
			m.nextUserID++
			userID := m.nextUserID
			m.quotas[userID] = 0
			m.mu.Unlock()
			_ = json.NewEncoder(w).Encode(map[string]any{
				"success": true,
				"data":    map[string]any{"id": userID, "username": "wallet", "quota": int64(0)},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/api/topup":
			var body struct {
				UserID int64 `json:"user_id"`
				Quota  int64 `json:"quota"`
			}
			_ = json.NewDecoder(r.Body).Decode(&body)
			m.mu.Lock()
			m.quotas[body.UserID] += body.Quota
			m.mu.Unlock()
			_ = json.NewEncoder(w).Encode(map[string]any{"success": true})
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/api/user/"):
			idStr := strings.TrimPrefix(r.URL.Path, "/api/user/")
			var userID int64
			_, _ = fmt.Sscanf(idStr, "%d", &userID)
			m.mu.Lock()
			quota := m.quotas[userID]
			m.mu.Unlock()
			_ = json.NewEncoder(w).Encode(map[string]any{
				"success": true,
				"data":    map[string]any{"id": userID, "quota": quota},
			})
		default:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{}"))
		}
	}))
	t.Cleanup(m.Server.Close)
	return m
}

func (m *NewAPIMock) SetQuota(userID, quota int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.quotas[userID] = quota
}

func (m *NewAPIMock) ApplyToConfig(cfg *config.Config) {
	cfg.NewAPIEnabled = true
	cfg.NewAPIBaseURL = m.Server.URL
	cfg.NewAPIAdminToken = "test-token"
}

type CreateCompanyHTTPResult struct {
	Company    store.Company
	InviteCode string
}

func NewRouter(t *testing.T, mock *NewAPIMock) http.Handler {
	t.Helper()
	appOpts := []app.Option{app.WithoutWorker()}
	if mock == nil {
		appOpts = append(appOpts, app.WithAdminClient(testutil.DefaultStubAdminClient()))
	}
	application := testutil.NewTestAppWithOptions(t, func(cfg *config.Config) {
		ApplyConfig(cfg)
		if mock != nil {
			mock.ApplyToConfig(cfg)
		}
	}, appOpts...)
	return application.Router
}

func LoginPlatform(t *testing.T, router http.Handler) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{
		"email": PlatformBootstrapEmail, "password": PlatformBootstrapPassword,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/platform/auth/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("platform login: expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	for _, c := range rec.Result().Cookies() {
		if c.Name == "tokenjoy_platform_session" && c.Value != "" {
			return "tokenjoy_platform_session=" + c.Value
		}
	}
	t.Fatal("platform session cookie not set")
	return ""
}

func CreateCompanyHTTP(t *testing.T, router http.Handler, platformCookie, slug, name, superAdminEmail string) CreateCompanyHTTPResult {
	t.Helper()
	body, _ := json.Marshal(map[string]string{
		"slug": slug, "name": name, "superAdminEmail": superAdminEmail,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/platform/companies", bytes.NewReader(body))
	req.Header.Set("Cookie", platformCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create company: expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}
	var created CreateCompanyHTTPResult
	if err := json.NewDecoder(rec.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}
	if created.InviteCode == "" {
		t.Fatal("expected invite token")
	}
	return created
}

func AcceptInviteHTTP(t *testing.T, router http.Handler, inviteToken, name, password string) (types.Member, string) {
	t.Helper()
	body, _ := json.Marshal(map[string]string{
		"inviteCode": inviteToken, "name": name, "password": password,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/accept-invite", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("accept invite: expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var member types.Member
	if err := json.NewDecoder(rec.Body).Decode(&member); err != nil {
		t.Fatal(err)
	}
	var sessionCookie string
	for _, c := range rec.Result().Cookies() {
		if c.Name == "tokenjoy_session_member" && c.Value != "" {
			sessionCookie = "tokenjoy_session_member=" + c.Value
			break
		}
	}
	if sessionCookie == "" {
		t.Fatal("member session cookie not set")
	}
	return member, sessionCookie
}

type ProvisionedCompany struct {
	Company      store.Company
	InviteCode   string
	Member       types.Member
	MemberCookie string
}

func ProvisionCompanyHTTP(t *testing.T, router http.Handler, platformCookie, slug, name, adminEmail, adminName, password string) ProvisionedCompany {
	t.Helper()
	created := CreateCompanyHTTP(t, router, platformCookie, slug, name, adminEmail)
	member, cookie := AcceptInviteHTTP(t, router, created.InviteCode, adminName, password)
	return ProvisionedCompany{
		Company: created.Company, InviteCode: created.InviteCode,
		Member: member, MemberCookie: cookie,
	}
}

func PlatformRechargeHTTP(t *testing.T, router http.Handler, platformCookie string, companyID int64, amount float64) {
	t.Helper()
	body, _ := json.Marshal(map[string]float64{"amount": amount})
	url := fmt.Sprintf("/api/platform/companies/%d/recharge", companyID)
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	req.Header.Set("Cookie", platformCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK && rec.Code != http.StatusNoContent {
		t.Fatalf("platform recharge: expected success, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func UpdateCompanyStatusHTTP(t *testing.T, router http.Handler, platformCookie string, companyID int64, status string) {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"status": status})
	url := fmt.Sprintf("/api/platform/companies/%d", companyID)
	req := httptest.NewRequest(http.MethodPatch, url, bytes.NewReader(body))
	req.Header.Set("Cookie", platformCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK && rec.Code != http.StatusNoContent {
		t.Fatalf("update company: expected success, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func UpdateBudgetNodeHTTP(t *testing.T, router http.Handler, memberCookie, nodeID string, budget float64) {
	t.Helper()
	body, _ := json.Marshal(map[string]float64{"budget": budget})
	url := fmt.Sprintf("/api/budget/departments/%s", nodeID)
	req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", memberCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("update budget node: expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func DefaultSeedMemberCookie(t *testing.T) string {
	t.Helper()
	return testutil.SessionCookie(t, contract.IDMemberAdmin)
}
