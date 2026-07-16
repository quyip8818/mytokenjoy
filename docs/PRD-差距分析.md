# PRD 与现有实现差距分析（按产品分类）

> **对照基准**：[PRD.md](./PRD.md)  
> **代码快照**：2026-07-16  
> **图例**：✅ 已对齐 · ⚠️ 部分实现 · ❌ 未实现 · 🚫 PRD 已排除

---

## 一、组织管理（P1 平台初始化）

### 1.1 US-01 配置第三方平台凭证

| PRD 要求 | 状态 | 现状 |
| --- | --- | --- |
| 飞书凭证（App ID + App Secret） | ✅ | `integration/datasource/feishu` 完整；测试连接 + 搜索验证 + 保存 |
| 钉钉凭证（CorpID + AppKey + AppSecret） | ❌ | 前端类型 `Credential.dingtalk` 与表单已有；后端 `factory.ForPlatform` → `platform not supported` |
| 企微凭证（CorpID + Secret + AgentID） | ❌ | 同上，`types.PlatformWecom` 枚举已定义，无 Provider 实现 |
| 切换平台确认弹窗 | ✅ | 前端已实现 |
| 覆盖修改凭证二次确认 | ✅ | 前端已实现 |

**缺失：** 钉钉/企微后端 Provider 实现（`datasource/dingtalk/`、`datasource/wecom/`）。

---

### 1.2 US-02 全量导入组织架构

| PRD 要求 | 状态 | 现状 |
| --- | --- | --- |
| 一键全量导入（部门树 + 成员） | ✅ | `POST /org/data-source/import` |
| 增量合并 | ✅ | 飞书 Diff 合并 |
| 失败详情表格 + 单条/批量重试 | ✅ | `ImportResult.failures[]` + `POST /org/data-source/import/retry` |

**缺失：** 无。

---

### 1.3 US-03 定时同步策略

| PRD 要求 | 状态 | 现状 |
| --- | --- | --- |
| 同步频率/开始时间配置 | ✅ | `SyncConfig` + Worker `org_sync` |
| Diff（新增/移除/改名） | ✅ | |
| 删除保护阈值 → 终止同步 | ✅ | |
| 超阈值**通知超管**（手机 + 邮箱 + IM） | ⚠️ | 仅 `NOTIFY_WEBHOOK_URL`；通知 infra 有 Email/SMS Channel 但无 IM Bot；未在同步场景中多渠道触发 |
| 同步日志（时间 + 类型 + 结果 + 变更详情） | ✅ | `SyncLog` |
| 手动数据不受同步影响 | ✅ | `source` 字段区分 |

**缺失：** 同步保护通知 → 邮箱/IM 多渠道投递。

---

### 1.4 US-04 手动管理组织架构

| PRD 要求 | 状态 | 现状 |
| --- | --- | --- |
| 部门多级 CRUD + 搜索 | ✅ | |
| 成员 CRUD + 批量操作 | ✅ | |
| 停用成员 → Key 同步失效 | ✅ | |
| **邀请成员**（发链接 → 未激活 → 激活） | ⚠️ | `POST /org/members/invite` 写入 `pending` 态；后端 `accept-invite` handler 存在；**无真实邮件/短信投递**；前端无 `/invite/accept` 路由页 |
| 批量导入 | ✅ | `POST /org/members/batch-import` |

**缺失：**
1. 邀请真实投递渠道（邮件/短信发出邀请链接）
2. 前端邀请激活独立页（`/invite/accept` 路由 + API 封装）

---

### 1.5 US-05 角色与权限管理

| PRD 要求 | 状态 | 现状 |
| --- | --- | --- |
| 预设角色 + 自定义角色 | ✅ | `manifest.json` + role CRUD |
| 角色分配成员 + 普通成员保底不可移除 | ✅ | |
| 角色变更即时生效 | ✅ | `authz_revision` + PDP + 前端 stale 策略 |
| 权限集动态下发 | ✅ | `GET /session` → `permissions[]` |

**缺失：** 无。

---

## 二、预算管控（P2 资源管控配置）

### 2.1 US-07 逐级预算分配

| PRD 要求 | 状态 | 现状 |
| --- | --- | --- |
| Company → 部门 → 子部门逐级下发 | ✅ | |
| 不允许超卖 | ✅ | 事务 + 非负校验 |
| 自然月重置 | ✅ | |
| 成员级预算 + 预留池 + 追加审批 | ✅ | |
| Budget Group（虚拟项目组） | ✅ | `projects` CRUD + 独立 Key |

**缺失：** 无。

---

### 2.2 US-08 用量预警与超限策略

