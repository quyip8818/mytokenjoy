# Backend 架构优化 — 剩余改进项

## 当前架构（重构后）

```
cmd/server/main.go
internal/
├── app/                — 依赖接线（compose_infra / compose_domain_wire / compose_domain / compose_worker）
├── config/             — 全局 Config（60+ 字段）
├── domain/
│   ├── port/           — ✅ 跨域共享 port interfaces (KeySyncPort, OverrunKeyControl)
│   ├── types/          — 共享值类型
│   ├── grants/         — 权限模型
│   └── {domain}/       — 各业务域
├── http/
│   ├── deps/           — Deps struct (25 字段)
│   ├── handler/{domain}/  — ✅ 每个 handler 有 Mount() 自注册
│   ├── middleware/
│   └── router.go       — ✅ mountAPI() 集中路由挂载
├── identity/           — ✅ authz 不再依赖 billing
├── store/              — Store interface + postgres/
├── adapter/
│   ├── enqueue/        — ✅ domain port → River job args
│   └── bridge/         — ✅ 跨域操作适配
├── infra/              — river / jobs / scheduler / budgetcheck / notification
├── integration/
│   ├── newapi/         — HTTP client + token store
│   ├── newapisync/     — ✅ 从 domain/ 移入
│   ├── datasource/
│   └── platform/
├── worker/pricingsync
└── pkg/                — ✅ 从 15 收到 9 包
    ├── baseurl/        — URL origin 解析
    ├── budget/         — 预算计算纯函数 (15 文件)
    ├── clock/          — 时间抽象 + 解析
    ├── common/         — 混合工具 (11 文件)
    ├── ctxcompany/     — context key
    ├── modelcatalog/   — 模型目录数据
    ├── org/            — 组织树纯函数
    ├── ratelimit/      — 限流算法接口
    └── tree/           — 通用树操作
```

---

## 已解决的问题

- ~~P0: 依赖方向违规~~ — `domain/port/` 解耦了 keys/budget → newapisync，ChargeRateResolver 注入解耦了 authz → billing
- ~~P1: Registry 三重冗余~~ — `domainServices` 已删除，Handler Mount 自注册
- ~~P2: adapter/ 混合职责~~ — 拆为 `enqueue/` + `bridge/`
- ~~P4: pkg/ 碎片化~~ — 15 → 9 包
- ~~newapisync 位置错误~~ — 已移至 `integration/`

---

## 剩余问题（按优先级排序）

### P1: Config 泄漏 — 所有 domain 直接 import config.Config

**现状**：19 个 domain 文件 import `internal/config`。每个 domain 的 `NewService()` 接收完整 `config.Config`（60+ 字段）。

**后果**：
- 测试需要构造完整 Config（或用 testutil 全局 helper）
- 改任何 config 字段 → 重编译所有 domain
- 不可能看 service 签名就知道它实际依赖哪些配置

**方案**：每个 domain 声明自己的 Config struct，wire 层负责从全局 Config 切片：

```go
// domain/budget/config.go
package budget

type Config struct {
    OverrunEnabled   bool
    PeriodKind       string
    Clock            func() time.Time
}

// app/compose_domain_wire.go
func wireBudget(cfg config.Config, i infra, enqueuer jobs.Enqueuer) domainbudget.Service {
    return domainbudget.NewService(domainbudget.Config{
        OverrunEnabled: cfg.OverrunEnabled,
        PeriodKind:     cfg.BudgetPeriodKind,
        Clock:          cfg.Clock(),
    }, i.store, i.delayer, enqueue.NewBudgetEnqueuer(enqueuer))
}
```

**优先改的 domain**（使用 config 字段最少的，收益/成本比高）：
1. `audit` — 只用 Clock
2. `memberanalytics` — 只用 Clock + BudgetPeriodKind
3. `dashboard` — 只用 Clock + NewAPIEnabled + BudgetPeriodKind
4. `budget` — 用 ~5 个字段
5. 其余逐步推进

**收益**：domain 包完全不 import `internal/config`，测试只需构造 3-5 字段的小 struct。

---

### P2: WithTx 回调泄漏全量 Store

**现状**：`store.Store.WithTx(ctx, func(Store) error)` 的回调参数是全量 `Store`。每个 domain 定义了 narrow Store interface，但事务回调里拿到的是 full Store。

```go
// domain/budget/service.go
type Store interface {
    Budget() store.BudgetRepository
    Org() store.OrgRepository
    WithTx(ctx context.Context, fn func(store.Store) error) error  // ← 泄漏全量 Store
}
```

**后果**：
- domain 的事务回调可以越权访问任何 repository
- Narrow Store interface 形同虚设（编译期不保护事务内行为）
- 测试 mock 需要实现全量 Store

**方案**：泛型 NarrowTxRunner 封装：

```go
// store/narrowtx.go
package store

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
// domain/budget/service.go
type service struct {
    store Store
    runTx store.NarrowTxRunner[Store]  // 事务内只暴露 narrow Store
}
```

**渐进式**：先在 budget domain 试点，验证模式后推广。不影响现有 WithTx 调用者。

---

### P3: ServiceRegistry 仍是 worker 和 HTTP 的共享管道

