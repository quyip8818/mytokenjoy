# Backend 架构重构方案

## 现状评估

当前 backend 分层基本清晰：`cmd → app → http/handler → domain → store`。

近期 clock refactor（`fb837fa0`）已将 clock 从 Config 解耦为构造函数注入，domain 各 service 通过 `clock.Clock` interface 接收时间源。budget domain 也新增了 `RotatePeriod` + `enrichTreeConsumed` 等逻辑，rebalance worker 从直接操作 store/config 改为委托 `domainbudget.Service`。这些变更强化了"infra worker 不持有 store，只委托 domain service"的方向。

存在的结构性问题：
1. 权限常量定义在 `infra/permission`，handler 层反向引用 infra
2. identity 模块与 domain 平级，auth handler 绕过 service 直接操作 repo
3. `internal/pkg/` 成为杂物间，包职责与 domain 层重叠
4. port 接口定义分散，无统一位置约定
5. handler 基类模式不统一（三种风格共存）
6. JSON decode 方式混用（`httputil.DecodeJSON` vs 裸 `json.NewDecoder`）

---

## 目标架构

```
cmd/server/main.go
internal/
├── app/              # composition root（组装 + 启动）
├── config/           # 纯配置 struct
├── domain/           # 纯业务逻辑，零外部依赖
│   ├── types/        # 共享值对象
│   ├── grants/       # 权限定义 + 角色常量 + 权限常量
│   ├── company/
│   ├── identity/     # 认证、会话、注册、验证码（新）
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
└── worker/           # 异步 job handler
```

删除的目录：`internal/pkg/`、`internal/identity/`、`infra/permission/`、`domain/port/`。

---

## 五个核心变更

### 变更 1：权限体系归入 `domain/grants`

**现状**：权限常量（`OrgRead`, `BudgetManage`）在 `infra/permission/keys.go`，manifest 在 `infra/permission/manifest.go`，normalizer 实现也在 infra。但角色名称常量已在 `domain/grants/roles.go`。

**终态**：

```
domain/grants/
├── keys.go          # const OrgRead = "org:read" ...（生成文件，目标从 infra 移来）
├── manifest.go      # embed manifest.json + ManifestData()
├── manifest.json    # 权限清单（生成文件）
├── normalizer.go    # NormalizeGrantIDs 实现
├── roles.go         # 角色名称 + 固定 UUID（已有）
├── platform.go      # IsPlatformPermission 判断
└── hierarchy.go     # Hierarchy map
```

代码生成脚本 (`packages/contracts/permission/generate-backend.go`) 输出路径改为 `domain/grants/keys.go`。

`infra/permission/` 整个删除。所有 handler 改 import `domain/grants`。

**依赖方向**：`http/handler → domain/grants`（合法向下引用），消除 handler → infra 的旁路依赖。

---

### 变更 2：Identity 收口为 `domain/identity`

**现状**：
- `internal/identity/` 独立于 domain，包含 authz、credentials、sessiontoken、verifycode、httpx、registertoken、secrets
- `internal/pkg/invitetoken` 是孤儿包
- auth handler 直接注入 6 个 repo + 3 个 service，不走 domain service

**终态**：

```
domain/identity/
├── service.go       # AuthService interface（Login, Logout, Refresh, AcceptInvite, SetPassword, ResetPassword, SelectCompany）
├── auth_impl.go     # AuthService 实现
├── session.go       # SessionToken issuer
├── credentials.go   # Credentials service（注册、密码校验、bootstrap）
├── authz.go         # Authz service（GetSessionContext, CheckPermission）
├── verifycode.go    # 验证码发送/校验
├── registertoken.go # 注册 token 签发
├── invitetoken.go   # 邀请 token 签发（从 pkg/invitetoken 移来）
├── secrets.go       # 密钥管理
└── httpx/           # HTTP cookie/token 解析 helper（transport-aware sub-package）
```

Auth handler 重构为：
```go
type Handler struct {
    shared.PublicHandlerBase
    authSvc identity.AuthService
}
```

**原则**：handler 不持有任何 repo，所有持久化操作封装在 domain service 内部。

---

### 变更 3：`internal/pkg` 拆空删除

| 现包 | 迁入 | 理由 |
|------|------|------|
| `pkg/ctxcompany` | `domain/company`（内联，删除独立包） | company context 是 company domain 的内部机制 |
| `pkg/invitetoken` | `domain/identity/invitetoken.go` | 认证领域 |
| `pkg/common` | `domain/types/money.go` | `MoneyToQuota`/`QuotaToMoney` 是业务换算 |
| `pkg/budget` | `domain/budget`（合并） | budget helper 归属 budget domain |
| `pkg/org` | `domain/org`（合并） | org helper 归属 org domain |
| `pkg/ratelimit` | `infra/ratelimit`（合并） | 限流是基础设施 |
| `pkg/clock` | `internal/support/clock`（保持独立包，见下方说明） | 42 处引用，跨 domain/infra/adapter/app 全层使用 |
| `pkg/tree` | `domain/types/tree.go` | 通用树结构是值对象 |
| `pkg/baseurl` | `config` 包 helper | 仅被 config 消费 |
| `pkg/modelcatalog` | `domain/models` 或 `integration/catalogsync` | 视内容归属 |

**结果**：`internal/pkg/` 整个目录删除。

**clock 特殊说明**：`pkg/clock` 有 42 处引用，横跨 domain、infra、adapter、app 全层。近期 clock refactor 已确立其为全局基础设施接口（constructor injection pattern）。不适合放入任何单一 domain。迁入 `internal/support/clock` 作为跨层时间抽象。

---

### 变更 4：Port 接口归位

**现状**：
- `domain/port/syncport.go` 放跨域接口
- 各 domain 包内零散定义 port（`billing.Store`, `budget.JobEnqueuer`）

