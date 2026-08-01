# Backend 配置架构

> **范围**：`apps/backend` 配置加载、生产契约、空库引导、Secure Cookie、时钟、凭证密钥、测试构造。  
> **索引**：env 快表见 [Backend-架构.md](./Backend-架构.md) §1.1；完整示例见 `apps/backend/.env.example`。

---

## 1. 原则

1. 一个开关只管一件事。
2. 开了就要配齐，缺了就不启动。
3. 密钥和 Cookie 显式配置，代码里无 dev fallback。
4. 没有 Profile 概念；以 `DEPLOY_ENV` + `SUPPORT_SAAS` 为准。

---

## 2. 核心环境变量

| 变量                         | 默认     | 职责                                                                       |
| ---------------------------- | -------- | -------------------------------------------------------------------------- |
| `DEPLOY_ENV`                 | `local`  | `local` / `staging` / `production`；仅 `production` 触发生产契约 fail-fast |
| `SUPPORT_SAAS`               | `false`  | `true`=SaaS 多租户（空库自动写 demo 快照）；`false`=单租户私有化（setup 流程一次性初始化） |
| `SECURE_COOKIE`              | `false`  | Set-Cookie Secure；`production` 下必须为 `true`                            |
| `CLOCK_ANCHOR`               | 空       | 可选 `YYYY-MM-DD`；空=系统时钟；固定看板「今天」；`production` 下禁止设置    |
| `SIMULATE_DELAY`             | `false`  | 模拟延迟；`production` 下必须为 `false`                                    |
| `SKIP_VERIFY_CODE`           | `false`  | 跳过验证码校验（本地/测试用）；`production` 下必须为 `false`               |
| `TOKENJOY_COMPANY_ID`        | 内置 UUID | 平台模型源公司 UUID；必须设置                                             |
| `DATA_SOURCE_CREDENTIAL_KEY` | **必填** | 数据源凭证加密；非法或不存在则启动失败                                     |

| `DEPLOY_ENV`        | 行为                                                             |
| ------------------- | ---------------------------------------------------------------- |
| `local` / `staging` | 启动日志标识；不强制生产契约（`staging` 可故意缺 NewAPI 做预发） |
| `production`        | `validate()` 强制 §7 生产契约；缺任一项即启动失败                |

典型本地：`DEPLOY_ENV=local` + `SUPPORT_SAAS=false` + 可选 `CLOCK_ANCHOR`。  
典型生产：`DEPLOY_ENV=production` + §7 全表。

---

## 3. `config` 包

源码：`internal/config/`（`config.go`、`deploy.go`、`validate.go`、`store_bootstrap.go` 等）。配置按能力域拆成多个子结构体（`DeployConfig`、`DatabaseConfig`、`NewAPIConfig`、`PlatformConfig` 等），嵌入到根 `Config` 中，字段访问仍是扁平的（如 `cfg.DatabaseURL`）。

### 3.1 关键字段

```go
type DeployConfig struct {
    SecureCookie   bool   `env:"SECURE_COOKIE" envDefault:"false"`
    ClockAnchor    string `env:"CLOCK_ANCHOR"`
    DeployEnv      string `env:"DEPLOY_ENV" envDefault:"local"`
    SimulateDelay  bool   `env:"SIMULATE_DELAY" envDefault:"false"`
    SkipVerifyCode bool   `env:"SKIP_VERIFY_CODE" envDefault:"false"`
}

// DatabaseConfig.StoreBootstrap — 仅测试/启动内部字段，非 env
type StoreBootstrap struct {
    SkipSchema          bool
    SkipSeed            bool
    TestPartitionMonths int
}
```

### 3.2 `Load()` 流程

```
env.Parse
  → 归一化 DeployEnv 小写
  → normalize()（NewAPIBaseURL 规整为 origin）
  → validate()（含 DATA_SOURCE_CREDENTIAL_KEY 校验）
  → 返回
```

### 3.3 `validate()` 要点

**始终必填 / 格式**：`DATABASE_URL`、`TOKENJOY_COMPANY_ID`、`SESSION_SECRET`、`DATA_SOURCE_CREDENTIAL_KEY`（格式校验）、`DEPLOY_ENV` 枚举、`CLOCK_ANCHOR` 格式。

**能力组合**（任意 deploy env）：

