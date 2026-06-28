package permission

const (
	OrgRead        = "org:read"
	OrgDatasource  = "org:datasource"
	OrgStructure   = "org:structure"
	OrgRoles       = "org:roles"
	OrgMembers     = "org:members"
	BudgetRead     = "budget:read"
	BudgetAllocate = "budget:allocate"
	BudgetApprove  = "budget:approve"
	BudgetPolicy   = "budget:policy"
	ModelManage    = "model:manage"
	ModelRead      = "model:read"
	ModelWhitelist = "model:whitelist"
	KeysAdmin      = "keys:admin"
	KeysRead       = "keys:read"
	KeysProvider   = "keys:provider"
	SelfKeys       = "self:keys"
	SelfApproval   = "self:approval"
	DashboardCost  = "dashboard:cost"
	DashboardUsage = "dashboard:usage"
	AuditRead      = "audit:read"
	APICall        = "api:call"
)

var AllPermissions = []string{
	OrgRead,
	OrgDatasource,
	OrgStructure,
	OrgRoles,
	OrgMembers,
	BudgetRead,
	BudgetAllocate,
	BudgetApprove,
	BudgetPolicy,
	ModelManage,
	ModelRead,
	ModelWhitelist,
	KeysAdmin,
	KeysRead,
	KeysProvider,
	SelfKeys,
	SelfApproval,
	DashboardCost,
	DashboardUsage,
	AuditRead,
	APICall,
}

var PermissionIDMap = map[string]string{
	"p-1":  OrgStructure,
	"p-2":  OrgMembers,
	"p-3":  OrgRoles,
	"p-4":  OrgDatasource,
	"p-5":  BudgetAllocate,
	"p-6":  BudgetApprove,
	"p-7":  ModelWhitelist,
	"p-8":  DashboardCost,
	"p-9":  DashboardUsage,
	"p-10": AuditRead,
	"p-11": APICall,
	"p-12": BudgetRead,
	"p-13": BudgetPolicy,
	"p-14": ModelManage,
	"p-15": KeysAdmin,
	"p-16": KeysProvider,
	"p-17": SelfKeys,
	"p-18": SelfApproval,
	"p-19": OrgRead,
	"p-20": KeysRead,
	"p-21": ModelRead,
}
