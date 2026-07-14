package newapi

import (
	"context"
	"fmt"
)

const (
	maxListPages = 20

	tokenListFirstPage   = 0
	channelListFirstPage = 1
)

type listPage[T any] struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Total    int `json:"total"`
	Items    []T `json:"items"`
}

// findLatestByName walks paginated NewAPI list endpoints and returns the
// matching item with the greatest ID (newest wins when duplicates exist).
func findLatestByName[T any](
	ctx context.Context,
	c *Client,
	name string,
	firstPage int,
	pathForPage func(page int) string,
	getName func(T) string,
	getID func(T) int64,
) (T, error) {
	var zero T
	if name == "" {
		return zero, fmt.Errorf("name required")
	}
	var best T
	var found bool
	for page := firstPage; page < firstPage+maxListPages; page++ {
		var list listPage[T]
		if err := c.do(ctx, "GET", pathForPage(page), nil, &list); err != nil {
			return zero, err
		}
		for _, item := range list.Items {
			if getName(item) != name {
				continue
			}
			if !found || getID(item) > getID(best) {
				best = item
				found = true
			}
		}
		if !listHasMore(list, page, firstPage) {
			break
		}
	}
	if !found {
		return zero, fmt.Errorf("%q not found", name)
	}
	return best, nil
}

func listHasMore[T any](list listPage[T], page, firstPage int) bool {
	if len(list.Items) == 0 {
		return false
	}
	pageSize := list.PageSize
	if pageSize <= 0 {
		pageSize = len(list.Items)
	}
	ordinal := page - firstPage + 1
	if list.Total > 0 && ordinal*pageSize >= list.Total {
		return false
	}
	return len(list.Items) >= pageSize
}
