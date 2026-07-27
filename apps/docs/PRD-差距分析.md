# PRD 与现有实现差距分析

> **对照基准**：[PRD.md](./PRD.md)  
> **代码快照**：2026-07-27  
> **图例**：✅ 已对齐 · ⚠️ 部分实现 · ❌ 未实现 · 🚫 PRD 已排除

---

## 一、组织管理（P1 平台初始化）

### 1.1 US-01 配置第三方平台凭证

| PRD 要求 | 状态 | 现状 |
|---------|------|------|
| 飞书凭证（App ID + App Secret） | ✅ | `integration/datasource/feishu` 完整；测试连接 + 搜索验证 + 保存 |
| 钉钉凭证（CorpID + AppKey + AppSecret） | ❌ | 前端类型与表单已有；后端 `factory.ForPlatform` → `platform not supported` |
| 企微凭证（CorpID + Secret + AgentID） | ❌ | 同上，`types.PlatformWecom` 枚举已定义，无 Provider 实现 |
| 切换平台确认弹窗 | ✅ | |
| 覆盖修改凭证二次确认 | ✅ | |

**缺失：** 钉钉/企微后端 Provider 实现（`datasource/dingtalk/`、`datasource/wecom/`）。

---

### 1.2 US-02 全量导入组织架构

| PRD 要求 | 状态 | 现状 |
|---------|------|------|
| 一键全量导入（部门树 + 成员） | ✅ | `POST /org/data-source/import` |
| 增量合并 | ✅ | 飞书 Diff 合并 |
| 失败详情表格 + 单条/批量重试 | ✅ | `ImportResult.failures[]` + `POST /org/data-source/import/retry` |

**缺失：** 无。

---

### 1.3 US-03 定时同步策略

| PRD 要求 | 状态 | 现状 |
|---------|------|------|
| 同步频率/开始时间配置 | ✅ | `SyncConfig` + Worker `org_sync` |
| Diff（新增/移除/改名） | ✅ | |
| 删除保护阈值 → 终止同步 | ✅ | |
| 超阈值通知超管（邮箱 + IM） | ⚠️ | `NOTIFY_WEBHOOK_URL` + Email/SMS/InApp 渠道可达；**IM Bot 未实现** |
| 同步日志 | ✅ | `SyncLog` |
| 手动数据不受同步影响 | ✅ | `source` 字段区分 |

**缺失：** IM Bot 通知渠道。Email/SMS 已有 Channel 实现（Resend + 阿里云），配置后即可投递。

---

### 1.4 US-04 手动管理组织架构

| PRD 要求 | 状态 | 现状 |
|---------|------|------|
| 部门多级 CRUD + 搜索 | ✅ | |
| 成员 CRUD + 批量操作 | ✅ | |
| 停用成员 → Key 同步失效 | ✅ | |
| 邀请成员（发链接 → 未激活 → 激活） | ✅ | `CreateMember` → `sendInviteNotifications` 通过 SMS/Email 真实投递；前端 `/invite/accept` 路由页存在；`AcceptInvite` handler 完整 |
| 批量导入 | ✅ | `POST /org/members/batch-import` |

**缺失：** 无。邀请全链路已打通（SMS via 阿里云 + Email via Resend + 前端激活页）。

---

### 1.5 US-05 角色与权限管理

| PRD 要求 | 状态 | 现状 |
|---------|------|------|
| 预设角色 + 自定义角色 | ✅ | `manifest.json` + role CRUD |
| 角色分配成员 + 普通成员保底不可移除 | ✅ | |
| 角色变更即时生效 | ✅ | `authz_revision` + PDP + 前端 stale 策略 |
| 权限集动态下发 | ✅ | `GET /session` → `permissions[]` |

**缺失：** 无。

---

## 二、预算管控（P2 资源管控配置）

### 2.1 US-07 逐级预算分配

| PRD 要求 | 状态 | 现状 |
|---------|------|------|
| Company → 部门 → 子部门逐级下发 | ✅ | |
| 不允许超卖 | ✅ | 事务 + 非负校验 |
| 自然月重置 | ✅ | |
| 成员级预算 + 预留池 + 追加审批 | ✅ | |
| Budget Group（虚拟项目组） | ✅ | `projects` CRUD + 独立 Key |

**缺失：** 无。

---

### 2.2 US-08 用量预警与超限策略

| PRD 要求 | 状态 | 现状 |
|---------|------|------|
| 配置多预警阈值 CRUD | ✅ | `AlertRules` CRUD + 启用/禁用 + 按部门/项目配置 |
| 运行时触发预警通知 | ✅ | `CheckBudgetAlerts` → `AlertPublisher` → `notification.Service.DispatchAsync`；按 category `budget_alert` 默认走 Email + InApp |
| 100% 阻断请求 | ✅ | `OverrunService` 评估 → 禁用 Key → Gateway `ErrBudgetExhausted` |
| 自定义阻断文案 `blockMessage` | ⚠️ | `overrun_policy.block_message` 存库；**Gateway 返回固定 error string，未读取该字段** |
| 通知方式：邮箱 + IM | ⚠️ | Email Channel (Resend) 就绪；InApp + Webhook 可达；**IM Bot 未实现** |

