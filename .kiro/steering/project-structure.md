# 文件放置规则

## apps（默认）

### 前端 (apps/frontend/)
- 页面：`routes/{domain}/{page}.tsx`（仅组合）
- 领域：`features/{domain}/`（含 hooks/components/lib/index.ts，外部禁止 deep import）
- 原子 UI：`components/ui/`（无业务语义）
- API：`api/{domain}.ts`，禁止直接 import——通过 useApis()
- 测试：`apps/frontend/tests/`（镜像 src/）
- 文档：`apps/docs/`

### 后端 (apps/backend/)
- cmd/ 仅入口，业务放 internal/
- domain 间禁止引用对方内部实现，通过 exported interface 协作
- 共享内核可自由引用：`domain/types`、`domain/grants`、`domain/company`、`domain/newapisync`
- 测试：`apps/backend/tests/`（镜像 internal/，外部测试包）

## sms

结构同 apps。测试 `sms/{frontend,backend}/tests/`，文档 `sms/docs/`。

## 通用

- 禁止在 src/internal/组件旁放测试文件
- 禁止在 frontend/backend/ 下新建 .md（README.md 除外）
