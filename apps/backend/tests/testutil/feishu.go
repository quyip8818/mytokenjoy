package testutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
)

type FeishuMockOpts struct {
	DeptName    string
	DeptNamePtr *string
	Users       []map[string]any
	AuthFails   bool
}

func DefaultFeishuUsers() []map[string]any {
	return []map[string]any{{
		"user_id": contract.IDFeishuExtUser1, "name": "Mock User", "email": "mock@example.com",
		"mobile": "13800000000", "department_ids": []string{contract.IDFeishuExtDept1}, "employee_no": "E001",
	}}
}

func StartFeishuMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	return StartFeishuMockServerWithOpts(t, FeishuMockOpts{
		DeptName: "Mock Dept",
		Users:    DefaultFeishuUsers(),
	})
}

func StartFeishuMockServerWithOpts(t *testing.T, opts FeishuMockOpts) *httptest.Server {
	t.Helper()
	if opts.DeptName == "" {
		opts.DeptName = "Mock Dept"
	}
	if len(opts.Users) == 0 {
		opts.Users = DefaultFeishuUsers()
	}
	resolveDeptName := func() string {
		if opts.DeptNamePtr != nil {
			return *opts.DeptNamePtr
		}
		if opts.DeptName != "" {
			return opts.DeptName
		}
		return "Mock Dept"
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/open-apis/auth/v3/tenant_access_token/internal":
			if opts.AuthFails {
				_ = json.NewEncoder(w).Encode(map[string]any{"code": 10003, "msg": "invalid app secret"})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0, "msg": "ok", "tenant_access_token": "mock-token", "expire": 7200,
			})
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/open-apis/contact/v3/departments/") && strings.HasSuffix(r.URL.Path, "/children"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"data": map[string]any{
					"items": []map[string]any{{
						"department_id": contract.IDFeishuExtDept1, "name": resolveDeptName(),
						"parent_department_id": "0", "leader_user_id": contract.IDFeishuExtUser1,
					}},
					"has_more": false,
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/open-apis/contact/v3/departments":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"data": map[string]any{
					"items": []map[string]any{{
						"department_id": contract.IDFeishuExtDept1, "name": resolveDeptName(),
						"parent_department_id": "0", "leader_user_id": contract.IDFeishuExtUser1,
					}},
					"has_more": false,
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/open-apis/contact/v3/users":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0, "data": map[string]any{"items": opts.Users, "has_more": false},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/open-apis/contact/v3/users/search":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"data": map[string]any{"users": opts.Users},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	return server
}

func StartFeishuAuthErrorServer(t *testing.T) *httptest.Server {
	t.Helper()
	return StartFeishuMockServerWithOpts(t, FeishuMockOpts{AuthFails: true})
}

func StartMutableFeishuServer(t *testing.T, deptName *string, users []map[string]any) *httptest.Server {
	t.Helper()
	if deptName == nil {
		name := "Mock Dept"
		deptName = &name
	}
	if len(users) == 0 {
		users = DefaultFeishuUsers()
	}
	return StartFeishuMockServerWithOpts(t, FeishuMockOpts{DeptNamePtr: deptName, Users: users})
}

func ConnectFeishuDataSource(t *testing.T, cfg *config.Config, st store.Store, baseURL string) {
	t.Helper()
	cfg.FeishuBaseURL = baseURL
	key := common.DevDefaultKey()
	payload, err := json.Marshal(types.FeishuCredential{
		Platform: types.PlatformFeishu, AppID: "cli_test", AppSecret: "secret_test",
	})
	if err != nil {
		t.Fatal(err)
	}
	encrypted, err := common.Encrypt(key, payload)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Org().SaveIntegrationCredential(Ctx(), types.PlatformFeishu, encrypted); err != nil {
		t.Fatal(err)
	}
	platform := types.PlatformFeishu
	integration, err := st.Org().Integration(Ctx())
	if err != nil {
		t.Fatal(err)
	}
	status := integration.ToDataSourceStatus()
	status.Connected = true
	status.Platform = &platform
	integration.ApplyDataSourceStatus(status)
	if err := st.Org().SetIntegration(Ctx(), integration); err != nil {
		t.Fatal(err)
	}
}
