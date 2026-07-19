package contract

import "github.com/google/uuid"

// remoteIDNamespace mirrors pkg/org.remoteIDNamespace for deterministic external→local ID mapping.
var remoteIDNamespace = uuid.MustParse("6ba7b814-9dad-11d1-80b4-00c04fd430c8")

// --- Companies ---

var (
	TokenJoyCompanyID = uuid.MustParse("00000000-0000-7000-8000-000000000001")
	LocalCompanyID    = uuid.MustParse("00000000-0000-7000-8000-000000000002")
	DefaultCompanyID  = LocalCompanyID
)

// --- Models ---

var (
	IDModelTest = uuid.MustParse("00000000-0000-7000-8000-0000000000a1")
	IDModel1    = uuid.MustParse("00000000-0000-7000-8000-0000000000b1") // deepseek-v4
	IDModel2    = uuid.MustParse("00000000-0000-7000-8000-0000000000b2") // deepseek-r1
	IDModel3    = uuid.MustParse("00000000-0000-7000-8000-0000000000b3") // qwen-3.5-plus
	IDModel4    = uuid.MustParse("00000000-0000-7000-8000-0000000000b4") // qwen-max-2026
	IDModel5    = uuid.MustParse("00000000-0000-7000-8000-0000000000b5") // glm-5
	IDModel6    = uuid.MustParse("00000000-0000-7000-8000-0000000000b6") // kimi-k3
	IDModel7    = uuid.MustParse("00000000-0000-7000-8000-0000000000b7") // doubao-pro-256k
	IDModel8    = uuid.MustParse("00000000-0000-7000-8000-0000000000b8") // minimax-m2
	IDModel9    = uuid.MustParse("00000000-0000-7000-8000-0000000000b9") // claude-sonnet-5
	IDModel10   = uuid.MustParse("00000000-0000-7000-8000-0000000000ba") // gpt-4o
)

// --- Departments ---

var (
	IDDept1 = uuid.MustParse("00000000-0000-7000-8000-000000000d01")
	IDDept2 = uuid.MustParse("00000000-0000-7000-8000-000000000d02")
	IDDept3 = uuid.MustParse("00000000-0000-7000-8000-000000000d03")
	IDDept4 = uuid.MustParse("00000000-0000-7000-8000-000000000d04")
	IDDept5 = uuid.MustParse("00000000-0000-7000-8000-000000000d05")
	IDDept6 = uuid.MustParse("00000000-0000-7000-8000-000000000d06")
	IDDept7 = uuid.MustParse("00000000-0000-7000-8000-000000000d07")
	IDDept8 = uuid.MustParse("00000000-0000-7000-8000-000000000d08")

	IDFeishuDept1 = uuid.NewSHA1(remoteIDNamespace, []byte("dept-feishu-od-1"))
)

// --- Members ---

var (
	IDMemberAdmin   = uuid.MustParse("00000000-0000-7000-8000-000000000e01")
	IDMember1       = uuid.MustParse("00000000-0000-7000-8000-000000000e02")
	IDMember2       = uuid.MustParse("00000000-0000-7000-8000-000000000e07")
	IDMember3       = uuid.MustParse("00000000-0000-7000-8000-000000000e03")
	IDMember4       = uuid.MustParse("00000000-0000-7000-8000-000000000e06")
	IDMember5       = uuid.MustParse("00000000-0000-7000-8000-000000000e09")
	IDMember6       = uuid.MustParse("00000000-0000-7000-8000-000000000e08")
	IDMember15      = uuid.MustParse("00000000-0000-7000-8000-000000000e15")
	IDMember16      = uuid.MustParse("00000000-0000-7000-8000-000000000e16")
	IDMemberPure    = uuid.MustParse("00000000-0000-7000-8000-000000000e04")
	IDMemberAuditor = uuid.MustParse("00000000-0000-7000-8000-000000000e05")
)

// --- Platform Keys ---

var (
	IDPlatformKey1 = uuid.MustParse("00000000-0000-7000-8000-000000000f01")
	IDPlatformKey2 = uuid.MustParse("00000000-0000-7000-8000-000000000f02")
	IDPlatformKey3 = uuid.MustParse("00000000-0000-7000-8000-000000000f03")
	IDPlatformKey4 = uuid.MustParse("00000000-0000-7000-8000-000000000f04")
	IDPlatformKey5 = uuid.MustParse("00000000-0000-7000-8000-000000000f05")
	IDPlatformKey6 = uuid.MustParse("00000000-0000-7000-8000-000000000f06")
)

