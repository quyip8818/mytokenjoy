# 测试用例实现计划

## 后端单测（Go）

### 1. 模型 CreateModel 新字段测试
- `TestCreateModelWithNewFields` — 验证 apiKey、endpointModelName、maxContext、maxTokens、capabilities 能正确传入并持久化

### 2. 模型 UpdateModel 新字段测试
- `TestUpdateModelNewFields` — 验证 apiKey、endpointModelName、maxTokens 能正确更新

### 3. Toggle 全局模型二次 toggle 测试
- `TestToggleGlobalModelTwice` — 第一次 toggle 创建覆盖，第二次 toggle 更新覆盖（走 update 分支）

### 4. 预算 memberAvgBudget 继承测试
- `TestMemberAvgBudgetInheritance` — 验证子节点未设置时继承父节点值

## 前端单测（Vitest）

### 1. use-model-list-page 测试扩展
- `filters to builtin only when not selfhosted` — SaaS 版只返回内置模型
- `returns isSelfHosted from session` — 验证 selfhosted 判断

### 2. use-model-routing-page 测试重写
- `loads rules, departments, and models` — 新 hook 同时加载三种数据
- `selects first department by default` — 默认选中第一个节点
- `finds selected rule for node` — 选中节点后匹配对应 rule
- `handleSave calls routingApi.updateRule` — 保存操作验证

### 3. use-budget-alert-rules-page 测试扩展
- `filters rules by type` — 类型筛选
- `filters rules by status` — 状态筛选
- `filters rules by search keyword` — 搜索筛选
- `computes stats correctly` — 统计数据验证

## 文件列表

- `apps/backend/tests/domain/models/service_test.go` — 追加测试
- `apps/backend/tests/domain/budget/tree_test.go` — 新建继承测试
- `apps/frontend/tests/features/models/use-model-list-page.test.ts` — 重写
- `apps/frontend/tests/features/models/use-model-routing-page.test.tsx` — 重写
- `apps/frontend/tests/features/budget/use-budget-alert-rules-page.test.ts` — 扩展
- `apps/frontend/tests/fixtures/models.ts` — 更新 fixture 加 maxTokens
