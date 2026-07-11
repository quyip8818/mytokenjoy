package types

type CostPeriod string

const (
	CostPeriodCurrentMonth CostPeriod = "current_month"
	CostPeriodLastMonth    CostPeriod = "last_month"
	CostPeriodLast7Days    CostPeriod = "last_7_days"
	CostPeriodCustom       CostPeriod = "custom"
)

type CostSummary struct {
	TotalCost            float64 `json:"totalCost"`
	TotalCostMom         float64 `json:"totalCostMom"`
	TotalTokens          float64 `json:"totalTokens"`
	TotalRequests        float64 `json:"totalRequests"`
	TotalRequestsMom     float64 `json:"totalRequestsMom"`
	AvgCostPerRequest    float64 `json:"avgCostPerRequest"`
	AvgCostPerRequestMom float64 `json:"avgCostPerRequestMom"`
	AvgCostPerMember     float64 `json:"avgCostPerMember"`
	AvgCostPerMemberMom  float64 `json:"avgCostPerMemberMom"`
}

type DepartmentCost struct {
	DepartmentID   string  `json:"departmentId"`
	DepartmentName string  `json:"departmentName"`
	Cost           float64 `json:"cost"`
	Percentage     float64 `json:"percentage"`
	HasChildren    bool    `json:"hasChildren,omitempty"`
}

type DepartmentCostMember struct {
	MemberID   string  `json:"memberId"`
	MemberName string  `json:"memberName"`
	Cost       float64 `json:"cost"`
	Requests   float64 `json:"requests"`
	Tokens     float64 `json:"tokens"`
}

type DailyCost struct {
	Date     string  `json:"date"`
	Cost     float64 `json:"cost"`
	Tokens   float64 `json:"tokens"`
	Requests float64 `json:"requests"`
}

type TopConsumer struct {
	MemberID   string  `json:"memberId"`
	MemberName string  `json:"memberName"`
	Department string  `json:"department"`
	Cost       float64 `json:"cost"`
	Tokens     float64 `json:"tokens"`
	Requests   float64 `json:"requests"`
}

type ModelUsage struct {
	CallType   string  `json:"callType"`
	ModelID    *int64  `json:"modelId,omitempty"`
	ModelName  string  `json:"modelName"`
	Provider   string  `json:"provider"`
	Requests   float64 `json:"requests"`
	Tokens     float64 `json:"tokens"`
	Cost       float64 `json:"cost"`
	Percentage float64 `json:"percentage"`
}

type TeamUsage struct {
	DepartmentID   string  `json:"departmentId"`
	DepartmentName string  `json:"departmentName"`
	Budget         float64 `json:"budget"`
	Consumed       float64 `json:"consumed"`
	MemberCount    float64 `json:"memberCount"`
	TopModel       string  `json:"topModel"`
}

type CostQueryParams struct {
	Period      string `json:"period"`
	StartDate   string `json:"startDate"`
	EndDate     string `json:"endDate"`
	Granularity string `json:"granularity"`
}
