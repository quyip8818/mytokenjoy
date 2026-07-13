package feishu

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/tokenjoy/backend/internal/domain/types"
)

type userItem struct {
	UserID            string   `json:"user_id"`
	Name              string   `json:"name"`
	Email             string   `json:"email"`
	Mobile            string   `json:"mobile"`
	OwnerDepartmentID []string `json:"department_ids"`
	EmployeeNo        string   `json:"employee_no"`
}

type usersPage struct {
	Items     []userItem `json:"items"`
	PageToken string     `json:"page_token"`
	HasMore   bool       `json:"has_more"`
}

type searchUsersPage struct {
	Users []userItem `json:"users"`
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
	if len(user.OwnerDepartmentID) > 0 {
		deptExternalID = user.OwnerDepartmentID[0]
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