**缺失：**

1. Gateway 消费 `blockMessage` 自定义文案返回给调用方
2. IM Bot 通知渠道

---

### 2.3 US-09 模型管理

| PRD 要求 | 状态 | 现状 |
|---------|------|------|
| 系统模型 + 企业自定义模型 | ✅ | `CreateModel` + `ToggleModel`（租户覆盖内置） |
| 白名单继承/自定义（只缩小不扩大） | ✅ | `ResolveRouting` + `UpdateRoutingRule` |
| 父级缩小 → 子级自动同步缩小 | ✅ | `ResolveDeptAllowedModelIDs` 递归解析 |
| API 未指定模型 → 错误 | ✅ | `EvaluateAt`: `model field is required` |
| 模型不在白名单 → 错误 | ✅ | `checkPlatformKey`: `model not allowed` |

**缺失：** 无。

---

## 三、密钥与审批（P3 成员接入与调用）

### 3.1 US-10 审批流（Key 申请 & 额度追加）

| PRD 要求 | 状态 | 现状 |
|---------|------|------|
| Key 申请 + 额度追加两种审批 | ✅ | `ApprovalTypeKey` / `ApprovalTypeMemberBudget` / `ProjectBudget` / `ProjectMemberBudget` |
| 通过 → 自动创建 Key / 扣预留池 | ✅ | Engine `OnApprovedTx` + `PostApprove` |
| 拒绝（可填理由） | ✅ | `Engine.Reject` + `rejectReason` |
| 预留池不足阻止通过 | ✅ | `PreApprove` 校验 |
| 审批人 IM/邮件通知 | ❌ | Engine 无通知 dispatch |
| 申请结果通知申请人 | ❌ | 同上 |

**缺失：** 审批全流程通知投递（提交 → 通知审批人；通过/拒绝 → 通知申请人）。Engine `PostApprove` 和 `Create` 均未调用 Notifier。

---

### 3.2 US-11 自主管理 Platform Key

| PRD 要求 | 状态 | 现状 |
|---------|------|------|
| 个人额度内自主创建多 Key | ✅ | |
| 选择绑定模型 + 分配额度 | ✅ | |
| 各 Key 独立计费 | ✅ | |
| Key 额度用完 → 该 Key 不可用 | ✅ | `OverrunService` 禁用单 Key |
| 禁用/启用/重新生成/删除/编辑 | ✅ | |
| Key 脱敏展示 + 复制完整值 | ✅ | `keyPrefix` + `fullKey` 仅 create/rotate 返回一次 |

**缺失：** 无。

---

### 3.3 US-12 API 调用

| PRD 要求 | 状态 | 现状 |
|---------|------|------|
| OpenAI API 格式（chat/completions、completions、embeddings） | ✅ | `allowedGatewayPaths` 含 4 条路径 |
| Anthropic API 格式（`/v1/messages`） | ❌ | `allowedGatewayPaths` 不含 `/v1/messages`；未做适配 |
| Key 无效 → 401 | ✅ | `http.StatusUnauthorized` |
| Key 禁用 → 403 | ✅ | Precheck `platform key inactive` → 403 |
| 模型不在绑定范围 → 403 | ✅ | Precheck `model not allowed` → 403 |
| 额度不足 → 429 | ⚠️ | 返回 HTTP **403** + `insufficient member or key quota`（PRD 要求 429） |
| 供应商不可用 → 502 | ✅ | Reverse proxy upstream error |
| 按实际 token 异步计费 | ✅ | Webhook → `usage_ledger` |

**缺失：**

1. Anthropic `/v1/messages` 路径白名单 + 请求/响应格式适配
2. 超限返回 HTTP 429（当前为 403）+ 自定义 `blockMessage` 文案

---

## 四、运营与合规（P4）

### 4.1 US-13 成本看板

| PRD 要求 | 状态 | 现状 |
|---------|------|------|
| 指标卡（总花费 + 环比、平均单次、人均、调用次数） | ✅ | `GET /dashboard/cost/summary` |
| 花费趋势折线图（天/周/月粒度） | ✅ | `GET /dashboard/cost/daily` + `granularity` |
| 部门花费占比饼图 | ✅ | `GET /dashboard/cost/departments` |
| 部门 → 子部门 → 成员下钻 | ✅ | `parentId` + `/departments/:deptId/members` |
| 时间维度（本月/上月/近7天/自定义） | ✅ | `CostQueryParams.period` |

