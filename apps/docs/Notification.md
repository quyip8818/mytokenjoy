# Notification Module

> 多渠道通知系统，支持 Email、SMS、In-App（Inbox）、Toast 投递，用户自选通知偏好，后端未配置时自动降级为前端 Toast。

## 架构概览

```
  业务代码                    Notification Service                      渠道
 ─────────                  ────────────────────                    ──────
                            ┌──────────────────┐
 notifier.Send() ──────────▶│  Dispatch 管线    │
                            │                  │
                            │ 1. Render        │──▶ Log Channel (slog)
                            │ 2. Load Prefs    │──▶ In-App Channel (DB + SSE)
                            │ 3. Quiet Hours   │──▶ Email Channel (Resend)
                            │ 4. Resolve Ch.   │──▶ SMS Channel (阿里云)
                            │ 5. Rate Limit    │──▶ Webhook Channel (HTTP)
                            │ 6. Deliver       │
                            └──────────────────┘
                                    │
                                    ▼ (async mode)
                            ┌──────────────────┐
                            │  RiverQueue Job  │
                            │  per channel     │
                            └──────────────────┘
```

## 核心流程

1. **业务代码触发** — 通过 `types.Notifier` 接口调用 `Send(ctx, Notification{...})`
2. **渲染** — 从 payload 提取 title/body，或使用事件类型默认标题
3. **偏好加载** — 查 `notification_preferences` 表获取用户设置，无记录则用默认
4. **Quiet Hours** — 非 critical 通知在用户设定的免打扰时段内被静默
5. **渠道选择** — 按优先级 fallback 链 × 用户偏好 × 已配置渠道取交集
6. **Rate Limit** — SMS 5条/小时/用户，Email 20条/小时/用户
7. **投递** — 同步 `Dispatch()` 或异步 `DispatchAsync()`（通过 RiverQueue）

## 渠道

| 渠道      | 实现                           | IsConfigured 条件                        | 备注                                   |
| --------- | ------------------------------ | ---------------------------------------- | -------------------------------------- |
| `log`     | slog.Info                      | 始终                                     | 审计日志，不面向用户                   |
| `in_app`  | 写 notification_log + SSE push | 始终                                     | 用户 Inbox 展示                        |
| `webhook` | HTTP POST                      | `NOTIFY_WEBHOOK_URL` 非空                | 兼容旧逻辑                             |
| `email`   | Resend API                     | `RESEND_API_KEY` + `RESEND_FROM` 非空    | 通过 RecipientResolver 查 member email；模板托管在 Resend Dashboard |
| `sms`     | 阿里云短信（dysmsapi）         | `ALIYUN_SMS_ACCESS_KEY_ID` + secret + `ALIYUN_SMS_SIGN_NAME` 非空 | 通过 RecipientResolver 查 member phone；模板为阿里云短信模板 Code |

### 优先级 Fallback 链

```
critical:  SMS → Email → In-App
high:      Email → In-App
normal:    In-App
low:       In-App
```

当某渠道未配置或用户关闭偏好时自动跳过，沿链路向下。critical 级别无渠道可用时强制 In-App。

**已知缺口：** 生产触发点（预算预警 `usage_alert.go`、超限阻断 `overrun.go`→`Notifier.Send()`、组织同步保护 `handler_admin.go` 同款 `Send()`）构造 `Event`/`Notification` 时都**硬编码** `Metadata.Priority = PriorityNormal`，并非按事件严重程度（如超限阻断本应是 critical）动态设置。`normal` 对应 fallback 链只有 `[InApp]`——这些事件目前**只投递站内通知**，即使 Email/SMS 配置齐全也不会触发。`webhook` 不在任何优先级链里，永远不会被 Dispatch 命中。

## 用户偏好

存储在 `notification_preferences` 表，按 category × channel 的矩阵：

| 类别                 | 含义     |
| -------------------- | -------- |
| `budget_alert`       | 预算告警 |
| `key_expiration`     | Key 到期 |
| `usage_report`       | 用量报告 |
| `security_event`     | 安全事件 |
| `system_maintenance` | 系统维护 |
| `overrun`            | 超支通知 |

用户未设置偏好时使用默认值（见 `domain/notification/types.go` 中 `CategoryDefaultChannels`）。

## 前端集成

```
┌─────────────────────────────────────────────────┐
│  NotificationProvider (SSE 连接管理)             │
│  ├─ NotificationInbox (Bell + Popover)          │
│  ├─ NotificationCenter (/notifications 页面)    │
│  └─ useNotify() hook (toast fallback)           │
└─────────────────────────────────────────────────┘
```