| 条件                           | 要求                                                   |
| ------------------------------ | ------------------------------------------------------ |
| `NEW_API_GATEWAY_ENABLED=true` | `NEW_API_ENABLED=true`                                 |
| `NEW_API_ENABLED=true`         | `NEW_API_BASE_URL` 必填                                |
| `LOG_DATABASE_URL` 非空        | `NEW_API_WEBHOOK_SECRET` 必填                          |
| `LogSchemaIsolated=true`（仅测试） | 需 `IngestEnabled()`                               |

**生产契约**：见 §7；实现为 `validateProductionContract()`，由 `IsProductionDeploy()` 触发。

### 3.4 辅助方法

```go
IsProductionDeploy()        // DeployEnv == "production"
AllowsDevHTTPRoutes()       // DeployEnv == "local"；/api/dev/* 唯一门禁
Clock()                     // 见 §4
IngestEnabled()             // LOG_DATABASE_URL != ""
CORSOriginList()
DBPoolMaxConns() / DBPoolMinConns()
InviteSecretKeys() / InviteExpireDuration()
```

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

| 组件                                  | 用法                                                                                               |
| ------------------------------------- | -------------------------------------------------------------------------------------------------- |
| `config.Config`                       | `Clock()` 解析 `CLOCK_ANCHOR`                                                                      |
| `domain/dashboard`、`memberanalytics` | 构造器内 `clock: cfg.Clock()`                                                                      |
| `domain/budget`、`keys`               | `Load*(..., cfg.Clock())`                                                                          |
| `integration/newapisync`              | `Load*(..., cfg.Clock())`                                                                          |
| `domain/gateway/precheck`             | `GatewayPrecheck.LoadPrecheckContext`；SQL 内按 `org_nodes.period` + `Clock` 算 `period_key`       |
| `domain/usage/ingest`                 | `OccurrenceDepartmentPeriod(..., OccurredAt)` + `OpenDepartmentPeriod(..., cfg.Clock())`           |
| `pkg/budget`                          | 开账工厂见 [Backend-业务时钟与账期.md](./Backend-业务时钟与账期.md)；`Load*` 收 `clock.Clock`      |
| `org/core` `BudgetPeriod()`           | 返回 `pkgbudget.PeriodMonthly`；实时 period_key 由 Clock 解析                                      |

账期语义、双轨与护栏全文见 [Backend-业务时钟与账期.md](./Backend-业务时钟与账期.md)。  
域代码不得直接读 `CLOCK_ANCHOR` env。

---

## 5. 数据引导（`seed/`）

入口：`seed.Init`（由 `store/postgres.New` 调用，schema DDL 之后）。**没有 `BOOTSTRAP_MODE` 开关**——引导策略完全由 `SUPPORT_SAAS` + 库是否为空决定：

| 步骤 | 行为                                                                                            |
| ---- | ------------------------------------------------------------------------------------------------ |
| 1    | **总是执行**：`bootstrap.ApplyBootstrap`（currencies/权限/角色/组织/模型，全部 `ON CONFLICT DO NOTHING` 幂等） + `seedGlobalPresetRoles` |
| 2    | `SUPPORT_SAAS=true` 且库为空（`SELECT COUNT(*) FROM members = 0`）→ 追加完整 demo 快照（`seed.Load` + `ApplyTables`）+ `runtime.ApplyDemo`（buckets/recharge/ledger 运行时演示数据） |
| —    | `SUPPORT_SAAS=false`（单租户私有化）→ 不写 demo 数据；`cfg.CompanyID` 由 `cmd/server` 的一次性 setup server 流程产出后再启动正常服务 |

`BOOTSTRAP_CONFIG_PATH`（可选 YAML，见 `seed/bootstrap/config.go`）：自定义 bootstrap 内容（如首个平台管理员账号）；未配置时用内嵌 `DefaultConfig()`。

`StoreBootstrap`（仅测试/启动内部字段，非 env）：`SkipSchema`（外部已 apply DDL 时跳过）、`SkipSeed`（模板 DB 已含 seed 数据时跳过）、`TestPartitionMonths`（默认 12，缩小测试模板分区范围；生产仍 2024–2032）。

克隆 schema 上 reopen store 须 `SkipSchema=true; SkipSeed=true`，否则会再跑 `apply partitions` 并在非分区父表上报错。

---

## 6. HTTP 与安全

- Cookie：`SecureCookie: d.Config.SecureCookie`（`http/deps/public.go`）。
- 凭证：`CredentialKey()` 只解析 `DATA_SOURCE_CREDENTIAL_KEY`；`DevDefaultKey()` 仅供单元测试直用，生产路径不可调用。

