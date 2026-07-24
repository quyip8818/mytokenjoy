# Backend 架构优化 — 长期目标架构

## 设计原则

1. **依赖方向单向**：`cmd → app → domain ← adapter/infra`。domain 不 import adapter、infra、integration、identity。
2. **Domain 自治**：每个 domain 只暴露 interface + value types，自己声明所需的 port interface。
3. **最少接触点**：新增一个完整 domain（包括 HTTP API）只碰 4 个文件（其中 2 个只加一行）。
4. **编译期安全**：保持编译时能发现接线遗漏。
5. **可测试性**：domain service 可以零基础设施单元测试。

---

## 现状架构

```
cmd/server/main.go
internal/
├── app/              — 依赖接线（compose_infra / compose_domain_wire / compose_worker / registry）
├── config/           — 全局环境配置（60+ 字段，embedded sub-config）
├── domain/           — 16 个子域
│   ├── types/        — 共享值类型 (shared kernel)
│   ├── grants/       — 权限模型
│   └── {domain}/     — 各业务域
├── http/
│   ├── deps/         — Deps(27 字段) + Protected struct
│   ├── handler/      — 16 个 handler 包 + Registry 中心注册
│   ├── middleware/   — 17 个中间件
│   └── router.go
├── identity/         — authz / credentials / sessiontoken / registertoken / verifycode
├── store/            — Repository 接口(25+) + postgres/ 实现(50+文件)
├── adapter/          — 跨域桥接 + Enqueuer 适配（混在一起）
├── infra/            — river / jobs / scheduler / budgetcheck / notification / ratelimit
├── integration/      — newapi / datasource / platform
├── worker/           — pricingsync
└── pkg/              — 15 个碎片工具包
```

---

## 问题清单（按严重性排序）

### P0: 依赖方向违规

| 位置 | 违规 | 后果 |
|------|------|------|
| `identity/authz` → `domain/billing` | 身份层反向依赖业务域 | authz 无法独立测试；billing 改动可能破坏登录 |
| `domain/keys` → `domain/newapisync` | 域间直接依赖实现包（获取接口定义） | newapisync 改动编译传导到 keys |
| `domain/budget` → `domain/newapisync` | 同上 | overrun 逻辑耦合同步实现包 |

**根因**：`newapisync` 包既定义接口又包含实现，消费方被迫 import 整个包只为拿接口类型。`billing.ResolveCompanyChargeRate` 是一个纯查询函数，被 authz 直接调用而非通过 port。

### P1: God Object — Deps / ServiceRegistry / domainServices 三重冗余

新增 domain 要改 5 个文件：`Deps` → `ServiceRegistry` → `domainServices` → `compose_domain_wire.go` → `handler/register.go`。

### P2: adapter/ 混合职责

Enqueuer 适配（domain port → River job args）和跨域编排（usage → budget delta 计算）共用一个目录，新人无法区分。

### P3: Config 泄漏

所有 domain 都 `import config`，接收完整 `config.Config`。测试构造成本高，改一个 env var 重新编译所有 domain。

### P4: pkg/ 碎片化

15 个包，多个只有单文件。`ctxcompany`、`baseurl`、`id`、`timeutil` 等粒度过细。

### P5: Store Tx 签名泄漏全量接口

`domain.Store.WithTx(ctx, func(store.Store) error)` — narrow store 定义事务但回调暴露全量 Store。

---

## 目标架构

