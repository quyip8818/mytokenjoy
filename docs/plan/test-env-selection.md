# 测试环境选择（SaaS / Local）

> 测试使用独立 PG 实例，与 dev 环境完全隔离。SaaS/Local 模式通过 `TEST_MODE` 环境变量切换，同一 PG 内用不同 template DB 区分。

---

## 现状

| 命令 | 范围 | PG | 模式 |
|------|------|------|------|
| `pnpm test` | 后端 unit + 前端 vitest | 5510 (dev PG) | `SupportSaas=false` |
| `pnpm test:integration` | 后端 integration | 5510 | 同上 |
| `pnpm test:e2e` | Playwright | 5510 (`tokenjoy_e2e` DB) | demo mode |

所有测试和开发共用一个 PG（5510），且只覆盖 Local 模式。

### 问题

1. 测试和开发共享 PG，template DB 的 create/drop 可能干扰正在调试的 dev 环境
2. 无法验证 SaaS 特有行为（多租户、platform admin、company provisioning）
3. 无法验证 Local 特有行为（setup flow、single-tenant middleware）
4. `DefaultDatabaseURL` 指向 dev PG，测试没有独立出口

---

## 目标

```
pnpm test              → SaaS 模式（默认）
pnpm test --local      → Local 模式
pnpm test:integration  → SaaS 模式（默认）
pnpm test:integration --local → Local 模式
pnpm test:e2e          → SaaS 模式（默认）
pnpm test:e2e --local  → Local 模式
```

`--saas` 可显式指定，等同于默认行为。`--local` 切换到私有化模式。

原则：
- 测试用独立 PG，不影响 dev 环境
- SaaS / Local 只是逻辑切换（flag + template），不拆 PG 实例
- 最小改动，不引入新依赖

---

## 设计

### 核心架构

```
┌─────────────────────────────────────────────┐
│            dev 环境（不受影响）                │
│  docker-compose.yml       PG:5510 Redis:6310 │
│  docker-compose.local.yml PG:5520 Redis:6320 │
└─────────────────────────────────────────────┘

┌─────────────────────────────────────────────┐
│            test 环境（独立）                  │
│  docker-compose.test.yml  PG:5530 Redis:6330 │
│                                             │
│  ┌─────────────────────────────────┐        │
│  │ test_template_local (SupportSaas=false)  │
│  │ test_template_saas  (SupportSaas=true)   │
│  └─────────────────────────────────┘        │
└─────────────────────────────────────────────┘
```

隔离轴：**dev vs test**（PG 实例级别）
模式轴：**saas vs local**（template DB 级别，同一 PG 内）

### 1. docker-compose.test.yml（新增）

```yaml
# 测试专用基础设施，与 dev 完全隔离
services:
  postgres:
    image: postgres:17-alpine
    command: ['postgres', '-c', 'max_locks_per_transaction=1024', '-c', 'max_connections=300']
    environment:
      POSTGRES_USER: tokenjoy
      POSTGRES_PASSWORD: tokenjoy
      POSTGRES_DB: tokenjoy
    ports:
      - '5530:5432'
    volumes:
      - test_pg:/var/lib/postgresql/data
      - ./scripts/postgres-init:/docker-entrypoint-initdb.d:ro
    healthcheck:
      test: ['CMD-SHELL', 'pg_isready -U tokenjoy -d tokenjoy']
      interval: 3s
      timeout: 3s
      retries: 5

  redis:
    image: redis:7-alpine
    ports:
      - '6330:6379'
    healthcheck:
      test: ['CMD', 'redis-cli', 'ping']
      interval: 3s
      timeout: 3s
      retries: 5

volumes:
  test_pg:
```

不含 newapi 服务——测试中 newapi 由 httptest mock 或按需启动。

---

### 2. 后端 testutil 改造

#### `config.go` — `defaultTestDatabaseURL`

```go
func defaultTestDatabaseURL() string {
    if v := os.Getenv("DATABASE_URL"); v != "" {
        return v
    }
    // 默认指向测试专用 PG（5530），不再用 dev 的 5510
    return "postgres://tokenjoy:tokenjoy@127.0.0.1:5530/tokenjoy?sslmode=disable"
}
```

