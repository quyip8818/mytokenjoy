package feishu

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
)

const (
	RootDepartmentExternalID = "0"
	defaultPageSize          = 50
)

type Department struct {
	ExternalID       string
	Name             string
	ParentExternalID string
	LeaderUserID     string
}

type Member struct {
	ExternalID           string
	Name                 string
	Email                string
	Mobile               string
	DepartmentExternalID string
	EmployeeNo           string
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	baseURL    string
	credential types.FeishuCredential
	httpClient HTTPClient

	tokenMu     sync.Mutex
	accessToken string
	tokenExpiry time.Time
}

func NewClient(baseURL string, credential types.FeishuCredential, httpClient HTTPClient) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		credential: credential,
		httpClient: httpClient,
	}
}

type apiResponse struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

type tokenResponse struct {
	TenantAccessToken string `json:"tenant_access_token"`
	Expire            int    `json:"expire"`
}

type departmentItem struct {
	DepartmentID       string `json:"department_id"`
	Name               string `json:"name"`
	ParentDepartmentID string `json:"parent_department_id"`
	LeaderUserID       string `json:"leader_user_id"`
}

type departmentsPage struct {
	Items     []departmentItem `json:"items"`
	PageToken string           `json:"page_token"`
	HasMore   bool             `json:"has_more"`
}

type userItem struct {
	UserID        string   `json:"user_id"`
	Name          string   `json:"name"`
	Email         string   `json:"email"`
	Mobile        string   `json:"mobile"`
	DepartmentIDs []string `json:"department_ids"`
	EmployeeNo    string   `json:"employee_no"`
}

type usersPage struct {
	Items     []userItem `json:"items"`
	PageToken string     `json:"page_token"`
	HasMore   bool       `json:"has_more"`
}

type searchUsersPage struct {
	Users []userItem `json:"users"`
}

func (c *Client) TestConnection(ctx context.Context) error {
	_, err := c.ensureToken(ctx)
	return err
}

func (c *Client) SearchMember(ctx context.Context, keyword string) (Member, error) {
	if strings.TrimSpace(keyword) == "" {
		return Member{}, fmt.Errorf("keyword is required")
	}
	if _, err := c.ensureToken(ctx); err != nil {
		return Member{}, err
	}

	var page searchUsersPage
	if err := c.post(ctx, "/open-apis/contact/v3/users/search", map[string]string{
		"query": keyword,
	}, &page); err != nil {
		return Member{}, err
	}
	if len(page.Users) == 0 {
		return Member{}, fmt.Errorf("member not found")
	}
	return mapUser(page.Users[0]), nil
}

func (c *Client) ListDepartments(ctx context.Context) ([]Department, error) {
	if _, err := c.ensureToken(ctx); err != nil {
		return nil, err
	}

	seen := make(map[string]struct{})
	seenResult := make(map[string]struct{})
	queue := []string{RootDepartmentExternalID}
	result := make([]Department, 0)

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if _, ok := seen[current]; ok {
			continue
		}
		seen[current] = struct{}{}

		pageToken := ""
		for {
			query := url.Values{}
			query.Set("page_size", fmt.Sprintf("%d", defaultPageSize))
			query.Set("department_id_type", "department_id")
			if pageToken != "" {
				query.Set("page_token", pageToken)
			}

			path := fmt.Sprintf("/open-apis/contact/v3/departments/%s/children?%s", current, query.Encode())
			var page departmentsPage
			if err := c.get(ctx, path, &page); err != nil {
				return nil, err
			}
			for _, item := range page.Items {
				if item.DepartmentID == "" {
					continue
				}
				if _, exists := seen[item.DepartmentID]; !exists {
					queue = append(queue, item.DepartmentID)
				}
				if _, exists := seenResult[item.DepartmentID]; exists {
					continue
				}
				seenResult[item.DepartmentID] = struct{}{}
				result = append(result, Department{
					ExternalID:       item.DepartmentID,
					Name:             item.Name,
					ParentExternalID: item.ParentDepartmentID,
					LeaderUserID:     item.LeaderUserID,
				})
			}
			if !page.HasMore || page.PageToken == "" {
				break
			}
			pageToken = page.PageToken
		}
	}
	return result, nil
}