- **SSE 连接** — `NotificationProvider` 在用户登录后建立到 `/api/notifications/stream` 的 EventSource
- **收到通知** — invalidate TanStack Query 缓存 + toast 提示（同 groupKey 10s 防刷屏）
- **Popover** — Header 中的 Bell icon（带边框 + 未读数字 badge），显示最近 8 条（分组后），hover 时时间戳替换为归档按钮
- **通知中心** — `/notifications` 独立页面，支持 Tabs（收件箱/已归档）、分类筛选、cursor 分页、归档/删除/撤销、全部已读
- **actionUrl** — 前端 `getActionUrl(notification)` 根据 event_type + payload 拼接跳转路由
- **降级** — 后端无 in_app channel 时（`capabilities` 返回无 in_app），所有通知走 toast

## 后端文件结构

```
internal/
├── domain/notification/
│   └── types.go              # 领域模型：Event, Channel, Priority, Category, Preference
├── infra/notification/
│   ├── service.go            # Service 构造、Notifier 接口实现
│   ├── dispatch.go           # Dispatch/DispatchAsync 管线逻辑
│   ├── channel.go            # Channel 接口定义
│   ├── registry.go           # Channel 注册表
│   ├── channel_log.go        # Log 渠道
│   ├── channel_inapp.go      # In-App 渠道（DB + SSE）
│   ├── channel_webhook.go    # Webhook 渠道
│   ├── channel_email.go      # Email 渠道（Resend）
│   ├── channel_sms.go        # SMS 渠道（阿里云）
│   ├── renderer.go           # 消息渲染
│   ├── recipient.go          # RecipientResolver（memberID → email/phone）
│   ├── ratelimit.go          # 频率限制
│   ├── quiethours.go         # 免打扰时段
│   └── sse_hub.go            # SSE 实时推送 Hub
├── infra/jobs/
│   └── kinds_notification.go # RiverQueue 异步投递 job
├── infra/river/workers/
│   └── notification_delivery.go
├── http/handler/notification/
│   ├── handler.go            # 路由注册
│   ├── handler_inbox.go      # 通知列表/已读/SSE/Capabilities
│   ├── handler_preferences.go # 偏好 CRUD
│   └── handler_admin.go      # 管理端日志/统计/测试发送
└── store/postgres/
    ├── notification_repo.go  # notification_log CRUD
    └── notification_preference_repo.go
```

## 前端文件结构

```
src/
├── api/
│   ├── notifications.ts             # API client (all endpoints)
│   └── types/notification.ts        # DTO types
├── features/notifications/
│   ├── components/
│   │   ├── notification-center.tsx      # /notifications 页面主体
│   │   ├── notification-list-item.tsx   # 单条通知组件
│   │   └── notification-empty-state.tsx # 空状态
│   ├── components/
│   │   └── notifications-page-shell.tsx # 偏好设置面板（渲染在设置页 tab 里）
│   ├── hooks/
│   │   ├── use-notification-connection.ts  # SSE EventSource
│   │   ├── use-notifications.ts            # Popover 列表 + 未读数
│   │   ├── use-notifications-page.ts       # 偏好设置页数据
│   │   ├── use-notification-inbox.ts       # 通知中心页面数据
│   │   ├── use-notification-capabilities.ts # 已配置渠道查询
│   │   └── use-notify.ts                   # toast fallback
│   ├── lib/
│   │   ├── category-config.ts         # 类别→图标/颜色映射
│   │   ├── format-time.ts            # 相对时间格式化
│   │   └── get-action-url.ts         # event_type+payload→路由
│   ├── notification-provider.tsx
│   └── index.ts                       # Barrel export
├── components/layout/
│   └── notification-inbox.tsx         # Bell icon + Popover
└── routes/notifications/
    └── index.tsx                       # 通知中心页面路由
```

## API 端点

