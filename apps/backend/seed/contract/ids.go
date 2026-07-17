package contract

import "github.com/tokenjoy/backend/internal/pkg/modelcatalog"

const (
	TokenJoyCompanyID int64 = 1
	LocalCompanyID    int64 = 2
	DefaultCompanyID  int64 = LocalCompanyID
	DemoBudgetPeriod        = "2026-06"

	IDModelLocalTest int64 = 1

	ProdCatalogModelIDStart int64 = modelcatalog.ProdCatalogModelIDStart
	IDModel1                int64 = ProdCatalogModelIDStart // deepseek-v4
	IDModel2                int64 = 101                     // deepseek-r1
	IDModel3                int64 = 102                     // qwen-3.5-plus
	IDModel4                int64 = 103                     // qwen-max-2026
	IDModel5                int64 = 104                     // glm-5
	IDModel6                int64 = 105                     // kimi-k3
	IDModel7                int64 = 106                     // doubao-pro-256k
	IDModel8                int64 = 107                     // minimax-m2
	IDModel9                int64 = 108                     // claude-sonnet-5
	IDModel10               int64 = 109                     // gpt-4o

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
	modelcatalog.DevCallTypeLocalTest: IDModelLocalTest,
	"deepseek-v4":                     IDModel1,
	"deepseek-r1":                     IDModel2,
	"qwen-3.5-plus":                   IDModel3,
	"qwen-max-2026":                   IDModel4,
	"glm-5":                           IDModel5,
	"kimi-k3":                         IDModel6,
	"doubao-pro-256k":                 IDModel7,
	"minimax-m2":                      IDModel8,
	"claude-sonnet-5":                 IDModel9,
	"gpt-4o":                          IDModel10,
}
