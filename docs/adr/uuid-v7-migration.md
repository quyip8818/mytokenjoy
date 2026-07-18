# UUID v7 迁移实现文档

## 目标

将所有业务实体 ID 从 `BIGINT`/`TEXT` 统一为 PostgreSQL 原生 `UUID` 类型 + Go `uuid.UUID`。

系统未上线，直接重写，不需要向后兼容、不需要数据迁移。

## 依赖

```
github.com/google/uuid   // uuid.UUID, uuid.NewV7(), uuid.Nil, uuid.MustParse()
```

---

## 转换规则

### 转换（白名单）

以下 Go struct 字段转为 `uuid.UUID`：

| 字段名 | 原类型 | DB 列 |
|--------|--------|-------|
| `ID` | `string`/`int64` | 主键（排除 `RawConsumeLog`、`RiverJobView`、`SSEEvent`、`internal/integration/` 包） |
| `CompanyID` | `int64` | `company_id` |
| `UserID` | `string` | `user_id` |
| `MemberID` | `string`/`*string` | `member_id` |
| `DepartmentID` | `string` | `department_id` |
| `ProjectID` | `string`/`*string` | `project_id` |
| `PlatformKeyID` | `string` | `platform_key_id` |
| `NodeID` | `string` | `node_id` |
| `RoleID` | `string` | `role_id` |
| `LotID` | `string` | `lot_id` |
| `RechargeOrderID` | `string` | `recharge_order_id` |
| `ApplicantID` | `string` | `applicant_id` |
| `OperatorID` | `string`（仅 `OperationLog`） | `operator_id` |
| `ManagerID` | `*string` | `manager_id` |
| `ParentID` | `*string` | `parent_id` |
| `DefaultModelID` | `*int64` | `default_model_id` |
| `FallbackModelID` | `*int64` | `fallback_model_id` |
| `OwnerDepartmentID` | `string` | `owner_department_id` |
| `FIFOHeadLotID` | `*string` | `fifo_head_lot_id` |
| `RootDeptID` | `*string` | `root_dept_id` |
| `LastLedgerID` | `*string` | `last_ledger_id` |
| `AxisID` | `string`（仅 `ConsumedDelta`） | `axis_id` |
| `OwnerID` | `string`（仅 `ModelAllowlistRow`） | `owner_id` |
| `ModelID` | `int64`（仅 `ModelAllowlistRow`/`ModelInfo`） | `model_id` |
| `CreatedBy` | `string`（仅 `RechargeOrder`） | `created_by` |

集合字段（匹配字段名+类型）：

| 字段名 | 原类型 | 转换为 |
|--------|--------|--------|
| `ModelWhitelist` | `[]int64` | `[]uuid.UUID` |
| `RequestedModels` | `[]int64` | `[]uuid.UUID` |
| `AllowedModelIDs` | `[]int64` | `[]uuid.UUID` |
| `NotifyRoleIDs` | `[]string` | `[]uuid.UUID` |
| `MemberIDs` | `[]string` | `[]uuid.UUID` |
| `MemberBudgets`（仅 `Project`） | `map[string]float64` | `map[uuid.UUID]float64` |

函数参数（精确参数名匹配）：

```
companyID int64 → uuid.UUID      memberID string → uuid.UUID
projectID string → uuid.UUID     departmentID string → uuid.UUID
nodeID string → uuid.UUID        platformKeyID string → uuid.UUID
keyID string → uuid.UUID         modelID int64 → uuid.UUID
ownerID string → uuid.UUID       lotID string → uuid.UUID
userID string → uuid.UUID        providerKeyID string → uuid.UUID
operatorID string → uuid.UUID    deptID string → uuid.UUID
```

### 不转换（黑名单）

| 字段/参数 | 保持类型 | 理由 |
|-----------|---------|------|
| `ExternalID` | `*string` | 飞书/钉钉外部 ID |
| `EmployeeID` | `string` | 外部员工编号 |
| `NewAPIWalletUserID` | `*int64` | NewAPI 外部系统 |
| `NewAPIKeyID` | `*int64` | NewAPI 外部系统 |
| `NewAPIChannelID` | `int` | NewAPI 外部系统 |
| `PackageID` | `*string` | 外部套餐标识 |
| `CallerID` | `string` | 调用详情显示用 |
| `AccessKeyID` | `string` | 阿里云 SMS |
| `AppID`/`CorpID`/`AgentID` | `string` | 三方平台凭证 |
| `TokenID` / `LogID` | `int64` | NewAPI 外部表 |
| `InviteCode` / `IdempotencyKey` | `string` | 非实体 ID |
| `KeyHash` / `KeyPrefix` | `string` | 密钥衍生值 |
| `Operator` | `string` | 显示名称 |
| `axisKind` / `ownerType` / `periodKey` | `string` | 枚举/标识值 |
| `logID` / `tokenID` 参数 | `int64` | NewAPI 外部 |
| `permissions.id` 相关 | `string` | TEXT 语义枚举 |

### 不变的 DB 列

`currencies.currency`、`permissions.id`、`role_permission_grants.permission_id`、`scheduler_locks.lock_name`、`reconcile_cursors.stream`、`platform_key_mappings.newapi_key_id`、`provider_keys.newapi_channel_id`、`companies.newapi_wallet_user_id`、`logs.id`、`org_nodes.path`、`org_nodes.external_id`、`members.external_id`/`employee_id`

---

## 需要删除的旧逻辑

### 1. `ModelID < 100` 数字范围判断

