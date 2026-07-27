# 站内通知中心 — 架构说明

> TokenJoy 站内通知系统技术架构与实现规范

---

## 1. 系统概览

通知中心由三层组成：

1. **投递层** — `infra/notification/` 下的 Service/Channel/Renderer，负责事件渲染与多渠道投递（in_app、email、sms）
2. **存储层** — `notification_log` 表 + `NotificationRepository` 接口，统一存储所有渠道投递记录
3. **展示层** — 前端 Popover（快捷预览）+ `/notifications` 页面（全功能管理）+ SSE 实时推送

```
Event 产生 → Dispatch → 渠道路由 → InAppChannel 写入 DB + SSE Push
                                  → EmailChannel 发邮件
                                  → 失败记录也写入 DB（send_ok=false）

前端 ← SSE 实时推送 ← SSEHub
前端 → REST API → notification_repo → notification_log
```

---

## 2. 数据模型

### 2.1 notification_log 表

```sql
CREATE TABLE IF NOT EXISTS notification_log (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id   UUID NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
    channel      TEXT NOT NULL,          -- in_app / email / sms / log
    event_type   TEXT NOT NULL,          -- budget_alert_reached / key_expired / ...
    user_id      UUID,                   -- 实际存 member_id
    title        TEXT NOT NULL DEFAULT '',
    body         TEXT NOT NULL DEFAULT '',
    payload      JSONB NOT NULL,
    send_ok      BOOLEAN NOT NULL DEFAULT true,
    error        TEXT,
    category     TEXT NOT NULL DEFAULT '',    -- budget_alert / key_expiration / ...
    group_key    TEXT NOT NULL DEFAULT '',    -- 分组键，空=不分组
    status       TEXT NOT NULL DEFAULT 'active',  -- active / archived / deleted
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    read_at      TIMESTAMPTZ,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

**关键设计**：
- `user_id` 存的是 `member_id`，不是 `users.id`。通知按 company 隔离，member 是用户在公司内的身份
- `send_ok` 布尔值替代 TEXT 状态，因为只有"成功/失败"两态
- `status` 控制用户侧生命周期（active → archived → deleted），替代 archived_at/deleted_at 双字段
- `read_at` 与 `status` 正交：归档的通知可以是未读的
- 所有状态变更同时更新 `updated_at`

### 2.2 索引策略

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

### 2.3 生命周期状态机

```
        新建 ──▶ active ──▶ archived ──▶ deleted（终态）
                   │                         ▲
                   └────────────────────────┘（可跳过 archived 直接删）

        active ← unarchive ← archived
        active ← undelete  ← deleted（恢复统一回 active）
```

- `read_at` 正交于生命周期：任何 status 下都可标记已读
- Undelete 恢复统一回 `active`（丢失"曾归档"信息，可接受）
- 所有删除为软删除，数据永久保留

---

## 3. API 端点

### 3.1 用户侧

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/notifications` | 列表查询（cursor 分页 + 分组） |
| GET | `/notifications/unread-count` | 未读数 |
| PATCH | `/notifications/{id}/read` | 标记已读 |
| POST | `/notifications/read-all` | 全部已读 |
| POST | `/notifications/{id}/archive` | 归档 |
| POST | `/notifications/archive-all` | 批量归档（支持 category 筛选） |
| POST | `/notifications/{id}/unarchive` | 取消归档 |
| POST | `/notifications/{id}/delete` | 软删除 |
| POST | `/notifications/{id}/undelete` | 撤销删除 |
| GET | `/notifications/capabilities` | 渠道能力查询 |
| GET | `/notifications/stream` | SSE 实时推送 |
| GET | `/notifications/preferences` | 通知偏好 |
| PUT | `/notifications/preferences` | 更新偏好 |
| POST | `/notifications/preferences/reset` | 重置偏好 |

### 3.2 管理员侧

| 方法 | 路径 | 权限 | 说明 |
|------|------|------|------|
| GET | `/notifications/admin/log` | audit:read | 全量日志查询（含失败记录） |
| GET | `/notifications/admin/stats` | audit:read | 按渠道统计 |
| POST | `/notifications/admin/test` | audit:read | 测试发送 |

### 3.3 GET /notifications 参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `limit` | int | 20 | 每页条数（max 100） |
| `cursor` | string | - | 游标（RFC3339Nano UTC 时间戳） |
| `category` | string | - | 按类别筛选 |
| `status` | string | all | `unread` / `read` / 空(all) |
| `archived` | bool | false | 查已归档列表 |
| `grouped` | bool | true | 是否分组返回 |
| `group_key` | string | - | 指定组内详情 |

### 3.4 响应体

```json
{
  "items": [{
    "id": "uuid",
    "eventType": "budget_alert_reached",
    "channel": "in_app",
    "category": "budget_alert",
    "title": "预算使用率达到 90%",
    "body": "项目 A 当前已用 ¥2,340 / ¥2,600",
    "payload": { "ruleID": "...", "projectID": "...", "threshold": 90 },
    "groupKey": "budget:rule-123:2026-07",
    "groupCount": 3,
    "status": "active",
    "createdAt": "2026-07-27T01:30:00Z",
    "readAt": null,
    "updatedAt": "2026-07-27T01:30:00Z"
  }],
  "nextCursor": "2026-07-26T06:00:00Z",
  "hasMore": true
}
```

注意：`nextCursor` 始终为 UTC 格式，避免 URL 中 `+` 号解析问题。

---

## 4. 分组机制

### 4.1 group_key 生成规则

由 InAppChannel 写入时从 event metadata 提取：

