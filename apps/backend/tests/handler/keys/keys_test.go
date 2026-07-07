package keys_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tokenjoy/backend/internal/store/seed"
	testhttp "github.com/tokenjoy/backend/tests/testutil/http"
)

func TestPlatformListEnrichedHTTP(t *testing.T) {
	router := testhttp.NewRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/api/keys/platform", nil)
	req.Header.Set("Cookie", testhttp.AdminCookie(t))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Items []struct {
			Type           string  `json:"type"`
			MemberName     *string `json:"memberName"`
			DepartmentID   string  `json:"departmentId"`
			DepartmentName string  `json:"departmentName"`
		} `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Items) == 0 {
		t.Fatal("expected platform keys in seed")
	}
	foundMember := false
	for _, item := range payload.Items {
		if item.Type == "member" && item.DepartmentID != "" {
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
}

func TestPlatformListTypeFilterHTTP(t *testing.T) {
	router := testhttp.NewRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/api/keys/platform?type=project", nil)
	req.Header.Set("Cookie", testhttp.AdminCookie(t))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Items []struct {
			Type string `json:"type"`
		} `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	for _, item := range payload.Items {
		if item.Type != "project" {
			t.Fatalf("expected only project keys, got %+v", item)
		}
	}
}

func TestApprovalsListApprovedTabHTTP(t *testing.T) {
	router := testhttp.NewRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/api/keys/approvals?tab=approved", nil)
	req.Header.Set("Cookie", testhttp.AdminCookie(t))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var approvals []struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &approvals); err != nil {
		t.Fatal(err)
	}
	for _, approval := range approvals {
		if approval.Status != "approved" {
			t.Fatalf("expected approved only, got %+v", approval)
		}
	}
}

func TestPlatformListDepartmentFilterHTTP(t *testing.T) {
	router := testhttp.NewRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/api/keys/platform?departmentId="+seed.IDDept3, nil)
	req.Header.Set("Cookie", testhttp.AdminCookie(t))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Items []struct {
			Type string `json:"type"`
		} `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Items) == 0 {
		t.Fatal("expected keys under dept-3 subtree")
	}
}

func TestApprovalsListPendingMemberIdHTTP(t *testing.T) {
	router := testhttp.NewRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/api/keys/approvals?tab=pending&memberId=m-5", nil)
	req.Header.Set("Cookie", testhttp.AdminCookie(t))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var approvals []struct {
		Status      string `json:"status"`
		ApplicantID string `json:"applicantId"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &approvals); err != nil {
		t.Fatal(err)
	}
	if len(approvals) != 1 {
		t.Fatalf("expected 1 pending approval for m-5, got %+v", approvals)
	}
	if approvals[0].Status != "pending" || approvals[0].ApplicantID != "m-5" {
		t.Fatalf("unexpected approval: %+v", approvals[0])
	}
}
