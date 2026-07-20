package newapi

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"

	"github.com/tokenjoy/backend/internal/domain/adminport"
)

const (
	manageActionAddQuota = "add_quota"
	manageModeAdd        = "add"
	manageModeSubtract   = "subtract"
)

// User is the JSON response from NewAPI user endpoints.
type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Quota    int64  `json:"quota"`
}

type manageUserRequest struct {
	ID     int64  `json:"id"`
	Action string `json:"action"`
	Value  int64  `json:"value"`
	Mode   string `json:"mode"`
}

func (c *Client) CreateUser(ctx context.Context, req adminport.CreateUserInput) (adminport.UserResult, error) {
	var user User
	createErr := c.do(ctx, "POST", "/api/user/", req, &user)
	if createErr == nil && user.ID > 0 {
		return adminport.UserResult{ID: user.ID}, nil
	}
	if createErr != nil && !isDuplicateUsernameError(createErr) {
		return adminport.UserResult{}, createErr
	}
	// Upstream often returns success with empty data; duplicates need lookup too.
	found, findErr := c.findUserByUsername(ctx, req.Username)
	if findErr != nil {
		if createErr != nil {
			return adminport.UserResult{}, fmt.Errorf("%w (lookup after duplicate: %v)", createErr, findErr)
		}
		return adminport.UserResult{}, fmt.Errorf("create user succeeded but id missing: %w", findErr)
	}
	if found.ID <= 0 {
		return adminport.UserResult{}, fmt.Errorf("create user succeeded but id missing")
	}
	return adminport.UserResult{ID: found.ID}, nil
}

func (c *Client) GetUserQuota(ctx context.Context, userID int64) (int64, error) {
	var user User
	path := "/api/user/" + strconv.FormatInt(userID, 10)
	if err := c.do(ctx, "GET", path, nil, &user); err != nil {
		return 0, err
	}
	return user.Quota, nil
}

func (c *Client) TopUp(ctx context.Context, req adminport.TopUpInput) error {
	if req.CompanyID <= 0 {
		return fmt.Errorf("user id required")
	}
	if req.Quota == 0 {
		return nil
	}
	mode := manageModeAdd
	value := req.Quota
	if value < 0 {
		if value == math.MinInt64 {
			return fmt.Errorf("topup quota delta out of range")
		}
		mode = manageModeSubtract
		value = -value
	}
	// NewAPI removed POST /api/topup; admin quota changes go through ManageUser.
	return c.do(ctx, "POST", "/api/user/manage", manageUserRequest{
		ID:     req.CompanyID,
		Action: manageActionAddQuota,
		Mode:   mode,
		Value:  value,
	}, nil)
}

func (c *Client) findUserByUsername(ctx context.Context, username string) (User, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return User{}, fmt.Errorf("username required")
	}
	var page listPage[User]
	path := "/api/user/search?keyword=" + url.QueryEscape(username)
	if err := c.do(ctx, "GET", path, nil, &page); err != nil {
		return User{}, err
	}
	for _, item := range page.Items {
		if item.Username == username && item.ID > 0 {
			return item, nil
		}
	}
	return User{}, fmt.Errorf("user %q not found", username)
}

func isDuplicateUsernameError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "users_username_key") ||
		(strings.Contains(msg, "duplicate") && strings.Contains(msg, "username"))
}
