# E2E 自动化测试方案设计

> Playwright + Go 内存后端，验证前端完整用户流程

---

## 1. 方案概述

使用 Playwright 驱动浏览器，对接 Go 后端（内存 Store 模式）。测试模拟真实用户操作，验证从登录到业务操作的完整链路。

**选择理由**：
- Go 后端已有内存 Store，启动快、数据隔离天然支持
- 验证前后端实际协作，不只是 Mock 契约
- Playwright 自动等待、语义定位、并行隔离，适合管理后台场景
- 本地和 CI 环境一致，仅需 Node + Go

---

## 2. 环境架构

```
pnpm test:e2e
  ├── 1. 启动 Go Backend (内存 Store, port 8080)
  ├── 2. 启动 Vite Dev Server (VITE_ENABLE_MOCKS=false, proxy → localhost:8080)
  └── 3. Playwright 执行测试 → 浏览器访问 localhost:5173
```

### 数据隔离

- 后端启动时自动 seed 初始数据（复用 `internal/seed/`）
- 测试间通过 `storageState` 隔离登录态
- 需要干净状态的测试可调用 `POST /api/test/reset`（仅测试模式可用）

### 环境变量

| 变量 | 值 | 说明 |
|------|-----|------|
| `VITE_ENABLE_MOCKS` | `false` | 关闭 MSW |
| `VITE_API_PROXY_TARGET` | `http://localhost:8080` | 代理到后端 |
| `TOKENJOY_MODE` | `test` | 后端开启测试模式（暴露 reset 端点） |

---

## 3. 项目结构

```
apps/frontend/
├── e2e/
│   ├── playwright.config.ts
│   ├── fixtures/
│   │   └── auth.ts                # 多角色登录 fixture
│   ├── pages/                     # Page Object Model
│   │   ├── login.page.ts
│   │   ├── org-structure.page.ts
│   │   ├── budget.page.ts
│   │   ├── keys-mine.page.ts
│   │   └── approval.page.ts
│   └── tests/
│       ├── smoke.spec.ts
│       ├── org/
│       │   └── structure.spec.ts
│       ├── budget/
│       │   └── allocation.spec.ts
│       └── keys/
│           ├── self-service.spec.ts
│           └── approval-flow.spec.ts
```

---

## 4. Playwright 配置

```ts
// e2e/playwright.config.ts
import { defineConfig } from '@playwright/test'

export default defineConfig({
  testDir: './tests',
  baseURL: 'http://localhost:5173',
  timeout: 30_000,
  retries: process.env.CI ? 1 : 0,
  workers: process.env.CI ? 2 : undefined,
  use: {
    screenshot: 'only-on-failure',
    trace: 'on-first-retry',
  },
  projects: [
    { name: 'chromium', use: { browserName: 'chromium' } },
  ],
  webServer: [
    {
      command: 'cd ../../apps/backend && TOKENJOY_MODE=test go run ./cmd/server',
      port: 8080,
      reuseExistingServer: !process.env.CI,
      timeout: 15_000,
    },
    {
      command: 'VITE_ENABLE_MOCKS=false VITE_API_PROXY_TARGET=http://localhost:8080 pnpm start:frontend',
      port: 5173,
      reuseExistingServer: !process.env.CI,
      timeout: 15_000,
    },
  ],
})
```

---

## 5. 多角色 Fixture

```ts
// e2e/fixtures/auth.ts
import { test as base, type Page } from '@playwright/test'

type AuthFixtures = {
  adminPage: Page
  memberPage: Page
}

export const test = base.extend<AuthFixtures>({
  adminPage: async ({ browser }, use) => {
    const context = await browser.newContext({ storageState: '.auth/admin.json' })
    const page = await context.newPage()
    await use(page)
    await context.close()
  },
  memberPage: async ({ browser }, use) => {
    const context = await browser.newContext({ storageState: '.auth/member.json' })
    const page = await context.newPage()
    await use(page)
    await context.close()
  },
})
```

预置两个登录态：
- `admin`：超级管理员（全部权限）
- `member`：普通成员（self:keys, self:approval, api:call）

