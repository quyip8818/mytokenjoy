package feishu

import (
	"context"
	"fmt"
	"net/url"
)

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
