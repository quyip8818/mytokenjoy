package contract

const (
	TokenJoyCompanyID int64 = 1
	LocalCompanyID    int64 = 2
	DefaultCompanyID  int64 = LocalCompanyID
	DemoBudgetPeriod        = "2026-06"

	IDModelLocalTest int64 = 1

	ProdCatalogModelIDStart int64 = 100
	IDModel1                int64 = ProdCatalogModelIDStart // gpt-4o
	IDModel2                int64 = 101                     // gpt-4o-mini
	IDModel3                int64 = 102                     // claude-opus-4-8
	IDModel4                int64 = 103                     // claude-sonnet-4-6
	IDModel5                int64 = 104                     // deepseek-v3
	IDModel6                int64 = 105                     // deepseek-r1
	IDModel7                int64 = 106                     // qwen-max
	IDModel8                int64 = 107                     // qwen-plus

	IDDept2          = "dept-2"
	IDDept3          = "dept-3"
	IDDept4          = "dept-4"
	IDDept5          = "dept-5"
	IDMemberAdmin    = "m-admin"
	IDMember1        = "m-1"
	IDMember3        = "m-3"
	IDMemberPure     = "m-pure"
	IDMemberAuditor  = "m-auditor"
	IDPlatformKey1   = "plk-1"
	IDApproval1      = "apv-1"
	IDApproval2      = "apv-2"
	IDProject4       = "proj-4"
	IDProject1       = "proj-1"
	IDFeishuExtDept1 = "od-1"
	IDFeishuExtUser1 = "ou-1"
	IDFeishuDept1    = "dept-feishu-od-1"
)

// ModelTypeToID maps demo seed model types to catalog model_id values.
var ModelTypeToID = map[string]int64{
	"local-test-model":  IDModelLocalTest,
	"gpt-4o":            IDModel1,
	"gpt-4o-mini":       IDModel2,
	"claude-opus-4-8":   IDModel3,
	"claude-sonnet-4-6": IDModel4,
	"deepseek-v3":       IDModel5,
	"deepseek-r1":       IDModel6,
	"qwen-max":          IDModel7,
	"qwen-plus":         IDModel8,
}