全局 setup 登录一次，保存 storageState 到 `.auth/` 目录，测试直接复用。

---

## 6. Page Object Model 设计

每个 Page Object 封装业务语义操作，隐藏选择器细节：

```ts
// e2e/pages/org-structure.page.ts
export class OrgStructurePage {
  constructor(private page: Page) {}

  async goto() {
    await this.page.goto('/org/structure')
  }

  async selectDepartment(name: string) {
    await this.page.getByRole('treeitem', { name }).click()
  }

  async addMember(info: { name: string; phone: string }) {
    await this.page.getByRole('button', { name: /添加成员/ }).click()
    await this.page.getByLabel('姓名').fill(info.name)
    await this.page.getByLabel('手机号').fill(info.phone)
    await this.page.getByRole('button', { name: /确定/ }).click()
  }

  async disableMember(name: string) {
    await this.page.getByRole('row', { name }).getByRole('button', { name: /更多/ }).click()
    await this.page.getByRole('menuitem', { name: /停用/ }).click()
    await this.page.getByRole('button', { name: /确认/ }).click()
  }

  async getMemberStatus(name: string): Promise<string> {
    const row = this.page.getByRole('row', { name })
    return row.getByTestId('member-status').innerText()
  }
}
```

定位策略优先级：
1. `getByRole` + accessible name（最稳定）
2. `getByLabel` / `getByText`（表单场景）
3. `getByTestId`（无语义标记时的兜底）

---

## 7. 测试场景（第一批）

### 7.1 smoke.spec.ts

验证：登录成功 → 侧边栏可见 → 依次导航到各页面 → 无白屏/错误

### 7.2 org/structure.spec.ts

验证：
- 创建部门（指定父节点）→ 树中可见
- 添加成员 → 列表中出现
- 停用成员 → 状态变更
- 删除非空部门 → 被拦截提示

### 7.3 budget/allocation.spec.ts

验证：
- 查看预算树 → 节点数据正确
- 向子部门分配额度 → 未分配余额减少
- 超额分配 → 错误提示阻止

### 7.4 keys/self-service.spec.ts

验证：
- 创建 Key（选模型 + 额度）→ 列表新增
- 轮转 Key → 前缀变更，配置保留
- 删除 Key → 额度释放回个人池

### 7.5 keys/approval-flow.spec.ts

验证：
- 成员提交额度申请
- 切换管理员账号 → 审批中心出现待审批
- 通过审批 → 成员额度增加

---

## 8. CI 集成

`ci.yml` 新增 job：

```yaml
e2e:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v5
    - uses: pnpm/action-setup@v4
    - uses: actions/setup-node@v5
      with:
        node-version: 24
        cache: pnpm
    - uses: actions/setup-go@v5
      with:
        go-version: '1.24'
    - run: pnpm install --frozen-lockfile
    - run: npx playwright install chromium
    - run: pnpm test:e2e
    - uses: actions/upload-artifact@v4
      if: failure()
      with:
        name: playwright-report
        path: apps/frontend/e2e/test-results/
```

预计 CI 耗时：30-60 秒（5 个 spec，2 workers 并行）。

---

## 9. 根目录脚本

`package.json` 新增：

```json
{
  "scripts": {
    "test:e2e": "pnpm --filter @tokenjoy/frontend test:e2e"
  }
}
```

`apps/frontend/package.json` 新增：

```json
{
  "scripts": {
    "test:e2e": "playwright test --config e2e/playwright.config.ts"
  }
}
```

---

## 10. 后续扩展（第二批）

确认第一批稳定后按需补充：
- `org/data-source.spec.ts` — 凭证配置 + 导入
- `org/roles.spec.ts` — 角色 CRUD + 权限生效
- `budget/alerts.spec.ts` — 预警规则管理
- `models/whitelist.spec.ts` — 白名单继承验证
- `audit/export.spec.ts` — 筛选 + CSV 导出
- `dashboard/cost.spec.ts` — 时间切换 + 下钻