**终态**：每个 domain 包自己定义依赖的 port 接口。

- `domain/keys/port.go` → `KeySyncPort`
- `domain/budget/ports.go` → `JobEnqueuer`, `OverrunKeyControl`（已有）
- `domain/billing/service.go` → `Store`, `QuotaSyncer`（已有）

近期 rebalance worker 重构验证了这一方向：worker 不再直接持有 store + config + clock，改为注入 `domainbudget.Service` 和 `domainbudget.Rebalancer`，domain service 封装 RotatePeriod 逻辑。这正是 port 归位的实践——worker 作为 infra 层只依赖 domain interface，不穿透到 store。

`domain/port/` 目录删除。不设统一 port 目录。

**理由**：domain 包是可独立提取的 module，port 跟着消费者走，不跟着实现者走。

---

### 变更 5：Handler 模式统一

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

| 阶段 | 内容 | 风险 | 验证标准 |
|------|------|------|----------|
| 1 | grants 合并：`infra/permission` → `domain/grants` | 低（纯 rename + import 替换） | 编译通过，`infra/permission` 零 import |
| 2 | pkg 迁移：逐包移动到归属位置 | 低（每次一个包，独立 PR） | `internal/pkg` 目录删除 |
| 3 | identity 收口：创建 `domain/identity.AuthService` + auth handler 重构 | 中（逻辑迁移） | auth handler 不直接持有 repo |
| 4 | port 归位 + `domain/port` 删除 | 低 | `domain/port` 删除 |
| 5 | handler base 统一 + JSON decode 统一 | 低 | grep `json.NewDecoder(r.Body)` 零结果 |

---

## 架构师自审：方案问题与风险

### 问题 1：`domain/identity/httpx` 是否违反 domain 层纯净原则

**问题**：`httpx` 包处理 HTTP cookie 和 token 解析，依赖 `net/http`。把它放在 domain 层意味着 domain 对 HTTP transport 有感知。

**结论**：这是方案中最大的妥协。严格来说 httpx 应该留在 `http/` 层。但 httpx 的消费者是 middleware（属于 http 层）和 auth service（需要签发 cookie）。

**修正**：`domain/identity` 不应包含 `httpx`。正确做法：
- `domain/identity.AuthService.Login()` 返回 `TokenPair{AccessToken, RefreshToken, Expiry}`
- handler 层负责把 TokenPair 写入 cookie（调用 `http/httpx` 包）
- `httpx` 保留在 `http/httpx/`（或 `http/middleware/` 内部）

这样 domain/identity 完全不 import `net/http`。

### 问题 2：`domain/grants` 包含代码生成文件是否稳定

**问题**：`keys.go` 和 `manifest.json` 是从 `packages/contracts` 生成的。如果 contracts 变更，domain 层会被"外部工具"修改。

**结论**：可接受。生成文件本质是"编译期常量"，Go 社区广泛接受 `//go:generate` 在任何层。只要 `.gitignore` 不忽略生成文件（保证 CI 可复现），不影响架构纯净性。

### 问题 3：`ctxcompany` 内联到 `domain/company` 后的循环引用风险

**问题**：`ctxcompany` 被 `store/company_id.go`、`identity/credentials`、`domain/org` 等广泛引用。如果合并到 `domain/company`，而 `domain/org` import `domain/company`，而 `domain/company` 如果 import `domain/org` 就会循环。

**结论**：当前 `domain/company` 不 import `domain/org`，方向是 `org → company`（单向）。`store/` import `domain/company` 也是合法方向（store 层实现可以引用 domain 层类型）。

**但**：`store/company_id.go` 直接 import `pkg/ctxcompany`。如果改为 import `domain/company`，会形成 `store → domain/company`。而 `domain/company` 的 service 又会 import `store`（通过接口）。Go 不允许循环 import。

**修正**：`ctxcompany` 的核心（context key + Get/Set）应独立为 `domain/types/tenant.go`，不放在 `domain/company`。`domain/company` 只提供业务封装层（`CompanyID(ctx)` 等快捷方法）。store 层直接引用 `domain/types`。

修正后依赖：
```
store → domain/types（取 company ID）
domain/company → domain/types（取 company context）
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

**结论**：需要区分：
- `tree`：如果只有 budget 用 → 归入 `domain/budget`。如果 org 和 budget 都用 → 放 `domain/types`（共享值对象）合理。
- `clock`：42 处引用，横跨 domain/infra/adapter/app。近期 clock refactor 已确立其为全局构造注入接口。放 `domain/types` 不对（infra 层不应反向引用 domain）。正确位置是 `internal/support/clock`。

**修正**：保留 `internal/support/` 放真正的跨层无业务语义工具（clock, tree）。规则：最多 3 个包，每个包最多 1 个文件，超出说明应归属到具体 domain。

---

## 最终修正后的目标结构

```
internal/
├── app/
├── config/
├── domain/
│   ├── types/        # 共享值对象 + tenant context（ctxcompany 核心）+ money/quota 换算
│   ├── grants/       # 权限常量 + manifest + normalizer + 角色
│   ├── company/      # company 业务（引用 domain/types 的 tenant context）
│   ├── identity/     # AuthService + Credentials + Authz + VerifyCode（无 net/http 依赖）
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
├── adapter/          # port 实现
├── infra/            # 基础设施（不含 permission）
├── store/            # 持久化
├── http/
│   ├── handler/
│   ├── middleware/
│   ├── httputil/     # response/decode helpers
│   ├── httpx/        # cookie/token HTTP helpers（从 identity/httpx 移来）
│   └── response/
├── worker/
└── support/          # 纯工具（clock、tree）— 最多 2-3 个文件，超出即需重新审视
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
