// TODO(real): replace in-memory field mappings with sync engine persistence.
package remote

import (
	"context"
	"fmt"
	"sync"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/types"
)

type fieldMappingKey struct {
	companyID int64
	platform  types.Platform
}

var (
	fieldMappingsMu      sync.RWMutex
	fieldMappingsStore   = make(map[fieldMappingKey][]types.FieldMapping)
	fieldMappingsSeeded  = make(map[fieldMappingKey]bool)
)

func defaultFieldMappings() []types.FieldMapping {
	return []types.FieldMapping{
		{SourceField: "user_name", SourceLabel: "用户姓名", TargetField: "name", Required: true},
		{SourceField: "mobile", SourceLabel: "手机号码", TargetField: "phone", Required: true},
		{SourceField: "user_email", SourceLabel: "邮箱地址", TargetField: "email", Required: false},
		{SourceField: "dept_name", SourceLabel: "部门名称", TargetField: "departmentName", Required: true},
		{SourceField: "dept_id", SourceLabel: "部门 ID", TargetField: "departmentId", Required: true},
		{SourceField: "user_status", SourceLabel: "用户状态", TargetField: "status", Required: false},
	}
}

func loadFieldMappings(companyID int64, platform types.Platform) []types.FieldMapping {
	key := fieldMappingKey{companyID: companyID, platform: platform}
	fieldMappingsMu.Lock()
	defer fieldMappingsMu.Unlock()
	if !fieldMappingsSeeded[key] {
		seeded := defaultFieldMappings()
		fieldMappingsStore[key] = append([]types.FieldMapping{}, seeded...)
		fieldMappingsSeeded[key] = true
	}
	return append([]types.FieldMapping{}, fieldMappingsStore[key]...)
}

func saveFieldMappings(companyID int64, platform types.Platform, mappings []types.FieldMapping) {
	key := fieldMappingKey{companyID: companyID, platform: platform}
	fieldMappingsMu.Lock()
	defer fieldMappingsMu.Unlock()
	fieldMappingsStore[key] = append([]types.FieldMapping{}, mappings...)
	fieldMappingsSeeded[key] = true
}

func parsePlatform(platform string) (types.Platform, error) {
	switch types.Platform(platform) {
	case types.PlatformFeishu, types.PlatformDingtalk, types.PlatformWecom:
		return types.Platform(platform), nil
	default:
		return "", domain.Validation("invalid platform")
	}
}

func (s *Service) GetFieldMappings(ctx context.Context, platform string) ([]types.FieldMapping, error) {
	p, err := parsePlatform(platform)
	if err != nil {
		return nil, err
	}
	return loadFieldMappings(company.CompanyID(ctx), p), nil
}

func (s *Service) SaveFieldMappings(ctx context.Context, config types.FieldMappingConfig) error {
	if config.Platform == "" {
		return domain.Validation("platform is required")
	}
	if len(config.Mappings) == 0 {
		return domain.Validation("mappings are required")
	}
	saveFieldMappings(company.CompanyID(ctx), config.Platform, config.Mappings)
	return nil
}

func (s *Service) TestFieldMapping(ctx context.Context, platform, keyword string) (types.MappingTestResult, error) {
	p, err := parsePlatform(platform)
	if err != nil {
		return types.MappingTestResult{}, err
	}
	if keyword == "" {
		return types.MappingTestResult{
			Success: false,
			Preview: map[string]string{},
			Errors:  []string{"请输入搜索关键词"},
		}, nil
	}
	_ = loadFieldMappings(company.CompanyID(ctx), p)
	name := keyword
	if keyword == "张三" {
		name = "张三"
	} else {
		name = fmt.Sprintf("%s（模拟）", keyword)
	}
	return types.MappingTestResult{
		Success: true,
		Preview: map[string]string{
			"姓名": name,
			"手机号": "138****1234",
			"邮箱":  "user@example.com",
			"部门":  "技术部 > 后端组",
			"状态":  "在职",
		},
		Errors: []string{},
	}, nil
}
