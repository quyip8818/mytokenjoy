package usage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"sort"
	"sync"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
)

type LogAggregator struct {
	client  newapi.AdminClient
	store   store.Store
	logger  *slog.Logger
	cacheMu sync.RWMutex
	cache   map[string]cachedSeries
}

type cachedSeries struct {
	response  types.UsageSeriesResponse
	expiresAt time.Time
}

func NewLogAggregator(client newapi.AdminClient, st store.Store, logger *slog.Logger) *LogAggregator {
	if logger == nil {
		logger = slog.Default()
	}
	return &LogAggregator{
		client: client,
		store:  st,
		logger: logger,
		cache:  make(map[string]cachedSeries),
	}
}

func newAPIUnavailableError() error {
	return domain.NewDomainErrorWithRetryAfter(
		domain.StatusServiceUnavailable,
		"NewAPI unavailable",
		types.UsageMinuteRetryAfterSecs,
	)
}

func (a *LogAggregator) Series(ctx context.Context, q types.UsageSeriesQuery) (types.UsageSeriesResponse, error) {
	if a.client == nil {
		return types.UsageSeriesResponse{}, newAPIUnavailableError()
	}
	cacheKey := a.cacheKey(q)
	if cached, ok := a.getCache(cacheKey); ok {
		return cached, nil
	}

	fetchCtx, cancel := context.WithTimeout(ctx, types.UsageNewAPILogsTimeout)
	defer cancel()

	unmappedCount := 0
	truncated := false
	totalLogs := 0
	aggregated := make(map[seriesAggKey]types.UsageSeriesPoint)
	models, err := a.store.Models().Models(ctx)
	if err != nil {
		return types.UsageSeriesResponse{}, fmt.Errorf("load models: %w", err)
	}
	loc, err := common.LoadLocation(q.Timezone)
	if err != nil {
		return types.UsageSeriesResponse{}, err
	}

	for page := 1; page <= types.UsageMaxLogPages; page++ {
		logs, err := a.client.ListLogs(fetchCtx, newapi.ListLogsParams{
			Page:          page,
			PageSize:      types.UsageLogPageSize,
			StartUnixTime: q.Start.Unix(),
			EndUnixTime:   q.End.Unix(),
		})
		if err != nil {
			if cached, ok := a.getCache(cacheKey); ok {
				resp := cached
				resp.Approximate = true
				return resp, nil
			}
			return types.UsageSeriesResponse{}, domain.NewDomainErrorWithRetryAfter(
				domain.StatusServiceUnavailable,
				"NewAPI logs unavailable",
				types.UsageMinuteRetryAfterSecs,
			)
		}
		if len(logs) == 0 {
			break
		}
		for _, entry := range logs {
			if totalLogs >= types.UsageMaxLogEntries {
				truncated = true
				break
			}
			totalLogs++
			createdAt := time.Unix(entry.CreatedAt, 0).UTC()
			if createdAt.Before(q.Start) || !createdAt.Before(q.End) {
				continue
			}
			mapping, err := a.store.Relay().GetMappingByNewAPITokenID(ctx, entry.TokenID)
			if err != nil {
				return types.UsageSeriesResponse{}, fmt.Errorf("lookup relay mapping: %w", err)
			}
			if mapping == nil {
				unmappedCount++
				continue
			}
			if q.DepartmentID != "" && mapping.DepartmentID != q.DepartmentID {
				continue
			}
			if len(q.ScopeDeptIDs) > 0 && !containsID(q.ScopeDeptIDs, mapping.DepartmentID) {
				continue
			}
			memberID := ""
			if mapping.MemberID != nil {
				memberID = *mapping.MemberID
				if q.MemberID != "" && memberID != q.MemberID {
					continue
				}
			}
			modelName := ResolveLogEntryModel(entry)
			cost := CostCNYFromLog(entry.Quota, modelName, models)
			bucket := common.FormatBucketISO(common.TruncateInTZ(createdAt, time.Minute, loc))
			key := seriesAggKey{bucket: bucket}
			switch q.GroupBy {
			case types.UsageGroupByDepartment:
				key.departmentID = mapping.DepartmentID
			case types.UsageGroupByMember:
				key.memberID = memberID
			case types.UsageGroupByModel:
				key.model = modelName
			}
			point := aggregated[key]
			if point.Bucket == "" {
				point.Bucket = bucket
				point.DepartmentID = key.departmentID
				point.MemberID = key.memberID
				point.Model = key.model
			}
			point.CostCNY += cost
			point.CallCount++
			aggregated[key] = point
		}
		if truncated || len(logs) < types.UsageLogPageSize {
			break
		}
	}

	points := make([]types.UsageSeriesPoint, 0, len(aggregated))
	for _, point := range aggregated {
		points = append(points, point)
	}
	sort.Slice(points, func(i, j int) bool { return points[i].Bucket < points[j].Bucket })
	if err := ValidateSeriesPointLimit(len(points)); err != nil {
		return types.UsageSeriesResponse{}, err
	}

	unmapped := unmappedCount
	trunc := truncated
	resp := types.UsageSeriesResponse{
		Granularity:   types.UsageGranularityMinute,
		Source:        types.UsageSourceLogs,
		Timezone:      q.Timezone,
		Approximate:   true,
		MappingAsOf:   types.UsageMappingAsOfQueryTime,
		UnmappedCount: &unmapped,
		Truncated:     &trunc,
		Points:        points,
	}
	a.setCache(cacheKey, resp)
	return resp, nil
}

type seriesAggKey struct {
	bucket       string
	departmentID string
	memberID     string
	model        string
}

func (a *LogAggregator) cacheKey(q types.UsageSeriesQuery) string {
	raw := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%v",
		q.Granularity, q.Start.UTC().Format(time.RFC3339), q.End.UTC().Format(time.RFC3339),
		q.GroupBy, q.DepartmentID, q.MemberID, q.ScopeDeptIDs,
	)
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func (a *LogAggregator) getCache(key string) (types.UsageSeriesResponse, bool) {
	a.cacheMu.RLock()
	defer a.cacheMu.RUnlock()
	entry, ok := a.cache[key]
	if !ok || time.Now().After(entry.expiresAt) {
		return types.UsageSeriesResponse{}, false
	}
	return entry.response, true
}

func (a *LogAggregator) setCache(key string, resp types.UsageSeriesResponse) {
	a.cacheMu.Lock()
	defer a.cacheMu.Unlock()
	a.cache[key] = cachedSeries{response: resp, expiresAt: time.Now().Add(types.UsageMinuteCacheTTL)}
}

func containsID(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