```
cmd/server/main.go              — 启动编排
internal/
├── app/
│   ├── app.go                  — App struct + lifecycle
│   ├── wire.go                 — 单文件：所有 domain service 构造
│   └── worker.go               — background workers 构造
├── config/                     — 全局 Config（不变）
├── domain/
│   ├── port/                   — ★ 新增：跨域共享 port interfaces
│   │   └── syncport.go         — KeySyncPort, OverrunKeyControl, ...
│   ├── types/                  — 值类型 (不变)
│   ├── grants/                 — 权限模型 (不变)
│   └── {domain}/
│       ├── service.go          — Service interface + constructor
│       ├── config.go           — domain-specific Config struct
│       ├── ports.go            — domain-private port interfaces (JobEnqueuer 等)
│       └── *.go                — 业务逻辑
├── http/
│   ├── deps.go                 — Deps struct (精简)
│   ├── router.go               — middleware + mount 调用
│   └── handler/
│       └── {domain}/
│           └── handler.go      — Handler + Mount() 函数
├── identity/                   — 不变，但移除 billing 依赖
├── store/                      — Repository 接口 + postgres/（不变）
├── adapter/
│   ├── enqueue/                — domain port → River job args
│   └── bridge/                 — 跨域操作适配（usage↔budget 等）
├── infra/                      — 不变
├── integration/
│   ├── newapi/                 — HTTP client + token store
│   ├── newapisync/             — ★ 从 domain/ 移入：实现 port/syncport 接口
│   ├── datasource/
│   └── platform/
├── worker/                     — 不变
└── pkg/
    ├── common/                 — 合并 baseurl, id, timeutil
    ├── ctxcompany/             — 保留（底层 context key，不可归入 domain）
    ├── clock/                  — 不变
    ├── budget/                 — 预算计算纯函数
    ├── org/                    — 组织树纯函数
    ├── modelcatalog/           — 模型目录
    ├── ratelimit/              — 限流算法
    └── tree/                   — 通用树操作
```

---

## 改动详解

### 1. 新增 `domain/port/` — 跨域共享接口注册表

**问题**：`domain/newapisync/lifecycle_iface.go` 里定义了 `KeysNewAPISync`、`OverrunKeyControl` 等接口，导致 `keys` 和 `budget` 必须 import `newapisync` 包。

**方案**：把跨域消费的 port interface 集中到 `domain/port/`：

```go
// internal/domain/port/syncport.go
package port

import (
    "context"
    "github.com/google/uuid"
    "github.com/tokenjoy/backend/internal/domain/types"
)

// KeySyncPort — keys domain 消费此接口，newapisync 实现它
type KeySyncPort interface {
    Enabled() bool
    SyncPlatformKeyCreate(ctx context.Context, key types.PlatformKey, departmentID uuid.UUID) (string, error)
    SyncCreatePlatformKey(ctx context.Context, key types.PlatformKey, departmentID uuid.UUID) error
    TrySyncCreate(ctx context.Context, platformKeyID uuid.UUID) (string, error)
    RollbackFailedCreate(ctx context.Context, platformKeyID uuid.UUID)
    SyncUpdatePlatformKey(ctx context.Context, platformKeyID uuid.UUID, targetActive *bool) error
    SyncRevokePlatformKey(ctx context.Context, platformKeyID uuid.UUID) error
    SyncRotatePlatformKey(ctx context.Context, platformKeyID uuid.UUID) (string, error)
    DisablePlatformKey(ctx context.Context, platformKeyID uuid.UUID) error
    EnqueueUpsertProviderKey(ctx context.Context, providerKeyID uuid.UUID) error
    SyncUpsertProviderKey(ctx context.Context, providerKeyID uuid.UUID) error
}

// OverrunKeyControl — budget/overrun 消费此接口
type OverrunKeyControl interface {
    Enabled() bool
    DisablePlatformKey(ctx context.Context, platformKeyID uuid.UUID) error
}
```

**变更影响**：
- `domain/keys/service.go`：`newapisync.KeysNewAPISync` → `port.KeySyncPort`
- `domain/budget/overrun.go`：`newapisync.OverrunKeyControl` → `port.OverrunKeyControl`

**不变**：`domain/newapisync/lifecycle_iface.go` 可以保留（内部子包用），但外部消费方统一从 `port/` 拿接口。

**关于 billingport**：authz → billing 的修复不需要放在 `domain/port/`（见改动 2），用函数签名注入即可。`domain/port/` 只放多消费者的跨域接口。

### 2. 修复 authz → billing 依赖违规

