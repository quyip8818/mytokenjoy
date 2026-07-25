# CLAUDE.md

## 语言
- 所有回复使用简体中文，设计文档使用中文
- 项目没有上线，不需要 migration，不需要向后兼容

## 架构

默认在 apps/。路径无前缀 = apps/。

- `apps/` — TokenJoy，客户侧 LLM API 管控平台。前端 React19+TailwindV4+TanStackQuery，后端 Go+PG+Redis。`SUPPORT_SAAS` flag 区分私有化/SaaS。
- `apps/web/` — 官网，纯展示 React+TailwindV3，无路由，iframe 嵌入 App 认证。
- `sms/` — 内部供应商管理系统，独立 Go module，制品不出门。
- apps/sms 互不 import，跨产品仅 HTTP API，共享契约放 `packages/contracts/`。

## 文件放置

### apps 前端 (apps/frontend/)
- 页面：`routes/{domain}/{page}.tsx`（仅组合）
- 领域：`features/{domain}/`（含 hooks/components/lib/index.ts，外部禁止 deep import）
- 原子 UI：`components/ui/`（无业务语义）
- API：`api/{domain}.ts`，禁止直接 import——通过 useApis()
- 测试：`apps/frontend/tests/`（镜像 src/）
- 文档：`apps/docs/`

### apps 后端 (apps/backend/)
- cmd/ 仅入口，业务放 internal/
- domain 间禁止引用对方内部实现，通过 exported interface 协作
- 共享内核可自由引用：`domain/types`、`domain/grants`、`domain/company`、`domain/newapisync`
- 测试：`apps/backend/tests/`（镜像 internal/，外部测试包）

### sms
结构同 apps。测试 `sms/{frontend,backend}/tests/`，文档 `sms/docs/`。

### 通用
- 禁止在 src/internal/组件旁放测试文件
- 禁止在 frontend/backend/ 下新建 .md（README.md 除外）

## 测试

```bash
pnpm test                 # apps 全量
pnpm test:sms             # sms 全量
# 单文件
cd apps/backend && go test -tags=testhook -run "TestXxx" ./tests/domain/xxx/ -v -count=1
cd sms/backend && go test -run "TestXxx" ./tests/domain/xxx/ -v
```

### 后端（apps）
- 改 schema.sql/seed → bump `testTemplateVersion`（`tests/testutil/pg/template.go`）
- 时钟固定 `2026-06-19`，不用 `time.Now()` 断言
- 必须 `-tags=testhook`
- DB: `postgres://tokenjoy:tokenjoy@127.0.0.1:5510/tokenjoy`

### 后端（sms）
- DB: `postgres://sms:sms@127.0.0.1:5510/sms`，无需 build tag

### 前端
- Vitest + @testing-library/react，API 层 vi.mock
- E2E: Playwright

## Ponytail — lazy senior dev

写代码前按顺序停在第一个成立的：YAGNI → 复用已有 → 标准库 → 平台特性 → 已装依赖 → 一行搞定 → 最小实现。

- 先理解问题再动手，trace 完整流程
- Bug fix = 修根因，grep 所有 caller，修共享函数一次
- 不加未请求的抽象/依赖/样板代码
- 删除优于新增，无聊优于聪明，最少文件
- 最短 diff 胜出（前提：改对地方）
- 刻意简化标 `ponytail:` 注释说明天花板和升级路径
- 非 trivial 逻辑留一个最小可运行验证（assert/小测试）

不偷懒：理解问题、信任边界校验、防数据丢失的错误处理、安全、无障碍。
