package adminport

// --- Token lifecycle (used by newapisync/platformkey) ---

type CreateTokenInput struct {
	UserID         int64  `json:"user_id,omitempty"`
	Name           string `json:"name"`
	RemainQuota    int64  `json:"remain_quota"`
	UnlimitedQuota bool   `json:"unlimited_quota"`
	Group          string `json:"group"`
	ExpiredTime    int64  `json:"expired_time"`
}

type UpdateTokenInput struct {
	ID             int64  `json:"id"`
	Name           string `json:"name,omitempty"`
	Status         *int   `json:"status,omitempty"`
	RemainQuota    *int64 `json:"remain_quota,omitempty"`
	UnlimitedQuota *bool  `json:"unlimited_quota,omitempty"`
	Group          string `json:"group,omitempty"`
	// ExpiredTime is optional; when nil, UpdateToken preserves the existing value
	// (NewAPI PUT replaces the whole token and zero means already-expired).
	ExpiredTime *int64 `json:"-"`
}

type TokenResult struct {
	ID          int64
	UserID      int64
	Key         string
	RemainQuota int64
	Group       string
}

// --- Channel lifecycle (used by newapisync/provider) ---

type UpsertChannelInput struct {
	ID     int    `json:"id,omitempty"`
	Type   int    `json:"type"`
	Name   string `json:"name"`
	Key    string `json:"key"`
	Status int    `json:"status"`
	Group  string `json:"group,omitempty"`
}

type ChannelResult struct {
	ID int
}

// --- User provisioning (used by company creation, bootstrap) ---

type CreateUserInput struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Password    string `json:"password"`
}

type UserResult struct {
	ID int64
}

// --- Pricing (used by models domain) ---

type ModelPricing struct {
	ModelName       string
	ModelRatio      float64
	CompletionRatio float64
}