当前 `authz.GetSessionContext()` 直接调用 `billing.ResolveCompanyChargeRate(ctx, s.store, companyID)`。这个函数只需要 `CurrencyStore`（= `Billing()` + `Company()`），authz 自己的 store 已满足此接口。问题纯粹是 **import 方向**——authz 包 import 了 billing 包。

**方案**：authz 声明自己需要的函数签名，wire 层注入实现。

```go
// identity/authz/service.go
type ChargeRateResolver func(ctx context.Context, companyID uuid.UUID) (currency string, quotaPerUnit int64, err error)

type service struct {
    store       Store
    chargeRate  ChargeRateResolver  // ← 新增
    cache       *LRUCache
    revCache    *revisionCache
}

func NewService(cfg config.Config, st Store, chargeRate ChargeRateResolver) Service {
    ...
}
```

wire 层注入：

```go
// app/compose_http.go
func wireIdentity(cfg config.Config, st store.Store) (...) {
    // billing.ResolveCompanyChargeRate 是包级函数，只需要 CurrencyStore
    // st 已满足 CurrencyStore 接口
    chargeRate := func(ctx context.Context, companyID uuid.UUID) (string, int64, error) {
        return billing.ResolveCompanyChargeRate(ctx, st, companyID)
    }
    authzSvc := authz.NewService(cfg, st, chargeRate)
    ...
}
```

**注意**：这里 `app/` 层 import billing 包完全合法（app 是组装层，允许 import 所有包）。关键是 `identity/authz` 不再 import `domain/billing`。

**替代方案**：如果不想用函数签名而想用 interface，可以把 `ChargeRateResolver` 放到 `domain/port/billingport.go`，billing service 实现它。两种方案都可以，函数签名更轻量（authz 不需要 import domain/port），interface 更正式。推荐函数签名——因为这是单方法、单消费者的场景。

### 3. newapisync 移入 `integration/`

**理由**：
- 它的核心是调外部 HTTP API + 管理 token 映射 — 这是 integration adapter 的定义
- 它有 8 个子包（platformkey, provider, provision, policy, ports, devapi, syncdeps, outbox）— 对 domain 来说太重了
- 移动后，`domain/` 通过 `domain/port/` 接口消费它，依赖方向正确

**移动路径**：`internal/domain/newapisync/` → `internal/integration/newapisync/`

**依赖方向验证**：
- `integration/newapisync` → `domain/company`（用 CompanyID context helpers）：✅ integration 允许 import domain
- `integration/newapisync` → `domain/types`：✅ 同上
- `integration/newapisync` → `domain/adminport`：✅ 同上
- `integration/newapisync` → `store`：✅ integration 允许 import store
- `domain/keys` → `integration/newapisync`：❌ 违规！— 但 Phase 1 已解决（keys 改为 import `domain/port`）

**前置条件**：必须先完成 Phase 1（提取 `domain/port/`），否则 `domain/keys` 和 `domain/budget` 还在直接 import newapisync。

**需要改的**：
- 所有 import path（~15 处，包括 app/, infra/river/, adapter/, tests/）
- newapisync 内部子包的互引路径（自动改 module prefix）
- tests/domain/newapisync/ 移到 tests/integration/newapisync/

**不需要改的**：domain 的业务逻辑代码本身（Phase 1 已经把 domain 的 import 切到 port/）。

### 4. Handler 自注册 — 消灭 Registry struct

每个 handler 包暴露 `Mount(chi.Router, deps.Deps)` 函数：

```go
// internal/http/handler/budget/handler.go
package budget

func Mount(r chi.Router, d deps.Deps) {
    h := &Handler{
        ProtectedHandlerBase: shared.NewProtectedHandlerBase(d.Protected()),
        service:              d.Budget,
    }
    r.Route("/budget", h.routes)
}

func (h *Handler) routes(r chi.Router) {
    r.Use(httpmiddleware.RequirePermission(h.AuthzSvc, permission.BudgetManage))
    r.Get("/tree", h.Tree)
    r.Put("/departments/{departmentId}", h.UpdateNode)
    // ...
}
```

