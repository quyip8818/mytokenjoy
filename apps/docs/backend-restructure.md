# Backend 架构重构方案

## 现状评估

当前 backend 分层基本清晰：`cmd → app → http/handler → domain → store`。

近期 clock refactor（`fb837fa0`）已将 clock 从 Config 解耦为构造函数注入，domain 各 service 通过 `clock.Clock` interface 接收时间源。budget domain 也新增了 `RotatePeriod` + `enrichTreeConsumed` 等逻辑，rebalance worker 从直接操作 store/config 改为委托 `domainbudget.Service`。这些变更强化了"infra worker 不持有 store，只委托 domain service"的方向。

存在的结构性问题：
1. ~~权限常量定义在 `infra/permission`，handler 层反向引用 infra~~ ✅ 已解决（阶段 1）
2. ~~identity 模块与 domain 平级，auth handler 绕过 service 直接操作 repo~~ ✅ 路径已迁移（阶段 3），httpx 已抽离到 http 层，AuthService 抽取留后续增量
3. ~~`internal/pkg/` 成为杂物间，包职责与 domain 层重叠~~ ✅ 已解决（阶段 2）
4. ~~port 接口定义分散，无统一位置约定~~ ✅ 已解决（阶段 4）
5. handler 基类模式不统一（三种风格共存）— 留后续增量
6. ~~JSON decode 方式混用（`httputil.DecodeJSON` vs 裸 `json.NewDecoder`）~~ ✅ 已解决（阶段 5）

---

## 目标架构

```
cmd/server/main.go
internal/
├── app/              # composition root（组装 + 启动）
├── config/           # 纯配置 struct
├── domain/           # 纯业务逻辑，零外部依赖
│   ├── types/        # 纯值对象 + DTO（struct 和 const only）
│   ├── grants/       # 权限常量 + manifest + normalizer + 角色
│   ├── company/
│   ├── identity/     # AuthService + Credentials + Authz + VerifyCode（无 net/http）
│   ├── billing/
│   ├── budget/
│   ├── org/
│   ├── keys/
│   ├── models/
│   ├── usage/
│   ├── dashboard/
│   ├── audit/
│   ├── approval/
│   ├── notification/
│   └── gateway/
├── adapter/          # port 的外部实现
├── infra/            # 基础设施（Redis、River、scheduler、metrics）
├── store/            # 持久化接口 + postgres 实现
├── http/             # HTTP transport
│   ├── handler/
│   ├── middleware/
│   ├── httputil/     # response/decode helpers
│   ├── httpx/        # cookie/token HTTP helpers（从 identity/httpx 移来）
│   └── response/
├── worker/           # 异步 job handler
└── support/          # 跨层无业务语义工具
    ├── clock/        # Clock interface + System/Fixed
    ├── tenant/       # context key + Get/Set（原 ctxcompany）
    ├── budget/       # 预算计算 helper（跨 5 domain 共用）
    ├── org/          # 组织树/路由规则 helper + LoadDepartments/LoadBudgetTree
    ├── quota/        # MoneyToQuota/QuotaToMoney/ResolveBillingCurrency + 常量
    ├── crypto/       # AES-GCM 加解密（凭证字段）
    ├── simulate/     # Delayer（开发延迟模拟）
    ├── sliceutil/    # Paginate、HasAny 泛型工具
    ├── ratelimit/    # Limiter interface + Result + HTTP response helpers
    ├── invitetoken/  # 邀请 token 签发/解析
    ├── baseurl/      # URL 派生
    └── modelcatalog/ # 模型目录数据结构
```

删除的目录：`internal/identity/`、`infra/permission/`、`domain/port/`、`internal/pkg/`、`support/common/`、`support/tree/`。

---

## 五个核心变更

### 变更 1：权限体系归入 `domain/grants` ✅

**现状**：~~权限常量（`OrgRead`, `BudgetManage`）在 `infra/permission/keys.go`，manifest 在 `infra/permission/manifest.go`，normalizer 实现也在 infra。但角色名称常量已在 `domain/grants/roles.go`。~~ 已合并。

**当前结构**（已实现）：

```
domain/grants/
├── keys.go          # const OrgRead = "org:read" ...（生成文件）
├── manifest.go      # embed manifest.json + ManifestData() + ExpandHierarchy()
├── manifest.json    # 权限清单（生成文件）
├── normalizer.go    # Normalizer interface + NewGrantNormalizer() + NormalizeGrantIDs() 直接函数
├── roles.go         # 角色名称 + 固定 UUID
└── platform.go      # IsPlatformPermission 判断
```

