package keys_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tokenjoy/backend/seed/contract"
	testhttp "github.com/tokenjoy/backend/tests/testutil/http"
)

type keysCase struct {
	name   string
	path   string
	assert func(t *testing.T, body []byte)
}

func keysCases() []keysCase {
	return []keysCase{
		{
			name: "platform list enriched",
			path: "/api/keys/platform",
			assert: func(t *testing.T, body []byte) {
				t.Helper()
				var payload struct {
					Items []struct {
						Scope          string  `json:"scope"`
						MemberName     *string `json:"memberName"`
						DepartmentID   string  `json:"departmentId"`
						DepartmentName string  `json:"departmentName"`
					} `json:"items"`
				}
				if err := json.Unmarshal(body, &payload); err != nil {
					t.Fatal(err)
				}
				if len(payload.Items) == 0 {
					t.Fatal("expected platform keys in seed")
				}
				foundMember := false
				for _, item := range payload.Items {
					if item.Scope == "member" && item.DepartmentID != "" {
						foundMember = true
						if item.MemberName == nil || *item.MemberName == "" {
							t.Fatalf("expected memberName from join, got %+v", item)
						}
						break
					}
				}
				if !foundMember {
					t.Fatalf("expected enriched member key with department, got %+v", payload.Items)
				}
			},
		},
		{
			name: "platform list scope filter",
			path: "/api/keys/platform?scope=project",
			assert: func(t *testing.T, body []byte) {
				t.Helper()
				var payload struct {
					Items []struct {
						Scope string `json:"scope"`
					} `json:"items"`
				}
				if err := json.Unmarshal(body, &payload); err != nil {
					t.Fatal(err)
				}
				for _, item := range payload.Items {
					if item.Scope != "project" {
						t.Fatalf("expected only project keys, got %+v", item)
					}
				}
			},
		},
		{
			name: "platform list department filter",
			path: "/api/keys/platform?departmentId=" + contract.IDDept3.String(),
			assert: func(t *testing.T, body []byte) {
				t.Helper()
				var payload struct {
					Items []struct {
						Scope string `json:"scope"`
					} `json:"items"`
				}
				if err := json.Unmarshal(body, &payload); err != nil {
					t.Fatal(err)
				}
				if len(payload.Items) == 0 {
					t.Fatal("expected keys under dept-3 subtree")
				}
			},
		},
	}
}

func TestKeysHTTPEndpoints(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	cookie := testhttp.AdminCookie(t)

	for _, tc := range keysCases() {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			req.Header.Set("Cookie", cookie)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
			}
			tc.assert(t, rec.Body.Bytes())
		})
	}
}