**缺失：** 无。

---

### 4.2 US-14 审计追踪

| PRD 要求 | 状态 | 现状 |
|---------|------|------|
| 操作审计（Key 增删、预算变更、权限变更等） | ✅ | `GET /audit/operations` |
| 调用审计（时间、调用人、Key、模型、Token、费用） | ✅ | `GET /audit/calls`（读 `usage_ledger`） |
| 筛选（时间、操作人、类型、关键词） | ✅ | |
| 导出 CSV | ✅ | 前端 `downloadCsv()` |
| 导出 Excel | ❌ | 仅 CSV，无 xlsx 格式 |
| prompt + response 全文留存（可配置关闭） | ⚠️ | `AuditSettings.contentRetentionEnabled` 存在；input 截断存储；**不存 output 原文** |
| 热存储 7 天 → 归档到对象存储 | ❌ | 全量存 Postgres；无归档管线 |
| 审计记录不可篡改、不可删除 | ⚠️ | Postgres 无 DELETE handler；但未做 WORM 级不可篡改保证 |
| 只读审计员权限 | ✅ | `audit:read` capability |

**缺失：**

1. Excel 导出格式
2. 对象存储归档管线（热 → 冷）
3. 调用全文留存（output 原文）

---

### 4.3 US-15 合规审查（敏感词）

| PRD 要求 | 状态 | 现状 |
|---------|------|------|
| 敏感词审查 | 🚫 | PRD 正文与附录**已明确排除产品范围** |

---

## 五、通知体系（横切能力）

PRD 在 US-03、US-08、US-10 多处要求通知。整体现状：

| PRD 要求渠道 | 后端基础设施 | 实际可达 |
|-------------|------------|---------|
| 邮箱 | `EmailChannel`（Resend API） | ✅ 代码就绪，需配置 `RESEND_API_KEY` |
| 手机（短信） | `SMSChannel`（阿里云短信） | ✅ 代码就绪，需配置 `ALIYUN_SMS_*` |
| IM（跟随数据源平台） | ❌ 无飞书/钉钉/企微 Bot Channel | ❌ |
| In-App（站内通知） | ✅ `InAppChannel` + SSE + 前端 `notificationApi` | ✅ |
| Webhook | ✅ `WebhookChannel` + `NOTIFY_WEBHOOK_URL` | ✅ |

**前端通知中心**：路由 `/me/settings`（含通知偏好管理）、SSE 实时推送。

**通知事件覆盖情况**：

| 场景 | 状态 | 说明 |
|------|------|------|
| 成员邀请 | ✅ | `sendInviteNotifications` → SMS + Email 真实投递 |
| 预算预警 | ✅ | `CheckBudgetAlerts` → `AlertPublisher` → Email + InApp |
| 超限阻断 | ✅ | `notifyOverrun` → InApp + Webhook |
| 同步保护阈值 | ⚠️ | Webhook 可达；Email/SMS 需配置 |
| 审批提交/通过/拒绝 | ❌ | Engine 无通知 dispatch |

**缺失汇总：**

1. IM Bot 投递渠道（飞书机器人/钉钉工作通知/企微应用消息）
2. 审批流程全事件通知（提交、通过、拒绝）

---

## 六、SaaS 与平台运营

### 6.1 平台运营端

| PRD 要求 | 后端 | 前端 |
|---------|------|------|
| 平台登录 `POST /platform/auth/login` | ✅ | ❌ 无 `/platform/login` 路由 |
| 企业列表 / 创建 / 状态变更 | ✅ 8 端点已实现 | ❌ 无 `/platform/*` 页面 |
| 代充 / 赠送 / 调账 | ✅ | ❌ |
| 全局 Channel 管理 | ✅ | ❌ |

**缺失：** 整个平台运营控制台前端（路由 + 页面 + `platformApi`）。

---

### 6.2 企业面 SaaS 扩展

| 能力 | 状态 | 现状 |
|------|------|------|
| 企业登录 + JWT Session | ✅ | `POST /auth/login` 前后端均接入 |
| 企业钱包 | ✅ | `/billing` 路由 + `billingApi` |
| 邀请激活 | ✅ | 后端 + 前端 `/invite/accept` 均完整 |
| 一人多企业 | ✅ | `routeByMembership` → `select_company` 流程已实现 |
| 企业自定义 Channel | ❌ | 未实现 |
| 真实支付渠道对接 | ❌ | 订单半真（`pending` → `confirm` 手动模拟） |

---

## 七、安全与技术债

| 项 | 状态 | 现状 |
|---|------|------|
| OIDC / SSO | ❌ | 仅邮箱/手机 + 密码登录 + 验证码登录 |
| 密钥明文存储 | ⚠️ | Provider Key `key` 列；Platform Key 已改 `key_hash` |
| Gateway HTTP 状态码规范化 | ⚠️ | PRD 定义 429（超限）；实际返回 403 |

