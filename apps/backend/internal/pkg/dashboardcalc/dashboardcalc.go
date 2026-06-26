package dashboardcalc

import (
	"math"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
)

var dailyCostFactors = []float64{
	0.82, 0.91, 0.88, 0.95, 1.02, 0.97, 0.85, 0.93, 1.05, 1.1, 0.98, 1.03, 0.89, 0.94, 1.08, 1.12,
	0.96, 1.0, 0.87, 0.92, 1.06, 1.15, 0.99, 1.04, 0.9, 0.86, 1.01, 1.09, 0.95, 1.07,
}

var periodScale = map[types.CostPeriod]float64{
	types.CostPeriodCurrentMonth: 1,
	types.CostPeriodLastMonth:    0.88,
	types.CostPeriodLast7Days:    0.28,
	types.CostPeriodCustom:       0.5,
}

type periodMom struct {
	totalCostMom         float64
	avgCostPerRequestMom float64
	avgCostPerMemberMom  float64
	totalRequestsMom     float64
}

var periodMomValues = map[types.CostPeriod]periodMom{
	types.CostPeriodCurrentMonth: {12.5, 8.1, 10.2, 15.3},
	types.CostPeriodLastMonth:    {8.2, 5.4, 7.1, 9.8},
	types.CostPeriodLast7Days:    {-3.1, -1.2, -2.5, 4.6},
	types.CostPeriodCustom:       {6.0, 3.5, 4.2, 7.8},
}

const totalCostTarget = 67500

type departmentCostRow struct {
	departmentID   string
	departmentName string
	cost           float64
	hasChildren    bool
}

var topLevelDepartments = []departmentCostRow{
	{departmentID: "dept-2", departmentName: "技术部", cost: 38200, hasChildren: true},
	{departmentID: "dept-6", departmentName: "产品部", cost: 14300},
	{departmentID: "dept-7", departmentName: "市场部", cost: 8500},
	{departmentID: "dept-8", departmentName: "行政部", cost: 6500},
}

var childDepartments = map[string][]departmentCostRow{
	"dept-2": {
		{departmentID: "dept-3", departmentName: "后端组", cost: 21000, hasChildren: true},
		{departmentID: "dept-4", departmentName: "前端组", cost: 11200, hasChildren: true},
		{departmentID: "dept-5", departmentName: "测试组", cost: 6000, hasChildren: true},
	},
}

var deptMemberCosts = map[string][]types.DepartmentCostMember{
	"dept-3": {
		{MemberID: "m-2", MemberName: "李四", Cost: 12500, Requests: 5200, Tokens: 8500000},
		{MemberID: "m-1", MemberName: "张三", Cost: 8700, Requests: 3800, Tokens: 5800000},
	},
	"dept-4": {
		{MemberID: "m-4", MemberName: "赵六", Cost: 6200, Requests: 2900, Tokens: 4100000},
		{MemberID: "m-5", MemberName: "钱七", Cost: 5000, Requests: 2200, Tokens: 3200000},
	},
	"dept-5": {
		{MemberID: "m-6", MemberName: "孙八", Cost: 3500, Requests: 1500, Tokens: 2300000},
		{MemberID: "m-7", MemberName: "周九", Cost: 2500, Requests: 1100, Tokens: 1800000},
	},
}

type topConsumerSpec struct {
	memberID string
	cost     float64
	tokens   float64
	requests float64
}

var topConsumerSpecs = []topConsumerSpec{
	{memberID: "m-2", cost: 12500, tokens: 8500000, requests: 5200},
	{memberID: "m-1", cost: 8700, tokens: 5800000, requests: 3800},
	{memberID: "m-4", cost: 6200, tokens: 4100000, requests: 2900},
	{memberID: "m-11", cost: 5400, tokens: 3600000, requests: 2500},
	{memberID: "m-14", cost: 4800, tokens: 3200000, requests: 2200},
	{memberID: "m-5", cost: 4200, tokens: 2800000, requests: 1900},
	{memberID: "m-18", cost: 3900, tokens: 2600000, requests: 1700},
	{memberID: "m-22", cost: 3500, tokens: 2300000, requests: 1500},
	{memberID: "m-3", cost: 3200, tokens: 2100000, requests: 1400},
	{memberID: "m-25", cost: 2800, tokens: 1800000, requests: 1200},
}

func ResolvePeriod(period string) types.CostPeriod {
	switch types.CostPeriod(period) {
	case types.CostPeriodLastMonth, types.CostPeriodLast7Days, types.CostPeriodCustom:
		return types.CostPeriod(period)
	default:
		return types.CostPeriodCurrentMonth
	}
}