| PRD 要求 | 状态 | 现状 |
| --- | --- | --- |
| 配置多预警阈值 CRUD | ✅ | `alert_rules` + `overrun_policy.thresholds[]` |
| 运行时触发 80%/90% 预警通知 | ⚠️ | `CheckBudgetAlerts` 已在 ingest 后端调用 → `AlertPublisher` → `notification.Service.DispatchAsync`；**投递渠道仅 in-app + webhook**，邮件需 SMTP 配置，SMS 需 Twilio 配置，**IM 无实现** |
| 100% 阻断请求 | ✅ | Gateway `ErrBudgetExhausted` |
| 自定义阻断文案 `blockMessage` | ⚠️ | `overrun_policy.block_message` 存库；**Gateway 返回固定错误信息 `budget exhausted`，未读取该字段** |
| 通知方式：邮箱 + 手机 + IM | ⚠️ | Email/SMS Channel 代码就绪但依赖外部凭证配置；**IM（飞书/钉钉 Bot）未实现** |

**缺失：**
1. Gateway 消费 `blockMessage` 自定义文案返回给调用方
2. IM 通知渠道（飞书/钉钉/企微 Bot 投递）
3. 预警通知可达性保障（当 Email/SMS 未配时的降级策略）

---

### 2.3 US-09 模型白名单管理

| PRD 要求 | 状态 | 现状 |
| --- | --- | --- |
| 系统模型 + 企业自定义模型 | ✅ | |
| 白名单继承/自定义（只缩小不扩大） | ✅ | |
| 父级缩小 → 子级自动同步缩小 | ✅ | |
| API 未指定模型 → 错误 | ✅ | |
| 模型不在白名单 → 错误 | ✅ | |

**缺失：** 无。

---

## 三、密钥与审批（P3 成员接入与调用）

### 3.1 US-10 审批流（Key 申请 & 额度追加）

| PRD 要求 | 状态 | 现状 |
| --- | --- | --- |
| Key 申请 + 额度追加两种审批 | ✅ | |
| 通过 → 自动创建 Key / 扣预留池 | ✅ | |
| 拒绝（可填理由） | ✅ | |
| 预留池不足阻止通过 | ✅ | 422 + budget-check |
| **审批人 IM 通知** | ❌ | 无投递 |
| **申请结果 IM 通知申请人** | ❌ | 无投递 |

**缺失：** 审批全流程的 IM/邮件通知（提交 → 通知审批人；通过/拒绝 → 通知申请人）。

---

### 3.2 US-11 自主管理 Platform Key

| PRD 要求 | 状态 | 现状 |
| --- | --- | --- |
| 个人额度内自主创建多 Key | ✅ | |
| 选择绑定模型 + 分配额度 | ✅ | |
| 各 Key 独立计费 | ✅ | |
| Key 额度用完 → 该 Key 不可用 | ✅ | |
| 禁用/启用/重新生成/删除/编辑 | ✅ | |
| Key 脱敏展示 + 复制完整值 | ✅ | `keyPrefix` + `fullKey` 仅 create/rotate 返回一次 |

**缺失：** 无。

---

### 3.3 US-12 API 调用

| PRD 要求 | 状态 | 现状 |
| --- | --- | --- |
| OpenAI API 格式（chat/completions、completions、embeddings） | ✅ | `allowedGatewayPaths` 含 4 条路径 |
| **Anthropic API 格式（`/v1/messages`）** | ❌ | `allowedGatewayPaths` 不含 `/v1/messages`；未做一等公民支持 |
| Key 无效 → 401 | ✅ | |
| Key 禁用 → 403 | ✅ | |
| 模型不在绑定范围 → 403 | ✅ | |
| 额度不足 → 429 | ⚠️ | 返回 Gateway 通用 error（非 HTTP 429 状态码） |
| 供应商不可用 → 502 | ✅ | |
| 按实际 token 异步计费 | ✅ | Webhook → `usage_ledger` |

**缺失：**
1. Anthropic `/v1/messages` 路径白名单 + 请求/响应格式适配
2. 超限返回 PRD 定义的 HTTP 429 + 自定义 `blockMessage` 文案

---

## 四、运营与合规（P4）

### 4.1 US-13 成本看板

| PRD 要求 | 状态 | 现状 |
| --- | --- | --- |
| 指标卡（总花费 + 环比、平均单次、人均、调用次数） | ✅ | `GET /dashboard/cost/summary` |
| 花费趋势折线图（天/周/月粒度） | ✅ | `GET /dashboard/cost/daily` + `granularity` |
| 部门花费占比饼图 | ✅ | `GET /dashboard/cost/departments` |
| 部门 → 子部门 → 成员下钻 | ✅ | `parentId` + `/departments/:deptId/members` |
| 时间维度（本月/上月/近7天/自定义） | ✅ | `CostQueryParams.period` |

