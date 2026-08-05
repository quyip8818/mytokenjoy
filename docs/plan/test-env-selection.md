# 测试环境选择（SaaS / Local）

> 测试使用独立 PG 实例（5530），与 dev 环境完全隔离。一条命令默认串行跑 SaaS + Local 两轮，`--saas` / `--local` 可只跑其中一个。

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
pnpm test              → 前端 vitest 跑一次 + 后端先 SaaS 再 Local 两轮
pnpm test --saas       → 前端 vitest + 后端只跑 SaaS
pnpm test --local      → 前端 vitest + 后端只跑 Local

pnpm test:integration  → 后端先 SaaS，再 Local
pnpm test:integration --saas  → 只跑 SaaS
pnpm test:integration --local → 只跑 Local

pnpm test:e2e          → 先 SaaS，再 Local
pnpm test:e2e --saas   → 只跑 SaaS
pnpm test:e2e --local  → 只跑 Local
```

`--saas` / `--local` 是过滤器，不带参数 = 两个都跑。
也可用环境变量：`TEST_MODE=local pnpm test`（flag 优先级高于 env）。

注意：前端 vitest 是纯 unit test（不连 PG），不参与 mode 循环，只跑一次。

原则：
- 测试用独立 PG，不影响 dev 环境
- SaaS / Local 只是逻辑切换（flag + template），不拆 PG 实例
- 最小改动，不引入新依赖

---

## 设计

### 核心架构

```
┌─────────────────────────────────────────────────┐
│            dev 环境（不受影响）                    │
│  docker-compose.yml       PG:5510 Redis:6310    │
│  docker-compose.local.yml PG:5520 Redis:6320    │
└─────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────┐
│            test 环境（独立）                      │
│  docker-compose.test.yml  PG:5530 Redis:6330    │
│                                                 │
│  ┌─────────────────────────────────────┐        │
│  │ template_saas  (SupportSaas=true)   │        │
│  │ template_local (SupportSaas=false)  │        │
│  └─────────────────────────────────────┘        │
└─────────────────────────────────────────────────┘
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

#### `mode.go`（新增）— `TestMode` 类型

```go
package testutil

import "os"

type TestMode string

const (
    ModeSaaS  TestMode = "saas"
    ModeLocal TestMode = "local"
)

func CurrentTestMode() TestMode {
    if os.Getenv("TEST_MODE") == "local" {
        return ModeLocal
    }
    return ModeSaaS // 默认
}
```

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

#### `config.go` — `TestConfig`

```go
func TestConfig(opts ...ConfigOption) config.Config {
    cfg := config.Config{
        // ... 现有字段不变
        PlatformConfig: config.PlatformConfig{
            SupportSaas:       CurrentTestMode() == ModeSaaS,
            CompanyID:         contract.DefaultCompanyID,
            TokenJoyCompanyID: contract.TokenJoyCompanyID,
            CompanyName:       "Demo Company",
        },
    }
    for _, opt := range opts {
        opt(&cfg)
    }
    return cfg
}
```

#### `pg/template.go` — `EnsureTemplateDB` 按 mode 参数化

```go
const testTemplateVersion = 47 // bump when schema/seed changes — 两个 template 共享版本号

// errOnce 支持失败重试的 once 包装
type errOnce struct {
    mu   sync.Mutex
    done bool
    err  error
}

func (o *errOnce) Do(f func() error) error {
    o.mu.Lock()
    defer o.mu.Unlock()
    if o.done {
        return o.err
    }
    o.err = f()
    if o.err == nil {
        o.done = true // 成功才标记完成，失败允许重试
    }
    return o.err
}

var templateOnces sync.Map // map[TestMode]*errOnce

func EnsureTemplateDB(ctx context.Context, baseURL string, mode TestMode) error {
    v, _ := templateOnces.LoadOrStore(mode, &errOnce{})
    return v.(*errOnce).Do(func() error {
        return buildOrVerifyTemplateDB(ctx, baseURL, mode)
    })
}

func templateDBName(mode TestMode) string {
    return "template_" + string(mode) // template_saas / template_local
}

func advisoryLockID(mode TestMode) int64 {
    switch mode {
    case ModeLocal:
        return 987654322
    default:
        return 987654321
    }
}
```

`buildOrVerifyTemplateDB` 内部逻辑不变，只是 DB 名和 lock ID 按 mode 分开。version 共享——bump 一次两个 template 都重建。失败时允许重试（errOnce 只在成功时标记 done）。

#### `pgschema.go` — `templateStoreConfig` 按 mode 参数化

**已实现（`apps/backend/tests/testutil/pgschema.go`）：**

