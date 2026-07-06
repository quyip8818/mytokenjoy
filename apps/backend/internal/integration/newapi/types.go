package newapi

type Token struct {
	ID                 int64  `json:"id"`
	Name               string `json:"name"`
	Key                string `json:"key"`
	Status             int    `json:"status"`
	RemainQuota        int64  `json:"remain_quota"`
	UsedQuota          int64  `json:"used_quota"`
	UnlimitedQuota     bool   `json:"unlimited_quota"`
	ModelLimitsEnabled bool   `json:"model_limits_enabled"`
	ModelLimits        string `json:"model_limits"`
	Group              string `json:"group"`
}

type CreateTokenRequest struct {
	UserID             int64  `json:"user_id,omitempty"`
	Name               string `json:"name"`
	RemainQuota        int64  `json:"remain_quota"`
	UnlimitedQuota     bool   `json:"unlimited_quota"`
	ModelLimitsEnabled bool   `json:"model_limits_enabled"`
	ModelLimits        string `json:"model_limits"`
	Group              string `json:"group"`
	ExpiredTime        int64  `json:"expired_time"`
}

type UpdateTokenRequest struct {
	ID                 int64  `json:"id"`
	Name               string `json:"name,omitempty"`
	Status             *int   `json:"status,omitempty"`
	RemainQuota        *int64 `json:"remain_quota,omitempty"`
	UnlimitedQuota     *bool  `json:"unlimited_quota,omitempty"`
	ModelLimitsEnabled *bool  `json:"model_limits_enabled,omitempty"`
	ModelLimits        string `json:"model_limits,omitempty"`
	Group              string `json:"group,omitempty"`
}

type Channel struct {
	ID   int    `json:"id"`
	Type int    `json:"type"`
	Name string `json:"name"`
	Key  string `json:"key"`
}

type UpsertChannelRequest struct {
	ID     int    `json:"id,omitempty"`
	Type   int    `json:"type"`
	Name   string `json:"name"`
	Key    string `json:"key"`
	Status int    `json:"status"`
}

const (
	TokenStatusEnabled  = 1
	TokenStatusDisabled = 2
)