同步修改 `config.DefaultDatabaseURL` 保持不变（仍为 5510，供 dev/生产使用），测试侧不再引用它。

#### `config.go` — `TestConfig`

```go
func TestConfig(opts ...ConfigOption) config.Config {
    cfg := config.Config{
        // ... 现有字段不变
        PlatformConfig: config.PlatformConfig{
            SupportSaas:       testModeSaas(), // 根据 TEST_MODE 决定
            CompanyID:         contract.DefaultCompanyID,
            TokenJoyCompanyID: contract.TokenJoyCompanyID,
            CompanyName:       "Demo Company",
        },
    }
    for _, opt := range opts {
        opt(&cfg)
    }
    // ...
    return cfg
}

func testModeSaas() bool {
    // 默认 saas；仅 TEST_MODE=local 时为 false
    return os.Getenv("TEST_MODE") != "local"
}
```

#### `pgschema.go` — template DB 按模式命名

```go
func templateDBName() string {
    if os.Getenv("TEST_MODE") == "local" {
        return "test_template_local"
    }
    return "test_template_saas"
}
```

`templateStoreConfig()` 相应调整 `BootstrapMode`：
- local → `BootstrapDemo`（现有行为）
- saas → `BootstrapNone` + `SupportSaas=true` + platform seed

#### `skip.go`（新增）

```go
package testutil

import "testing"

func SkipUnlessSaaS(t *testing.T) {
    t.Helper()
    if os.Getenv("TEST_MODE") == "local" {
        t.Skip("requires TEST_MODE=saas (default)")
    }
}

func SkipUnlessLocal(t *testing.T) {
    t.Helper()
    if os.Getenv("TEST_MODE") != "local" {
        t.Skip("requires TEST_MODE=local")
    }
}
```

#### 删除 `ApplyLocalEnv` / `ApplyProductionEnv` 中的 `DATABASE_URL`

`ApplyLocalEnv` 不再设置 `DATABASE_URL`（测试侧由 `defaultTestDatabaseURL` 统一管理）。其余 env 保留。

---

### 3. 前端 E2E (Playwright)

#### 环境切换

```typescript
// playwright.config.ts
const mode = process.env.TEST_MODE ?? 'saas'  // 默认 saas
const isSaaS = mode !== 'local'

const TEST_PG_PORT = '5530'
const E2E_BACKEND_PORT = isSaaS ? 9420 : 9410
const E2E_DATABASE_URL = `postgres://tokenjoy:tokenjoy@127.0.0.1:${TEST_PG_PORT}/tokenjoy_e2e?sslmode=disable`
```

注意：E2E 的 PG port 始终是 5530（测试 PG），通过 `SUPPORT_SAAS` flag 区分模式。

#### webServer 启动

```typescript
webServer: [
  {
    command: `SUPPORT_SAAS=${isSaaS} PORT=${E2E_BACKEND_PORT} DATABASE_URL=${E2E_DATABASE_URL} go run ./cmd/server`,
    port: E2E_BACKEND_PORT,
    cwd: '../../backend',
  },
]
```

#### global-setup

```typescript
const email = isSaaS ? 'admin@tokenjoy.me' : 'demo@tokenjoy.me'
const password = isSaaS ? 'admin1234' : 'demo1234'
```

E2E global-setup 需要负责：
1. 连接 5530 创建 `tokenjoy_e2e` DB（如不存在）
2. 按 `TEST_MODE` 执行对应 migration + seed
3. 用对应账户完成登录

---

### 4. 脚本层

#### `scripts/dev/test.sh`（改造）

```bash
#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "$0")" && pwd)/../lib/common.sh"

# 解析 --saas / --local flag，默认 saas
MODE="saas"
nocache=false
for arg in "$@"; do
  case "${arg}" in
    --local) MODE="local" ;;
    --saas)  MODE="saas" ;;
    --nocache) nocache=true ;;
  esac
