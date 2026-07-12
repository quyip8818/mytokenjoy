# Backend 结构优化

> **目的：** 记录 `apps/backend/` **当前结构基线**与**剩余分层债务**（非上线阻塞）。  
> **相关：** [Backend.md](./Backend.md) · [Backend-架构.md](./Backend-架构.md) · [Backend-计费模式.md](./Backend-计费模式.md) · [Backend-测试优化.md](./Backend-测试优化.md) · [工程收口.md](./工程收口.md)  
> **维护：** 结构变化先更新本文，再同步 [Backend-架构.md §3](./Backend-架构.md#3-项目结构)。

**读者速览：** domain 零 `infra/*` import；六域 Job 端口（`ports.go` + `app/*_enqueuer.go`）；`types.Notifier` SSOT；lot 写 SSOT 在 `domain/billing/lot/`；业务测在 `tests/`。§1 为现状；§2 为剩余债务；§3 为 PR 自检。

---

## 1. 当前架构

### 1.1 分层

```text
HTTP（handler / middleware）
  ↓
Domain（Service）
  ├→ Store
  └→ Port → Infra / Integration（经 app/ 注入）
```

**不变量：**

- 业务 Handler 调 `domain.Service`，不直访 Store（health / metrics 等基础设施 handler 除外）
- Handler **零业务规则**（如 budget 非负校验在 `domain/budget`）；跨域编排在 `app/`
- DTO SSOT：`domain/types/`
- domain **禁止** import 具体 integration 实现；NewAPI 经 `adminport.Port`；数据源经 `integration/datasource.Provider`
- **`WalletService`** 依赖 `company.QuotaReader`；由 `adminport.Port` 满足
- **`httpdeps.Deps`** 不携带 `store.Store`
- **dashboard scope** 经 `pkg/authzscope`；`usage/scope.go` 不 import `identity/authz`
- **Job 类端口**：六域 `ports.go` + `app/*_enqueuer.go`（billing / budget / usage / dashboard / newapisync / org-remote）
- middleware 读 authz 修订经 `identity/authz.RevisionReader`
- 业务测在 `tests/`（`internal/` 无普通 `*_test.go`）

**硬约束（CI / 本地可验）：**

```bash
rg 'internal/infra/' apps/backend/internal/domain/
rg 'integration/newapi|integration/datasource/feishu' apps/backend/internal/domain/
rg '\.Store\b' apps/backend/internal/http/handler/
```

### 1.2 目录职责

见 [Backend-架构.md §3](./Backend-架构.md#3-项目结构)。Store 按域拆 `*_repo_*.go`；org structure 拆 `domain/org/structure/` 多文件。

### 1.3 领域端口

**Job enqueuer（6 域）：** 各域 `ports.go` + `app/*_enqueuer.go`。

| 端口 | 定义 | 适配器 | 说明 |
| --- | --- | --- | --- |
| `billing.JobEnqueuer` | `domain/billing/ports.go` | `app/billing_enqueuer.go` | wallet_sync / 充值后 rebalance |
| `budget.JobEnqueuer` | `domain/budget/ports.go` | `app/budget_enqueuer.go` | 预算投影 / overrun / rebalance |
| `usage.IngestJobEnqueuer` | `domain/usage/ports.go` | `app/usage_enqueuer.go` | 入账后 enqueue；**须透传 `store.Tx`** |
| `dashboard.JobEnqueuer` | `domain/dashboard/ports.go` | `app/dashboard_enqueuer.go` | 看板投影 / reconcile |
| `newapisync.SyncJobEnqueuer` | `domain/newapisync/ports.go` | `app/newapisync_enqueuer.go` | PlatformKey 生命周期 |
| `remote.JobEnqueuer` | `domain/org/remote/ports.go` | `app/org_enqueuer.go` | org sync job |

**其它端口：**

| 端口 | 定义 | 注入 |
| --- | --- | --- |
| `adminport.Port` | `domain/adminport/` | `wiring_infra.go` |
| `types.Notifier` | `domain/types/notifier.go` | `infra/notification` |
| `GatewaySoftCache` | `domain/budget/gateway_soft_cache.go` | `budgetcheck.WrapStore` |
| `datasource.Provider` | `integration/datasource/` | `wire_domain_services.go` |
| `authz.RevisionReader` | `identity/authz/revision.go` | `authz.Service` |

**注入 SSOT：** `wire_domain_services.go`、`wire_river.go`、`wiring_infra.go`、`registry.go`。

**规则：**

- Job adapter **必须**在 `app/`，不可放 `infra/jobs`
- `usage.IngestJobEnqueuer.EnqueueAfterIngest` **必须透传 `store.Tx`**
- org/budget 通知经 `types.Notifier.Send`；domain 不 import `infra/notification`

### 1.4 钱包与 lot 边界

| 名称 | 路径 |
| --- | --- |
| **Lot 写 SSOT** | `domain/billing/lot/`（FIFO 消费、`wallet_remain`） |
| **Billing 域** | `domain/billing/`（充值、展示、wallet_sync） |
| **WalletService** | `domain/company/`（NewAPI quota 读；依赖 `QuotaReader`） |
| **Usage 聚合** | `store/postgres/usage_aggregate.go` → `UsageRepository` |

计费语义见 [Backend-计费模式.md](./Backend-计费模式.md)。

---

## 2. 剩余债务

上线 P0 见 [工程收口.md](./工程收口.md)。下列为低优先级、互不阻塞项。

| 序 | 类型 | 项 |
| ---: | --- | --- |
| 1 | 可读性 | [2.1 大文件机械拆分](#21-大文件机械拆分) |
| 2 | 性能 | [2.2 schema clone 性能](#22-schema-clone-性能) |
| 3 | 分层 | [2.3 端口定义位置收敛](#23-端口定义位置收敛) |

### 2.1 大文件机械拆分

| 文件 | 行数 | 拆法 |
| --- | ---: | --- |
| `integration/datasource/feishu/client.go` | ~391 | `auth.go`、`departments.go`、`members.go` |
| `infra/jobs/args.go` | ~253 | 按 kind 拆 `args_wallet_sync.go` 等；`Insert*` 仍集中在 `enqueue.go` |

### 2.2 schema clone 性能

范围：`tests/testutil/pg/clone.go`。详见 [Backend-测试优化.md](./Backend-测试优化.md)。

### 2.3 端口定义位置收敛

端口分散在 `domain/*/ports.go`、`adminport/`、`integration/datasource/`、域内具名文件。新外部系统端口优先 `domain/<域>/ports.go` 或 `adminport/`。

---

## 3. PR 自检

- [ ] 新异步入队：域内 `ports.go` + `app/*_enqueuer.go`；domain 不 import `infra/jobs`
- [ ] lot 写路径只经 `domain/billing/lot/`
- [ ] usage 聚合只经 `UsageRepository` / `usage_aggregate.go`
- [ ] middleware 读 authz 修订经 `RevisionReader`
- [ ] domain 无新增 `infra/*` / 具体 integration import（`datasource` Provider 接口包除外）
- [ ] 业务 handler 不直访 store
- [ ] `make test-unit` 全绿

---

*§1 随代码基线更新；§2 随债务收口删减。*