---

## 八、差距优先级汇总

### P0 — 上线阻塞

| # | 差距 | 关联 US | 说明 |
|---|------|---------|------|
| 1 | Gateway 自定义 `blockMessage` 文案返回 | US-08 | 存库但 Gateway 返回固定 error string |
| 2 | Anthropic `/v1/messages` 路径支持 | US-12 | PRD 明确要求双格式 |
| 3 | Gateway 超限返回 HTTP 429 | US-12 | 当前返回 403，需状态码规范化 |

### P1 — 核心体验

| # | 差距 | 关联 US | 说明 |
|---|------|---------|------|
| 4 | 审批通知（审批人 + 申请人） | US-10 | Engine 无通知 dispatch，审批流程断裂 |
| 5 | IM Bot 通知渠道 | 横切 | 所有通知场景均缺 IM |

### P2 — 产品完整性

| # | 差距 | 关联 US | 说明 |
|---|------|---------|------|
| 6 | 钉钉 Provider 实现 | US-01 | 前端就绪，后端缺 |
| 7 | 企微 Provider 实现 | US-01 | 同上 |
| 8 | SaaS 平台运营前端 | 平台运营 | 后端 8 端点已有 |
| 9 | 审计 Excel 导出 | US-14 | 当前仅 CSV |

### P3 — 长期演进

| # | 差距 | 关联 | 说明 |
|---|------|------|------|
| 10 | 审计归档（热存 → 对象存储） | US-14 | 全在 Postgres |
| 11 | 调用全文留存（output 原文） | US-14 | 首版有意不做 |
| 12 | OIDC / SSO | 安全 | |
| 13 | 真实支付渠道 | SaaS | |
| 14 | 企业自定义 Channel | SaaS | |

### 🚫 明确不做

- US-15 敏感词合规审查（PRD 已排除）

---

## 九、PRD 未要求但已实现的能力

| 能力 | 说明 |
|------|------|
| 成员工作台 `/me/*` | 3 路由（我的 Key / 我的用量 / 设置） |
| 站内通知中心 | SSE 实时推送 + 通知偏好管理 |
| NewAPI 拓扑与同步 | adminport → NewAPISync → Gateway 数据面 |
| Provider Key 管理 | `/keys/provider`（PRD 聚焦 Platform Key） |
| 用量时间序列 | `/dashboard/usage`（minute/hour/day 多粒度） |
| 企业钱包 lot 体系 | 双轴计费（point + 展示币） |
| Identity JWT + PDP | 强于 PRD 静态权限表描述 |
| 验证码登录（手机/邮箱） | PRD 仅提邮箱密码登录 |
| 一人多企业 | `select_company` 流程已实现 |
| 预算 Budget Group | 含项目级预算 + 项目成员预算审批 |
| Gateway 限流 | Per-Key Token Bucket（Redis） |
| 通知偏好与静默时段 | 用户级 channel × category 精细控制 |

---

## 十、有意与 PRD 不同（避免误判为 bug）

| 主题 | PRD 表述 | 实现选择 | 理由 |
|------|---------|---------|------|
| 计费单位 | 人民币（元） | 内部 **point** + lot 钱包；UI `÷ PPU` 换算展示 | 精度与多币种扩展 |
| Key 存储 | `key_value` | `key_hash` 鉴权；`fullKey` 仅创建/轮转时返回一次 | 安全 |
| 超限行为 | 80%/90% 预警 + 100% 阻断 | 预警通知已投递（Email + InApp）；阻断为禁用 Key | 渠道暂缺 IM ≠ 逻辑缺 |
| 审批人 | 直属 TL | 拥有 `budget:approve` 权限的管理员 | 更灵活 |
| 登录方式 | 邮箱密码 | 手机/邮箱 + 密码 + 验证码（双因子就绪） | 国内场景适配 |
| API 契约 | PRD 附录列举 | **权威**：[Frontend.md](./Frontend.md) §5 + `api/types/` | PRD 附录已声明 |
| Gateway 拒绝 | 不同错误不同 status code | 统一 403（除 401/rate limit 429） | 简化实现，P0 修复中 |

---

## 十一、相关文档

| 文档 | 用途 |
|------|------|
| [PRD.md](./PRD.md) | 产品需求（对照基准） |
| [Roadmap.md](./Roadmap.md) | 差距状态简表 |
| [Frontend.md](./Frontend.md) | 页面路由与 API 契约权威来源 |
| [Backend-架构.md](./Backend-架构.md) | 分层、NewAPI、Gateway |
| [权限管理.md](./权限管理.md) | 鉴权与 RBAC |
