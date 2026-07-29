# 测试环境选择（SaaS / Local）

> 支持 `pnpm test`、`test:integration`、`test:e2e` 在 SaaS 或 Local 模式下运行，覆盖两种部署形态的差异行为。

---

## 现状

| 命令 | 范围 | PG | 模式 |
|------|------|------|------|
| `pnpm test` | 后端 unit + 前端 vitest | 5510 (SaaS PG) | `SupportSaas=false`, `CompanyID=DefaultCompanyID` |
| `pnpm test:integration` | 后端 integration (race, serial) | 5510 | 同上 |
| `pnpm test:e2e` | Playwright (browser) | 5510 (`tokenjoy_e2e` DB) | 后端以 demo mode 启动 |

所有测试目前**固定运行在 SaaS 的 PG 实例上**（port 5510），且 `testutil.TestConfig()` 硬编码 `SupportSaas=false`。E2E 用独立 DB（`tokenjoy_e2e`）但也在 5510 上。

### 问题

1. 无法验证 SaaS 特有行为（多租户、platform admin、company provisioning）
2. 无法验证 Local 特有行为（setup flow、single-tenant middleware、selfhosted company type）
3. E2E global-setup 硬编码 admin 账户，不区分模式

---

## 目标

```
pnpm test              → 默认跑 local 模式（向后兼容）
pnpm test --saas       → 跑 SaaS 模式
pnpm test:integration  → 默认 local
pnpm test:integration --saas
pnpm test:e2e          → 默认 local
pnpm test:e2e --saas
```

不引入新依赖。最小改动原则。

---

## 设计

### 核心思路

环境差异通过 **`TEST_MODE=local|saas`** 环境变量传递。各层根据此变量选择：
- 使用哪个 PG（5510 vs 5520）
- Config 中 `SupportSaas` / `CompanyID` 的值
- E2E 启动的后端实例配置

### 1. 后端 unit / integration

#### `testutil.TestConfig()`

```go
func TestConfig(opts ...ConfigOption) config.Config {
    mode := os.Getenv("TEST_MODE") // "saas" | "" (default=local)
    supportSaas := mode == "saas"

    cfg := config.Config{
        PlatformConfig: config.PlatformConfig{
            SupportSaas:       supportSaas,
            CompanyID:         contract.DefaultCompanyID,
            TokenJoyCompanyID: contract.TokenJoyCompanyID,
            CompanyName:       "Demo Company",
        },
        // ...
    }
    // ...
}
```

#### 数据库选择

```go
func defaultTestDatabaseURL() string {
    if v := os.Getenv("DATABASE_URL"); v != "" {
        return v
    }
    // TEST_MODE=saas → port 5510, 否则 5520
    if os.Getenv("TEST_MODE") == "saas" {
        return "postgres://tokenjoy:tokenjoy@127.0.0.1:5510/tokenjoy?sslmode=disable"
    }
    return config.DefaultDatabaseURL // 5520
}
```

#### `ApplyLocalEnv` / `ApplySaasEnv`

```go
func ApplyLocalEnv(t *testing.T) {
    t.Setenv("SUPPORT_SAAS", "false")
    // ... 其余同现有
}

func ApplySaasEnv(t *testing.T) {
    t.Setenv("SUPPORT_SAAS", "true")
    t.Setenv("DATABASE_URL", "postgres://tokenjoy:tokenjoy@127.0.0.1:5510/tokenjoy?sslmode=disable")
    // ... SaaS 特有配置
}
```

#### 条件跳过

```go
func SkipUnlessSaaS(t *testing.T) {
    if os.Getenv("TEST_MODE") != "saas" {
        t.Skip("requires TEST_MODE=saas")
    }
}

func SkipUnlessLocal(t *testing.T) {
    if os.Getenv("TEST_MODE") == "saas" {
        t.Skip("requires TEST_MODE=local")
    }
}
```

用于标记只在某种模式下有意义的测试（如 setup flow 只 local，platform company provisioning 只 SaaS）。

---

### 2. 前端 E2E (Playwright)

#### 环境切换

`playwright.config.ts` 读 `TEST_MODE`：