**缺失：** 无。另有 PRD 未单列的 `/dashboard/usage` 用量分析页（超出 PRD 范围）。

---

### 4.2 US-14 审计追踪

| PRD 要求 | 状态 | 现状 |
| --- | --- | --- |
| 操作审计（Key 增删、预算变更、权限变更等） | ✅ | `GET /audit/operations` |
| 调用审计（时间、调用人、Key、模型、Token、费用） | ✅ | `GET /audit/calls`（读 `usage_ledger`） |
| 筛选（时间、操作人、类型、关键词） | ✅ | |
| 导出 CSV | ✅ | 前端 `downloadCsv()` |
| **导出 Excel** | ❌ | 仅 CSV，无 xlsx 格式 |
| prompt + response 全文留存（可配置关闭） | ⚠️ | `AuditSettings.contentRetentionEnabled` 控制 `previewSnippet`（input 截断 ~200 字）；**不存 output 原文**；首版有意不提供全文 |
| 热存储 7 天 → **归档到对象存储** | ❌ | 全量存 Postgres；无 S3/OSS 归档管线 |
| 审计记录不可篡改、不可删除 | ⚠️ | Postgres 账本无 DELETE handler；但未做 WAL/WORM 级不可篡改保证 |
| 只读审计员权限 | ✅ | `audit:read` capability |

**缺失：**
1. Excel 导出格式
2. 对象存储归档管线（热 → 冷）
3. 调用全文留存（output 原文）

---

### 4.3 US-15 合规审查（敏感词）

| PRD 要求 | 状态 | 现状 |
| --- | --- | --- |
| 敏感词审查 | 🚫 | PRD 正文与附录**已明确排除产品范围** |

---

## 五、通知体系（横切能力）

PRD 在 US-03、US-08、US-10 多处要求通知。整体现状：

| PRD 要求渠道 | 后端基础设施 | 实际可达 |
| --- | --- | --- |
| 邮箱 | `EmailChannel`（SMTP） | ⚠️ 代码就绪，需配置 `EMAIL_HOST` 等环境变量 |
| 手机（短信） | `SMSChannel`（Twilio） | ⚠️ 代码就绪，需配置 `TWILIO_*` 环境变量 |
| IM（跟随数据源平台） | ❌ 无飞书/钉钉/企微 Bot Channel | ❌ |
| In-App（站内通知） | ✅ `InAppChannel` + SSE + 前端 `notificationApi` | ✅ |
| Webhook | ✅ `WebhookChannel` + `NOTIFY_WEBHOOK_URL` | ✅ |

**前端通知中心**已有：路由 `/me/notifications`、`notificationApi`（list / unreadCount / markRead / preferences）、SSE 实时推送。

**缺失汇总：**
1. IM Bot 投递渠道（飞书机器人/钉钉工作通知/企微应用消息）
2. Email/SMS 在生产环境的凭证配置与验证
3. 各业务场景（审批、预警、同步保护）→ 通知事件的完整触发点接入

---

## 六、SaaS 与平台运营

### 6.1 平台运营端

| PRD 要求 | 后端 | 前端 |
| --- | --- | --- |
| 平台登录 `POST /platform/auth/login` | ✅ | ❌ 无 `/platform/login` 路由 |
| 企业列表 / 创建 / 状态变更 | ✅ 8 端点已实现 | ❌ 无 `/platform/*` 页面 |
| 代充 / 赠送 / 调账 | ✅ | ❌ |
| 全局 Channel 管理 | ✅ | ❌ |

**缺失：** 整个平台运营控制台前端（路由 + 页面 + `platformApi`）。

---

### 6.2 企业面 SaaS 扩展

| 能力 | 状态 | 现状 |
| --- | --- | --- |
| 企业登录 + JWT Session | ✅ | `POST /auth/login` 前后端均接入 |
| 企业钱包 | ✅ | `/wallet` 路由 + `billingApi` |
| 邀请激活 `POST /auth/accept-invite` | ⚠️ | 后端已实现；前端无独立路由页 |
| 一人多企业 | ❌ | 未实现 |
| 企业自定义 Channel | ❌ | 未实现 |
| 真实支付渠道对接 | ❌ | 订单半真（`pending` → `confirm` 手动模拟） |
| `package_id` 自动改配额 | ❌ | MVP 仅展示 |

---

## 七、安全与技术债

| 项 | 状态 | 现状 |
| --- | --- | --- |
| OIDC / SSO | ❌ | 仅邮箱密码登录 |
| 密钥明文存储 | ⚠️ | Provider Key `key` 列；Platform Key 已改 `key_hash`（安全评估 H5） |
| Gateway HTTP 状态码规范化 | ⚠️ | PRD 定义 401/403/429/502；实际超限非标准 429 |

