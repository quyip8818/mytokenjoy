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
                            │ 3. Quiet Hours   │──▶ Email Channel (SMTP)
                            │ 4. Resolve Ch.   │──▶ SMS Channel (Twilio)
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

| 渠道 | 实现 | IsConfigured 条件 | 备注 |
|------|------|-------------------|------|
| `log` | slog.Info | 始终 | 审计日志，不面向用户 |
| `in_app` | 写 notification_log + SSE push | 始终 | 用户 Inbox 展示 |
| `webhook` | HTTP POST | `NOTIFY_WEBHOOK_URL` 非空 | 兼容旧逻辑 |
| `email` | net/smtp | `SMTP_HOST` + `SMTP_FROM` 非空 | 通过 RecipientResolver 查 member email |
| `sms` | Twilio REST API | `TWILIO_ACCOUNT_SID` + token + from 非空 | 通过 RecipientResolver 查 member phone |

### 优先级 Fallback 链

```
critical:  SMS → Email → In-App
high:      Email → In-App
normal:    In-App
low:       In-App
```

当某渠道未配置或用户关闭偏好时自动跳过，沿链路向下。critical 级别无渠道可用时强制 In-App。

## 用户偏好

存储在 `notification_preferences` 表，按 category × channel 的矩阵：

| 类别 | 含义 |
|------|------|
| `budget_alert` | 预算告警 |
| `key_expiration` | Key 到期 |
| `usage_report` | 用量报告 |
| `security_event` | 安全事件 |
| `system_maintenance` | 系统维护 |
| `overrun` | 超支通知 |

用户未设置偏好时使用默认值（见 `domain/notification/types.go` 中 `CategoryDefaultChannels`）。

## 前端集成

```
┌─────────────────────────────────────────────────┐
│  NotificationProvider (SSE 连接管理)             │
│  ├─ NotificationInbox (Bell + Popover)          │
│  └─ useNotify() hook (toast fallback)           │
└─────────────────────────────────────────────────┘
```

- **SSE 连接** — `NotificationProvider` 在用户登录后建立到 `/api/notifications/stream` 的 EventSource
- **收到通知** — invalidate TanStack Query 缓存 + toast 提示
- **Inbox** — Header 中的 Bell icon，Popover 展示通知列表，支持标记已读
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
│   ├── channel_email.go      # Email 渠道（SMTP）
│   ├── channel_sms.go        # SMS 渠道（Twilio）
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
│   ├── notification.ts          # API client
│   └── types/notification.ts    # DTO types
├── features/notifications/
│   ├── hooks/
│   │   ├── use-notification-connection.ts  # SSE EventSource
│   │   ├── use-notifications.ts            # 列表 + 未读数 query
│   │   ├── use-notify.ts                   # 统一入口 (toast fallback)
│   │   └── use-notification-capabilities.ts
│   ├── notification-provider.tsx
│   └── index.ts
├── components/layout/
│   └── notification-inbox.tsx   # Bell icon + Popover
└── routes/member/
    ├── notifications.tsx        # 偏好设置页
    └── hooks/use-notifications-page.ts
```

## API 端点

| Method | Path | 描述 |
|--------|------|------|
| GET | `/api/notifications` | 通知列表（分页） |
| GET | `/api/notifications/unread-count` | 未读数量 |
| PATCH | `/api/notifications/:id/read` | 标记已读 |
| POST | `/api/notifications/read-all` | 全部已读 |
| GET | `/api/notifications/capabilities` | 已配置渠道查询 |
| GET | `/api/notifications/stream` | SSE 实时推送 |
| GET | `/api/notifications/preferences` | 获取偏好 |
| PUT | `/api/notifications/preferences` | 更新偏好 |
| POST | `/api/notifications/preferences/reset` | 恢复默认 |
| GET | `/api/notifications/admin/log` | 投递日志（管理端） |
| GET | `/api/notifications/admin/stats` | 投递统计 |
| POST | `/api/notifications/admin/test` | 测试发送 |

## 环境变量

```env
# Webhook (可选)
NOTIFY_WEBHOOK_URL=

# Email (配置后自动启用 email channel)
SMTP_HOST=
SMTP_PORT=587
SMTP_USER=
SMTP_PASS=
SMTP_FROM=

# SMS (配置后自动启用 sms channel)
TWILIO_ACCOUNT_SID=
TWILIO_AUTH_TOKEN=
TWILIO_FROM_NUMBER=
```

## 数据库表

```sql
-- 通知日志（兼 In-App Inbox 存储）
notification_log (
  id, company_id, channel, event_type, recipient, user_id,
  title, body, payload, status, error, created_at, read_at
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