func BuildCostSummary(period types.CostPeriod, members []types.Member) types.CostSummary {
	scale := periodScale[period]
	mom := periodMomValues[period]
	totalCost := math.Round(totalCostTarget * scale)
	activeCount := 0
	for _, member := range members {
		if member.Status == "active" {
			activeCount++
		}
	}
	totalRequests := math.Round(28500 * scale)
	avgCostPerRequest := 0.0
	if totalRequests > 0 {
		avgCostPerRequest = math.Round((totalCost/totalRequests)*100) / 100
	}
	avgCostPerMember := 0.0
	if activeCount > 0 {
		avgCostPerMember = math.Round(totalCost / float64(activeCount))
	}
	return types.CostSummary{
		TotalCost:            totalCost,
		TotalCostMom:         mom.totalCostMom,
		TotalTokens:          math.Round(45000000 * scale),
		TotalRequests:        totalRequests,
		TotalRequestsMom:     mom.totalRequestsMom,
		AvgCostPerRequest:    avgCostPerRequest,
		AvgCostPerRequestMom: mom.avgCostPerRequestMom,
		AvgCostPerMember:     avgCostPerMember,
		AvgCostPerMemberMom:  mom.avgCostPerMemberMom,
	}
}

func GetDepartmentCostsForParent(parentID string, period types.CostPeriod) []types.DepartmentCost {
	scale := periodScale[period]
	var rows []departmentCostRow
	if parentID == "" {
		rows = topLevelDepartments
	} else {
		rows = childDepartments[parentID]
	}
	total := 0.0
	for _, row := range rows {
		total += row.cost
	}
	if total == 0 {
		total = 1
	}
	result := make([]types.DepartmentCost, 0, len(rows))
	for _, row := range rows {
		item := types.DepartmentCost{
			DepartmentID:   row.departmentID,
			DepartmentName: row.departmentName,
			Cost:           math.Round(row.cost * scale),
			Percentage:     math.Round((row.cost/total)*1000) / 10,
		}
		if row.hasChildren {
			item.HasChildren = true
		}
		result = append(result, item)
	}
	return result
}

func GetDepartmentMemberCosts(deptID string, period types.CostPeriod) []types.DepartmentCostMember {
	scale := periodScale[period]
	rows := deptMemberCosts[deptID]
	if rows == nil {
		return []types.DepartmentCostMember{}
	}
	result := make([]types.DepartmentCostMember, len(rows))
	for i, row := range rows {
		result[i] = types.DepartmentCostMember{
			MemberID:   row.MemberID,
			MemberName: row.MemberName,
			Cost:       math.Round(row.Cost * scale),
			Requests:   math.Round(row.Requests * scale),
			Tokens:     math.Round(row.Tokens * scale),
		}
	}
	return result
}

func BuildDailyCosts(period types.CostPeriod) []types.DailyCost {
	scale := periodScale[period]
	factors := dailyCostFactors
	startDay := 1
	if period == types.CostPeriodLast7Days {
		factors = dailyCostFactors[len(dailyCostFactors)-7:]
		startDay = 24
	}
	factorSum := 0.0
	for _, factor := range factors {
		factorSum += factor
	}
	baseDaily := (totalCostTarget * scale) / factorSum
	result := make([]types.DailyCost, len(factors))
	for i, factor := range factors {
		date := time.Date(2026, time.June, startDay+i, 0, 0, 0, 0, time.UTC)
		cost := math.Round(baseDaily*factor*100) / 100
		result[i] = types.DailyCost{
			Date:     date.Format("2006-01-02"),
			Cost:     cost,
			Tokens:   math.Round(cost * 700),
			Requests: math.Round(cost / 2.37),
		}
	}
	return result
}

func GetTopConsumers(limit int, period types.CostPeriod, members []types.Member) []types.TopConsumer {
	if limit <= 0 {
		limit = 5
	}
	scale := periodScale[period]
	if limit > len(topConsumerSpecs) {
		limit = len(topConsumerSpecs)
	}
	result := make([]types.TopConsumer, 0, limit)
	for _, spec := range topConsumerSpecs[:limit] {
		memberName := spec.memberID
		department := ""
		for _, member := range members {
			if member.ID == spec.memberID {
				memberName = member.Name
				department = member.DepartmentName
				break
			}
		}
		result = append(result, types.TopConsumer{
			MemberID:   spec.memberID,
			MemberName: memberName,
			Department: department,
			Cost:       math.Round(spec.cost * scale),
			Tokens:     math.Round(spec.tokens * scale),
			Requests:   math.Round(spec.requests * scale),
		})
	}
	return result
}