---

## 7. 生产契约（`DEPLOY_ENV=production`）

`validate()` fail-fast，运维 checklist 与代码同表：

| 变量                         | 要求         |
| ---------------------------- | ------------ |
| `SECURE_COOKIE`              | `true`       |
| `NEW_API_ENABLED`            | `true`       |
| `NEW_API_GATEWAY_ENABLED`    | `true`       |
| `LOG_DATABASE_URL`           | 已设置       |
| `NEW_API_WEBHOOK_SECRET`     | 已设置       |
| `SIMULATE_DELAY`             | `false`      |
| `SKIP_VERIFY_CODE`           | `false`      |
| `CLOCK_ANCHOR`               | 未设置       |

（`DATA_SOURCE_CREDENTIAL_KEY`、`TOKENJOY_COMPANY_ID` 属「始终必填」，见 §3.3，不重复在此列。）

---

## 8. 应用装配

| 位置                | 约定                                                                                                                           |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------------ |
| `compose_domain.go` | 构造器只收 `cfg`；账期路径内部 `cfg.Clock()`                                                                                   |
| `compose_http.go`   | `wireGateway` / `wireIdentity`：`GatewayPrecheck()` + `cfg.Clock()` → `PrecheckService`                                        |
| `compose_infra.go`  | `buildAdminPort` 直连 NewAPI 库读 token；`SimulateDelay` 读 `cfg.SimulateDelay`                                                |

---

## 9. 测试约定

`tests/testutil/config.go` 默认：`DeployEnv=local`、`SkipSchema=true`（clone 时）、`SkipSeed=true`（clone 时）、合法 `DataSourceCredentialKey`。`CompanyID=contract.DefaultCompanyID`。

常用 option：`WithClockAnchor`、`WithDeployEnv`、`WithSecureCookie`、`WithSupportSaas`、`WithProductionContract`、`WithIngestEnabled`、`WithNewAPIEnabled`。

| Helper                                             | 用途                        |
| -------------------------------------------------- | --------------------------- |
| `NewSecureCookieRouter`                            | 仅 `SECURE_COOKIE=true`     |
| `NewTestStoreWithDemoRuntime` / `ApplyDemoRuntime` | 显式写入 usage/充值演示数据 |
| `WithProductionContract`                           | 填满 §7 以测生产契约加载    |

**开发循环：**

| 命令             | 用途                                      |
| ---------------- | ----------------------------------------- |
| `make test-fast` | 仅 `tests/pkg/...`，无 `DATABASE_URL`     |
| `make test-unit` | 全量 `go test -tags=testhook ./tests/...` |


---

## 10. 源码索引

| 路径                                      | 职责                                                         |
| ----------------------------------------- | ------------------------------------------------------------ |
| `internal/config/config.go`               | Load / normalize / 子结构体定义                             |
| `internal/config/deploy.go`                | IsProductionDeploy / Clock / validateProductionContract      |
| `internal/config/validate.go`             | validate 主流程                                              |
| `internal/config/store_bootstrap.go`      | StoreBootstrap（测试专用）                                   |
| `internal/pkg/clock/clock.go`             | Clock 接口                                                   |
| `seed/init.go`                            | `seed.Init` — 数据引导总入口                                 |
| `seed/bootstrap/`                         | `ApplyBootstrap` / `seedGlobalPresetRoles` / bootstrap config |
| `seed/runtime/demo.go`                    | `ApplyDemo`（demo 运行时数据）                               |
| `internal/http/deps/public.go`            | SecureCookie                                                 |
| `internal/domain/org/core/credentials.go` | CredentialKey                                                |
| `tests/testutil/config.go`                | TestConfig + options                                         |
| `tests/testutil/pg/`                      | 测试 schema 模板与 clone                                     |
| `tests/config/config_test.go`             | 生产 / local / staging 分层校验                              |
| `apps/backend/.env.example`               | 本地与生产样例                                               |

---

## 11. 一句话

没有 Profile。本地 `DEPLOY_ENV=local` + `SUPPORT_SAAS=false`（或 `true` 走 demo 快照）；生产 `DEPLOY_ENV=production` 强制 §7 fail-fast；账期业务时间走 `Clock()`；密钥缺则死；bootstrap 幂等执行、SaaS 空库才写 demo 数据。
