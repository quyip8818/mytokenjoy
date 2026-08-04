# 结构化错误码

## 设计决策

| 决策 | 结论 |
|------|------|
| 谁生成用户可见消息？ | 后端。`message` 是最终展示文案，前端直接 toast |
| 前端用 code 做什么？ | 匹配 deep link action、触发特殊 UI（如确认弹窗） |
| 所有错误都要 code？ | 不。只有"前端要做不同事情"的错误才加 code |
| meta 用途？ | 前端逻辑判断用（如数值比较），每个 code 的 meta schema 在清单中定义 |
| 命名风格？ | UPPER_SNAKE_CASE，`{DOMAIN}_{SPECIFIC_ERROR}` |
| 多语言？ | 不考虑。项目只有中文，message 由后端硬编码中文 |

## HTTP 错误响应格式

```json
{
  "code": "BUDGET_RESERVED_POOL_INSUFFICIENT",
  "message": "预留池余额不足，当前剩余 0.00",
  "meta": { "remaining": 0, "requested": 3000 }
}
```

- `code` — 可选。无 code 的普通错误只返回 `{"message": "..."}`
- `message` — 必须。人类可读，前端直接展示
- `meta` — 可选。每个 code 的 meta schema 固定，见清单

## 后端实现

```go
// domain/errors.go
type DomainError struct {
    Status  int
    Code    string         // 空串 = 无 code，不输出
    Message string
    Meta    map[string]any // nil = 不输出
}

func Validation(msg string) *DomainError {
    return &DomainError{Status: StatusUnprocessable, Message: msg}
}

func ValidationCode(code, msg string, meta ...map[string]any) *DomainError {
    var m map[string]any
    if len(meta) > 0 { m = meta[0] }
    return &DomainError{Status: StatusUnprocessable, Code: code, Message: msg, Meta: m}
}

func ForbiddenCode(code, msg string, meta ...map[string]any) *DomainError { ... }
```

序列化逻辑（`httputil/write.go`）：Code 非空时输出 `code`，Meta 非 nil 时输出 `meta`。

## 前端实现

```ts
// api/client.ts
class ApiError extends Error {
  status: number
  code?: string
  meta?: Record<string, unknown>
}

// lib/toast.ts
// toast.error 接受 string 或 ApiError
// 传入 ApiError 时按 code 匹配 deep link，展示 error.message
toast.error(apiError)  // 推荐：自动处理 code + message
toast.error('手动消息') // 兜底：无 code 匹配
```

## 错误码清单

只收录前端需要特殊处理的错误。普通校验错误不需要 code。

### BUDGET — 预算

| Code | 触发场景 | message | action | meta |
|------|---------|---------|--------|------|
| `BUDGET_RESERVED_POOL_INSUFFICIENT` | 审批 member_budget，预留池不足 | 预留池余额不足，当前剩余 X | → /budget | `{remaining, requested}` |
| `BUDGET_DEPT_POOL_INSUFFICIENT` | 审批 project_budget，部门预留池不足 | 部门预留池余额不足，当前剩余 X | → /budget | `{remaining, requested}` |
| `BUDGET_PROJECT_UNALLOCATED_INSUFFICIENT` | 审批 project_member_budget，项目余额不足 | 项目未分配余额不足，当前剩余 X | → /budget | `{remaining, requested, projectId}` |
| `BUDGET_EXCEED_PARENT` | 修改部门预算超出上级 | 超出上级可分配预算，当前剩余约 X | → /budget | `{remaining, nodeId}` |
| `BUDGET_BELOW_ALLOCATED` | 修改部门预算低于已分配 | 部门预算不能低于已分配总额（...） | → /budget | `{allocated, nodeId}` |
| `BUDGET_DEPT_NOT_SET` | 部门没有额度 | 请先给该部门分配额度 | → /budget | `{nodeId}` |

### KEY — Key 管理

| Code | 触发场景 | message | action | meta |
|------|---------|---------|--------|------|
| `KEY_BUDGET_INSUFFICIENT` | 创建 key 时额度不足 | 额度不足 | → /budget | — |
| `KEY_MODEL_DISABLED` | 所选模型已停用 | 模型已停用 | → /models | `{modelId}` |
| `KEY_MODEL_NOT_FOUND` | 模型不存在 | 模型不存在 | → /models | `{modelId}` |

### TRIAL — 试用限制

| Code | 触发场景 | message | action | meta |
|------|---------|---------|--------|------|
| `TRIAL_MEMBER_LIMIT` | 超出试用成员上限 | 试用环境成员上限为 N 人，升级后可扩容 | → /upgrade | `{limit}` |
| `TRIAL_NO_TOPUP` | 试用不支持充值 | 试用环境不支持充值，升级后可使用 | → /upgrade | — |

## 不需要 code 的错误

前端只做 `toast.error(message)`，无需区分类型：

- 表单校验：amount must be positive, name is required
- 权限不足：403 有全局拦截
- 资源不存在：404 有全局兜底
- 资源冲突：model already exists

## 实现步骤

1. 后端：扩展 `DomainError` + `WriteError` 序列化
2. 后端：新增 `ValidationCode` / `ForbiddenCode` 构造函数
3. 后端：清单中所有错误一次性改用 code 版构造函数
4. 前端：`ApiError` 解析 code + meta
5. 前端：`toast.error` 支持接收 `ApiError`，按 code 匹配 deep link