**删除**：
- `ProdCatalogModelIDStart int64 = 100`
- `func IsDevCatalogModelID(id int64) bool { return id > 0 && id < 100 }`

**替换为** model `Type` 字段 `dev-` 前缀判断：

```go
func IsDevModel(m types.ModelInfo) bool {
    return strings.HasPrefix(m.Type, "dev-")
}
```

原 `DevCallTypeLocalTest = "local-test-model"` 改为 `"dev-local-test"`。

### 2. `CompanyID` 自增 + 范围校验

**删除**：
- `var nextID int64 = 1; for _, t := range companies { if t.ID >= nextID { nextID = t.ID + 1 } }`
- `if c.TokenJoyCompanyID <= 0 || c.LocalCompanyID <= 0 { error }`
- `if c.TokenJoyCompanyID >= 1000000 || c.LocalCompanyID >= 1000000 { error }`

**替换为**：
- Company ID 生成：`uuid.Must(uuid.NewV7())`
- Config 校验改为：`if c.TokenJoyCompanyID == uuid.Nil { error }`

### 3. 旧 ID 生成函数

**全部删除**：
- `internal/domain/org/structure/id.go` → `generateID(prefix string)`
- `internal/domain/budget/service.go` → `generateBudgetID(prefix string)`
- 所有 `fmt.Sprintf("prefix-%d", time.Now().UnixMilli())` 调用
- 所有 `fmt.Sprintf("prefix-%d-%x", time.Now().UnixMilli(), rand)` 调用
- 所有 `fmt.Sprintf("prefix-%d-%d", companyID, time.Now().UnixNano())` 调用

**统一替换为**：`uuid.Must(uuid.NewV7())`

---

## 批处理策略：AST 重写 + 编译器驱动

### 核心思路

用 Go AST 重写脚本精确修改类型声明（struct 字段 + 函数参数），然后让编译器驱动剩余修复。

### 为什么用 AST 而不是 regex

| | regex/sed | AST 重写 |
|---|---|---|
| struct field `CompanyID int64` | ✅ | ✅ |
| func param `companyID int64` | ✅ | ✅ |
| 局部变量 `companyID := int64(0)` | ❌ 误改 | ✅ 跳过 |
| 注释 `// CompanyID int64` | ❌ 误改 | ✅ 跳过 |
| 黑名单 `NewAPIKeyID *int64` | ❌ 可能误改 | ✅ 精确跳过 |
| 字符串 `"CompanyID"` | ❌ 可能误改 | ✅ 跳过 |

### 步骤

#### Step 1: AST 重写脚本

写 `scripts/uuid-migrate.go`，基于 `go/ast` + `go/parser` + `go/printer`：

- 遍历所有 `.go` 文件
- 对 struct field：字段名在白名单且类型匹配旧类型 → 改为 `uuid.UUID`（或 `*uuid.UUID`）
- 对 func/method 参数：参数名在白名单且类型匹配 → 改类型
- 自动添加 `"github.com/google/uuid"` import
- 跳过黑名单字段、`internal/integration/` 包、排除列表中的 struct

一次运行，覆盖所有文件。

#### Step 2: goimports

```bash
goimports -w ./...
```

清理多余/缺失的 import。

#### Step 3: 编译器驱动修复

```bash
go build ./... 2>&1 | grep -v "_test.go" > /tmp/errors.txt
```

AST 脚本只改声明，不改函数体内的赋值/比较/调用。编译器报出剩余错误，按类型分类批处理：

| 错误模式 | 修复方式 |
|---------|---------|
| `cannot use "" as uuid.UUID` | 改为 `uuid.Nil` |
| `cannot use 0 as uuid.UUID` | 改为 `uuid.Nil` |
| `invalid operation: x == ""` | 改为 `x == uuid.Nil` |
| `invalid operation: x <= 0` | 改为 `x == uuid.Nil`（删除范围校验） |
| `undefined: generateID` | 改为 `uuid.Must(uuid.NewV7())` |
| `fmt.Sprintf("prefix-...")` | 改为 `uuid.Must(uuid.NewV7())` |
| `cannot use x (string) as uuid.UUID` | 改变量声明类型或加 `.String()` |

每类错误用 grep + sed 或编辑器批量修复，**只改编译器报出来的行**。

#### Step 4: seed 常量化

`seed/contract/ids.go` 集中定义所有硬编码 UUID：

```go
var (
    TokenJoyCompanyID = uuid.MustParse("...")
    LocalCompanyID    = uuid.MustParse("...")
    IDModel1          = uuid.MustParse("...")
    ...
)
```

seed 文件中的字符串字面量（`"dept-1"` 等）统一替换为常量引用。

#### Step 5: 前端

TypeScript 类型文件精确 sed（已知字段列表，安全）。

### 预估

| 步骤 | 工作量 |
|------|--------|
| 写 AST 脚本 | 100-150 行 Go，15 min |
| 运行脚本 | 1 秒 |
| goimports | 1 秒 |
| 编译器修复（3-4 轮） | 30 min |
| seed 常量化 | 15 min |
| 前端 | 5 min |
| **总计** | **~65 min** |

---

## 前端

- `companyId: number` → `companyId: string`
- `modelId: number` → `id: string`（字段重命名）
- `modelWhitelist: number[]` → `modelWhitelist: string[]`
- `allowedModelIds: number[]` → `allowedModelIds: string[]`
- `defaultModelId: number | null` → `defaultModelId: string | null`
- `requestedModels: number[]` → `requestedModels: string[]`
- 其他 ID 字段已经是 `string`，值格式从 `"plk-xxx"` 变为 UUID 格式
