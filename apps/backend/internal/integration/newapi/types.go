package newapi

// Token is the JSON response from NewAPI token endpoints.
type Token struct {
	ID                 int64  `json:"id"`
	UserID             int64  `json:"user_id"`
	Name               string `json:"name"`
	Key                string `json:"key"`
	Status             int    `json:"status"`
	RemainQuota        int64  `json:"remain_quota"`
	UsedQuota          int64  `json:"used_quota"`
	UnlimitedQuota     bool   `json:"unlimited_quota"`
	ModelLimitsEnabled bool   `json:"model_limits_enabled"`
	ModelLimits        string `json:"model_limits"`
	Group              string `json:"group"`
	ExpiredTime        int64  `json:"expired_time"`
}

// Channel is the JSON response from NewAPI channel endpoints.
type Channel struct {
	ID     int    `json:"id"`
	Type   int    `json:"type"`
	Name   string `json:"name"`
	Key    string `json:"key"`
	Status int    `json:"status"`
	Group  string `json:"group"`
}