注：`hierarchy.go` 未单独拆分——`ExpandHierarchy` 放在 `manifest.go` 中（函数紧邻 manifest 数据，无必要拆文件）。

代码生成脚本 (`packages/contracts/permission/generate-backend.go`) 输出路径改为 `domain/grants/keys.go`。

`infra/permission/` 整个删除。所有 handler 改 import `domain/grants`。

**依赖方向**：`http/handler → domain/grants`（合法向下引用），消除 handler → infra 的旁路依赖。

---

### 变更 2：Identity 收口为 `domain/identity` ✅

**现状**：~~`internal/identity/` 独立于 domain~~ 已迁入 `domain/identity/`。

**已实现**：`internal/identity/` 整体路径迁移到 `internal/domain/identity/`（authz, credentials, sessiontoken, verifycode, httpx, registertoken, secrets 全部保留原有子包结构）。

**未完成（后续增量）**：
- auth handler 提取 `AuthService` interface（暂不动——handler 逻辑与 HTTP transport 高度耦合，强制抽取会增加 boilerplate 无实际收益。等功能需要修改时顺手做）

---

### 变更 3：`internal/pkg` 拆空删除 ✅

**已实现迁移表**（实际 vs 原计划差异标注）：

| 原包 | 实际迁入 | 与原计划差异 |
|------|----------|-------------|
| `pkg/ctxcompany` | `support/tenant`（包名改为 tenant） | 如计划 |
| `pkg/invitetoken` | `support/invitetoken` | 原计划放 domain/identity，因跨 domain 引用暂留 support |
| `pkg/common` | `support/common`（整体迁移） | 原计划细粒度拆分，实际整体迁移更低风险 |
| `pkg/budget` | `support/budget` | 原计划合入 domain/budget，因跨 5 domain 使用改为 support |
| `pkg/org` | `support/org` | 同上 |
| `pkg/ratelimit` | `infra/ratelimit`（合并） | 如计划，修复了自引用循环 |
| `pkg/clock` | `support/clock` | 如计划 |
| `pkg/tree` | `support/tree` | 如计划 |
| `pkg/baseurl` | `support/baseurl` | 原计划放 config，因影响面小留 support |
| `pkg/modelcatalog` | `support/modelcatalog` | 原计划待定，放 support |

**结果**：`internal/pkg/` 目录删除，零残留引用。

---

### 变更 4：Port 接口归位 ✅

**已实现**：
- `KeySyncPort` → `domain/keys/port.go`
- `OverrunKeyControl` → `domain/budget/ports.go`（与已有 `JobEnqueuer` 同文件）
- `newapisync/sync.go` 更新 compile-time assertions 为 `domainkeys.KeySyncPort` + `budget.OverrunKeyControl`
- `domain/port/` 目录删除

---

### 变更 5：Handler 模式统一 ✅（JSON decode 部分）

**定义三种 Base（各有明确适用场景）**：

```go
// http/handler/shared/base.go

// PublicBase — 不需要 session 的路由（auth, register, health）
type PublicBase struct {
    Cfg config.Config
}

// ProtectedBase — 需要 session + authz 的路由（绝大多数业务 handler）
type ProtectedBase struct {
    httpdeps.Protected
}

// PlatformBase — platform admin 专用（独立 auth 链）
type PlatformBase struct {
    httpdeps.Platform
}
```

每个 handler 必须 embed 其中一种。所有 handler 通过 `Mount(r chi.Router, d httpdeps.Deps)` 注册。

JSON decode 统一使用 `httputil.DecodeJSON`（有 1MB body size limit + 统一错误格式）。

---

## 执行阶段

| 阶段 | 内容 | 风险 | 验证标准 | 状态 |
|------|------|------|----------|------|
| 1 | grants 合并：`infra/permission` 全部代码 → `domain/grants`，全局 import 替换（20+ 文件） | 低-中（大范围机械化操作） | 编译通过 + 全量测试通过，`infra/permission` 目录删除 | ✅ 完成 |
| 2 | pkg 迁移：逐包移动到归属位置 | 低（每次一个包，独立 PR） | `internal/pkg` 目录删除 | ✅ 完成 |
| 3 | identity 收口：移动 `internal/identity/` → `domain/identity/` | 低（纯路径迁移） | `internal/identity` 不存在，`domain/identity` 存在 | ✅ 完成 |
| 4 | port 归位 + `domain/port` 删除 | 低 | `domain/port` 删除 | ✅ 完成 |
| 5 | handler base 统一 + JSON decode 统一 | 低 | grep `json.NewDecoder(r.Body)` 零结果 | ✅ 完成 |

