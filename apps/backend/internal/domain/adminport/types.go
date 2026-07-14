package adminport

type CreateTokenInput struct {
	UserID             int64
	Name               string
	RemainQuota        int64
	UnlimitedQuota     bool
	ModelLimitsEnabled bool
	ModelLimits        string
	Group              string
	ExpiredTime        int64
}

type UpdateTokenInput struct {
	ID                 int64
	Name               string
	Status             *int
	RemainQuota        *int64
	UnlimitedQuota     *bool
	ModelLimitsEnabled *bool
	ModelLimits        string
	Group              string
}

type TokenResult struct {
	ID          int64
	Key         string
	RemainQuota int64
	Group       string
}

type UpsertChannelInput struct {
	ID     int
	Type   int
	Name   string
	Key    string
	Status int
	Group  string
}

type ChannelResult struct {
	ID int
}

type CreateUserInput struct {
	Username    string
	DisplayName string
	Password    string
	Quota       int64
}

type UserResult struct {
	ID int64
}

type TopUpInput struct {
	UserID int64
	Quota  int64
}
