# API 优化审查

> 2026-07-22 · 已关闭

## 已完成

| 事项 | 状态 |
|------|------|
| 合并两套审批系统为统一 `/approvals` 引擎 | ✅ |
| 前端审批路由从 `/keys/approval` 迁移到 `/approvals` | ✅ |
| 删除 `POST /auth/sms/send`、`/auth/sms/verify`、`/auth/sms/select` legacy aliases | ✅ |
| 删除 `GET /auth/invites/pending` 路由 + handler | ✅ |
| 删除 `GET /dashboard/usage/series` 路由 + handler + `UsageSeriesFromQuery` | ✅ |
| budget-summary 从 keys 域搬到 budget 域：`GET /budget/members/:memberId/summary` | ✅ |
| 合并 `accountApi` + `meApi` 为单一 `meApi`（`api/me.ts`），删除 `api/account.ts` + `api/member.ts` | ✅ |
| 清理 `memberanalytics` 对 keys domain 的死依赖 | ✅ |

---

## 不动

| 项目 | 理由 |
|------|------|
| `sessionApi` 只有一个端点 | session 是独立领域概念，合并进 meApi 语义模糊 |
| Org 5 个子 API 对象 | 改动面广，当前不痛 |
| Toggle 端点命名（PUT 语义） | 前端传绝对值 `enabled: bool`，不会因命名出 bug |
| Dashboard 并发请求数 | TanStack Query 已管理，无性能问题 |
| `POST /auth/refresh` | 被 `api/client.ts` token 刷新逻辑隐式调用，非死代码 |
| `GET /notifications/stream` | 被 `use-notification-connection.ts` SSE 连接使用，非死代码 |
| Notification admin 端点（3 个） | 有权限守卫，保留供未来通知管理后台使用 |

---

## 审查结论

所有可操作项已完成。`pnpm lint` + `tsc --noEmit` + `go build ./...` 全部通过。剩余项目经评估均为"不值得投入时间"。本轮审查关闭。