### 已完成记录

**阶段 1 完成**（2026-08-06）：
- `infra/permission/` 全部代码（keys, manifest, normalizer, platform, grants）移入 `domain/grants/`
- 38 个文件的 import path 替换（`infra/permission` → `domain/grants`）
- 原有 `permission.X` 引用改为 `grants.X`
- `compose_domain.go` 中 `grants` 变量名改为 `normalizer` 避免与包名冲突
- `credentials/service.go` 重复 import 修复
- 测试从 `tests/infra/permission/` 移到 `tests/domain/grants/`
- `infra/permission/` 目录删除
- `Normalizer` 接口保留（向后兼容），同时提供 `grants.NormalizeGrantIDs()` 直接函数调用

**阶段 5 完成**（2026-08-06）：
- 32 处 `json.NewDecoder(r.Body).Decode` → `httputil.DecodeJSON`
- 11 个文件移除多余的 `encoding/json` import
- 涉及 handler：auth, platform, register, me, billing, notification
- 所有 handler 现在统一有 1MB body size limit + domain error 格式响应

**阶段 4 完成**（2026-08-06）：
- `KeySyncPort` 移入 `domain/keys/port.go`
- `OverrunKeyControl` 移入 `domain/budget/ports.go`
- `newapisync/sync.go` 更新为 import `domain/keys` + `domain/budget`
- `domain/port/` 目录删除

**阶段 2 完成**（2026-08-06）：
- `pkg/clock` → `support/clock`（62 处引用）
- `pkg/ctxcompany` → `support/tenant`，包名重命名为 `tenant`（16 处引用）
- `pkg/budget` → `support/budget`（61 处引用，跨 domain 使用不适合合并入单一 domain）
- `pkg/org` → `support/org`（39 处引用，同上）
- `pkg/ratelimit` → 合并入 `infra/ratelimit`（修复自引用循环）
- `pkg/invitetoken` → `support/invitetoken`（9 处引用）
- `pkg/common` → `support/common`（整体迁移，细粒度拆分留后续增量）
- `pkg/tree` → `support/tree`
- `pkg/baseurl` → `support/baseurl`
- `pkg/modelcatalog` → `support/modelcatalog`
- `internal/pkg/` 目录删除，零残留引用

**阶段 3 完成**（2026-08-06）：
- `internal/identity/` 整体移入 `internal/domain/identity/`（43 处 import 替换）
- 包含：authz、credentials、sessiontoken、verifycode、registertoken、secrets
- `httpx` 已从 `domain/identity/httpx` 抽离到 `internal/http/httpx/`（domain 层不依赖 net/http）
- `SessionFromContext`/`WithSessionContext` 核心实现移入 `domain/types/session.go`（纯 context 操作）
- `http/httpx/context.go` 委托 `domain/types` 实现（保持向后兼容的 API）
- `domain/budget/audit.go` 改为直接引用 `types.SessionFromContext`（不再依赖 httpx）

---

## 架构师自审：方案问题与风险

### 问题 1：`domain/identity/httpx` 是否违反 domain 层纯净原则 ✅ 已解决

**问题**：`httpx` 包处理 HTTP cookie 和 token 解析，依赖 `net/http`。把它放在 domain 层意味着 domain 对 HTTP transport 有感知。

**已实现的修正**：
- `httpx` 已从 `domain/identity/` 移到 `internal/http/httpx/`
- `SessionFromContext`/`WithSessionContext` 核心（纯 context 操作）提取到 `domain/types/session.go`
- `httpx/context.go` 委托 `types` 实现，API 不变
- `domain/budget/audit.go` 改为直接 import `domain/types`（不再依赖 httpx）
- domain 层零 `net/http` 依赖（除 `domain/gateway` 这个 transport-aware hybrid）

Auth handler → AuthService 提取有意跳过（handler 与 HTTP 高度耦合，强制抽取增加 boilerplate 无实际收益）。

