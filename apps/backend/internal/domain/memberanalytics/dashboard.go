package memberanalytics

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
)

type AccountStats struct {
	BudgetRemaining float64 `json:"budgetRemaining"`
	TotalSpent      float64 `json:"totalSpent"`
}

type UsageStats struct {
	RequestCount int `json:"requestCount"`
	TotalCount   int `json:"totalCount"`
}

type ResourceConsumption struct {
	TotalCost   float64 `json:"totalCost"`
	TotalTokens int64   `json:"totalTokens"`
}

type PerformanceStats struct {
	AvgRPM float64 `json:"avgRPM"`
	AvgTPM int64   `json:"avgTPM"`
}

type TimeSeriesPoint struct {
	Time  string  `json:"time"`
	Value float64 `json:"value"`
}

type NamedValue struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

type ModelRank struct {
	Model string `json:"model"`
	Count int    `json:"count"`
}

type DashboardView struct {
	Account                 AccountStats        `json:"account"`
	UsageStats              UsageStats          `json:"usageStats"`
	ResourceConsumption     ResourceConsumption `json:"resourceConsumption"`
	Performance             PerformanceStats    `json:"performance"`
	ConsumptionTrend        []TimeSeriesPoint   `json:"consumptionTrend"`
	ConsumptionDistribution []TimeSeriesPoint   `json:"consumptionDistribution"`
	CallDistribution        []NamedValue        `json:"callDistribution"`
	CallRanking             []ModelRank         `json:"callRanking"`
}

func (s *service) GetDashboard(ctx context.Context, memberID uuid.UUID) (DashboardView, error) {
	if memberID == uuid.Nil {
		return DashboardView{}, domain.BadRequest("memberId is required")
	}
	loc, err := time.LoadLocation(types.UsageDefaultTimezone)
	if err != nil {
		return DashboardView{}, fmt.Errorf("load timezone: %w", err)
	}
	end := s.dashboardWindowEnd()
	start := end.AddDate(0, 0, -30)
	summary, err := s.reader.QuerySummary(ctx, types.UsageAggregateQuery{
		Start:    start,
		End:      end,
		Timezone: types.UsageDefaultTimezone,
		MemberID: memberID,
	})
	if err != nil {
		return DashboardView{}, fmt.Errorf("query usage summary: %w", err)
	}
	budgetSummary, err := s.keys.BudgetSummary(ctx, memberID)
	if err != nil {
		return DashboardView{}, fmt.Errorf("budget summary: %w", err)
	}
	trendPoints, err := s.reader.QuerySeries(ctx, types.UsageSeriesQuery{
		Granularity: types.UsageGranularityHour,
		Start:       start,
		End:         end,
		GroupBy:     types.UsageGroupByNone,
		MemberID:    memberID,
		Timezone:    types.UsageDefaultTimezone,
	})
	if err != nil {
		return DashboardView{}, fmt.Errorf("query usage trend: %w", err)
	}
	modelPoints, err := s.reader.QuerySeries(ctx, types.UsageSeriesQuery{
		Granularity: types.UsageGranularityHour,
		Start:       start,
		End:         end,
		GroupBy:     types.UsageGroupByModel,
		MemberID:    memberID,
		Timezone:    types.UsageDefaultTimezone,
	})
	if err != nil {
		return DashboardView{}, fmt.Errorf("query usage by model: %w", err)
	}
	consumptionTrend := seriesCostByBucket(trendPoints, loc)
	consumptionDistribution := seriesCostByBucket(modelPoints, loc)
	callDistribution, callRanking := modelCallBreakdown(modelPoints)
	avgRPM, avgTPM := performanceStats(summary, start, end)
	return DashboardView{
		Account: AccountStats{
			BudgetRemaining: budgetSummary.Remaining,
			TotalSpent:      summary.Spend(),
		},
		UsageStats: UsageStats{
			RequestCount: summary.CallCount,
			TotalCount:   summary.CallCount,
		},
		ResourceConsumption: ResourceConsumption{
			TotalCost:   summary.Spend(),
			TotalTokens: summary.InputTokens + summary.OutputTokens,
		},
		Performance: PerformanceStats{
			AvgRPM: avgRPM,
			AvgTPM: avgTPM,
		},
		ConsumptionTrend:        consumptionTrend,
		ConsumptionDistribution: consumptionDistribution,
		CallDistribution:        callDistribution,
		CallRanking:             callRanking,
	}, nil
}

func (s *service) dashboardWindowEnd() time.Time {
	anchor := s.clock.Now().UTC()
	return time.Date(anchor.Year(), anchor.Month(), anchor.Day(), 0, 0, 0, 0, time.UTC).Add(24 * time.Hour)
}

func seriesCostByBucket(points []types.UsageSeriesPoint, loc *time.Location) []TimeSeriesPoint {
	byBucket := make(map[string]float64)
	for _, point := range points {
		label := formatChartTime(point.Bucket, loc)
		byBucket[label] += point.Cost
	}
	labels := make([]string, 0, len(byBucket))
	for label := range byBucket {
		labels = append(labels, label)
	}
	sort.Strings(labels)
	result := make([]TimeSeriesPoint, 0, len(labels))
	for _, label := range labels {
		result = append(result, TimeSeriesPoint{Time: label, Value: byBucket[label]})
	}
	return result
}

func formatChartTime(bucket string, loc *time.Location) string {
	parsed, err := time.Parse(time.RFC3339, bucket)
	if err != nil {
		return bucket
	}
	return parsed.In(loc).Format("01-02 15:04")
}

func modelCallBreakdown(points []types.UsageSeriesPoint) ([]NamedValue, []ModelRank) {
	byModel := make(map[string]int)
	for _, point := range points {
		if point.Model == "" {
			continue
		}
		byModel[point.Model] += point.CallCount
	}
	distribution := make([]NamedValue, 0, len(byModel))
	ranking := make([]ModelRank, 0, len(byModel))
	for model, count := range byModel {
		distribution = append(distribution, NamedValue{Name: model, Value: float64(count)})
		ranking = append(ranking, ModelRank{Model: model, Count: count})
	}
	sort.Slice(distribution, func(i, j int) bool {
		return distribution[i].Value > distribution[j].Value
	})
	sort.Slice(ranking, func(i, j int) bool {
		return ranking[i].Count > ranking[j].Count
	})
	return distribution, ranking
}

func performanceStats(totals types.UsageSummaryTotals, start, end time.Time) (float64, int64) {
	minutes := end.Sub(start).Minutes()
	if minutes <= 0 {
		return 0, 0
	}
	avgRPM := float64(totals.CallCount) / minutes
	tokenTotal := totals.InputTokens + totals.OutputTokens
	avgTPM := int64(float64(tokenTotal) / minutes)
	return avgRPM, avgTPM
}