路由注册收拢到一个文件：

```go
// internal/http/router.go
func mountAPI(api chi.Router, d deps.Deps) {
    budget.Mount(api, d)
    keys.Mount(api, d)
    models.Mount(api, d)
    org.Mount(api, d)
    billing.Mount(api, d)
    dashboard.Mount(api, d)
    audit.Mount(api, d)
    approval.Mount(api, d)
    me.Mount(api, d)
    notification.Mount(api, d)
    session.Mount(api, d)
    auth.Mount(api, d)
    register.Mount(api, d)
    ingest.Mount(api, d)
    if d.Config.SupportSaas {
        platform.Mount(api, d)
    }
    if d.Config.AllowsDevHTTPRoutes() {
        dev.Mount(api, d)
    }
}
```

**编译期安全**：如果忘记调用 `Mount`，该包的 import 就没用，`go vet` 或 IDE unused import 检测会提示。加一个 integration test 验证所有 `/api/*` prefix 返回非 404 更稳。

**删除**：`handler/register.go` 里的 `Registry` struct、`NewRegistry()`、`RegisterAPIRoutes()`。

### 5. Deps 精简 — 消除三层冗余

当前链路：`domainServices` → `ServiceRegistry.Deps` → `httpdeps.Deps`。三层说的是同一件事。

**但 ServiceRegistry 还持有 worker-only 的依赖**（`OrgSync`、`Overrun`、`Rebalance`、`Infra`），这些不属于 HTTP handler。

**目标**：拆成两个独立结构，各管各的：

```go
// internal/http/deps.go — 只给 HTTP handler 用
type Deps struct {
    Config         config.Config
    Logger         *slog.Logger

    // Identity
    AuthzSvc       authz.Service
    Credentials    credentials.Service
    SessionToken   sessiontoken.Issuer
    VerifyCodeSvc  *verifycode.Service

    // Infrastructure (handler 需要的)
    Store          store.Store
    RateLimiter    pkgrl.Limiter
    IngestEnqueuer jobs.Enqueuer
    IngestMetrics  ingestmetrics.Recorder

    // Domain services
    Org            domainorg.Service
    Budget         domainbudget.Service
    Keys           domainkeys.Service
    Models         domainmodels.Service
    Dashboard      domaindashboard.Service
    Audit          domainaudit.Service
    Billing        domainbilling.Service
    Company        domaincompany.Service
    MemberAnalytics domainmemberanalytics.Service
    Approval       *domainapproval.Engine
    ReadModel      domainusage.ReadModel
    IngestSvc      domainusage.Ingestor

    // Integration
    Gateway        domaingateway.GatewayService
    DevBearer      devapi.BearerResolver
    DevReadiness   devapi.ReadinessChecker
    Notification   *notification.Service
    CompanyGate    *domaincompany.Gate
}

func (d Deps) Protected() Protected { ... }
```

```go
// internal/app/app.go — worker 依赖单独持有
type App struct {
    Config  config.Config
    Store   store.Store
    Router  http.Handler
    Workers *backgroundWorkers
    closers []func()
}
```

`app/wire.go` 构造 `deps.Deps`（给 router）和 worker 依赖（Overrun、Rebalance、OrgSync）分别传递，不再用 `ServiceRegistry` 做中转。

**删除**：`domainServices` struct、`ServiceRegistry` struct、`buildServiceRegistry()` 函数。用一个 `buildDeps()` + 直接传参给 `buildBackgroundWorkers()` 替代。

### 6. Config 分片

每个 domain 声明自己的 Config struct：

```go
// domain/budget/config.go
package budget

type Config struct {
    OverrunEnabled   bool
    PeriodKind       string
    OverrunThreshold float64
}
```

wire 函数负责 slice：

```go
// app/wire.go
budgetCfg := budget.Config{
    OverrunEnabled:   cfg.OverrunEnabled,
    PeriodKind:       cfg.BudgetPeriodKind,
    OverrunThreshold: cfg.OverrunThreshold,
}
budgetSvc := budget.NewService(budgetCfg, st, delayer, budgetEnqueuer)
```