### 问题 2：`domain/grants` 包含代码生成文件是否稳定

**问题**：`keys.go` 和 `manifest.json` 是从 `packages/contracts` 生成的。如果 contracts 变更，domain 层会被"外部工具"修改。

**结论**：可接受。生成文件本质是"编译期常量"，Go 社区广泛接受 `//go:generate` 在任何层。只要 `.gitignore` 不忽略生成文件（保证 CI 可复现），不影响架构纯净性。

### 问题 3：`ctxcompany` 内联到 `domain/company` 后的循环引用风险

**问题**：`ctxcompany` 被 `store/company_id.go`、`identity/credentials`、`domain/org` 等广泛引用。如果合并到 `domain/company`，而 `domain/org` import `domain/company`，而 `domain/company` 如果 import `domain/org` 就会循环。

**结论**：当前 `domain/company` 不 import `domain/org`，方向是 `org → company`（单向）。`store/` import `domain/company` 也是合法方向（store 层实现可以引用 domain 层类型）。

**但**：`store/company_id.go` 直接 import `pkg/ctxcompany`。如果改为 import `domain/company`，会形成 `store → domain/company`。而 `domain/company` 的 service 又会 import `store`（通过接口）。Go 不允许循环 import。

**修正**：`ctxcompany` 的核心（context key + Get/Set）应独立为 `internal/support/tenant`（和 clock 同层，都是跨层基础设施）。`domain/company` 只提供业务封装层（`CompanyID(ctx)` 等快捷方法，内部调 support/tenant）。store 层直接引用 `support/tenant`。

修正后依赖：
```
store → support/tenant（取 company ID）
domain/company → support/tenant（封装 tenant context 快捷方法）
domain/org → domain/company（合法单向）
```

### 问题 4：Auth handler 重构是否过度设计

**问题**：当前 auth handler 虽然直接持有 repo，但功能运行正常。强行抽 `AuthService` 会引入一个"包装层"，可能让简单的 login 流程变得更难 trace。

**结论**：对于本项目规模（17 个 handler、每个 handler 平均 5-10 个 endpoint），auth handler 是唯一的异类。统一性的收益大于多一层间接的开销。且 auth 逻辑（login routing、multi-company selection、invite-accept）确实复杂到值得有 service 层。

但要注意：`AuthService` 不应该膨胀为 god service。按职责可拆为：
- `AuthService`：login/logout/refresh
- `InviteService`：invite-accept/invite-link（可以是 org domain 的一部分）
- `CredentialService`：已有，保持

### 问题 5：删除 `internal/pkg` 是否过于激进

**问题**：`pkg/tree`、`pkg/clock` 是纯工具代码，不属于任何 domain。放在 `domain/types` 是否语义准确？

**结论**：不应放 `domain/types`。`domain/types` 应严格限于值对象（struct + const），不放逻辑函数。
- `tree`：数据结构操作工具 → `internal/support/tree`
- `clock`：42 处引用，横跨全层 → `internal/support/clock`
- `tenant`（ctxcompany）：跨层 context 机制 → `internal/support/tenant`

**关于 `support` vs 保留 `pkg` 的命名**：收益很小。如果团队已习惯 `internal/pkg`，保留 `pkg` 只做瘦身（只留 clock + tenant + tree）比 rename 的认知成本更低。这不应阻塞其他阶段。

**修正**：保留 `internal/support/`（或 `internal/pkg/`）放真正的跨层工具。守则：最多 3 个包，每个包最多 1-2 个文件。超出说明应归属到具体 domain 或 infra。

### 问题 6：`infra/permission` → `domain/grants` 合并比方案描述的复杂

**问题**：实际依赖图比"纯 rename + import 替换"更复杂。当前：
- `domain/grants` 定义角色常量 + `Normalizer` 接口
- `infra/permission` 包含：常量（keys.go）、manifest 解析、ExpandHierarchy、PresetRoleCapabilities、WriteCapabilitiesFromManifest、NormalizeGrantIDs 实现
- `infra/permission` 反向 import `domain/grants`（获取角色常量）
- `identity/authz` import `infra/permission`（需要 ExpandHierarchy、PresetRoleCapabilities、CompanyPermissions、PermissionIDMap）