// --- Approvals ---

var (
	IDApproval1 = uuid.MustParse("00000000-0000-7000-8000-0000000000c1")
	IDApproval2 = uuid.MustParse("00000000-0000-7000-8000-0000000000c2")
)

// --- Projects ---

var (
	IDProject1 = uuid.MustParse("00000000-0000-7000-8000-000000000101")
	IDProject4 = uuid.MustParse("00000000-0000-7000-8000-000000000104")
)

// --- Roles ---

var (
	IDRole1 = uuid.MustParse("00000000-0000-7000-8000-00000000a101")
	IDRole2 = uuid.MustParse("00000000-0000-7000-8000-00000000a102")
	IDRole3 = uuid.MustParse("00000000-0000-7000-8000-00000000a103")
	IDRole4 = uuid.MustParse("00000000-0000-7000-8000-00000000a104")
	IDRole5 = uuid.MustParse("00000000-0000-7000-8000-00000000a105")
	IDRole6 = uuid.MustParse("00000000-0000-7000-8000-00000000a106")
)

// --- Misc seed IDs ---

var (
	IDSyncLog1       = uuid.MustParse("00000000-0000-7000-8000-0000000aa001")
	IDBudgetAppr4    = uuid.MustParse("00000000-0000-7000-8000-000000000a04")
	IDBudgetAppr5    = uuid.MustParse("00000000-0000-7000-8000-000000000a05")
	IDMemberFallback = uuid.MustParse("00000000-0000-7000-8000-000000000e45")

	IDAlertRule1 = uuid.MustParse("00000000-0000-7000-8000-0000000ab001")
	IDAlertRule2 = uuid.MustParse("00000000-0000-7000-8000-0000000ab002")
	IDAlertRule3 = uuid.MustParse("00000000-0000-7000-8000-0000000ab003")
	IDAlertRule4 = uuid.MustParse("00000000-0000-7000-8000-0000000ab004")
	IDAlertRule5 = uuid.MustParse("00000000-0000-7000-8000-0000000ab005")
	IDAlertRule6 = uuid.MustParse("00000000-0000-7000-8000-0000000ab006")
	IDAlertRule7 = uuid.MustParse("00000000-0000-7000-8000-0000000ab007")
	IDAlertRule8 = uuid.MustParse("00000000-0000-7000-8000-0000000ab008")

	IDBudgetApproval1 = uuid.MustParse("00000000-0000-7000-8000-000000000a01")
	IDBudgetApproval2 = uuid.MustParse("00000000-0000-7000-8000-000000000a02")
	IDBudgetApproval3 = uuid.MustParse("00000000-0000-7000-8000-000000000a03")

	// Seed lot/order used by usage_ledger seed data.
	IDSeedLotOrder = uuid.MustParse("00000000-0000-7000-8000-000000000aa0")
	IDSeedLot      = uuid.MustParse("00000000-0000-7000-8000-000000000aa1")
)

// --- Constants ---

const DemoBudgetPeriod = "2026-06"

const (
	// External IDs remain strings — they're not internal entity IDs.
	IDFeishuExtDept1 = "od-1"
	IDFeishuExtUser1 = "ou-1"
)

// ModelTypeToID maps demo seed model types to catalog model_id values.
// NOTE: Cannot reference modelcatalog.TestCallType here due to import cycle
// (contract is imported by modelcatalog's test). The literal must stay in sync.
var ModelTypeToID = map[string]uuid.UUID{
	"test-model":      IDModelTest,
	"deepseek-v4":     IDModel1,
	"deepseek-r1":     IDModel2,
	"qwen-3.5-plus":   IDModel3,
	"qwen-max-2026":   IDModel4,
	"glm-5":           IDModel5,
	"kimi-k3":         IDModel6,
	"doubao-pro-256k": IDModel7,
	"minimax-m2":      IDModel8,
	"claude-sonnet-5": IDModel9,
	"gpt-4o":          IDModel10,
}