```typescript
const mode = process.env.TEST_MODE ?? 'local'
const isSaaS = mode === 'saas'

export const PG_PORT = isSaaS ? '5510' : '5520'
export const E2E_BACKEND_PORT = isSaaS ? 9420 : 9410
export const E2E_DATABASE_URL = `postgres://tokenjoy:tokenjoy@127.0.0.1:${PG_PORT}/tokenjoy_e2e?sslmode=disable`
```

#### webServer 启动

```typescript
webServer: [
  {
    command: `SUPPORT_SAAS=${isSaaS} PORT=${E2E_BACKEND_PORT} DATABASE_URL=${E2E_DATABASE_URL} ... go run ./cmd/server`,
    port: E2E_BACKEND_PORT,
  },
  {
    command: `VITE_SUPPORT_SAAS=${isSaaS} vite preview --port ${E2E_PREVIEW_PORT}`,
    port: E2E_PREVIEW_PORT,
  },
]
```

#### global-setup

```typescript
// SaaS: platform admin login
// Local: demo admin login (setup 阶段创建的 admin)
const email = isSaaS ? 'admin@tokenjoy.me' : 'demo@tokenjoy.me'
const password = isSaaS ? 'admin1234' : 'demo1234'
```

---

### 3. 脚本层

#### `scripts/dev/test.sh`

```bash
MODE="${1:-local}"  # pnpm test [local|saas]
shift || true

export TEST_MODE="${MODE}"

case "${MODE}" in
  saas)  PG_COMPOSE=tokenjoy-saas; PG_FILE="${ROOT}/docker-compose.yml" ;;
  local) PG_COMPOSE=tokenjoy-local; PG_FILE="${ROOT}/docker-compose.local.yml" ;;
esac

docker compose -p "${PG_COMPOSE}" -f "${PG_FILE}" up postgres -d --wait
# ... 运行测试
```

#### `package.json`（root）

```json
{
  "test": "bash scripts/dev.sh test",
  "test:saas": "bash scripts/dev.sh test saas",
  "test:local": "bash scripts/dev.sh test local",
  "test:integration": "bash scripts/dev.sh test:integration",
  "test:integration:saas": "bash scripts/dev.sh test:integration saas",
  "test:e2e": "bash scripts/dev.sh test:e2e",
  "test:e2e:saas": "bash scripts/dev.sh test:e2e saas"
}
```

或更简洁：`pnpm test -- saas`（传参方式）。

---

## 文件变更

| 操作 | 路径 | 说明 |
|------|------|------|
| 修改 | `scripts/dev/test.sh` | 接受 `local`/`saas` 参数，export `TEST_MODE`，选择对应 PG compose |
| 修改 | `tests/testutil/config.go` | `TestConfig` 读 `TEST_MODE` 设 `SupportSaas` |
| 修改 | `tests/testutil/config.go` | `defaultTestDatabaseURL` 按 mode 选 port |
| 修改 | `tests/testutil/env.go` | 新增 `ApplySaasEnv`；`ApplyLocalEnv` 保持不变 |
| 新增 | `tests/testutil/skip.go` | `SkipUnlessSaaS` / `SkipUnlessLocal` |
| 修改 | `apps/frontend/e2e/e2e-db.ts` | `PG_PORT` 和端口按 `TEST_MODE` 切换 |
| 修改 | `apps/frontend/playwright.config.ts` | webServer env 注入 `SUPPORT_SAAS` |
| 修改 | `apps/frontend/e2e/global-setup.ts` | 按模式选登录账户 |
| 修改 | `package.json` | 新增 `test:saas` / `test:e2e:saas` 等便捷命令 |

---

## 不做的事

- 不做 matrix CI（CI 后续再加 saas/local 并行 job）
- 不做 testcontainers（现有 docker compose 够用）
- 不改 template DB 机制（SaaS 模式的 template 用 5510 创建，local 用 5520）
- 不为 SMS 子系统添加模式选择（SMS 只有一种模式）

---

## 默认行为（向后兼容）

- `pnpm test` 不带参数 = `TEST_MODE=local`（`DefaultDatabaseURL` 指向 5510 保持不变，但 `TestConfig` 中 `SupportSaas=false`）
- 现有所有测试在 local 模式下行为不变
- SaaS 模式是 opt-in：明确传 `saas` 或设 `TEST_MODE=saas`