done
export TEST_MODE="${MODE}"

# 测试专用 compose（独立于 dev）
TEST_COMPOSE=(docker compose -p tokenjoy-test -f "${ROOT}/docker-compose.test.yml")

"${TEST_COMPOSE[@]}" up postgres redis -d --wait

if [[ "${nocache}" == "true" ]]; then
  pnpm -F @tokenjoy/frontend -F @tokenjoy/backend --parallel test:nocache
else
  pnpm -F @tokenjoy/frontend -F @tokenjoy/backend --parallel test
fi
```

#### `package.json`（root）

```json
{
  "test": "bash scripts/dev.sh test",
  "test:local": "bash scripts/dev.sh test --local",
  "test:integration": "bash scripts/dev.sh test:integration",
  "test:integration:local": "bash scripts/dev.sh test:integration --local",
  "test:e2e": "pnpm -F @tokenjoy/frontend test:e2e",
  "test:e2e:local": "TEST_MODE=local pnpm -F @tokenjoy/frontend test:e2e"
}
```

`pnpm test` 默认 SaaS 模式。`pnpm test --local` 或 `pnpm test:local` 切换到 Local 模式。也可手动 `TEST_MODE=local pnpm test`。

---

## 文件变更

| 操作 | 路径 | 说明 |
|------|------|------|
| 新增 | `docker-compose.test.yml` | 测试专用 PG:5530 + Redis:6330 |
| 修改 | `apps/backend/tests/testutil/config.go` | `defaultTestDatabaseURL` 指向 5530；`TestConfig` 读 `TEST_MODE` |
| 修改 | `apps/backend/tests/testutil/pgschema.go` | template DB 名称按 mode 区分 |
| 修改 | `apps/backend/tests/testutil/env.go` | 移除 `DATABASE_URL` 硬编码 |
| 新增 | `apps/backend/tests/testutil/skip.go` | `SkipUnlessSaaS` / `SkipUnlessLocal` |
| 修改 | `scripts/dev/test.sh` | 使用 `docker-compose.test.yml`，export `TEST_MODE` |
| 修改 | `apps/frontend/playwright.config.ts` | 端口和 flag 按 `TEST_MODE` 切换 |
| 修改 | `apps/frontend/e2e/global-setup.ts` | createdb + 按模式选登录账户 |
| 修改 | `package.json` | 新增 `test:local` / `test:integration:local` / `test:e2e:local` |

---

## template DB 策略

| template 名 | TEST_MODE | BootstrapMode | SupportSaas | seed 内容 |
|---|---|---|---|---|
| `test_template_saas` | saas（默认） | `none` | true | TokenJoyCompany + platform admin + company provisioning 数据 |
| `test_template_local` | local | `demo` | false | DefaultCompany + demo admin + 模型 |

两个 template 共存于同一个 PG（5530），由 `testTemplateVersion` + mode 后缀管理生命周期。schema 变更时 bump version，两个 template 都会重建。

---

## 不做的事

- 不为 dev 环境引入任何变更（5510/5520 保持原样）
- 不做 CI matrix（后续再加 saas/local 并行 job）
- 不做 testcontainers（compose 够用）
- 不为 SMS 子系统添加模式选择（SMS 只有一种模式）
- 不新增 `ApplySaasEnv` 辅助函数——统一用 `TestConfig(WithSupportSaas(true/false))` option 模式

---

## 向后兼容

- `pnpm test` 不带任何参数 = `TEST_MODE=saas`（SupportSaas=true, BootstrapNone + platform seed）
- `pnpm test --local` = `TEST_MODE=local`（SupportSaas=false, BootstrapDemo）
- 唯一基础设施变化：PG 端口从 5510 → 5530。首次运行需 `docker compose -f docker-compose.test.yml up -d`
- 现有测试需确认在 SaaS 默认模式下通过（SupportSaas=true）；纯 Local 行为的测试加 `SkipUnlessLocal`
- Local 模式是 opt-in：明确 `pnpm test --local` 或 `TEST_MODE=local`
