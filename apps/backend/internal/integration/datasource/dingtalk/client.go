package dingtalk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type Credential struct {
	AppKey    string `json:"appKey"`
	AppSecret string `json:"appSecret"`
}

type Department struct {
	ExternalID       string
	Name             string
	ParentExternalID string
}

type Member struct {
	ExternalID           string
	Name                 string
	Mobile               string
	Email                string
	DepartmentExternalID string
	EmployeeNo           string
}

type Client struct {
	cred       Credential
	httpClient *http.Client

	mu          sync.Mutex
	accessToken string
	tokenExpiry time.Time
}

func NewClient(cred Credential, httpClient *http.Client) *Client {
	return &Client{cred: cred, httpClient: httpClient}
}

// TestConnection 验证凭证有效性（获取 token）
func (c *Client) TestConnection(ctx context.Context) error {
	_, err := c.getToken(ctx)
	return err
}

// ListDepartments 获取所有部门（根部门 + 递归子部门）
func (c *Client) ListDepartments(ctx context.Context) ([]Department, error) {
	token, err := c.getToken(ctx)
	if err != nil {
		return nil, err
	}

	// 获取根部门信息
	rootName, err := c.getDepartment(ctx, token, 1)
	if err != nil {
		return nil, fmt.Errorf("get root department: %w", err)
	}

	all := []Department{
		{ExternalID: "1", Name: rootName, ParentExternalID: ""},
	}
	if err := c.walkDepartments(ctx, token, 1, "1", &all); err != nil {
		return nil, err
	}
	return all, nil
}

// ListMembers 获取所有部门下的成员
func (c *Client) ListMembers(ctx context.Context) ([]Member, error) {
	token, err := c.getToken(ctx)
	if err != nil {
		return nil, err
	}
	depts, err := c.ListDepartments(ctx)
	if err != nil {
		return nil, err
	}

	var all []Member
	seen := make(map[string]bool)
	for _, dept := range depts {
		members, err := c.listDeptMembers(ctx, token, dept.ExternalID)
		if err != nil {
			continue // best-effort
		}
		for _, m := range members {
			if !seen[m.ExternalID] {
				seen[m.ExternalID] = true
				m.DepartmentExternalID = dept.ExternalID
				all = append(all, m)
			}
		}
	}
	return all, nil
}

func (c *Client) getToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.accessToken != "" && time.Now().Add(60*time.Second).Before(c.tokenExpiry) {
		return c.accessToken, nil
	}

	body, _ := json.Marshal(map[string]string{
		"appKey":    c.cred.AppKey,
		"appSecret": c.cred.AppSecret,
	})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.dingtalk.com/v1.0/oauth2/accessToken", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("dingtalk get token: %w", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	var result struct {
		AccessToken string `json:"accessToken"`
		ExpireIn    int    `json:"expireIn"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("dingtalk parse token: %w", err)
	}
	if result.AccessToken == "" {
		return "", fmt.Errorf("dingtalk auth failed: %s", string(respBody))
	}

	c.accessToken = result.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(result.ExpireIn) * time.Second)
	return c.accessToken, nil
}

func (c *Client) getDepartment(ctx context.Context, token string, deptID int) (string, error) {
	body, _ := json.Marshal(map[string]any{"dept_id": deptID})
	respBody, err := c.oapi(ctx, token, "/topapi/v2/department/get", body)
	if err != nil {
		return "", err
	}
	var resp struct {
		ErrCode int `json:"errcode"`
		Result  struct {
			Name string `json:"name"`
		} `json:"result"`
		ErrMsg string `json:"errmsg"`
	}
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return "", err
	}
	if resp.ErrCode != 0 {
		return "", fmt.Errorf("dingtalk get dept %d: %s", deptID, resp.ErrMsg)
	}
	return resp.Result.Name, nil
}

func (c *Client) walkDepartments(ctx context.Context, token string, deptID int, parentExtID string, out *[]Department) error {
	body, _ := json.Marshal(map[string]any{"dept_id": deptID})
	respBody, err := c.oapi(ctx, token, "/topapi/v2/department/listsub", body)
	if err != nil {
		return err
	}

	var resp struct {
		ErrCode int `json:"errcode"`
		Result  []struct {
			DeptID   int    `json:"dept_id"`
			Name     string `json:"name"`
			ParentID int    `json:"parent_id"`
		} `json:"result"`
	}
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return err
	}

	for _, d := range resp.Result {
		extID := fmt.Sprintf("%d", d.DeptID)
		*out = append(*out, Department{
			ExternalID:       extID,
			Name:             d.Name,
			ParentExternalID: parentExtID,
		})
		_ = c.walkDepartments(ctx, token, d.DeptID, extID, out)
	}
	return nil
}

func (c *Client) listDeptMembers(ctx context.Context, token string, deptExtID string) ([]Member, error) {
	var all []Member
	cursor := 0
	for {
		body, _ := json.Marshal(map[string]any{
			"dept_id": deptExtID,
			"cursor":  cursor,
			"size":    100,
		})
		respBody, err := c.oapi(ctx, token, "/topapi/v2/user/list", body)
		if err != nil {
			return all, err
		}

		var resp struct {
			ErrCode int `json:"errcode"`
			Result  struct {
				HasMore bool `json:"has_more"`
				NextCur int  `json:"next_cursor"`
				List    []struct {
					UserID string `json:"userid"`
					Name   string `json:"name"`
					Mobile string `json:"mobile"`
					Email  string `json:"email"`
					JobNum string `json:"job_number"`
				} `json:"list"`
			} `json:"result"`
		}
		if err := json.Unmarshal(respBody, &resp); err != nil {
			return all, err
		}

		for _, u := range resp.Result.List {
			all = append(all, Member{
				ExternalID: u.UserID,
				Name:       u.Name,
				Mobile:     u.Mobile,
				Email:      u.Email,
				EmployeeNo: u.JobNum,
			})
		}
		if !resp.Result.HasMore {
			break
		}
		cursor = resp.Result.NextCur
	}
	return all, nil
}

func (c *Client) oapi(ctx context.Context, token, path string, body []byte) ([]byte, error) {
	url := "https://oapi.dingtalk.com" + path + "?access_token=" + token
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("dingtalk oapi %s: %w", path, err)
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