```go
func templateStoreConfig(mode TestMode) config.Config {
    switch mode {
    case ModeSaaS:
        cfg := TestConfig(
            WithSupportSaas(true),
            WithPlatformBootstrap("admin@tokenjoy.me", "admin1234"),
            WithIngestEnabled(true),
        )
        cfg.StoreBootstrap.TestPartitionMonths = 12
        return cfg
    default: // ModeLocal
        // ponytail: local template 仍用 SupportSaas=true 写 demo 数据，
        // 让 template 里有数据可用；单个测试再按需覆盖 SupportSaas。
        cfg := TestConfig(
            WithSupportSaas(true),
            WithIngestEnabled(true),
        )
        cfg.StoreBootstrap.TestPartitionMonths = 12
        return cfg
    }
}
```

签名显式接收 mode 参数，和 `EnsureTemplateDB` 对齐，不依赖隐式 env 读取。**实现与原设计的差异**：Local 模板并未用 `WithSupportSaas(false)`，而是同样用 `true` 写入 demo 数据以保证模板内有数据可测，单个测试用例再按需覆盖回 `SupportSaas=false`。

#### `skip.go`（新增）

```go
package testutil

import "testing"

func SkipUnlessSaaS(t *testing.T) {
    t.Helper()
    if CurrentTestMode() != ModeSaaS {
        t.Skip("requires TEST_MODE=saas")
    }
}

func SkipUnlessLocal(t *testing.T) {
    t.Helper()
    if CurrentTestMode() != ModeLocal {
        t.Skip("requires TEST_MODE=local")
    }
}
```

---

### 3. 前端 E2E (Playwright)

#### env preset 文件

```typescript
// e2e/env/saas.ts
export const saasEnv = {
  SUPPORT_SAAS: 'true',
  BOOTSTRAP_MODE: 'none',
  PLATFORM_BOOTSTRAP_EMAIL: 'admin@tokenjoy.me',
  PLATFORM_BOOTSTRAP_PASSWORD: 'admin1234',
  COMPANY_NAME: '',
}

// e2e/env/local.ts
export const localEnv = {
  SUPPORT_SAAS: 'false',
  BOOTSTRAP_MODE: 'demo',
  COMPANY_NAME: 'Demo Company',
}
```

#### playwright.config.ts

```typescript
import { defineConfig } from '@playwright/test'
import { saasEnv } from './e2e/env/saas'
import { localEnv } from './e2e/env/local'

const mode = (process.env.TEST_MODE ?? 'saas') as 'saas' | 'local'
const modeEnv = mode === 'saas' ? saasEnv : localEnv

const TEST_PG_PORT = '5530'
const E2E_BACKEND_PORT = 9420
const E2E_SMS_PORT = 9421  // 偏移避免和 dev E2E (9411) 冲突
const E2E_PREVIEW_PORT = 9422
const E2E_HOST = '127.0.0.1'
const E2E_BASE_URL = `http://${E2E_HOST}:${E2E_PREVIEW_PORT}`
const E2E_DATABASE_URL = `postgres://tokenjoy:tokenjoy@127.0.0.1:${TEST_PG_PORT}/tokenjoy_e2e_${mode}?sslmode=disable`

export default defineConfig({
  // ...
  webServer: [
    {
      command: 'make run',
      cwd: '../backend',
      url: `http://${E2E_HOST}:${E2E_BACKEND_PORT}/healthz`,
      reuseExistingServer: !process.env.CI,
      timeout: 120_000,
      env: {
        ...modeEnv,
        DATABASE_URL: E2E_DATABASE_URL,
        PORT: String(E2E_BACKEND_PORT),
        SESSION_SECRET: 'e2e-test-session-secret',
        DATA_SOURCE_CREDENTIAL_KEY: 'dGV2LWNyZWRlbnRpYWwta2V5LWZvci1sb2NhbC1kZXY=',
        DEPLOY_ENV: 'local',
        NEW_API_BASE_URL: 'http://127.0.0.1:3010',
      },
    },
    {
      command: 'make seed && make run',
      cwd: '../../sms/backend',
      url: `http://${E2E_HOST}:${E2E_SMS_PORT}/api/health`,
      reuseExistingServer: !process.env.CI,
      timeout: 60_000,
      env: {
        DATABASE_URL: `postgres://tokenjoy:tokenjoy@127.0.0.1:${TEST_PG_PORT}/sms?sslmode=disable`,
        JWT_SECRET: 'e2e-sms-jwt-secret',
        PORT: String(E2E_SMS_PORT),
      },
    },
    {
      command: `pnpm build && pnpm exec vite preview --port ${E2E_PREVIEW_PORT} --strictPort --host ${E2E_HOST}`,
      url: E2E_BASE_URL,
      reuseExistingServer: !process.env.CI,
      timeout: 180_000,
      env: {
        VITE_API_PROXY_TARGET: `http://${E2E_HOST}:${E2E_BACKEND_PORT}`,
      },
    },
  ],
})
```

E2E DB 按模式命名：`tokenjoy_e2e_saas` / `tokenjoy_e2e_local`。
E2E 端口统一偏移到 942x 段，避免和 dev 的 941x 冲突。

#### global-setup

```typescript
const credentials = {
  saas: { email: 'admin@tokenjoy.me', password: 'admin1234' },
  local: { email: 'demo@tokenjoy.me', password: 'demo1234' },
}

