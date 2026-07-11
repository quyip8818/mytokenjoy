# Backend 配置架构

> **范围**：`apps/backend` 配置加载、生产契约、空库引导、Secure Cookie、时钟、凭证密钥、测试构造。  
> **索引**：env 快表见 [Backend.md](./Backend.md) §3 / [Backend-架构.md](./Backend-架构.md) §1.1；完整示例见 `apps/backend/.env.example`。

---

## 1. 原则

1. 一个开关只管一件事。  
2. 开了就要配齐，缺了就不启动。  
3. 密钥和 Cookie 显式配置，代码里无 dev fallback。  
4. 不使用 `APP_PROFILE` / `DEMO_TODAY` / `MinimalSeed` 等旧开关；以 `DEPLOY_ENV` + `BOOTSTRAP_MODE` 为准。

---

## 2. 核心环境变量

| 变量 | 默认 | 职责 |
| --- | --- | --- |
| `DEPLOY_ENV` | `local` | `local` / `staging` / `production`；仅 `production` 触发生产契约 fail-fast |
| `BOOTSTRAP_MODE` | `none` | `none` / `minimal` / `demo`；空库引导策略 |
| `SECURE_COOKIE` | `false` | Set-Cookie Secure；`production` 下必须为 `true` |
| `CLOCK_ANCHOR` | 空 | 可选 `YYYY-MM-DD`；空=系统时钟；固定看板「今天」与种子参考日期 |
| `SIMULATE_DELAY` | `false` | 模拟延迟；`production` 下必须为 `false` |
| `DATA_SOURCE_CREDENTIAL_KEY` | **必填** | 数据源凭证加密；非法或不存在则启动失败 |

| `DEPLOY_ENV` | 行为 |
| --- | --- |
| `local` / `staging` | 启动日志标识；不强制生产契约（`staging` 可故意缺 NewAPI 做预发） |
| `production` | `validate()` 强制 §7 生产契约；缺任一项即启动失败 |

典型本地：`DEPLOY_ENV=local` + `BOOTSTRAP_MODE=demo` + 可选 `CLOCK_ANCHOR`。  
典型生产：`DEPLOY_ENV=production` + §7 全表。

---

## 3. `config` 包

源码：`internal/config/`（`config.go`、`bootstrap.go`、`deploy.go`、`validate.go`、`store_bootstrap.go` 等）。

### 3.1 关键字段

```go
BootstrapMode string `env:"BOOTSTRAP_MODE" envDefault:"none"`
SecureCookie  bool   `env:"SECURE_COOKIE" envDefault:"false"`
ClockAnchor   string `env:"CLOCK_ANCHOR"`
DeployEnv     string `env:"DEPLOY_ENV" envDefault:"local"`
StoreBootstrap StoreBootstrap // 仅测试构造，非 env
```

### 3.2 `Load()` 流程

```
env.Parse
  → 归一化 BootstrapMode、DeployEnv 小写
  → validate()（含 DATA_SOURCE_CREDENTIAL_KEY ParseKey）
  → 返回
```

`cmd/server` 启动日志带 `deploy_env`、`bootstrap_mode`。  
`BOOTSTRAP_MODE=demo` 且 `CLOCK_ANCHOR` 为空时打 **warn**（`DemoWithoutClockAnchor()`）。

### 3.3 `validate()` 要点

**始终必填 / 格式**：`DATABASE_URL`、`SESSION_SECRET`、`DATA_SOURCE_CREDENTIAL_KEY`（及 ParseKey）、SaaS/单租户规则、`BOOTSTRAP_MODE` / `DEPLOY_ENV` 枚举、`CLOCK_ANCHOR` 格式。

**能力组合**（任意 deploy env）：

| 条件 | 要求 |
| --- | --- |
| `NEW_API_GATEWAY_ENABLED=true` | `NEW_API_ENABLED=true` |
| `NEW_API_ENABLED=true` | `NEW_API_BASE_URL`、`NEW_API_ADMIN_TOKEN`；URL 无 path |
| `LOG_DATABASE_URL` 非空 | `NEW_API_WEBHOOK_SECRET` 必填 |

**生产契约**：见 §7；实现为 `validateProductionContract()`，由 `IsProductionDeploy()` 触发。

### 3.4 辅助方法

```go
BootstrapIsNone / BootstrapIsMinimal / BootstrapIsDemo
IsProductionDeploy
Clock()
SeedReferenceDate()   // clock.NowUTC(Clock()).Format("2006-01-02")；种子展示日
DemoWithoutClockAnchor()
IngestEnabled / CORSOriginList
```

`SeedReferenceDate` **不是**业务「当前时间」；账期路径一律 `Clock()`（或 `clock.NowUTC(clk)`）。

---

## 4. 时钟（`internal/pkg/clock`）

```go
type Clock interface { Now() time.Time }
func System() Clock
func Fixed(t time.Time) Clock
func OrDefault(clk Clock) Clock
func NowUTC(clk Clock) time.Time
```

`Config.Clock()`：`CLOCK_ANCHOR` 空 → `System()`；否则 `Fixed(锚定日 UTC 零点)`。  
包级 `clock.NowUTC(clk)`：业务「现在」的 UTC 瞬时。

### 4.1 调用约定

