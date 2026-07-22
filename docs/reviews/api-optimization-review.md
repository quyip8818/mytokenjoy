# API 优化审查

## 当前 API 全景

| 域 | 前端文件 | 后端 handler | 端点数 |
|---|---|---|---|
| Auth | `auth.ts` | `auth/`, `register/` | 14 |
| Session | `session.ts` | `session/` | 1 |
| Me/Account | `account.ts`, `member.ts` | `me/` | 6 |
| Org | `org.ts` | `org/` | 26 |
| Budget | `budget.ts` | `budget/` | 16 |
| Keys | `keys.ts` | `keys/` | 17 |
| Models | `models.ts` | `models/` | 8 |
| Dashboard | `dashboard.ts` | `dashboard/` | 8 |
| Audit | `audit.ts` | `audit/` | 6 |
| Billing | `billing.ts` | `billing/` | 4 |
| Notification | `notification.ts` | `notification/` | 12 |
| Platform (SaaS) | — | `platform/` | 8 |
| Dev | `dev.ts` | `dev/` | 2 |
| Internal | — | `ingest/` | 2 |
| **合计** | | | **~130** |

---

## 问题与优化建议

### 1. 两套审批系统需合并

**问题**：`/keys/approvals/*` 和 `/budget/approvals/*` 是两套独立的审批流，但业务本质相同（成员申请额度/模型 → 管理员批准/拒绝）。

| 现有 | 端点 |
|---|---|
| Keys Approvals | `GET /keys/approvals`, `POST /keys/approvals`, `PUT /keys/approvals/:id/approve`, `PUT /keys/approvals/:id/reject`, `GET /keys/approvals/:id/budget-check` |
| Budget Approvals | `GET /budget/approvals`, `PUT /budget/approvals/:id` |

**建议**：统一到 `/approvals` 域或择一保留。如果 keys approval 是"成员自助申请 key"、budget approval 是"预算超支审批"，则至少后端用同一审批 engine；如果本质相同，直接合并。

**收益**：删除 ~4 个端点，前端减少一套 query/mutation。

---

### 2. `/keys/platform/budget-summary` 放错域

**问题**：`GET /keys/platform/budget-summary?memberId=xxx` 查的是成员预算信息，属于 budget 域，却挂在 keys 下面。

**建议**：移到 `GET /budget/members/:memberId/summary`。已有 `/budget/departments/:id/member-budgets` 等类似路径。

**收益**：keys 域更聚焦于密钥管理；budget 域数据集中。

---

### 3. Dashboard `usage/series` 只有后端，前端未接

**问题**：`GET /dashboard/usage/series` 后端已注册但前端 `dashboardApi` 没有对应调用。

**建议**：
- 如果已废弃 → 删除后端端点
- 如果计划使用 → 补上前端 API

---

### 4. 认证域遗留别名路由

**问题**：后端 auth handler 注册了 `/auth/sms/send`、`/auth/sms/verify`、`/auth/sms/select` 作为 legacy alias，但前端没有任何调用方使用 `auth/sms/` 路径。

**建议**：删除 3 个 legacy alias 路由。前端已全部使用 `/auth/verify-code/*`。

**收益**：减少 3 个冗余路由，消除歧义。

---

### 5. Notification admin 路由未暴露给前端

**问题**：后端 notification handler 有 `GET /notifications/admin/log`、`GET /notifications/admin/stats`、`POST /notifications/admin/test` 三个管理端点，前端没有对应 API client。

**建议**：
- 如果仅供运维/调试 → 保留但标记为 internal
- 如果需要前端展示 → 补上 `notificationApi` 调用

---

### 6. `accountApi` 和 `meApi` 可合并

**问题**：前端有两个 API 模块指向 `/me/*`：
- `account.ts` → `getProfile`, `changePassword`, `changePhone`, `changeEmail`, `revokeSessions`
- `member.ts` → `getDashboard`

两者都命中后端 `/me` handler，却分成两个文件/两个 DI 名称。

**建议**：合并为单一 `meApi`，包含 `getProfile`, `getDashboard`, `changePassword`, `changePhone`, `changeEmail`, `revokeSessions`。

**收益**：前端少一个 API module、`AppApis` 接口少一个成员。

---

### 7. `sessionApi` 只有一个端点，可内联到 session feature