func (c *Client) ListMembers(ctx context.Context) ([]Member, []types.ImportFailure, error) {
	departments, err := c.ListDepartments(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Always include root department in the fetch list so members directly
	// under root are not missed (they map to the local root/company dept).
	type deptRef struct {
		ExternalID string
		Name       string
	}
	fetchList := []deptRef{{ExternalID: RootDepartmentExternalID, Name: "root"}}
	for _, dept := range departments {
		fetchList = append(fetchList, deptRef{ExternalID: dept.ExternalID, Name: dept.Name})
	}

	seenUsers := make(map[string]struct{})
	members := make([]Member, 0)
	failures := make([]types.ImportFailure, 0)

	for _, dept := range fetchList {
		pageToken := ""
		for {
			query := url.Values{}
			query.Set("department_id", dept.ExternalID)
			query.Set("department_id_type", "department_id")
			query.Set("page_size", fmt.Sprintf("%d", defaultPageSize))
			if pageToken != "" {
				query.Set("page_token", pageToken)
			}

			var page usersPage
			if err := c.get(ctx, "/open-apis/contact/v3/users?"+query.Encode(), &page); err != nil {
				failures = append(failures, types.ImportFailure{
					ID:     fmt.Sprintf("dept-%s", dept.ExternalID),
					Name:   dept.Name,
					Reason: err.Error(),
				})
				break
			}
			for _, user := range page.Items {
				if user.UserID == "" {
					continue
				}
				if _, ok := seenUsers[user.UserID]; ok {
					continue
				}
				seenUsers[user.UserID] = struct{}{}
				member := mapUser(user)
				if member.DepartmentExternalID == "" {
					member.DepartmentExternalID = dept.ExternalID
				}
				members = append(members, member)
			}
			if !page.HasMore || page.PageToken == "" {
				break
			}
			pageToken = page.PageToken
		}
	}
	return members, failures, nil
}

func mapUser(user userItem) Member {
	deptExternalID := ""
	if len(user.DepartmentIDs) > 0 {
		deptExternalID = user.DepartmentIDs[0]
	}
	return Member{
		ExternalID:           user.UserID,
		Name:                 user.Name,
		Email:                user.Email,
		Mobile:               user.Mobile,
		DepartmentExternalID: deptExternalID,
		EmployeeNo:           user.EmployeeNo,
	}
}

func (c *Client) ensureToken(ctx context.Context) (string, error) {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()
	if c.accessToken != "" && time.Now().Before(c.tokenExpiry) {
		return c.accessToken, nil
	}

	body := map[string]string{
		"app_id":     c.credential.AppID,
		"app_secret": c.credential.AppSecret,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/open-apis/auth/v3/tenant_access_token/internal",
		bytes.NewReader(raw),
	)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	payload, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	if res.StatusCode >= 400 {
		return "", fmt.Errorf("feishu token request failed: status %d", res.StatusCode)
	}

	var tokenResp struct {
		Code              int    `json:"code"`
		Msg               string `json:"msg"`
		TenantAccessToken string `json:"tenant_access_token"`
		Expire            int    `json:"expire"`
	}
	if err := json.Unmarshal(payload, &tokenResp); err != nil {
		return "", err
	}
	if tokenResp.Code != 0 {
		if tokenResp.Msg == "" {
			tokenResp.Msg = "authentication failed"
		}
		return "", fmt.Errorf("%s", tokenResp.Msg)
	}
	if tokenResp.TenantAccessToken == "" {
		return "", fmt.Errorf("empty company access token")
	}
	c.accessToken = tokenResp.TenantAccessToken
	ttl := tokenResp.Expire
	if ttl <= 60 {
		ttl = 7200
	}
	c.tokenExpiry = time.Now().Add(time.Duration(ttl-60) * time.Second)
	return c.accessToken, nil
}

func (c *Client) get(ctx context.Context, path string, out any) error {
	return c.doRequest(ctx, http.MethodGet, path, nil, out)
}

func (c *Client) post(ctx context.Context, path string, body any, out any) error {
	return c.doRequest(ctx, http.MethodPost, path, body, out)
}

func (c *Client) doRequest(ctx context.Context, method, path string, body any, out any) error {
	token, err := c.ensureToken(ctx)
	if err != nil {
		return err
	}
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	payload, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode >= 400 {
		return fmt.Errorf("feishu request failed: status %d", res.StatusCode)
	}
	var wrapped apiResponse
	if err := json.Unmarshal(payload, &wrapped); err != nil {
		return err
	}
	if wrapped.Code != 0 {
		if wrapped.Msg == "" {
			wrapped.Msg = "request failed"
		}
		return fmt.Errorf("%s", wrapped.Msg)
	}
	if out == nil || len(wrapped.Data) == 0 {
		return nil
	}
	return json.Unmarshal(wrapped.Data, out)
}