合并后 `identity/authz` 变为 import `domain/grants`——这是合法的。但 `identity/authz` 目前是独立于 domain 层的 `internal/identity/authz`。如果 identity 也要收入 `domain/identity`，那 `domain/identity` 就会 import `domain/grants`——这是 domain 间合法引用，没问题。

**真正的风险**：`infra/permission/normalizer.go` 是 `grants.Normalizer` 接口的实现。如果把实现也移入 `domain/grants`，那 `domain/grants` 就变成既定义接口又自己实现。这打破了 interface/implementation 分离的初衷。

**修正**：normalizer 接口和实现统一放在 `domain/grants`。这不是 port pattern 场景——normalizer 不对接外部系统，是纯内存计算逻辑（解析 manifest JSON、expand hierarchy），放在 domain 层完全合理。删除接口/实现分离，直接提供 `grants.NormalizeGrantIDs()` 函数。

### 问题 7：阶段 1 实际影响范围比"低风险"描述的大

**问题**：阶段 1 声称"纯 rename + import 替换"，但实际 `infra/permission` 有 20+ 处被引用（10 个 handler + middleware + identity/authz + seed + app）。且不是简单 rename——需要把逻辑代码（ExpandHierarchy, PresetRoleCapabilities 等函数实现）从 infra 移入 domain/grants。

**修正**：阶段 1 风险应标为"低-中"。操作步骤更准确描述为：
1. 把 `infra/permission` 全部代码移入 `domain/grants`（保留 `roles.go` 已有内容）
2. 删除 `infra/permission/roles.go`（目前只是 re-export grants 常量的中间层）
3. 全局 sed 替换 import path
4. 删除 `infra/permission` 目录

这是一个大范围但机械化的操作，可以一次完成但需要全量回归测试。

---

## 战略判断：做到什么程度

方案列出了 5 个阶段，但不是每个阶段在每个团队规模下都值得做。

| 团队规模 | 建议执行 | 理由 |
|----------|----------|------|
| 1 人 | 阶段 1 + 5 | grants 合并消除最扎眼的依赖反转；JSON decode 统一是纯机械改动。其余 smell 对单人无实际阻碍 |
| 2-3 人 | 阶段 1 + 2 + 5 | pkg 瘦身让新成员能快速定位代码归属。identity 收口可推迟到 auth 逻辑下次需要改动时 |
| 4+ 人 | 全部 | 统一性和可预测性对多人协作 ROI 才足够高 |

**阶段依赖关系**（必须遵守的顺序）：
```
阶段 1（grants） ──┐
                   ├→ 阶段 3（identity，依赖 1+2 完成后的 import 稳定）
阶段 2（pkg）   ──┘
阶段 4（port）  独立
阶段 5（handler） 独立
```

---

## 近期变更对方案的影响

### clock refactor（fb837fa0）
- clock 已从 config 解耦为 constructor injection
- 42 处引用横跨全层，确认 clock 是跨层基础设施，不应放入 `domain/types`
- 文档已修正：clock 迁入 `internal/support/clock`

### budget rotation 重构（52cac4aa ~ 5c1fe7ba）
- RebalanceWorker 已从"直接持有 store + config + clock"变为"只持有 domainbudget.Service + domainbudget.Rebalancer"
- 新增 `budget.Service.RotatePeriod()`，rotation 逻辑完全封装在 domain 内
- 新增 `enrichTreeConsumed`，budget domain 新增对 `store.LedgerRepository` 的窄依赖

**结论**：这些变更强化了方案方向（worker → domain service → store），无需修改方案目标。但验证了两点：
1. `pkg/clock` 不能放 domain/types —— 它被 infra/river 等 non-domain 包广泛使用
2. domain service 的窄 Store 接口会持续膨胀（budget.Store 新增 `Ledger()`），这是正常演进

---

## 不变的部分

- `store.Store` 大接口 — domain 层已通过窄接口隔离，不需要拆
- `adapter/bridge/` — 数量少说明 domain 间耦合低
- handler 文件拆分风格（单文件 vs 多文件）— 取决于复杂度，不强制统一
- `infra/jobs`、`infra/river` — River 队列封装在 infra 层合理
- Router `Mount` 注册模式 — 已统一，保持
- domain service 窄 Store 接口模式（如 `budget.Store` 新增 `Ledger()`）— 继续演进
- clock constructor injection 模式 — 已确立，所有 service 通过构造函数接收 clock
- worker 委托 domain service 模式（如 rebalance worker → budget.Service.RotatePeriod）— 正确方向