export default async function globalSetup() {
  const mode = (process.env.TEST_MODE ?? 'saas') as 'saas' | 'local'
  const { email, password } = credentials[mode]
  await loginAndSave(email, password, '.auth/admin.json')
}
```

---

### 4. 脚本层

#### `scripts/dev/test.sh`（改造）

```bash
#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "$0")" && pwd)/../lib/common.sh"

# 解析过滤器（flag 优先级高于 TEST_MODE env）
modes=()
nocache=false
for arg in "$@"; do
  case "${arg}" in
    --saas)    modes+=("saas") ;;
    --local)   modes+=("local") ;;
    --nocache) nocache=true ;;
  esac
done

# 无 flag 时：读 TEST_MODE env，仍无则两个都跑
if [[ ${#modes[@]} -eq 0 ]]; then
  if [[ -n "${TEST_MODE:-}" ]]; then
    modes=("${TEST_MODE}")
  else
    modes=("saas" "local")
  fi
fi

# 测试专用 compose（独立于 dev）
TEST_COMPOSE=(docker compose -p tokenjoy-test -f "${ROOT}/docker-compose.test.yml")
"${TEST_COMPOSE[@]}" up postgres redis -d --wait

# 前端 vitest 只跑一次（纯 unit test，不依赖 PG/mode）
if [[ "${nocache}" == "true" ]]; then
  pnpm -F @tokenjoy/frontend test:nocache
else
  pnpm -F @tokenjoy/frontend test
fi

# 后端按 mode 循环
for mode in "${modes[@]}"; do
  echo ""
  echo "════════════════════════════════════════════"
  echo "  TEST_MODE=${mode}"
  echo "════════════════════════════════════════════"
  echo ""
  if [[ "${nocache}" == "true" ]]; then
    TEST_MODE="${mode}" pnpm -F @tokenjoy/backend test:nocache
  else
    TEST_MODE="${mode}" pnpm -F @tokenjoy/backend test
  fi
done
```

#### `scripts/dev/test-integration.sh`（新增）

```bash
#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "$0")" && pwd)/../lib/common.sh"

modes=()
for arg in "$@"; do
  case "${arg}" in
    --saas)  modes+=("saas") ;;
    --local) modes+=("local") ;;
  esac
done
if [[ ${#modes[@]} -eq 0 ]]; then
  if [[ -n "${TEST_MODE:-}" ]]; then
    modes=("${TEST_MODE}")
  else
    modes=("saas" "local")
  fi
fi

TEST_COMPOSE=(docker compose -p tokenjoy-test -f "${ROOT}/docker-compose.test.yml")
"${TEST_COMPOSE[@]}" up postgres redis -d --wait

for mode in "${modes[@]}"; do
  echo "=== TEST_MODE=${mode} ==="
  TEST_MODE="${mode}" pnpm -F @tokenjoy/backend test:integration
done
```

#### `scripts/dev/test-e2e.sh`（新增）

```bash
#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "$0")" && pwd)/../lib/common.sh"

modes=()
for arg in "$@"; do
  case "${arg}" in
    --saas)  modes+=("saas") ;;
    --local) modes+=("local") ;;
  esac
done
if [[ ${#modes[@]} -eq 0 ]]; then
  if [[ -n "${TEST_MODE:-}" ]]; then
    modes=("${TEST_MODE}")
  else
    modes=("saas" "local")
  fi
fi

TEST_COMPOSE=(docker compose -p tokenjoy-test -f "${ROOT}/docker-compose.test.yml")
"${TEST_COMPOSE[@]}" up postgres redis -d --wait

for mode in "${modes[@]}"; do
  echo "=== TEST_MODE=${mode} ==="
  TEST_MODE="${mode}" pnpm -F @tokenjoy/frontend test:e2e
done
```

#### `scripts/dev.sh`（补充 case）

```bash
case "${cmd}" in
  # ... 现有 case 不变
  test)             exec bash "${DEV}/test.sh" "$@" ;;
  test:integration) exec bash "${DEV}/test-integration.sh" "$@" ;;
  test:e2e)         exec bash "${DEV}/test-e2e.sh" "$@" ;;
esac
```

#### `package.json`（root）

```json
{
  "test": "bash scripts/dev.sh test",
  "test:integration": "bash scripts/dev.sh test:integration",
  "test:e2e": "bash scripts/dev.sh test:e2e"
}
```

命令不膨胀。模式选择通过 flag（`-- --saas`）或环境变量（`TEST_MODE=local`）。

> **注意**：pnpm 可能不透传未知 flag。稳妥用法是 `TEST_MODE=local pnpm test` 或 `pnpm test -- --local`。脚本同时支持两种方式（flag 优先于 env）。

---

## 文件变更

| 操作 | 路径 | 说明 |
|------|------|------|
| 新增 | `docker-compose.test.yml` | 测试专用 PG:5530 + Redis:6330 |
| 新增 | `apps/backend/tests/testutil/mode.go` | `TestMode` 类型 + `CurrentTestMode()` |
| 修改 | `apps/backend/tests/testutil/config.go` | `defaultTestDatabaseURL` 指向 5530；`TestConfig` 读 `CurrentTestMode()` |
| 修改 | `apps/backend/tests/testutil/pg/template.go` | `EnsureTemplateDB(ctx, url, mode)` 参数化；`errOnce`（失败可重试）；advisory lock 按 mode 分 |
| 修改 | `apps/backend/tests/testutil/pgschema.go` | `templateStoreConfig(mode)` 显式接收 mode 参数 |
| 新增 | `apps/backend/tests/testutil/skip.go` | `SkipUnlessSaaS` / `SkipUnlessLocal` |
| 修改 | `apps/backend/tests/testutil/env.go` | 移除 `DATABASE_URL` 硬编码 |
| 修改 | `scripts/dev.sh` | 新增 `test:integration` / `test:e2e` case |
| 修改 | `scripts/dev/test.sh` | 使用 `docker-compose.test.yml`；前端跑一次、后端循环 modes |
| 新增 | `scripts/dev/test-integration.sh` | integration 测试脚本（循环 modes） |
| 新增 | `scripts/dev/test-e2e.sh` | E2E 测试脚本（循环 modes） |
| 新增 | `apps/frontend/e2e/env/saas.ts` | SaaS 模式 env preset |
| 新增 | `apps/frontend/e2e/env/local.ts` | Local 模式 env preset |
| 修改 | `apps/frontend/playwright.config.ts` | 读 `TEST_MODE`，按 mode 选 env preset；端口偏移到 942x |
| 修改 | `apps/frontend/e2e/global-setup.ts` | 按 mode 选登录账户 |
| 修改 | `package.json` | `test:integration` / `test:e2e` 路由到 dev.sh |

---

## template DB 策略

| template 名 | TEST_MODE | BootstrapMode | SupportSaas | seed 内容 |
|---|---|---|---|---|
| `template_saas` | saas | `demo` | true | demo seed + SupportSaas=true（CompanyType=demo，provider key 由 platform 管理） |
| `template_local` | local | `demo` | false | demo seed + SupportSaas=false（CompanyType=selfhosted，provider key 自管理） |

两个 template 共存于同一个 PG（5530），共享 `testTemplateVersion`。schema 变更时 bump version，两个 template 都重建。

---

## 测试分类策略

| 分类 | 标记方式 | 何时跑 | 举例 |
|------|----------|--------|------|
| 通用 | 无标记（默认） | 两轮都跑 | CRUD、billing 计算、权限校验 |
| SaaS-only | `testutil.SkipUnlessSaaS(t)` | 仅 SaaS 轮 | platform admin、company provisioning、多租户隔离 |
| Local-only | `testutil.SkipUnlessLocal(t)` | 仅 Local 轮 | setup flow、single-tenant middleware、selfhosted company type |

### 原则

1. **大多数测试应为通用**——业务逻辑不应依赖部署模式
2. **SaaS-only / Local-only 仅用于验证模式差异行为**——即 `if cfg.SupportSaas` 分支里的逻辑
3. **新测试默认不加 Skip**——除非明确测试某种模式特有的功能路径

---

## CI 策略

```yaml
strategy:
  matrix:
    mode: [saas, local]
env:
  TEST_MODE: ${{ matrix.mode }}
steps:
  - run: TEST_MODE=${{ matrix.mode }} pnpm test --${{ matrix.mode }}
```

CI 用 matrix 并行跑两个 mode。通用测试跑两遍验证模式无关性，only 测试通过 Skip 自动过滤。
本地 `pnpm test` 串行跑两轮（先 saas 再 local）。

---

## 不做的事

- 不为 dev 环境引入任何变更（5510/5520 保持原样）
- 不用 testcontainers（compose 够用）
- 不为 saas/local 拆两个 PG 实例（同一 PG 内 template 隔离足够）
- 不为 SMS 子系统添加模式选择（SMS 只有一种模式）
- 不新增 `ApplySaasEnv` 辅助函数——统一用 `TestConfig` + `CurrentTestMode()` 自动决定
- 前端 vitest（纯 unit test）不参与 mode 循环，只跑一次