**收益**：
- Domain 包不再 `import "github.com/tokenjoy/backend/internal/config"`
- 测试只构造 domain-local Config
- 配置变更的编译范围缩小到相关 domain

### 7. adapter/ 按职责拆分

```
internal/adapter/
├── enqueue/
│   ├── budget.go          — budget.JobEnqueuer → River args
│   ├── org.go             — org.SyncEnqueuer → River args
│   ├── newapisync.go      — newapisync.SyncJobEnqueuer → River args
│   ├── usage.go           — usage.IngestEnqueuer → River args
│   └── dashboard.go       — dashboard.ProjectionEnqueuer → River args
└── bridge/
    ├── usage_budget.go    — usage.BudgetOps adapter (wraps budget pure functions)
    ├── usage_alert.go     — usage alert → notification adapter
    └── usage_lot.go       — lot consumer adapter
```

### 8. Narrow Tx 泛型包装

解决 `WithTx(ctx, func(store.Store) error)` 泄漏全量 Store 的问题：

```go
// internal/store/narrowtx.go
package store

import "context"

// TxFunc wraps WithTx so narrow Store interfaces don't leak the full Store.
// Usage: in domain's service constructor, inject a NarrowTxRunner.
type NarrowTxRunner[S any] func(ctx context.Context, fn func(S) error) error

func NewNarrowTxRunner[S any](st Store, narrow func(Store) S) NarrowTxRunner[S] {
    return func(ctx context.Context, fn func(S) error) error {
        return st.WithTx(ctx, func(txSt Store) error {
            return fn(narrow(txSt))
        })
    }
}
```

Domain 使用：

```go
// domain/billing/service.go
type Store interface {
    Billing() store.BillingRepository
    Company() store.CompanyRepository
}

type service struct {
    store Store
    runTx store.NarrowTxRunner[Store]
}
```

### 9. pkg/ 收拢

| 当前 | → 目标 | 理由 |
|------|--------|------|
| `pkg/baseurl` | `pkg/common` | 单函数 |
| `pkg/id` | `pkg/common` | UUID 工具函数 |
| `pkg/timeutil` | `pkg/clock` | 时间相关 |
| `pkg/companyids` | `pkg/common` | 公司 ID 解析工具 |
| `pkg/secrets` | `identity/secrets` | 密钥生成属于身份层 |
| `pkg/newapiunits` | `integration/newapi` | NewAPI 计量单位属于集成层 |
| `pkg/authzscope` | `identity/authz` | 授权范围属于 authz |

**不动**：
- `pkg/ctxcompany` — 被 store、middleware、identity 等底层包使用，不可归入 domain（会循环依赖）
- `pkg/budget` — 预算计算纯函数，被 domain/budget 和 adapter 共用
- `pkg/org` — 组织树纯函数
- `pkg/clock` — 时间抽象
- `pkg/common` — 通用工具
- `pkg/ratelimit` — 限流算法接口
- `pkg/tree` — 通用树操作
- `pkg/modelcatalog` — 模型目录数据

收拢后 pkg/ 从 15 包降到 **8 包**。

---

## 目标架构下新增 Domain 的步骤

以新增 `foo` domain 为例：

| # | 动作 | 文件 |
|---|------|------|
| 1 | 创建 domain 包 | `internal/domain/foo/` (service.go, config.go, ports.go, *.go) |
| 2 | 创建 handler 包 | `internal/http/handler/foo/handler.go` (含 `Mount()`) |
| 3 | 在 wire.go 加构造 | `internal/app/wire.go` 加 `wireFoo()` + 赋值到 Deps |
| 4 | 在 router.go 挂载 | `internal/http/router.go` 加 `foo.Mount(api, d)` |

**总计 4 个文件**（当前 6-7 个）。其中 3、4 各只加一行。

---

## 执行计划

分三个阶段，每阶段可独立提交、独立验证。

### Phase 1: 依赖方向修复（优先级最高）

