# Frontend PRD 与 API 差距

本文档记录 [TokenJoy-PRD.md](./TokenJoy-PRD.md) 与当前前端实现 / [Frontend-API契约.md](./Frontend-API契约.md) 之间的**剩余**差异。

**范围说明：**

- 对比基准：PRD User Story（US-01～US-14）及附录数据模型
- 实现基准：`apps/frontend/src/api/`、`api/types/`、MSW handlers
- **不含 US-15（合规审查）**

**状态图例：**

| 标记 | 含义 |
| --- | --- |
| ➕ 超前 | 契约/实现已有，PRD 未单独描述 |
| ➖ 范围外 | PRD 提及，但不应由前端管理 API 承担 |

**最近对齐（2026-06-25）：** US-07 成员额度与 BG Key、US-11 `modelWhitelist`、US-13 成本看板（环比 / 自定义日期 / 趋势粒度）、US-14 审计筛选（时间 / 操作人 / 关键词）均已写入契约与 Mock。细节见 [Frontend-API契约.md](./Frontend-API契约.md)。

---

## 1. 剩余差距总览

US-01～US-14 与契约**已对齐**（除下表）。此前 P0～P2 差距项已从本文档移除。

| 标记 | 主题 | 摘要 |
| --- | --- | --- |
| ➕ | 供应商密钥等超前能力 | `providerKeyApi`、`defaultModel` / `fallbackModel`、`revoke`、用量看板等 — PRD 附录「实现扩展」可继续补全 |
| ➖ | 租户创建 | 部署 / SaaS 超管，非管理端 API |
| ➖ | US-12 LLM 网关 | 独立服务，非 `src/api/` 管理面 |
| ➖ | IM/邮件通知 | 审批、预警的副作用，非 REST 资源 |

---

## 2. 已对齐的 User Story（摘要）

| PRD | 契约要点 |
| --- | --- |
| US-01～US-06 | 组织、同步、审批、模型白名单等与契约一致 |
| US-07 | `GET/PUT /budget/members`、预留池、`PlatformKey.budgetGroupId`、`member-quota-config` workflow |
| US-08～US-10 | 超限策略、白名单、Key 审批 |
| US-11 | `modelWhitelist[]`；额度用尽由 `quota <= used` 判断；PRD ER 已同步 |
| US-13 | `CostSummary` 四指标 `*Mom`；`CostQueryParams`（含 `custom`、粒度）；下钻 ✅ |
| US-14 | `from`/`to`、`operatorId`/`callerId`、`keyword`；留存开关与客户端导出 ✅ |

---

## 3. 数据模型说明（非差距）

以下为**有意设计**，无需再改契约：

| PRD / ER | 契约 | 说明 |
| --- | --- | --- |
| `MEMBER.personal_quota` | `MemberBudgetQuota` / `MemberQuotaSummary` | 额度不在 `Member` 实体上，经预算与 Key 侧路 API |
| `MEMBER.used_quota` | `MemberQuotaSummary.used` | 同上 |
| `MODEL_WHITELIST` | `RoutingRule` | 命名不同，部门级白名单语义一致 |

---

## 4. 契约超前于 PRD（可选文档补全）

以下已在 [Frontend-API契约.md](./Frontend-API契约.md) 实现；若 PRD 需完整 ER，可写入附录「实现扩展」：

| 能力 | 端点 / 类型 |
| --- | --- |
| 供应商密钥管理 | `providerKeyApi` |
| 路由默认 / 降级模型 | `RoutingRule.defaultModel`、`fallbackModel` |
| Platform Key 吊销 | `PUT /keys/platform/:id/revoke` |
| Demo Session | `GET /session?memberId=`（生产换真实鉴权） |
| 用量看板 | `dashboard/usage/*` |

---

## 5. PRD 有、契约合理不覆盖

| 项 | PRD 出处 | 原因 |
| --- | --- | --- |
| 租户创建 / 多租户行隔离 | P1 主线、§1.2 | 部署或 SaaS 运营后台，非业务管理 SPA |
| LLM API 网关（US-12） | OpenAI / Anthropic 兼容 | 独立网关服务；契约仅管理配置与观测 |
| 审批 / 预警 IM 通知 | US-03、US-08、US-10 | 异步通知，非前端 REST 资源 |
| 审计热存储 / 归档 / 留存周期 | US-14 存储策略 | 后端基础设施 |
| 同步定时触发 | US-03 | 服务端 cron；前端仅配置与手动 `trigger` |

---

## 6. 相关文档

| 文档 | 职责 |
| --- | --- |
| [TokenJoy-PRD.md](./TokenJoy-PRD.md) | 产品需求与验收标准 |
| [Frontend-API契约.md](./Frontend-API契约.md) | 当前前端 REST 契约（实现真相） |
| [Frontend-代码结构.md](./Frontend-代码结构.md) | 前端分层与 Mock 架构 |

出现新的 PRD 变更或契约扩展时，先更新契约与实现，再回本表维护「剩余差距」。
