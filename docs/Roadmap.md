# TokenJoy Roadmap

PRD 与当前实现的差距，以及计划优化项。**现状**见 [Backend.md](./Backend.md)（索引）、[Frontend.md](./Frontend.md)。

**图例：** ✅ 已实现 · ⚠️ 部分实现 · ❌ 未实现 · 🔥 破坏性替换（见 [权限管理.md](./权限管理.md)）

---

## 1. 组织与数据源

| PRD / 契约              | 状态 | 说明                                                |
| ----------------------- | ---- | --------------------------------------------------- |
| US-01 飞书凭证          | ✅   | `integration/datasource/feishu`                     |
| US-01 钉钉 / 企微       | ❌   | 类型与前端表单已支持；后端 `platform not supported` |
| US-02–03 导入与定时同步 | ✅   | 飞书全量/增量；Worker `org_sync`                    |
| US-03 删除保护阈值通知  | ⚠️   | 超阈值终止同步已实现；仅 Webhook 通知               |
| US-04 邀请成员激活      | ⚠️   | API 存邀请态；无真实邮件/短信                       |
| SaaS 企业超管邀请       | ✅   | `company_invites` + `accept-invite`                 |

---

## 2. 预算与预警

| PRD                           | 状态 | 说明                                          |
| ----------------------------- | ---- | --------------------------------------------- |
| US-07 逐级预算 / Budget Group | ✅   |                                               |
| US-08 阈值预警 80%/90%        | ❌   | `alert_rules` 仅 CRUD；无 Worker              |
| US-08 超限阻断文案            | ⚠️   | `overrun_policy.blockMessage` 仅存库          |
| 运行时超限封禁                | ⚠️   | `consumed >= budget` 硬封 Key；不读阈值百分比 |

---

## 3. 密钥与审批

| PRD                    | 状态 | 说明                            |
| ---------------------- | ---- | ------------------------------- |
| US-10 Key / 额度审批   | ✅   |                                 |
| US-10 IM 通知审批人    | ❌   |                                 |
| US-11 成员自主 Key     | ✅   |                                 |
| US-12 API 调用与 Relay | ⚠️   | 需 `NEW_API_ENABLED` + Relay 栈 |

---

## 4. 运营与审计

| PRD                       | 状态 | 说明                      |
| ------------------------- | ---- | ------------------------- |
| US-13 成本看板            | ✅   |                           |
| US-14 操作 / 调用审计     | ✅   | 调用审计读 `usage_ledger` |
| US-14 热存 → 对象存储归档 | ❌   | 账本全在 Postgres         |
| US-14 输出正文留存        | ❌   | 首版不提供 output 全文    |
| US-15 合规审查            | ❌   | 产品范围已排除            |

---

## 5. 通知

| 能力                   | 状态 | 说明                           |
| ---------------------- | ---- | ------------------------------ |
| 预警 / 超限 / 同步通知 | ⚠️   | `NOTIFY_WEBHOOK_URL` 出站 only |
| IM 跟随数据源平台      | ❌   |                                |

---

## 6. SaaS 与前端

| 能力                       | 后端 | 前端                                                    |
| -------------------------- | ---- | ------------------------------------------------------- |
| `POST /auth/login`         | ✅   | ✅                                                      |
| `POST /auth/logout`        | ✅   | ✅（`authApi.logout`）                                  |
| `POST /auth/accept-invite` | ✅   | ⚠️ 无独立页 / API 封装                                  |
| `/billing/*`               | ✅   | ⚠️ `/billing` 页；缺 `confirm`；`WalletView` 字段待对齐 |
| `/platform/*`              | ✅   | ❌                                                      |

---

## 7. 权限与鉴权

**目标架构**见 [权限管理.md](./权限管理.md) §11–§12。后端 identity 收口与 HTTP/DI 简化已完成（`internal/identity/`、`deps.Public` / `Protected` / `Platform`；middleware 与 handler 统一 `deps.Protected`）。

| 项                                | 状态 | 说明                                                   |
| --------------------------------- | ---- | ------------------------------------------------------ |
| 方案 B（Identity JWT + PDP）      | ✅   | `identity/sessiontoken` + `identity/authz`             |
| `manifest.json` 单一契约          | ✅   | `pnpm generate:permissions` + `manifest_test`          |
| `authz_revision` + PDP 缓存       | ✅   | LRU；BatchImport / import / sync 软删除均 bump         |
| 全部业务路由 Session + capability | ✅   | 统一 `ReadRoutes`                                      |
| UI stale 策略                     | ✅   | revision 头 / focus / broadcast / 403                  |
| `POST /api/auth/login`            | ✅   | 前后端均已接入                                         |
| Billing 前端                      | ⚠️   | `/billing` + `PermissionGate`；缺 `confirm` 与类型对齐 |
| 平台 JWT                          | ✅   | 独立 `PLATFORM_SESSION_SECRET`                         |
| E2E 完整登录流程 spec             | ⚠️   | `auth.ts` helper 已有；spec 仅未登录重定向             |

---

## 8. 架构演进（非阻塞）

| 项                   | 说明                                                                                   |
| -------------------- | -------------------------------------------------------------------------------------- |
| 消耗多写点           | Ingest 同时写 ledger + `used`/`consumed`/`usage_buckets`；有意设计，对账以 ledger 为准 |
| 部门 consumed rollup | 祖先含子孙花费；UI 需与此一致                                                          |
| 预算组 vs 成员轴     | Key 挂组时不走成员个人超限分支                                                         |

---

## 9. 其他

| 项                      | 状态 |
| ----------------------- | ---- |
| OIDC / SSO              | ❌   |
| 支付渠道真实对接        | ❌   |
| `package_id` 自动改配额 | ❌   |
| 一人多企业              | ❌   |
| 企业自定义 Channel      | ❌   |

---

## 10. 变更约定

1. 实现后更新本文状态或移除条目
2. API 变更同步 [Frontend.md](./Frontend.md) §5 + `api/types/`
3. 权限变更同步 [权限管理.md](./权限管理.md) + `manifest.json`
4. 产品新需求先更新 [TokenJoy-PRD.md](./TokenJoy-PRD.md)