| 任务 | 描述 | 预估 |
|------|------|------|
| 1.1 | 新建 `domain/port/` 包，提取跨域接口 | 1h |
| 1.2 | `domain/keys` 改用 `port.KeySyncPort` | 0.5h |
| 1.3 | `domain/budget` 改用 `port.OverrunKeyControl` | 0.5h |
| 1.4 | `identity/authz` 改为函数注入 `ChargeRateResolver`，去掉 billing import | 0.5h |
| 1.5 | 验证：编译通过 + 现有测试全绿 | 0.5h |

**阶段产出**：依赖图干净，domain 不再横向 import 其他 domain 实现包。

### Phase 2: 接线简化

| 任务 | 描述 | 预估 |
|------|------|------|
| 2.1 | 每个 handler 包加 `Mount()` 函数 | 2h |
| 2.2 | 在 `router.go` 用 `mountAPI()` 替代 Registry | 0.5h |
| 2.3 | 删除 `handler/register.go` 里的 Registry | 0.5h |
| 2.4 | 合并 `ServiceRegistry` + `domainServices` 为单个 `buildDeps()` | 2h |
| 2.5 | 验证 | 0.5h |

**阶段产出**：新增 handler 接线从碰 5 文件降到碰 2 文件（wire.go + router.go）。

### Phase 3: 解耦 + 整理

| 任务 | 描述 | 预估 |
|------|------|------|
| 3.1 | newapisync 移到 `integration/newapisync/`，更新 import | 2h |
| 3.2 | adapter/ 拆为 `enqueue/` + `bridge/` | 0.5h |
| 3.3 | Config 分片（逐个 domain，可分多次 PR） | 4-6h |
| 3.4 | pkg/ 收拢 7 个碎片包 | 2h |
| 3.5 | Narrow Tx 泛型封装（可选，不阻塞） | 2h |

**阶段产出**：最终目标架构完成。

---

## 不动的部分（当前已经做对的）

| 模式 | 为什么好 |
|------|---------|
| Domain 定义 narrow Store interface | 依赖倒置，domain 不耦合全量 store |
| Domain 定义 port interface (JobEnqueuer, AlertPublisher) | adapter 实现，domain 可独立测试 |
| `shared.ProtectedHandlerBase` | 统一鉴权 + 权限检查模式 |
| `httputil.WriteJSON(w, status, data, err)` | 统一错误→HTTP 状态码映射 |
| `DomainError` 携带 status code | domain 能表达业务异常语义但不 import net/http |
| River + Holder 延迟绑定 | 避免初始化循环依赖 |
| `domain/types/` 共享内核 | 跨域值对象集中管理 |
| Middleware 组合（租户解析→限流→超时→CORS） | 分层清晰，插拔方便 |
| 测试外部包 `tests/` 镜像 `internal/` | 黑盒测试，不依赖内部实现 |
| `store.Store` 聚合全部 Repository | 事务边界统一，Postgres 实现简洁 |

---

## 风险与缓解

| 风险 | 缓解 |
|------|------|
| newapisync 移动 import path 改动多 | IDE rename + `goimports`；一次 PR，CI 验证 |
| Handler 自注册后遗漏 Mount 不报编译错误 | unused import 检测 + integration test 覆盖路由 prefix |
| Config 分片后 wire.go 变长 | 每个 domain 的 Config 构造可提取为 `cfg.ForBudget()` helper |
| domain/port/ 膨胀 | 只放跨域消费的接口；domain-private port 仍在各域 ports.go |

---

## 总结

当前架构基础扎实，domain 隔离和 port/adapter 模式已经到位。核心改动是：

1. **修正依赖方向**（新增 `domain/port/`）—— 解决 domain 横向 import 和 identity→domain 反向依赖
2. **精简接线仪式**（Handler Mount + 合并 Registry）—— 从 5 文件降到 2 文件
3. **重新定位 newapisync**（移入 integration）—— 反映其真实本质

三个阶段可独立交付，Phase 1 最重要且风险最低——纯粹的接口提取，不改任何业务逻辑。
