# 项目结构重构计划

基于 `.kiro/steering/project-structure.md` 规则，对现有代码进行合规性审计。

---

## 一、已合规（无需动作）

### 前端

- 页面入口 `routes/{domain}/{page}.tsx` 仅组合 — ✅
- 领域特性包 `features/{domain}/` 含 hooks/components/lib/index.ts — ✅
- 横切特性包 `features/{concern}/` — ✅ session、query、workflow
- 原子组件 `components/ui/` 无业务语义 — ✅
- 布局组件 `components/layout/` — ✅
- HTTP 客户端 `api/{domain}.ts` — ✅
- 纯工具函数 `lib/` — ✅
- 禁止直接 import API 函数 — ✅ 通过 useApis()/useInjectedApis()
- 测试在 `apps/frontend/tests/`，src/ 无测试 — ✅
- 共享合约在 `packages/contracts/` — ✅

### 后端

- cmd/ 仅 main 入口 — ✅
- 集成测试在 `apps/backend/tests/` — ✅
- 单元测试允许就近 `_test.go` — ✅

---

## 二、规则自身需补充的决策

在执行重构前，以下项需要先在规则层面达成共识：

### 2.1 后端共享内核范围

现状：`domain/company` 被 11 个 domain 引用，`domain/newapisync` 被 5 个 domain 引用。这两个在系统中扮演的角色类似 `domain/types` 和 `domain/grants`，属于基础设施型 domain。

**建议：** 将共享内核列表扩展为：

```
domain/types       — 公共值对象
domain/grants      — 权限模型
domain/company     — 租户上下文（company.Context、company.Clock）
domain/newapisync  — NewAPI 同步原语（PlatformKey/ProviderKey 类型定义）
```

更新后的 `project-structure.md` 共享内核例外行：
```
共享内核例外：domain/types、domain/grants、domain/company、domain/newapisync 可被自由引用
```

**决策点：** 是否接受？如接受，跨域违规数从 10 项降至 4 项。

### 2.2 前端 features 间引用的边界

规则写"features 之间只通过对方 index.ts 引用"，但存在两种特殊情况：

**情况 A：横切 feature 聚合各领域 query-keys**

`features/query/query-keys.ts` 直接引用 `@/features/budget/query-keys`（非 barrel）。这是因为 query-keys 聚合器需要收集所有 domain 的 key factory，但各 feature 的 `index.ts` 不一定导出 query-keys（避免循环）。

**建议：** query-keys 文件视为"注册点"，允许横切 feature 引用各 domain 的 `query-keys.ts`（仅此文件，不扩展到其他 deep path）。在规则中补充：

```
例外：features/query/query-keys.ts 允许引用各 feature 的 query-keys.ts
```

**情况 B：跨 feature deep import（真正违规）**

| 调用方 | 违规引用 | 原因 |
|--------|---------|------|
| `features/workflow/workflows/approval-review.tsx` | `@/features/models/hooks/use-model-labels` | 需要 model 标签 |
| `features/workflow/workflows/key-form/index.tsx` | `@/features/models/hooks/use-model-labels` | 同上 |
| `features/workflow/workflows/approval-submit.tsx` | `@/features/models/hooks/use-model-labels` | 同上 |
| `features/keys/components/platform-keys-page-shell.tsx` | `@/features/models/hooks/use-model-labels` | 同上 |

**修复方案：** 将 `useModelLabels` 加入 `features/models/index.ts` 的 barrel export，调用方改为 `import { useModelLabels } from '@/features/models'`。

---

## 三、待修复项

### 3.1 前端：跨 feature deep import（4 处）

上述 `useModelLabels` 相关的 4 处违规，修复步骤：

1. 在 `features/models/index.ts` 中添加 `export { useModelLabels } from './hooks/use-model-labels'`
2. 将 4 处 import 路径改为 `from '@/features/models'`

**验证命令：**
```bash
# 查找所有跨 feature deep import（feature 内部自引不算）
grep -rn "from '@/features/" apps/frontend/src/ --include="*.ts" --include="*.tsx" \
  | grep -v "from '@/features/[^']*'" \
  | grep -v "query-keys'" \
  | awk -F: '{print $1}' | sort -u
```

### 3.2 前端：其他 deep import（基础设施类）

以下 deep import 虽然违反字面规则，但属于"同层基础设施引用"：