| Method | Path                                   | 描述               |
| ------ | -------------------------------------- | ------------------ |
| GET    | `/api/notifications`                   | 通知列表（cursor 分页 + 分组） |
| GET    | `/api/notifications/unread-count`      | 未读数量           |
| PATCH  | `/api/notifications/:id/read`          | 标记已读           |
| POST   | `/api/notifications/:id/archive`       | 归档               |
| POST   | `/api/notifications/:id/unarchive`     | 取消归档           |
| POST   | `/api/notifications/:id/delete`        | 软删除             |
| POST   | `/api/notifications/:id/undelete`      | 撤销删除           |
| GET    | `/api/notifications/capabilities`      | 已配置渠道查询     |
| GET    | `/api/notifications/stream`            | SSE 实时推送       |
| GET    | `/api/notifications/preferences`       | 获取偏好           |
| PUT    | `/api/notifications/preferences`       | 更新偏好           |
| POST   | `/api/notifications/preferences/reset` | 恢复默认           |
| GET    | `/api/notifications/admin/log`         | 投递日志（管理端） |
| GET    | `/api/notifications/admin/stats`       | 投递统计           |
| POST   | `/api/notifications/admin/test`        | 测试发送           |

## 环境变量

```env
# Webhook (可选)
NOTIFY_WEBHOOK_URL=

# Email (配置后自动启用 email channel，Resend)
RESEND_API_KEY=
RESEND_FROM=

# SMS (配置后自动启用 sms channel，阿里云短信)
ALIYUN_SMS_ACCESS_KEY_ID=
ALIYUN_SMS_ACCESS_KEY_SECRET=
ALIYUN_SMS_SIGN_NAME=
ALIYUN_SMS_ENDPOINT=dysmsapi.aliyuncs.com
```

## 数据库表

```sql
-- 通知日志（兼 In-App Inbox 存储）
notification_log (
  id, company_id, channel, event_type, user_id,
  title, body, payload JSONB,
  send_ok BOOLEAN DEFAULT true, error,
  category, group_key,
  status DEFAULT 'active',  -- active / archived / deleted
  created_at, read_at, updated_at
)

-- 用户偏好矩阵
notification_preferences (
  id, company_id, user_id, category, channel, enabled,
  created_at, updated_at
  UNIQUE(company_id, user_id, category, channel)
)
```


## 扩展新渠道

1. 在 `infra/notification/` 创建 `channel_xxx.go`，实现 `Channel` 接口（Name/IsConfigured/Send）
2. 在 `config.go` 添加对应环境变量
3. 在 `service.go` 的 `NewService` 中 `registry.Register(NewXxxChannel(...))`
4. 在 `domain/notification/types.go` 添加 `ChannelXxx` 常量和 fallback 链配置
5. 前端 `contracts/notification/types.ts` 同步更新

无需改动 Dispatch 逻辑、HTTP handler 或前端 Inbox 组件。


---

## 数据模型详细规范

### notification_log 索引策略

```sql
-- 收件箱查询（partial index，只索引 active 记录）
CREATE INDEX idx_notification_log_inbox
    ON notification_log (company_id, user_id, created_at DESC)
    WHERE status = 'active';

-- 分类筛选
CREATE INDEX idx_notification_log_category
    ON notification_log (company_id, user_id, category, created_at DESC)
    WHERE status = 'active';

-- 分组聚合
CREATE INDEX idx_notification_log_group
    ON notification_log (company_id, user_id, group_key, created_at DESC)
    WHERE group_key != '' AND status = 'active';

-- 未读计数
CREATE INDEX idx_notification_log_unread
    ON notification_log (company_id, user_id)
    WHERE read_at IS NULL AND status = 'active';
```

### 生命周期状态机

```
新建 ──▶ active ──▶ archived ──▶ deleted（终态）
           │                         ▲
           └────────────────────────┘（可跳过 archived 直接删）

active ← unarchive ← archived
active ← undelete  ← deleted（恢复统一回 active）
```

- `read_at` 正交于生命周期：任何 status 下都可标记已读
- 所有删除为软删除，数据永久保留

### 分组机制

#### group_key 生成规则

| 事件类型 | group_key 格式 | 效果 |
|----------|-----------|------|
| budget_alert_reached | `budget:{ruleID}:{period}` | 同规则同周期合并 |
| overrun_blocked | `overrun:{projectID}:{date}` | 同项目同天合并 |
| key_expiring_soon / key_expired | `key_expiry:{keyID}` | 同 Key 合并 |
| security_login_new_device | `""` | 不分组 |
| system_maintenance_scheduled | `maintenance:{eventID}` | 同事件合并 |
| usage_weekly_report | `""` | 不分组 |

#### 分组查询

使用 PG `DISTINCT ON` 两步查询：找每组代表行（最新一条）→ cursor 分页 → JOIN 回原表 + 子查询 group_count。空 group_key 用 `id::text` 作 fallback key。

### SSE 实时推送

