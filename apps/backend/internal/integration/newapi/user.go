package newapi

import (
	"context"
	"strconv"
)

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Quota    int64  `json:"quota"`
}

type CreateUserRequest struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Password    string `json:"password"`
	Quota       int64  `json:"quota"`
}

type TopUpRequest struct {
	UserID int64  `json:"user_id"`
	Quota  int64  `json:"quota"`
	Remark string `json:"remark"`
}

func (c *Client) CreateUser(ctx context.Context, req CreateUserRequest) (User, error) {
	var user User
	if err := c.do(ctx, "POST", "/api/user/", req, &user); err != nil {
		return User{}, err
	}
	return user, nil
}

func (c *Client) GetUserQuota(ctx context.Context, userID int64) (int64, error) {
	var user User
	path := "/api/user/" + strconv.FormatInt(userID, 10)
	if err := c.do(ctx, "GET", path, nil, &user); err != nil {
		return 0, err
	}
	return user.Quota, nil
}

func (c *Client) TopUp(ctx context.Context, req TopUpRequest) error {
	return c.do(ctx, "POST", "/api/topup", req, nil)
}