| 模式 | 数量 | 处理 |
|------|------|------|
| `@/features/session/use-session` | ~8 处 | 加入 session index.ts 的 barrel export |
| `@/features/session/session-gate` | 1 处 | 同上 |
| `@/features/query/use-injected-query` | ~10 处 | 加入 query index.ts 的 barrel export |
| `@/features/query/use-injected-mutation` | 1 处 | 同上 |
| `@/features/notifications/notification-provider` | 1 处 | 加入 notifications index.ts 的 barrel export |
| `@/features/dev/components/simulate-consume-dialog` | 1 处 | 加入 dev index.ts 的 barrel export |

**修复步骤：** 确认各 feature 的 index.ts 已导出被外部使用的 symbol，然后统一替换 import 路径。

**验证命令（修复后应为 0 结果）：**
```bash
grep -rn "from '@/features/[^/]*/[^']*'" apps/frontend/src/ --include="*.ts" --include="*.tsx" \
  | grep -v "/features/[^/]*/index" \
  | grep -v "query-keys'"
```

### 3.3 前端：添加 lint 守护规则

修复完成后，添加 eslint 规则防止回退：

```js
// eslint.config.js 中增加
{
  rules: {
    'no-restricted-imports': ['error', {
      patterns: [{
        group: ['@/features/*/*', '!@/features/*/index'],
        message: '禁止 deep import features，请通过 barrel export 引用'
      }]
    }]
  }
}
```

需针对 `query-keys` 聚合点配置例外。

### 3.4 后端：跨 domain 直接引用（去除共享内核后）

假设 `domain/company` 和 `domain/newapisync` 被纳入共享内核，剩余违规：

| 调用方 | 引用的 domain | 优先级 |
|--------|--------------|--------|
| `domain/usage` | `domain/billing/lot` | P1 |
| `domain/usage` | `domain/budget` | P1 |
| `domain/memberanalytics` | `domain/keys` | P2 |
| `domain/memberanalytics` | `domain/usage` | P2 |
| `domain/dashboard` | `domain/usage` | P2 |
| `domain/models` | `domain/adminport` | P3 |

**解耦策略：**
- `usage → budget`：usage 定义 `BudgetChecker` interface，budget 实现，app 层注入
- `usage → billing/lot`：lot 计算逻辑考虑下沉到 `domain/types/` 或 `pkg/`
- `dashboard/memberanalytics → usage`：定义 `UsageQuerier` interface
- `models → adminport`：定义 `AdminPort` interface

### 3.5 后端：根目录 .test 文件清理

```bash
# 确认是 go test 编译产物
file apps/backend/*.test
# 清理
echo "*.test" >> apps/backend/.gitignore
git rm --cached apps/backend/*.test
```

### 3.6 文档：seed/FIXES.md 迁移

将 `apps/backend/seed/FIXES.md` 内容合并到 `docs/Backend-测试优化.md` 或 `docs/todos/` 中，然后删除原文件。

---

## 四、优先级与执行顺序

| 阶段 | 任务 | 工作量 | 风险 | 前置 |
|------|------|--------|------|------|
| **P0** | 规则更新：共享内核加入 company/newapisync | 改 3 处 md | 无 | — |
| **P0** | 清理 backend *.test 文件 | 5 分钟 | 无 | — |
| **P0** | seed/FIXES.md 迁移 | 5 分钟 | 无 | — |
| **P1** | 前端 barrel export 补全 + deep import 修复 | ~25 处改 import | 低 | — |
| **P1** | 添加 eslint no-restricted-imports 守护 | 1 处配置 | 无 | P1 修复完 |
| **P2** | usage → budget/billing 接口解耦 | 1 PR | 中 | 设计评审 |
| **P2** | dashboard/memberanalytics → usage 接口解耦 | 1 PR | 中 | P2 上一步 |
| **P3** | models → adminport 接口解耦 | 1 PR | 中 | — |

---

## 五、验收标准

- [ ] `grep` 命令（3.2 节）输出为空
- [ ] eslint `no-restricted-imports` 规则通过 `pnpm verify`
- [ ] 后端 `go build ./...` 成功（跨域解耦后）
- [ ] 后端 `*.test` 文件不在 git 追踪中
- [ ] `project-structure.md` 三处（kiro/cursor/CLAUDE.md）共享内核列表一致
