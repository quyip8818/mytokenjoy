package types

type Platform string

const (
	PlatformFeishu   Platform = "feishu"
	PlatformDingtalk Platform = "dingtalk"
	PlatformWecom    Platform = "wecom"
)

type DataSourceStatus struct {
	Platform         *Platform     `json:"platform"`
	Connected        bool          `json:"connected"`
	LastImport       *string       `json:"lastImport"`
	LastImportResult *ImportResult `json:"lastImportResult"`
}

type ImportResult struct {
	SuccessMembers     int             `json:"successMembers"`
	SuccessDepartments int             `json:"successDepartments"`
	Failures           []ImportFailure `json:"failures"`
}

type ImportFailure struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	EmployeeID string `json:"employeeId"`
	Reason     string `json:"reason"`
}

type SyncConfig struct {
	Enabled                   bool   `json:"enabled"`
	StartTime                 string `json:"startTime"`
	FrequencyHours            int    `json:"frequencyHours"`
	DeleteMemberThreshold     int    `json:"deleteMemberThreshold"`
	DeleteDepartmentThreshold int    `json:"deleteDepartmentThreshold"`
	NotifyPhone               bool   `json:"notifyPhone"`
	NotifyEmail               bool   `json:"notifyEmail"`
	NotifyIm                  bool   `json:"notifyIm"`
}

type SyncLog struct {
	ID     string `json:"id"`
	Time   string `json:"time"`
	Type   string `json:"type"`
	Result string `json:"result"`
	Detail string `json:"detail"`
}

type Department struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	ParentID    *string      `json:"parentId"`
	Children    []Department `json:"children,omitempty"`
	MemberCount int          `json:"memberCount"`
}

type BatchImportRow struct {
	Name           string `json:"name"`
	Phone          string `json:"phone"`
	Email          string `json:"email"`
	DepartmentName string `json:"departmentName"`
}

type MemberBatchImportResult struct {
	Imported int                        `json:"imported"`
	Failures []MemberBatchImportFailure `json:"failures"`
}

type MemberBatchImportFailure struct {
	Row    int    `json:"row"`
	Reason string `json:"reason"`
}

type Permission struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Group string `json:"group"`
}

type DataSourceSearchResult struct {
	Name       string `json:"name"`
	Department string `json:"department"`
	MappingOK  bool   `json:"mappingOk"`
}

type DataSourceTestResult struct {
	Success bool    `json:"success"`
	Message *string `json:"message,omitempty"`
}

type BatchInviteResult struct {
	Sent int `json:"sent"`
}
