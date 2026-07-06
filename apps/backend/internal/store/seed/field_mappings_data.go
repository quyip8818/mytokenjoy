package seed

import "github.com/tokenjoy/backend/internal/domain/types"

func buildDefaultFieldMappings() []types.FieldMapping {
	return []types.FieldMapping{
		{SourceField: "user_name", SourceLabel: "用户姓名", TargetField: "name", Required: true},
		{SourceField: "mobile", SourceLabel: "手机号码", TargetField: "phone", Required: true},
		{SourceField: "user_email", SourceLabel: "邮箱地址", TargetField: "email", Required: false},
		{SourceField: "dept_name", SourceLabel: "部门名称", TargetField: "departmentName", Required: true},
		{SourceField: "dept_id", SourceLabel: "部门 ID", TargetField: "departmentId", Required: true},
		{SourceField: "user_status", SourceLabel: "用户状态", TargetField: "status", Required: false},
	}
}