---

## 八、差距优先级汇总

### P0 — 上线阻塞

| # | 差距 | 关联 US | 说明 |
| --- | --- | --- | --- |
| 1 | Gateway 自定义 `blockMessage` 文案返回 | US-08 | 存库但未消费 |
| 2 | Anthropic `/v1/messages` 路径支持 | US-12 | PRD 明确要求双格式 |
| 3 | Gateway 超限返回 HTTP 429 | US-12 | 状态码规范化 |

### P1 — 核心体验

| # | 差距 | 关联 US | 说明 |
| --- | --- | --- | --- |
| 4 | 审批 IM/邮件通知（审批人 + 申请人） | US-10 | 审批无通知则流程断裂 |
| 5 | 预警通知真实到达（Email 配置 + IM Bot） | US-08 | Worker 已触发但渠道不通 |
| 6 | 邀请成员真实投递 + 前端激活页 | US-04 | API 存在但无渠道无页面 |
| 7 | 同步保护超阈值多渠道通知 | US-03 | 仅 Webhook |

### P2 — 产品完整性

| # | 差距 | 关联 US | 说明 |
| --- | --- | --- | --- |
| 8 | 钉钉 Provider 实现 | US-01 | 前端就绪，后端缺 |
| 9 | 企微 Provider 实现 | US-01 | 同上 |
| 10 | SaaS 平台运营前端 | 平台运营 | 后端 8 端点已有 |
| 11 | 审计 Excel 导出 | US-14 | 当前仅 CSV |
| 12 | IM Bot 通知渠道（飞书/钉钉/企微） | 横切 | 所有通知场景 |

### P3 — 长期演进

| # | 差距 | 关联 | 说明 |
| --- | --- | --- | --- |
| 13 | 审计归档（热存 → 对象存储） | US-14 | 全在 Postgres |
| 14 | 调用全文留存（output 原文） | US-14 | 首版有意不做 |
| 15 | OIDC / SSO | 安全 | |
| 16 | 真实支付渠道 | SaaS | |
| 17 | 一人多企业 | SaaS | |
| 18 | 企业自定义 Channel | SaaS | |

### 🚫 明确不做

- US-15 敏感词合规审查（PRD 已排除）

---

## 九、PRD 未要求但已实现的能力

| 能力 | 说明 |
| --- | --- |
| 成员工作台 `/me/*` | 3 路由 + `meApi`（PRD 隐含未单列） |
| 站内通知中心 | `/me/notifications` + SSE + 偏好管理 |
| NewAPI 拓扑与同步 | adminport → NewAPISync → Gateway 数据面 |
| Provider Key 管理 | `/keys/provider`（PRD 聚焦 Platform Key） |
| 用量时间序列 | `/dashboard/usage/series`（minute/hour/day 多粒度） |
| 企业钱包 lot 体系 | 双轴计费（point + 展示币） |
| Identity JWT + PDP | 强于 PRD 静态权限表描述 |
| Dev Popup 模拟消耗 | 仅开发环境 |
| 通知基础设施 | 完整的 dispatch → channel → SSE 管线（超出 PRD 纯 IM 描述） |

---

## 十、有意与 PRD 不同（避免误判为 bug）

| 主题 | PRD 表述 | 实现选择 | 理由 |
| --- | --- | --- | --- |
| 计费单位 | 人民币（元） | 内部 **point** + lot 钱包；UI `÷ PPU` 换算展示 | 精度与多币种扩展 |
| Key 存储 | `key_value` | `key_hash` 鉴权；`fullKey` 仅创建/轮转时返回一次 | 安全 |
| 超限行为 | 80%/90% 预警 + 100% 阻断 | 预警 Worker 已触发通知（in-app）；阻断为硬封 | 渠道暂缺≠逻辑缺 |
| 审批人 | 直属 TL | 拥有 `budget:approve` 权限的管理员 | 更灵活 |
| API 契约 | PRD 附录列举 | **权威**：[Frontend.md](./Frontend.md) §5 + `api/types/` | PRD 附录已声明 |

---

## 十一、相关文档

| 文档 | 用途 |
| --- | --- |
| [PRD.md](./PRD.md) | 产品需求（对照基准） |
| [Roadmap.md](./Roadmap.md) | 差距状态简表 |
| [Frontend.md](./Frontend.md) | 页面路由与 API 契约权威来源 |
| [Backend-架构.md](./Backend-架构.md) | 分层、NewAPI、Gateway |
| [权限管理.md](./权限管理.md) | 鉴权与 RBAC |
| [工程收口.md](./工程收口.md) | 联调/架构未完成项 |