| 事件类型 | group_key 格式 | 效果 |
|----------|-----------|------|
| budget_alert_reached | `budget:{ruleID}:{period}` | 同规则同周期合并 |
| overrun_blocked | `overrun:{projectID}:{date}` | 同项目同天合并 |
| key_expiring_soon / key_expired | `key_expiry:{keyID}` | 同 Key 合并 |
| security_login_new_device | `""` | 不分组 |
| system_maintenance_scheduled | `maintenance:{eventID}` | 同事件合并 |
| usage_weekly_report | `""` | 不分组 |

### 4.2 分组查询逻辑

使用 PG `DISTINCT ON` 两步查询：

1. 找每组代表行（最新一条）
2. 对代表行做 cursor 分页
3. JOIN 回原表取完整字段 + 子查询计算 group_count

空 group_key 的条目用 `id::text` 作 fallback key，视为独立组。

---

## 5. SSE 实时推送

### 5.1 后端 SSEHub

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

### 5.2 前端处理

- `useNotificationConnection` hook 管理 EventSource 连接
- 收到事件 → invalidate TanStack Query → 弹 toast
- 同 groupKey 10s 内防刷屏（Map 记录最后展示时间）
- Toast 可点击跳转（前端 `getActionUrl` 计算目标路由）

---

## 6. 投递管道

### 6.1 流程

```
Event → Service.Dispatch()
  → Renderer.Render()：从 payload 提取 title/body，注入 _groupKey/_category
  → resolveChannels()：根据 priority + user preferences + rate limit 决定渠道
  → Channel.Send()：各渠道分别投递
  → recordFailure()：失败时写入 send_ok=false 的记录
```

### 6.2 渠道优先级

| Priority | 渠道链 |
|----------|--------|
| critical | sms → email → in_app |
| high | email → in_app |
| normal | in_app |
| low | in_app |

### 6.3 category 映射

由 `EventCategory()` 函数根据 event_type 确定：

| event_type | category |
|-----------|----------|
| budget_alert_reached, sync_threshold_exceeded | budget_alert |
| overrun_blocked, overdraft_expanded | overrun |
| key_expired, key_expiring_soon | key_expiration |
| usage_weekly_report | usage_report |
| security_login_new_device | security_event |
| system_maintenance_scheduled | system_maintenance |

---

## 7. 前端架构

### 7.1 模块结构

```
features/notifications/
├── components/
│   ├── notification-center.tsx      // /notifications 页面主体
│   ├── notification-list-item.tsx   // 单条通知组件
│   └── notification-empty-state.tsx // 空状态
├── hooks/
│   ├── use-notifications.ts         // Popover 数据（top 8）
│   ├── use-notification-inbox.ts    // 通知中心页面数据
│   └── use-notification-connection.ts // SSE 连接管理
├── lib/
│   ├── category-config.ts           // 类别→图标/颜色映射
│   ├── format-time.ts               // 时间格式化（相对时间）
│   └── get-action-url.ts            // event_type+payload→路由
├── notification-provider.tsx        // Context provider
└── index.ts                         // Barrel export
```

### 7.2 入口组件

- `NotificationInbox`（Popover）：放在 Header，显示最近 8 条（分组后）
- `NotificationCenter`（页面）：`/notifications` 路由，Tabs + 筛选 + 无限滚动

### 7.3 actionUrl 映射

前端 `getActionUrl(notification)` 根据 event_type + payload 计算跳转路由：

```ts
switch (notification.eventType) {
  case 'budget_alert_reached':
    return `/budget/alerts` // 带 ruleID query param
  case 'key_expired':
  case 'key_expiring_soon':
    return `/keys/platform`
  case 'overrun_blocked':
    return `/budget`
  // ...
}
```

---

## 8. 设计决策

| 决策 | 选择 | 理由 |
|------|------|------|
| 生命周期字段 | `status` 枚举 + `updated_at` | 替代双时间戳，索引更简洁 |
| 投递状态 | `send_ok` BOOLEAN | 只有两态，比 TEXT 更简洁 |
| 删除策略 | 软删除（status='deleted'） | 表混存多渠道记录，不能真删 |
| 分页 | cursor（基于 created_at UTC） | offset 在新增数据时漂移 |
| 分组 | DISTINCT ON + cursor | PG 原生，无需额外表 |
| actionUrl | 前端计算 | 路由变更不需动后端 |
| category | 入库持久化 | 索引直接命中，避免反查 |
| channel 隔离 | 用户操作限定 `channel='in_app'` | 不影响运维日志 |
| 多租户 | company_id + member_id 联合过滤 | 索引命中 + 数据隔离 |

---

## 9. UI 规范

### 9.1 通知类别视觉

| 类别 | 图标 (lucide-react) | 颜色 |
|------|------|---------|
| budget_alert | `TrendingUp` | `text-orange-500` |
| key_expiration | `Key` | `text-amber-500` |
| usage_report | `BarChart3` | `text-blue-500` |
| security_event | `ShieldAlert` | `text-red-500` |
| system_maintenance | `Settings` | `text-slate-500` |
| overrun | `AlertTriangle` | `text-rose-500` |

### 9.2 状态视觉

| 状态 | 样式 |
|------|------|
| 未读 | `border-l-2 border-blue-500` + `bg-blue-50/40` + `font-medium` |
| 已读 | 无边框 + 透明背景 + `font-normal text-muted-foreground` |
| hover | `bg-muted/60` + 浮出操作按钮 |

### 9.3 尺寸

| 元素 | 规格 |
|------|------|
| Popover 宽度 | 360px，max-h=400px |
| 列表页宽度 | max-w-3xl (768px) 居中 |
| 图标 | category 16px, action 14px |
| 文字 | title text-sm, body text-xs |
