# Notification Module 架构设计文档

> TokenJoy 通用通知模块设计，支持多渠道（Email、SMS、In-App、Toast）投递，用户自选通知偏好，后端未配置时自动降级为前端 Toast。

## 1. 背景与目标

### 1.1 现状

| 层级 | 当前能力 | 局限 |
|------|---------|------|
| Backend | `infra/notification/service.go` — 仅 log + webhook 两种 channel | 无 Email/SMS 真实投递、无用户偏好 |
| Frontend | `sonner` Toast 组件 | 仅临时提示，无持久化、无 Inbox |
| 共享 | `packages/contracts` 仅有 permissions | 无通知相关类型契约 |

### 1.2 设计目标

1. **多渠道投递** — Email、SMS、In-App（Inbox）、Toast，可独立启停
2. **用户偏好** — 用户可按通知类别选择接收渠道
3. **优雅降级** — 后端渠道未配置时，前端自动 fallback 到 Toast
4. **可扩展** — 新增渠道（Push、Slack、Webhook）只需实现一个 interface
5. **项目共享** — 通知类型、事件定义通过 `packages/contracts` 全项目复用

---

## 2. 行业调研：同行方案对比

| 平台 | 核心模式 | 我们可借鉴的点 |
|------|---------|---------------|
| [Novu](https://novu.co) | Workflow 编排引擎 + Provider 抽象 + Inbox 组件 | Provider 插件化、Workflow step 概念、嵌入式 Inbox |
| [Knock](https://knock.app) | Trigger → Workflow → Channel Steps + Preference Center | 用户偏好 per-category、Digest/Batch、Channel Fallback |
| [SuprSend](https://suprsend.com) | Unified API + Template Engine + Preference Matrix | 按类别 × 渠道的偏好矩阵、渠道优先级链 |

### 行业共识的关键设计模式

```
┌─────────────────────────────────────────────────────┐
│  Strategy Pattern  — 渠道选择                        │
│  Factory Pattern   — Provider 实例化                 │
│  Observer/Event    — 事件驱动解耦                    │
│  Chain of Resp.    — Fallback 降级链                 │
│  Repository        — 通知日志持久化                  │
└─────────────────────────────────────────────────────┘
```

---

## 3. 整体架构

```
                    ┌─────────────────────────────────────┐
                    │         Event Producer              │
                    │  (业务代码 emit NotificationEvent)   │
                    └──────────────┬──────────────────────┘
                                   │
                                   ▼
                    ┌─────────────────────────────────────┐
                    │       Notification Service          │
                    │  ┌───────────────────────────────┐  │
                    │  │  1. Resolve Recipient         │  │
                    │  │  2. Load User Preferences     │  │
                    │  │  3. Select Channels           │  │
                    │  │  4. Execute Delivery Pipeline │  │
                    │  └───────────────────────────────┘  │
                    └──────────────┬──────────────────────┘
                                   │
                    ┌──────────────┼──────────────────────┐
                    │              │                      │
                    ▼              ▼                      ▼
             ┌───────────┐  ┌───────────┐        ┌───────────┐
             │   Email   │  │    SMS    │        │  In-App   │
             │  Channel  │  │  Channel  │  ...   │  Channel  │
             └─────┬─────┘  └─────┬─────┘        └─────┬─────┘
                   │              │                      │
                   ▼              ▼                      ▼
             ┌───────────┐  ┌───────────┐        ┌───────────┐
             │ Provider  │  │ Provider  │        │   Store   │
             │ (SendGrid │  │ (Twilio/  │        │ + WebSocket│
             │  /SES/...)│  │  Vonage)  │        │  Push      │
             └───────────┘  └───────────┘        └───────────┘

                              Frontend
                    ┌─────────────────────────────────────┐
                    │  Notification Client (React)        │
                    │  ┌──────────┐  ┌─────────────────┐  │
                    │  │  Toast   │  │  Inbox Center   │  │
                    │  │ (sonner) │  │  (持久化列表)    │  │
                    │  └──────────┘  └─────────────────┘  │
                    │       ▲ fallback       ▲ primary    │
                    │       └────────────────┘            │
                    └─────────────────────────────────────┘
```

---

## 4. 核心概念模型

### 4.1 NotificationEvent（事件）

业务代码触发的通知意图，与渠道无关。

```typescript
// packages/contracts/src/notification.ts
interface NotificationEvent {
  eventType: string;           // e.g. "budget_alert", "key_expired"
  recipientId: string;         // user ID
  tenantId: string;            // org/tenant scope
  payload: Record<string, any>;
  metadata?: {
    deduplicationKey?: string;
    priority?: 'critical' | 'high' | 'normal' | 'low';
    groupKey?: string;         // for digest/batch
  };
}
```

### 4.2 Channel（渠道）

```go
// internal/domain/notification/channel.go
type Channel interface {
    Name() string
    IsConfigured() bool
    Send(ctx context.Context, msg RenderedMessage) error
}
```

### 4.3 UserPreference（用户偏好）

```typescript
// 偏好矩阵: 每个事件类别 × 渠道
interface NotificationPreference {
  userId: string;
  preferences: {
    [category: string]: {
      email: boolean;
      sms: boolean;
      inApp: boolean;
    };
  };
  globalMute: boolean;
  quietHours?: { start: string; end: string; timezone: string };
}
```

### 4.4 渠道优先级与 Fallback 链

```
critical:  SMS → Email → In-App → Toast
high:      Email → In-App → Toast  
normal:    In-App → Toast
low:       In-App only (batch/digest)
```

**降级规则**: 当某渠道 `IsConfigured() == false` 时，自动跳过该渠道，沿链路向下传递。如果所有后端渠道均未配置，最终由前端 Toast 兜底。

---

## 5. 后端设计（Go）

### 5.1 包结构

```
apps/backend/internal/
├── domain/
│   └── notification/
│       ├── event.go           // NotificationEvent, Category 定义
│       ├── preference.go      // UserPreference 领域模型
│       └── types.go           // Channel 枚举, Status, Priority
├── infra/
│   └── notification/
│       ├── service.go         // 核心调度器 (现有，待扩展)
│       ├── channel_email.go   // Email Channel 实现
│       ├── channel_sms.go     // SMS Channel 实现
│       ├── channel_inapp.go   // In-App Channel (写 DB + push)
│       ├── channel_log.go     // Log Channel (现有)
│       ├── channel_webhook.go // Webhook Channel (现有)
│       ├── renderer.go        // 模板渲染
│       ├── preference.go      // 偏好加载逻辑
│       └── registry.go        // Channel 注册表
└── store/
    └── notification_repo.go   // NotificationRepository 扩展
```

### 5.2 Service 核心流程

```go
func (s *Service) Dispatch(ctx context.Context, event NotificationEvent) error {
    // 1. 解析接收者
    recipient, err := s.resolveRecipient(ctx, event.RecipientID)
    if err != nil {
        return err
    }

    // 2. 加载用户偏好
    prefs, err := s.preferenceStore.Get(ctx, event.RecipientID)
    if err != nil {
        prefs = defaultPreferences() // 降级为默认偏好
    }

    // 3. 确定目标渠道 (偏好 ∩ 已配置渠道)
    channels := s.resolveChannels(event, prefs)

    // 4. 渲染消息
    msg, err := s.renderer.Render(event)
    if err != nil {
        return err
    }

    // 5. 逐渠道投递 (可并行, 通过 riverqueue 异步)
    for _, ch := range channels {
        if err := s.enqueueDelivery(ctx, ch, recipient, msg); err != nil {
            s.logger.Warn("channel delivery failed, trying next",
                "channel", ch.Name(), "error", err)
            continue
        }
        break // 如果是 fallback 模式则 break；如果是 fan-out 模式则 continue
    }

    return nil
}
```

### 5.3 Channel Registry（注册表模式）

```go
type Registry struct {
    channels map[string]Channel
}

func NewRegistry(cfg config.Config) *Registry {
    r := &Registry{channels: make(map[string]Channel)}

    // 按配置动态注册
    if cfg.SMTPHost != "" {
        r.Register(NewEmailChannel(cfg))
    }
    if cfg.TwilioSID != "" {
        r.Register(NewSMSChannel(cfg))
    }
    // In-App 始终注册 (依赖 DB, 无需外部配置)
    r.Register(NewInAppChannel(store))

    return r
}

func (r *Registry) Configured() []Channel { ... }
func (r *Registry) Get(name string) (Channel, bool) { ... }
```

### 5.4 异步投递 (RiverQueue)

利用已有的 riverqueue 基础设施，将通知投递作为 background job：

```go
type DeliveryJob struct {
    Channel     string
    RecipientID string
    Message     RenderedMessage
    Attempt     int
    MaxRetries  int
}

// Worker 处理 — 自带重试、死信队列
func (w *DeliveryWorker) Work(ctx context.Context, job *river.Job[DeliveryJob]) error {
    ch, ok := w.registry.Get(job.Args.Channel)
    if !ok {
        return fmt.Errorf("channel %s not registered", job.Args.Channel)
    }
    return ch.Send(ctx, job.Args.Message)
}
```

---

## 6. 前端设计（React）

### 6.1 模块结构

```
apps/frontend/src/features/notifications/
├── components/
│   ├── notification-inbox.tsx       // Inbox 面板 (bell icon + dropdown)
│   ├── notification-item.tsx        // 单条通知渲染
│   ├── notification-preferences.tsx // 偏好设置 UI
│   └── notification-toast-bridge.tsx // Toast fallback 桥接
├── hooks/
│   ├── use-notifications.ts         // 通知列表 query + realtime
│   ├── use-notification-preferences.ts
│   └── use-notification-connection.ts // WebSocket/SSE 连接
├── lib/
│   ├── notification-client.ts       // API client
│   └── notification-store.ts        // Zustand store
└── index.ts
```

### 6.2 降级策略（核心亮点）

```typescript
// notification-toast-bridge.tsx
import { toast } from 'sonner';

/**
 * 前端通知分发器
 * - 后端 In-App channel 已配置 → 通过 WebSocket 接收, 展示在 Inbox
 * - 后端未配置任何渠道 → 前端直接 toast 展示
 */
export function useNotificationDispatch() {
  const { isBackendConnected } = useNotificationConnection();

  const notify = useCallback((event: NotificationEvent) => {
    if (isBackendConnected) {
      // 后端会处理投递, 前端通过 WS 收到后展示在 Inbox
      return;
    }

    // Fallback: 后端未配置, 直接 toast
    toast[mapPriorityToToastType(event.metadata?.priority)](
      event.payload.title,
      { description: event.payload.body }
    );
  }, [isBackendConnected]);

  return { notify };
}
```

### 6.3 实时推送连接

```typescript
// use-notification-connection.ts
export function useNotificationConnection() {
  const [isConnected, setIsConnected] = useState(false);
  const queryClient = useQueryClient();

  useEffect(() => {
    // 尝试建立 SSE 连接 (比 WebSocket 更轻量, 适合单向推送)
    const eventSource = new EventSource('/api/notifications/stream');

    eventSource.onopen = () => setIsConnected(true);
    eventSource.onerror = () => setIsConnected(false);

    eventSource.addEventListener('notification', (e) => {
      const notification = JSON.parse(e.data);
      // 更新本地缓存
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
      // 同时 toast 提示新消息
      toast.info(notification.title);
    });

    return () => eventSource.close();
  }, []);

  return { isBackendConnected: isConnected };
}
```

### 6.4 Inbox UI 组件

```typescript
// notification-inbox.tsx — 概念示例
export function NotificationInbox() {
  const { data: notifications, isLoading } = useNotifications();
  const { markAsRead, markAllAsRead } = useNotificationActions();

  const unreadCount = notifications?.filter(n => !n.readAt).length ?? 0;

  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="ghost" className="relative">
          <BellIcon className="h-4 w-4" />
          {unreadCount > 0 && (
            <Badge className="absolute -top-1 -right-1 h-4 w-4 p-0 text-[10px]">
              {unreadCount}
            </Badge>
          )}
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-80 p-0">
        <div className="flex items-center justify-between p-3 border-b">
          <span className="text-sm font-semibold">Notifications</span>
          <Button variant="ghost" size="sm" onClick={markAllAsRead}>
            Mark all read
          </Button>
        </div>
        <ScrollArea className="h-80">
          {notifications?.map(n => (
            <NotificationItem key={n.id} notification={n} onRead={markAsRead} />
          ))}
        </ScrollArea>
      </PopoverContent>
    </Popover>
  );
}
```

---

## 7. 用户偏好中心

### 7.1 偏好矩阵 UI

```
┌─────────────────────────────────────────────────────────────┐
│  Notification Preferences                                    │
├──────────────────────┬─────────┬───────┬────────┬───────────┤
│  Category            │  Email  │  SMS  │ In-App │   说明     │
├──────────────────────┼─────────┼───────┼────────┼───────────┤
│  Budget Alerts       │   ✓     │   ✓   │   ✓    │ 超预算通知 │
│  Key Expiration      │   ✓     │       │   ✓    │ Key 到期   │
│  Usage Reports       │   ✓     │       │   ✓    │ 周报/月报  │
│  Security Events     │   ✓     │   ✓   │   ✓    │ 安全告警   │
│  System Maintenance  │         │       │   ✓    │ 维护公告   │
└──────────────────────┴─────────┴───────┴────────┴───────────┘
```

### 7.2 后端存储

```sql
CREATE TABLE notification_preferences (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id),
    tenant_id   UUID NOT NULL,
    category    TEXT NOT NULL,
    channel     TEXT NOT NULL,   -- 'email' | 'sms' | 'in_app'
    enabled     BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ DEFAULT now(),
    updated_at  TIMESTAMPTZ DEFAULT now(),
    UNIQUE(user_id, category, channel)
);

CREATE TABLE notification_log (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    user_id     UUID NOT NULL,
    event_type  TEXT NOT NULL,
    channel     TEXT NOT NULL,
    status      TEXT NOT NULL,   -- 'pending' | 'sent' | 'failed' | 'read'
    payload     JSONB,
    error       TEXT,
    created_at  TIMESTAMPTZ DEFAULT now(),
    read_at     TIMESTAMPTZ
);

CREATE INDEX idx_notification_log_user ON notification_log(user_id, created_at DESC);
CREATE INDEX idx_notification_log_unread ON notification_log(user_id) WHERE read_at IS NULL;
```

---

## 8. API 设计

### 8.1 通知相关 API

| Method | Path | 描述 |
|--------|------|------|
| GET | `/api/notifications` | 获取当前用户通知列表 (分页) |
| GET | `/api/notifications/unread-count` | 未读数量 |
| PATCH | `/api/notifications/:id/read` | 标记已读 |
| POST | `/api/notifications/read-all` | 全部已读 |
| GET | `/api/notifications/stream` | SSE 实时推送 |

### 8.2 偏好相关 API

| Method | Path | 描述 |
|--------|------|------|
| GET | `/api/notification-preferences` | 获取用户偏好矩阵 |
| PUT | `/api/notification-preferences` | 更新偏好 |
| POST | `/api/notification-preferences/reset` | 恢复默认 |

### 8.3 管理端 API（Admin）

| Method | Path | 描述 |
|--------|------|------|
| GET | `/api/admin/notifications/log` | 查询投递日志 |
| GET | `/api/admin/notifications/stats` | 投递统计 |
| POST | `/api/admin/notifications/test` | 发送测试通知 |

---

## 9. 降级方案详解

### 9.1 降级决策流

```
                      ┌──────────────────┐
                      │ Event Triggered  │
                      └────────┬─────────┘
                               │
                      ┌────────▼─────────┐
                      │ Backend Channel  │
                      │   Configured?    │
                      └────────┬─────────┘
                         Yes ┌─┴─┐ No
                             │   │
                    ┌────────▼┐  ├────────────────┐
                    │ Dispatch │  │ Emit to Frontend│
                    │ via      │  │ via Response    │
                    │ Backend  │  │ Header/Meta     │
                    └─────────┘  └────────┬────────┘
                                          │
                                 ┌────────▼────────┐
                                 │  Frontend Toast  │
                                 │  (sonner)        │
                                 └─────────────────┘
```

### 9.2 前端如何感知后端状态

**方案 A: 启动时配置发现（推荐）**

```typescript
// 应用启动时请求一次后端能力
GET /api/notifications/capabilities
→ { channels: ["in_app"], email: false, sms: false }
```

前端据此决定：
- 有 `in_app` → 建立 SSE 连接, 展示 Inbox
- 无任何 channel → 所有通知走 Toast

**方案 B: 请求级 Header**

后端在 API response 中附带 `X-Notification-Mode: toast-only`，前端据此降级。

### 9.3 配置示例

```env
# .env — 后端渠道开关
NOTIFICATION_EMAIL_ENABLED=false
NOTIFICATION_SMS_ENABLED=false
NOTIFICATION_INAPP_ENABLED=true

# Email Provider (配置后自动启用)
SMTP_HOST=
SMTP_PORT=
SMTP_USER=
SMTP_PASS=

# SMS Provider
TWILIO_ACCOUNT_SID=
TWILIO_AUTH_TOKEN=
TWILIO_FROM_NUMBER=
```

---

## 10. 渐进式实施路线

### Phase 1: Toast 增强（1-2 天）

- [ ] 在 `packages/contracts` 定义 NotificationEvent 类型
- [ ] 前端封装 `useNotify()` hook，统一调用入口
- [ ] 所有现有 toast 调用迁移到 `useNotify()`
- [ ] 支持 priority 映射到 toast variant

### Phase 2: In-App Inbox（3-5 天）

- [ ] 后端新增 `notification_log` 表
- [ ] 实现 In-App Channel (写入 DB)
- [ ] SSE endpoint `/api/notifications/stream`
- [ ] 前端 Inbox 组件 + 未读 badge
- [ ] 通知列表 API + 标记已读

### Phase 3: 用户偏好（2-3 天）

- [ ] `notification_preferences` 表
- [ ] 偏好 CRUD API
- [ ] 前端偏好设置 UI（矩阵形式）
- [ ] Dispatch 流程集成偏好查询

### Phase 4: Email Channel（2-3 天）

- [ ] Email Channel 实现 (SMTP / SendGrid SDK)
- [ ] 邮件模板引擎 (Go `html/template`)
- [ ] RiverQueue 异步投递 + 重试
- [ ] 投递日志记录

### Phase 5: SMS Channel（1-2 天）

- [ ] SMS Channel 实现 (Twilio / Vonage)
- [ ] 短信模板 (短文本, 变量替换)
- [ ] 频率限制 (防止 SMS 滥用)

### Phase 6: 管理与可观测（2-3 天）

- [ ] Admin 投递日志面板
- [ ] 投递成功率统计
- [ ] 失败告警 (dogfooding — 通知自己)
- [ ] Digest/Batch 合并通知

---

## 11. 技术选型建议

| 关注点 | 推荐方案 | 备选 |
|--------|---------|------|
| Email Provider | AWS SES (低成本) | SendGrid, Resend |
| SMS Provider | Twilio | Vonage, AWS SNS |
| 实时推送 | SSE (单向，轻量) | WebSocket (双向需求时) |
| 异步队列 | RiverQueue (已有) | — |
| 模板引擎 (后端) | Go `html/template` | MJML + 预编译 |
| 前端状态 | Zustand + TanStack Query | — |
| 前端 Toast | Sonner (已有) | — |

---

## 12. 与现有代码的集成点

| 现有模块 | 集成方式 |
|---------|---------|
| `infra/notification/service.go` | 扩展为核心 Dispatcher，保留现有 log/webhook channel |
| `domain/types` — Notification | 扩展 Channel 枚举 (`email`, `sms`, `in_app`) |
| `store.NotificationRepository` | 扩展方法：List, MarkRead, GetUnreadCount |
| `sonner` Toaster | 保留作为最终 fallback；新增 Inbox 作为主要 In-App 渠道 |
| `riverqueue` | 新增 `DeliveryWorker` job type |
| `packages/contracts` | 新增 `notification.ts` — 共享事件类型和偏好接口 |

---

## 13. 关键设计决策

### Q: 为什么选 SSE 而不是 WebSocket？

通知推送是单向场景（server → client），SSE 原生支持自动重连，实现简单，与 Go 标准库 HTTP handler 兼容性好。如果未来需要双向交互（如 typing indicators），再升级 WebSocket。

### Q: 为什么偏好存在后端而不是前端 localStorage？

用户可能在多设备登录，偏好需要跨设备同步。后端存储还能与 Dispatch 逻辑在同一进程内完成，避免前端同步延迟。

### Q: 为什么不直接用 Novu/Knock 等 SaaS？

TokenJoy 作为 API 管理平台，通知量可控（非海量消息场景），自建更灵活、成本可控，且避免额外外部依赖。架构设计上借鉴了它们的 Provider 抽象和 Workflow 概念。

### Q: 如何防止通知轰炸？

- **Digest**: 同类事件在时间窗口内合并（如 5 分钟内多次 budget alert 合为一条）
- **Rate Limit**: 每用户每渠道每小时上限（如 SMS ≤ 5 条/小时）
- **Quiet Hours**: 用户可设置免打扰时段，该时段内仅 critical 通知不受限

---

## 14. 参考资料

- [Novu — How Novu Works](https://docs.novu.co/platform/how-novu-works) — 开源通知基础设施的架构说明
- [Knock — Introduction to Notification Infrastructure](https://knock.app/manuals/notification-infrastructure/introduction-to-notification-infrastructure) — 通知基础设施从简到繁的演进
- [SuprSend — Notification Preference Center](https://www.suprsend.com/post/notification-preference-center) — 用户偏好中心的 UX 设计模式
- [System Design — Notification Service](https://serhatgiydiren.com/system-design-interview-notification-service/) — 通知服务系统设计面试经典方案
- [Knock — Build vs Buy Notifications](https://knock.app/blog/build-v-buy-notifications) — 自建与外购的权衡分析

> Content was rephrased for compliance with licensing restrictions.