```go
type SSEEvent struct {
    ID        string          `json:"id"`
    EventType string          `json:"eventType"`
    Title     string          `json:"title"`
    Body      string          `json:"body"`
    GroupKey  string          `json:"groupKey,omitempty"`
    Category  string          `json:"category,omitempty"`
    Payload   json.RawMessage `json:"payload,omitempty"`
}
```

- 每用户一个 subscriber channel（按 member_id 订阅）
- InAppChannel 写入 DB 后同步 Publish 到 Hub
- 无持久化，断连重连后靠 REST API 补数据
- 前端同 groupKey 10s 内防刷屏

### GET /notifications 查询参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `limit` | int | 20 | 每页条数（max 100） |
| `cursor` | string | — | 游标（RFC3339Nano UTC） |
| `category` | string | — | 按类别筛选 |
| `status` | string | `""` | `unread` / `read` / 空表示全部 |
| `archived` | bool | false | 查已归档列表 |
| `grouped` | bool | true | 是否分组返回 |
| `group_key` | string | — | 指定组内详情 |

---

## 已实现通知事件

> `templates/*.html` 仅为本地备份参考，代码不直接渲染这些文件——真实 Email 模板托管在 Resend Dashboard，SMS 模板是阿里云短信模板 Code，代码里只维护 ID/Code 映射（`templates.go`）。
>
> 由于当前所有触发点都硬编码 `Metadata.Priority = normal`（见 §渠道"已知缺口"），下列"渠道"一栏描述的是**代码逻辑上打算投递的目标渠道**；实际运行时均**只会投递到 InApp**。

### 组织同步删除保护超阈值

**触发**：定时/手动同步 Diff 计算完毕，待删除成员数 > `deleteMemberThreshold` 或部门数 > `deleteDepartmentThreshold`。

**行为**：
1. 不执行任何变更
2. 写入 SyncLog（result=failure）
3. 调 `NotifySyncThresholdExceeded` 发送通知

**已知缺口**：`Send()` 调用未设置 `RecipientID`（默认零值 UUID），代码里并没有查询"本企业所有超级管理员+组织管理员"再逐个发送的逻辑；payload 里塞的 `notifyPhone`/`notifyEmail`/`notifyIm` 全代码库无处读取，是死数据。实际效果是产生一条收件人为空 UUID 的日志记录，真实用户看不到。

**代码入口**：`domain/org/core/notify.go` → `NotifySyncThresholdExceeded`。

### 预算预警

**触发**：Ingest commit 后 `CheckBudgetAlerts` 检测 touched department 的阈值。

**收件人解析**：按 `alert_rules.notify_role_ids` 解析收件人，category 为 `budget_alert`。

**代码入口**：`domain/budget/alert_publisher.go`。

### 超限阻断

**触发**：OverrunService 禁用 Key 后。

**代码入口**：`domain/budget/overrun.go` → `notifyOverrun`。

### 成员邀请

**触发**：CreateMember / BatchImport / BatchInvite。

**渠道**：SMS（阿里云）+ Email（Resend），含注册链接（此路径走 `SendDirect`，不经过优先级 fallback 链，不受上述缺口影响）。

**代码入口**：`domain/org/structure/member_mutate.go` → `sendInviteNotifications`。

---

## UI 规范

### 通知类别视觉

| 类别 | 图标 (lucide-react) | 颜色 |
|------|------|---------|
| budget_alert | `TrendingUp` | `text-orange-500` |
| key_expiration | `Key` | `text-amber-500` |
| usage_report | `BarChart3` | `text-blue-500` |
| security_event | `ShieldAlert` | `text-red-500` |
| system_maintenance | `Settings` | `text-slate-500` |
| overrun | `AlertTriangle` | `text-rose-500` |

### 铃铛按钮

- 按钮 h-9 w-9，`rounded-md border border-border`
- 未读 badge：`absolute -right-1 -top-1`，蓝色圆形，白字数字，>99 显示 "99+"

### 列表项状态

| 状态 | 样式 |
|------|------|
| 未读 | `border-l-[3px] border-l-primary` + `bg-accent/40` + `font-medium` |
| 已读 | 无边框 + 透明背景 + `font-normal text-muted-foreground` |
| hover | `bg-muted/60` + 时间戳 → 归档按钮 |

### actionUrl 映射

前端 `getActionUrl(notification)` 根据 event_type + payload 计算跳转路由：

| event_type | 路由 |
|-----------|------|
| budget_alert_reached | `/budget/alerts` |
| key_expired / key_expiring_soon | `/keys/platform` |
| overrun_blocked | `/budget` |
| sync_threshold_exceeded | `/org/data-source` |