**问题**：`session.ts` 只导出 `getCurrent()`，一行代码。

**建议**：合并进 `meApi`（改名为 `getSession`），或直接写在 session feature 内部的 query 函数里（如果只有一个消费点）。

**收益**：少一个文件；但如果多处使用则保持独立也无害，此项为低优先级。

---

### 8. `/auth/invites/pending` 前端未调用

**问题**：后端注册了 `GET /auth/invites/pending`，但前端没有任何 API 调用此端点（grep 无结果）。

**建议**：确认是否有外部调用方；否则删除。

---

### 9. Dashboard 域查询参数过度重复

**问题**：7 个 dashboard 端点都接受相同的 `CostQueryParams` + `departmentId` 参数。前端每次传相同 filter。

**建议**：考虑批量端点 `POST /dashboard/query`，body 里指定需要哪些 facets：

```json
{
  "period": "2024-01",
  "departmentId": "xxx",
  "facets": ["summary", "daily", "topConsumers", "modelUsage"]
}
```

一次请求返回多维数据，减少前端并发 7 个请求。

**收益**：网络 round-trip 从 N 降为 1（Dashboard 页面加载提速明显）。

**折中**：可保留细粒度 GET 用于局部刷新，新增聚合端点用于首屏。

---

### 10. Org 域子资源拆分过细

**问题**：`org.ts` 导出 5 个 API 对象（`dataSourceApi`, `syncApi`, `departmentApi`, `memberApi`, `roleApi`），全部指向后端同一个 `/org/` 前缀、同一个 handler。

**建议**：前端可以合并为 `orgApi` 一个对象（后端已经是单一 handler），用 namespace 分组：

```ts
export const orgApi = {
  dataSource: { ... },
  sync: { ... },
  departments: { ... },
  members: { ... },
  roles: { ... },
}
```

保持一个 barrel export，减少 `AppApis` 接口膨胀。

**收益**：`AppApis` 少 4 个顶层成员，DI mock 更简洁。

---

### 11. Toggle 端点统一语义

**问题**：`/keys/provider/:id/toggle`、`/keys/platform/:id/toggle`、`/models/:id/toggle` 都是 PUT + `{ enabled: bool }` body，但 toggle 暗示翻转当前值。

**建议**：重命名为更显式的 `PATCH /:id` + body `{ enabled: bool }`，或统一用现有 `PUT /:id` 里包含 enabled 字段。

**收益**：语义更清晰，减少 3 个端点（合并进 update）。

---

### 12. Platform 管理端点不需要前端 API client

**问题**：`/platform/*` 只有 SaaS 模式下的内部运营人员使用，且当前没有前端 API 文件。

**建议**：现状合理，无需变动。如果将来有管理界面再补。

---

## 优先级排序

| 优先级 | 建议 | 影响 | 难度 |
|---|---|---|---|
| **P0** | #4 删除 legacy `/auth/sms/*` | 消除歧义 | 极低 |
| **P0** | #8 删除未使用 `/auth/invites/pending` | 减少死代码 | 极低 |
| **P1** | #3 处理 `usage/series` | 代码一致性 | 低 |
| **P1** | #6 合并 `accountApi` + `meApi` | 简化前端 DI | 低 |
| **P1** | #2 迁移 `budget-summary` 到 budget 域 | 领域边界清晰 | 中 |
| **P2** | #1 合并审批系统 | 架构简化 | 高 |
| **P2** | #9 Dashboard 聚合查询 | 性能 | 中 |
| **P2** | #10 合并 org 子 API | 代码简洁 | 低 |
| **P3** | #7 合并 `sessionApi` | 微优化 | 极低 |
| **P3** | #11 Toggle → PATCH 合并 | API 设计规范 | 中 |
| **P3** | #5 Notification admin 决定去留 | 清理 | 低 |

---

## 总结

当前 API 约 130 个端点，结构清晰、域划分合理。主要优化方向：

1. **删除死代码**（#4, #8, 可能 #3）— 立即可做
2. **合并前端 API 模块**（#6, #10, #7）— 减少 DI 接口膨胀
3. **域边界修正**（#2, #1）— 让 budget 归 budget
4. **性能聚合**（#9）— 减少 Dashboard 页面网络请求