| 组件 | 用法 |
| --- | --- |
| `config.Config` | `Clock()` 解析 `CLOCK_ANCHOR` |
| `domain/dashboard`、`memberanalytics` | 构造器内 `clock: cfg.Clock()` |
| `domain/budget`、`keys`、`newapisync` | `Load*(..., cfg.Clock())` |
| `domain/gateway/precheck` | `GatewayPrecheck.LoadPrecheckContext`；SQL 内按 `org_nodes.period` + `Clock` 算 `period_key` |
| `domain/usage/ingest` | `OccurrenceDepartmentPeriod(..., OccurredAt)` + `OpenDepartmentPeriod(..., cfg.Clock())` → `Apply` |
| `pkg/budget` | 开账工厂见 [Backend-业务时钟与账期.md](./Backend-业务时钟与账期.md)；`Load*` 收 `clock.Clock` |
| `org/core` `BudgetPeriod()` | 返回 `pkgbudget.PeriodMonthly`；实时 period_key 由 Clock 解析 |

账期语义、双轨与护栏全文见 [Backend-业务时钟与账期.md](./Backend-业务时钟与账期.md)。  
域代码不得直接读 `CLOCK_ANCHOR` env。

---

## 5. 数据引导（`store/postgres`）

入口：`loadOrSeedDomain`（`members` 计数为 0 视为空库）。

| `BOOTSTRAP_MODE` | 空库行为 |
| --- | --- |
| `none` | 启动失败，提示设置 `minimal` / `demo` 或外部灌库 |
| `minimal` | `seed.LoadMinimal` → `ApplyTables` |
| `demo` | `seed.Load` → `ApplyTables` → `runtime.ApplyDemo`（buckets / recharge / ledger） |

非空库：**永不**写入 seed / runtime。  
非空库补演示 runtime：测试用 `NewTestStoreWithDemoRuntime` / `ApplyDemoRuntime`；运维用迁移脚本。

`StoreBootstrap`（仅测试）：`SchemaPrepared`、`TestPartitionMonths`（默认 12，缩小测试模板分区范围；生产仍 2024–2032）。无 `RuntimeSeed` / `SkipRuntimeSeed`。

克隆 schema 上 reopen store 须 `PreparedConfig(schemaURL)`（`SchemaPrepared=true`），否则会再跑 `apply partitions` 并在非分区父表上报错。见 [Backend.md](./Backend.md) §5.0。

---

## 6. HTTP 与安全

- Cookie：`SecureCookie: d.Config.SecureCookie`（`http/deps/public.go`）。  
- 凭证：`CredentialKey()` 只解析 `DATA_SOURCE_CREDENTIAL_KEY`；`DevDefaultKey()` 仅供单元测试直用，生产路径不可调用。

---

## 7. 生产契约（`DEPLOY_ENV=production`）

`validate()` fail-fast，运维 checklist 与代码同表：

| 变量 | 要求 |
| --- | --- |
| `BOOTSTRAP_MODE` | `none` |
| `SECURE_COOKIE` | `true` |
| `NEW_API_ENABLED` | `true` |
| `NEW_API_GATEWAY_ENABLED` | `true` |
| `LOG_DATABASE_URL` | 已设置 |
| `NEW_API_WEBHOOK_SECRET` | 已设置 |
| `DATA_SOURCE_CREDENTIAL_KEY` | 已设置且合法 |
| `SIMULATE_DELAY` | `false` |
| `CLOCK_ANCHOR` | 未设置 |

---

## 8. 应用装配

| 位置 | 约定 |
| --- | --- |
| `wire_domain_services` / `wiring_domain` | 构造器只收 `cfg`；账期路径内部 `cfg.Clock()` |
| `wire_gateway` | `GatewayPrecheck()` + `cfg.Clock()` → `PrecheckService`（无 `WalletService` / `AsyncJobs`） |
| `wiring_infra` | `newapisync.New(cfg, ...)`；`SimulateDelay` 读 `cfg.SimulateDelay` |

---

## 9. 测试约定

`tests/testutil/config.go` 默认：`DeployEnv=local`、`BootstrapMode=minimal`、`SchemaPrepared=true`、合法 `DataSourceCredentialKey`。

常用 option：`WithBootstrapMode`、`WithClockAnchor`、`WithDeployEnv`、`WithSecureCookie`、`WithProductionContract`。

| Helper | 用途 |
| --- | --- |
| `NewSecureCookieRouter` | 仅 `SECURE_COOKIE=true` |
| `NewTestStoreWithDemoRuntime` / `ApplyDemoRuntime` | 显式写入 usage/充值演示数据 |
| `WithProductionContract` | 填满 §7 以测生产契约加载 |

测试构造以 `WithProductionContract`、`NewSecureCookieRouter`、`ApplyDemoRuntime` 等为准；无 `WithProfile` / `WithMinimalSeed` 类 helper。

---

## 10. 源码索引

| 路径 | 职责 |
| --- | --- |
| `internal/config/*.go` | Load / validate / Clock / Bootstrap / SeedReferenceDate |
| `internal/pkg/clock/clock.go` | Clock 接口 |
| `internal/store/postgres/postgres.go` | `loadOrSeedDomain` |
| `seed/runtime/demo.go` | `ApplyDemo` |
| `internal/http/deps/public.go` | SecureCookie |
| `internal/domain/org/core/credentials.go` | CredentialKey |
| `tests/testutil/config.go` | TestConfig + options |
| `tests/testutil/pg/` | 测试 schema 模板与 clone |
| `tests/config/config_test.go` | 生产 / local / staging 分层校验 |
| `apps/backend/.env.example` | 本地与生产样例 |

---

## 11. 一句话

没有 Profile。本地 `DEPLOY_ENV=local` + 显式 `BOOTSTRAP_MODE=demo`；生产 `DEPLOY_ENV=production` 强制 §7 fail-fast；账期业务时间走 `Clock()`；密钥缺则死；空库 `none` 则死。
