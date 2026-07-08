package remote

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
)

var defaultFieldMappings = map[types.Platform][]types.FieldMapping{
	types.PlatformFeishu: {
		{SourceField: "name", SourceLabel: "姓名", TargetField: "name", Required: true},
		{SourceField: "mobile", SourceLabel: "手机号", TargetField: "phone", Required: true},
		{SourceField: "email", SourceLabel: "邮箱", TargetField: "email", Required: false},
		{SourceField: "department_id", SourceLabel: "部门 ID", TargetField: "departmentId", Required: true},
		{SourceField: "user_id", SourceLabel: "用户 ID", TargetField: "_ignore", Required: false},
		{SourceField: "status", SourceLabel: "状态", TargetField: "status", Required: false},
	},
	types.PlatformDingtalk: {
		{SourceField: "name", SourceLabel: "姓名", TargetField: "name", Required: true},
		{SourceField: "mobile", SourceLabel: "手机号", TargetField: "phone", Required: true},
		{SourceField: "email", SourceLabel: "邮箱", TargetField: "email", Required: false},
		{SourceField: "dept_id", SourceLabel: "部门 ID", TargetField: "departmentId", Required: true},
		{SourceField: "userid", SourceLabel: "用户 ID", TargetField: "_ignore", Required: false},
	},
	types.PlatformWecom: {
		{SourceField: "name", SourceLabel: "姓名", TargetField: "name", Required: true},
		{SourceField: "mobile", SourceLabel: "手机号", TargetField: "phone", Required: true},
		{SourceField: "email", SourceLabel: "邮箱", TargetField: "email", Required: false},
		{SourceField: "department", SourceLabel: "部门 ID", TargetField: "departmentId", Required: true},
		{SourceField: "userid", SourceLabel: "用户 ID", TargetField: "_ignore", Required: false},
	},
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
	integration, err := s.d.Store.Org().Integration(ctx)
	if err != nil {
		return nil, err
	}
	if integration.Connected && integration.Platform != nil && *integration.Platform != p {
		return nil, domain.Validation("platform mismatch")
	}
	mappings, err := s.d.Store.Org().FieldMappings(ctx)
	if err != nil {
		return nil, err
	}
	// Return platform defaults if no custom mappings are stored
	if len(mappings) == 0 {
		if defaults, ok := defaultFieldMappings[p]; ok {
			return defaults, nil
		}
	}
	return append([]types.FieldMapping{}, mappings...), nil
}

func (s *Service) SaveFieldMappings(ctx context.Context, config types.FieldMappingConfig) error {
	if config.Platform == "" {
		return domain.Validation("platform is required")
	}
	if len(config.Mappings) == 0 {
		return domain.Validation("mappings are required")
	}
	integration, err := s.d.Store.Org().Integration(ctx)
	if err != nil {
		return err
	}
	if integration.Connected && integration.Platform != nil && *integration.Platform != config.Platform {
		return domain.Validation("platform mismatch")
	}
	return s.d.Store.Org().SetFieldMappings(ctx, config.Mappings)
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
	_ = p
	name := keyword
	if keyword != "张三" {
		name = fmt.Sprintf("%s（模拟）", keyword)
	}
	return types.MappingTestResult{
		Success: true,
		Preview: map[string]string{
			"姓名":  name,
			"手机号": "138****1234",
			"邮箱":  "user@example.com",
			"部门":  "技术部 > 后端组",
			"状态":  "在职",
		},
		Errors: []string{},
	}, nil
}