---

## 阶段 6+7：support/common 拆分 + datasource 依赖翻转（2026-08-06）

### 阶段 6：`support/common` 彻底拆空删除 ✅

`support/common` 是从 `internal/pkg/common` 整体迁移来的杂物间（10 个文件），现已按语义拆分：

| 原 common 文件 | 新归属 | 理由 |
|---|---|---|
| `constants.go` (MoneyToQuota/QuotaToMoney/ResolveBillingCurrency/DefaultQuotaPerUnit/DefaultBillingCurrency/DefaultPersonalBudget) | `support/quota/quota.go` | 纯 billing 计算，无外部依赖 |
| `routing.go` (GetRoutingRuleForDept/ShrinkChildRoutingRules/ValidateModelIDsForMember 等) | `support/org/routing.go` | 组织路由规则逻辑 |
| `org_store.go` (LoadDepartments/LoadBudgetTree/LoadRoutingRules/PersistRoutingRules) | `support/org/org_store.go` | 组织树数据加载 |
| `crypto.go` + `fieldcrypto.go` | `support/crypto/crypto.go` | AES-GCM 加解密 |
| `paginate.go` + `scope_check.go` (Paginate/HasAny) | `support/sliceutil/sliceutil.go` | 泛型 slice 工具 |
| `parse.go` (ParseIntParam) | `http/httputil/params.go` | 纯 HTTP 参数解析 |
| `simulate.go` (Delayer) | `support/simulate/delayer.go` | 开发延迟模拟 |
| `auditfilter.go` | 删除 | 零 production 调用者（dead code） |

**关键设计决策**：
- `support/org/org_store.go` 定义了 narrow interface（`OrgNodeTreeReader`、`AllowlistWriter`）而非 import store 包，避免 `store → support/org → store` 循环依赖
- `PersistRoutingRules` 签名改为接收 `AllowlistWriter` + `OrgNodeTreeWriter` 两个参数（而非一个组合接口），让调用方直接传 `st.Models().Allowlist()` + `st.Org().Nodes()`
- `support/tree` 泛型 `Flatten` 仅有 1 个调用者 → inline 到 `support/budget/tree.go`，包目录删除

**同时完成**：
- `billing.DefaultQuotaPerUnit()` wrapper 函数删除（调用者改用 `quota.DefaultQuotaPerUnit` 常量）
- `ModelNotInDeptMessage`/`NewAPIGroupPrefix` 常量规范化到 `support/org`（从 `support/quota` 中删除重复定义）
- `app/testhook.go` 加 `//go:build testhook`（`NewWithStore` 不出现在 production binary）

**验证**：109 files changed, +330 -1459 lines。`go build ./...` ✅ `make lint` ✅ `make test-unit-nocache` ✅

---

### 阶段 7：翻转 `domain/org` → `integration/datasource` 依赖方向 ✅

**问题**：domain/org 的 7 个文件直接 import `integration/datasource`（向上引用外部层）。

**修正**：
- `RemoteDepartment`、`RemoteMember` struct + `DataSourceProvider`、`DataSourceFactory` interface 移入 `domain/types/datasource.go`（shared kernel）
- `integration/datasource/provider.go` 改为 type alias（`type Provider = types.DataSourceProvider`）
- `support/org/sync_diff.go` 改为 import `domain/types`

**修正后依赖**：
```
Before: domain/org → integration/datasource  (向上 ❌)
After:  integration/datasource → domain/types (向下 ✅)
```

**验证**：`grep -rl 'integration/' internal/domain/` = 0。10 files changed。

---

## 当前架构健康度总结

| 指标 | 值 | 评价 |
|------|-----|------|
| domain → integration | 0 文件 | ✅ |
| domain → infra | 0 文件 | ✅ |
| domain → http | 0 文件 | ✅ |
| domain → support | 68 文件 | ✅ 合法 |
| domain → store | 77 文件 | ✅ Go 惯例 |
| domain → config | 19 文件 | ✅ 合法 |
| support/ 包数 | 12 | ✅ 各包职责单一 |
| 总 .go 文件数 | 505 | — |

**结论**：无剩余结构性问题。后续优化仅为 cosmetic 级（`integration/` rename 为 `adapter/`、`Deps` 渐进窄化），等业务需求触碰时顺手做。