**现状**：`ServiceRegistry` 嵌入 `httpdeps.Deps` + 额外持有 `Infra`, `OrgSync`, `Overrun`, `Rebalance`。`buildBackgroundWorkers` 通过 `reg.Infra.*` 和 `reg.Overrun` 等字段获取 worker 依赖。

**后果**：
- 改 HTTP 层的 Deps 可能影响 worker 编译
- 新增 worker-only 依赖需要加到 ServiceRegistry

**方案**：`buildBackgroundWorkers` 改为直接接参数，不走 ServiceRegistry：

```go
type workerDeps struct {
    IngestSvc     domainusage.Ingestor
    Overrun       domainbudget.OverrunProcessor
    Rebalance     domainbudget.Rebalancer
    OrgSync       domainorg.SyncService
    NewAPISync    newapisync.OutboxHandler
    Notification  *notification.Service
    BudgetCheck   budgetcheck.Store
}

func buildBackgroundWorkers(cfg config.Config, logger *slog.Logger, st store.Store, 
    wd workerDeps, holder *jobs.Holder, orgAdmin *enqueue.OrgRiverAdminHolder) (*backgroundWorkers, error) {
    ...
}
```

`ServiceRegistry` 退化为仅暴露 `httpdeps.Deps`（给测试），或完全消除。

---

### P4: pkg/common 是万能垃圾桶

**现状**：11 个文件，职责混杂：
- `auditfilter.go` — 审计过滤逻辑（应属于 domain/audit 或 store/）
- `constants.go` — 全局常量（NewAPIGroupPrefix 等）
- `crypto.go` / `fieldcrypto.go` — 加密工具（应属于 identity/ 或独立 pkg/crypto）
- `org_store.go` — Org Store 辅助函数（应属于 domain/org 或 store/）
- `paginate.go` — 分页工具（通用，保留）
- `parse.go` — 字符串解析工具（通用，保留）
- `routing.go` — 模型路由算法（应属于 domain/models 或 pkg/modelcatalog）
- `scope_check.go` — HasAny 权限检查（通用，保留）
- `simulate.go` — Delayer 模拟延迟（保留）
- `time.go` — 时区相关工具（应合入 pkg/clock）

**方案**：逐步拆分，目标让 common 只剩真正的通用工具（paginate, parse, simulate, scope_check, constants）。

| 文件 | → 目标 |
|------|--------|
| `auditfilter.go` | `domain/audit` 或 `store/` |
| `crypto.go` + `fieldcrypto.go` | `pkg/crypto/` |
| `org_store.go` | `domain/org` 或 `store/` |
| `routing.go` | `pkg/modelcatalog/` |
| `time.go` | `pkg/clock/` |

---

### P5: pkg/budget 体量过大（15 文件）— 考虑内聚性

**现状**：`pkg/budget/` 有 15 个文件，全是纯函数（无 DB、无 side effect），被 `domain/budget`、`adapter/bridge`、`tests/` 共同引用。

**问题**：
- 文件数过多，新人不知道从哪入手
- 部分文件只被 `domain/budget` 使用（如 `validate.go`, `scope_validate.go`），不需要暴露到 pkg

**方案**（低优先级）：
- 仅被 `domain/budget` 使用的文件移回 `domain/budget` 内部（作为私有函数）
- 保留真正被多方共用的纯函数在 `pkg/budget`

---

### P6: River Deps 与 ServiceRegistry 字段重复

**现状**：`riverinfra.Deps` 有 17 个字段，其中 `Overrun`, `Rebalance`, `OrgSync`, `NewAPISync`, `Ingest` 与 ServiceRegistry 完全重复。两处都从 `buildBackgroundWorkers` 传入。

**方案**：如果实现 P3（消除 ServiceRegistry），River Deps 可以直接接收 `workerDeps` struct，或者 compose_worker.go 直接构造 `riverinfra.Deps` 而不经过中转。当前不阻塞，仅是信息重复的整洁问题。

---

## 不做的事项

| 项目 | 原因 |
|------|------|
| 拆分 `store.Store` interface | 当前聚合设计简洁（事务边界统一），narrow Store 在 domain 层已解决消费侧问题 |
| Middleware 重构 | 当前 17 个中间件已足够模块化 |
| 引入 DI 框架 (wire/fx) | 当前手动接线清晰，项目规模不需要自动注入 |
| 前端 API 层变更 | 本次纯后端重构，API 路径不变 |

---

## 执行优先级建议

| 顺序 | 任务 | 预估 | 风险 |
|------|------|------|------|
| 1 | P1: Config 分片（先做 audit/memberanalytics 试点） | 2h | 极低 |
| 2 | P2: NarrowTxRunner（先在 budget 试点） | 2h | 低 |
| 3 | P3: 消除 ServiceRegistry → workerDeps | 1h | 低 |
| 4 | P4: 拆分 pkg/common | 2h | 低 |
| 5 | P5: pkg/budget 内聚整理 | 1h | 极低 |
| 6 | P6: River Deps 整理 | 0.5h | 极低 |

所有任务独立可交付，不互相阻塞。推荐按顺序做 P1 → P2，收益最大。
