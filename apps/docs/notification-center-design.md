# 站内通知中心 — UI/UX 设计文档

> TokenJoy 站内通知系统完整设计方案
> 版本: v1.1 | 日期: 2026-07-27

---

## 1. 设计参考与灵感来源

### 1.1 行业标杆分析

| 产品 | 亮点 | 我们借鉴什么 |
|------|------|-------------|
| **Linear** | Split Inbox（左列表+右详情）、高密度排版、Snooze 暂缓机制 | 高密度列表设计、已读/未读对比度 |
| **GitHub** | 多维筛选（repo/reason/unread）、Done 归档、Save 收藏 | 归档即"完成"的心智模型、分类筛选 |
| **Vercel** | 极简黑白设计、Geist 字体、notification bell + slide panel | 极简视觉语言、popover 轻量设计 |
| **Stripe** | 每条通知带明确 action link、已读用字重区分 | 通知关联操作跳转、字重而非颜色区分状态 |
| **阿里云/腾讯云** | 左侧分类 Tab、详情弹窗 | 分类 Tab 筛选 |

### 1.2 核心设计原则

1. **信息密度优先** — 参考 Linear，单行承载更多信息，减少视觉噪音
2. **渐进式披露** — Popover 做快捷预览，独立页做深度管理
3. **操作零阻力** — hover 显示操作按钮
4. **分组不丢信息** — 同 group_key 显示最新一条 + 数量 badge，类别 dropdown 可筛选查看

---

## 2. 信息架构

```
┌─ Header ──────────────────────────────────────────────────────┐
│                                    [🔔 Bell + Badge]          │
│                                         │                     │
│                                    Popover Panel              │
│                                    (快捷预览)                  │
│                                         │                     │
│                                    "查看全部 →"               │
└────────────────────────────────────────│──────────────────────┘
                                         ▼
┌─ /notifications ─────────────────────────────────────────────┐
│                                                               │
│  通知中心页面（独立路由，全功能管理）                            │
│  ├── 收件箱 Tab                                               │
│  ├── 已归档 Tab                                               │
│  ├── 分类筛选 / 状态筛选（dropdown）                           │
│  └── 扁平列表（分组显示最新一条 + badge）                      │
│                                                               │
└───────────────────────────────────────────────────────────────┘

┌─ /settings (通知 Tab) ───────────────────────────────────────┐
│  通知偏好矩阵（已有，不改）                                    │
└───────────────────────────────────────────────────────────────┘
```

---

## 3. 铃铛 Popover（Header 快捷面板）

### 3.1 视觉规格

```
┌─────────────────────────────────────── 360px ──┐
│                                                 │
│  ┌─ Header ──────────────────────────────────┐ │
│  │  通知                        [全部已读]    │ │  h=44px
│  └───────────────────────────────────────────┘ │
│                                                 │
│  ┌─ Filter Chips ────────────────────────────┐ │
│  │  [全部] [未读●3]                           │ │  h=36px
│  └───────────────────────────────────────────┘ │
│                                                 │
│  ┌─ Notification List ───────────────────────┐ │
│  │                                            │ │
│  │  ┌─ Item (unread) ─────────────────────┐  │ │
│  │  │  🟠  预算告警 (×3)        2 分钟前   │  │ │  h=72px
│  │  │      项目 A 预算使用率达到 90%       │  │ │
│  │  └─────────────────────────────────────┘  │ │
│  │                                            │ │
│  │  ┌─ Item (unread) ─────────────────────┐  │ │
│  │  │  🔑  Key 到期              1 小时前  │  │ │
│  │  │      prod-gw 将于 7 天后到期         │  │ │
│  │  └─────────────────────────────────────┘  │ │
│  │                                            │ │
│  │  ┌─ Item (read) ──────────────────────┐   │ │
│  │  │  ⚙️  系统维护              3 天前   │   │ │
│  │  │      计划维护窗口已结束             │   │ │
│  │  └────────────────────────────────────┘   │ │
│  │                                            │ │
│  └─── ScrollArea max-h=400px─────────────────┘ │
│                                                 │
│  ┌─ Footer ─────────────────────────────────┐  │
│  │           查看全部通知 →                   │  │  h=40px
│  └───────────────────────────────────────────┘  │
│                                                 │
└─────────────────────────────────────────────────┘
```

### 3.2 交互细节

| 行为 | 设计 |
|------|------|
| 铃铛 Badge | 未读数 >0 显示红色圆点 + 数字（>99 显示 "99+"） |
| 列表条数 | 最多显示最近 8 条（分组后），不做无限滚动 |
| 未读样式 | 左侧 2px 实色竖线 + 标题 font-medium + 浅色背景 |
| 已读样式 | 无竖线、标题 font-normal、透明背景 |
| 点击通知 | 标记已读 → 跳转关联资源页面（前端 `getActionUrl` 计算） → 关闭 popover |
| 分组显示 | 同 group_key 只显示最新一条，右上角显示 "×3" badge |
| hover 态 | 背景色加深 (muted)，右侧浮现 "归档" 图标按钮 |
| "全部已读" | 仅标记收件箱内所有为已读，不归档 |
| "查看全部" | 路由跳转到 /notifications |
| 空状态 | 居中灰色文字 "暂无新通知" + 铃铛描边图标 |

### 3.3 通知类别视觉映射

| 类别 | 图标 | 颜色标记 |
|------|------|---------|
| budget_alert（预算告警） | `TrendingUp` | `text-orange-500` |
| key_expiration（Key 到期） | `Key` | `text-amber-500` |
| usage_report（用量报告） | `BarChart3` | `text-blue-500` |
| security_event（安全事件） | `ShieldAlert` | `text-red-500` |
| system_maintenance（系统维护） | `Settings` | `text-slate-500` |
| overrun（超支通知） | `AlertTriangle` | `text-rose-500` |

图标使用 lucide-react，尺寸 16px，放在通知条目左侧作为类别标识。

---

## 4. 通知中心页面（/notifications）

### 4.1 整体布局

参考 Linear 的 Inbox 设计：高密度、干净、操作内敛。

```
┌─────────────────────────────────────────────────────────────────┐
│  通知中心                                                        │
│                                                                  │
│  ┌─ Tabs ──────────────────────────────────────────────────────┐│
│  │  [收件箱 ●12]    [已归档]                                    ││
│  └──────────────────────────────────────────────────────────────┘│
│                                                                  │
│  ┌─ Toolbar ───────────────────────────────────────────────────┐│
│  │  类别: [全部 ▾]   状态: [全部 ▾]        [全部已读] [全部归档]││
│  └──────────────────────────────────────────────────────────────┘│
│                                                                  │
│  ┌─ Notification List (扁平，分组只显示最新一条 + badge) ──────┐│
│  │                                                              ││
│  │  ┌─ Item (unread, grouped ×3) ───────────────────────────┐  ││
│  │  │  🟠 预算告警                    ×3         2 分钟前    │  ││
│  │  │     项目 A 预算使用率达到 90%                          │  ││
│  │  │                                 [已读] [归档] [删除]   │  ││
│  │  └───────────────────────────────────────────────────────┘  ││
│  │                                                              ││
│  │  ┌─ Item (unread) ───────────────────────────────────────┐  ││
│  │  │  🔑 Key 到期                              1 小时前    │  ││
│  │  │     API Key "prod-gw" 将于 7 天后到期                  │  ││
│  │  │                                 [已读] [归档] [删除]   │  ││
│  │  └───────────────────────────────────────────────────────┘  ││
│  │                                                              ││
│  │  ┌─ Item (read) ─────────────────────────────────────────┐  ││
│  │  │  ⚙️ 系统维护                               3 天前     │  ││
│  │  │     计划维护窗口已结束                                  │  ││
│  │  │                                 [归档] [删除]           │  ││
│  │  └───────────────────────────────────────────────────────┘  ││
│  │                                                              ││
│  │  ... 无限滚动加载更多 ...                                    ││
│  │                                                              ││
│  └──────────────────────────────────────────────────────────────┘│
│                                                                  │
│  ┌─ Footer Link ───────────────────────────────────────────────┐│
│  │  ⚙️ 通知偏好设置                                             ││
│  └──────────────────────────────────────────────────────────────┘│
└──────────────────────────────────────────────────────────────────┘
```

### 4.2 列表项设计

#### 单条通知 — 未读态

```
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│  🟠  预算告警                                    2 分钟前   │
│      ───────────                                            │
│      项目 A 预算使用率达到 90%，当前已用 ¥2,340             │
│                                                             │
│  ┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄  hover 时浮出 ┄┄┄┄┄┄┄┄┄┄┄┄  │
│                                 [✓已读] [📦归档] [🗑删除]    │
│                                                             │
│ ← 2px blue-500 left border                                 │
└─────────────────────────────────────────────────────────────┘
```

视觉要素：
- **左边框**：2px solid blue-500，标识未读
- **类别图标**：16px，对应颜色
- **标题**：font-medium text-sm text-foreground
- **摘要**：text-xs text-muted-foreground，max 2 行 line-clamp
- **时间**：text-xs text-muted-foreground，右对齐
- **操作按钮**：ghost icon buttons，仅 hover 时显示（移动端始终显示）

#### 单条通知 — 已读态

```
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│  ⚙️  系统维护                                     3 天前    │
│      ────────                                               │
│      计划维护窗口已结束                                      │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

- 无左边框
- 标题 font-normal
- 整行 opacity 略低（text-muted-foreground/80）
- 背景透明（未读用 bg-blue-50/30 dark:bg-blue-950/10）

#### 分组通知 — 扁平显示

```
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│  🟠  预算告警                       ┌───┐       2 分钟前   │
│      ────────                       │ 3 │                  │
│      项目 A 预算使用率达到 90%       └───┘                  │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

- 同 group_key 只显示最新一条，右侧 badge 显示组内总数
- 不做折叠/展开。点击通知直接跳转资源页面
- 如需查看同组历史通知，通过类别 dropdown 筛选 + 时间排序自然可见
- 后端照存 group_key 和 group_count，供 Popover 防重和 badge 显示

### 4.3 筛选与排序

#### 类别筛选 (Dropdown)

```
┌──────────────────┐
│  全部类别     ▾  │
├──────────────────┤
│  ✓ 全部          │
│  · 预算告警      │
│  · Key 到期      │
│  · 用量报告      │
│  · 安全事件      │
│  · 系统维护      │
│  · 超支通知      │
└──────────────────┘
```

#### 状态筛选 (Dropdown)

```
┌──────────────────┐
│  全部状态     ▾  │
├──────────────────┤
│  ✓ 全部          │
│  · 未读          │
│  · 已读          │
└──────────────────┘
```

- 排序固定为时间倒序（最新在前），不提供排序切换
- 筛选使用 URL query params，支持浏览器后退/书签

### 4.4 归档 Tab

```
┌─────────────────────────────────────────────────────────────────┐
│  [收件箱 ●12]    [已归档 ←当前]                                  │
│                                                                  │
│  ┌─ Toolbar ──────────────────────────────────────────────────┐ │
│  │  类别: [全部 ▾]                              [清空已归档]   │ │
│  └────────────────────────────────────────────────────────────┘ │
│                                                                  │
│  ┌─ Archived List ────────────────────────────────────────────┐ │
│  │  🟠  预算告警            归档于 1 小时前          [恢复]    │ │
│  │  🔑  Key 到期            归档于 昨天              [恢复]    │ │
│  └────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

- 归档列表样式同收件箱，但操作只有 [恢复] 和 [删除]
- "恢复" = 将 status 设回 'active'，回到收件箱
- "清空已归档" = 批量软删除（设置 status='deleted'），前端不再展示，需确认 dialog

### 4.5 空状态

#### 收件箱空 — 无任何通知

```
        ┌──────────────────────────┐
        │                          │
        │      🔔 (描边 48px)      │
        │                          │
        │    暂无通知               │
        │    有新动态时会在这里提醒   │
        │                          │
        └──────────────────────────┘
```

#### 收件箱空 — 有通知但全部归档

```
        ┌──────────────────────────┐
        │                          │
        │      ✅ (描边 48px)      │
        │                          │
        │    全部处理完了            │
        │    查看已归档通知 →       │
        │                          │
        └──────────────────────────┘
```

#### 筛选无结果

```
        ┌──────────────────────────┐
        │                          │
        │      🔍 (描边 48px)      │
        │                          │
        │    没有符合条件的通知      │
        │    试试调整筛选条件       │
        │                          │
        └──────────────────────────┘
```

---

## 5. Toast 实时推送（优化）

### 5.1 触发时机

SSE 收到新通知时：
1. 铃铛 badge 数字 +1（TanStack Query invalidate）
2. 弹出 Toast 通知

### 5.2 SSE 事件结构扩展

现有 SSEEvent 需要扩展以支持分组防刷屏：

```json
{
  "id": "uuid",
  "eventType": "budget_alert_reached",
  "title": "项目 A 预算使用率达到 90%",
  "body": "当前已用 ¥2,340 / ¥2,600",
  "groupKey": "budget_alert:rule-123:2026-07",
  "category": "budget_alert",
  "payload": { "ruleID": "rule-123", "projectID": "proj-456" }
}
```

> 新增 `groupKey`、`category`、`payload` 字段。前端据 groupKey 做防刷屏去重，据 payload 计算跳转 URL。
> `payload` 在 Go 侧使用 `json.RawMessage` 类型（即 `[]byte`），避免 InAppChannel 已 marshal 的 payload 被二次 marshal。SSEEvent 结构体定义：
>
> ```go
> type SSEEvent struct {
>     ID        string          `json:"id"`
>     EventType string          `json:"eventType"`
>     Title     string          `json:"title"`
>     Body      string          `json:"body"`
>     GroupKey  string          `json:"groupKey,omitempty"`
>     Category  string          `json:"category,omitempty"`
>     Payload   json.RawMessage `json:"payload,omitempty"`
> }
> ```

### 5.3 Toast 样式

```
┌─────────────────────────────────────────────┐
│  🟠  预算告警                          ✕    │
│      项目 A 预算使用率达到 90%              │
│                            ──── 5s 后消失   │
└─────────────────────────────────────────────┘
```

- 使用 sonner 的 toast，position: top-right
- 带类别图标 + 颜色
- 可点击：点击后跳转对应资源页面（前端 `getActionUrl(notification)` 计算），标记已读
- 5s 自动消失
- 同一 group_key 的多条快速推送，只弹最新一条（防刷屏）

---

## 6. 删除确认设计

> **设计原则**：所有"删除"操作均为软删除（设置 `status = 'deleted'`），记录对用户永久不可见，但数据保留在数据库中。原因：`notification_log` 表同时承载用户收件箱（in_app）和投递失败记录（admin 侧可见），用户操作不应销毁平台运维数据。行业惯例（GitHub/Linear/Gmail/阿里云）均采用类似策略。

### 6.1 单条删除 — 软删除 + Undo Toast

```
┌─────────────────────────────────────────────┐
│  已删除 1 条通知                    [撤销]   │
│                            ──── 5s 后消失   │
└─────────────────────────────────────────────┘
```

- 点击删除 → 立即发 `POST /notifications/{id}/delete` 请求（设置 status = 'deleted'）
- 列表 UI 立即移除该条目（乐观更新）
- Toast 显示 5s，期间可点击 "撤销" → 发请求恢复 status = 'active'，条目恢复
- 5s 后 toast 消失，软删除生效，前端永不再展示
- 数据保留在数据库，admin 侧仍可在日志中查看

> **Undo 实现**：删除后弹出 sonner toast（5s），期间可点击 "撤销" 发 `undelete` 请求（恢复 status='active'）。toast 消失或被 dismiss 即视为确认。路由跳转时 toast 自动 dismiss，不做 undo 队列持久化——数据已软删无损，这是可接受的简单方案。

### 6.2 "清空已归档"

- 对已归档列表中所有记录批量设置 status = 'deleted'
- 需确认 dialog（"确定清空所有已归档通知？"）
- 前端不可见，数据保留

---

## 7. 视觉规范

### 7.1 配色 Token

| 用途 | Light Mode | Dark Mode |
|------|-----------|-----------|
| 未读背景 | `bg-blue-50/40` | `bg-blue-950/15` |
| 未读左边框 | `border-l-2 border-blue-500` | `border-l-2 border-blue-400` |
| 未读标题 | `text-foreground font-medium` | 同 |
| 已读标题 | `text-muted-foreground font-normal` | 同 |
| hover 背景 | `bg-muted/60` | `bg-muted/40` |
| 操作图标 | `text-muted-foreground hover:text-foreground` | 同 |
| 删除按钮 | `text-destructive hover:bg-destructive/10` | 同 |
| Badge 计数 | `bg-muted text-muted-foreground text-[11px]` | 同 |
| 铃铛未读点 | `bg-red-500` | 同 |

### 7.2 间距与尺寸

| 元素 | 规格 |
|------|------|
| 通知条目高度 | 自适应，min-h-16 (64px) |
| 条目内边距 | `px-4 py-3` |
| 条目间分割 | `border-b border-border` (1px) |
| 图标尺寸 | 16px (category icon)，14px (action icons) |
| 标题字号 | `text-sm` (14px) |
| 摘要字号 | `text-xs` (12px) |
| 时间字号 | `text-xs` (12px) |
| Popover 宽度 | 360px |
| 列表页最大宽度 | `max-w-3xl` (768px) 居中 |

### 7.3 动画

| 交互 | 动画 |
|------|------|
| 删除条目 | `animate-out fade-out slide-out-to-left` 300ms |
| 归档条目 | `animate-out fade-out slide-out-to-right` 300ms |
| 新通知插入 | `animate-in fade-in slide-in-from-top` 200ms |
| Popover 出现 | shadcn 默认 popover animation |

---

## 8. 响应式设计

### 8.1 桌面端（≥1024px）

- 通知中心页面 max-w-3xl 居中，两侧留白
- hover 显示操作按钮

### 8.2 平板端（768px - 1024px）

- 全宽，去掉两侧留白
- 操作按钮始终可见（触摸设备无 hover）

### 8.3 移动端（<768px）

- Popover 改为全屏 sheet（从底部滑出）
- 通知中心页面列表全宽
- 操作按钮始终显示（无 hover）
- 时间戳缩写（"2分钟前" → "2分"）

---

## 9. 数据模型与 API

### 9.1 数据库 Schema 变更

> **身份标识约定**：`notification_log.user_id` 存储的是 **member_id**（即 `members.id`），不是 `users.id`。原因：member 是用户在某个 company 内的身份，通知天然按 company 隔离，用 member_id 可直接与 session context 中的 `Member.ID` 对应，无需额外 join。InAppChannel 写入时 `recipientID` 传的也是 member_id。SSE Hub 的 Subscribe/Publish 同样以 member_id 为 key。

```sql
-- notification_log 表（直接改 schema.sql 重建，无需 migration）
CREATE TABLE IF NOT EXISTS notification_log (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id   UUID NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
    channel      TEXT NOT NULL,
    event_type   TEXT NOT NULL,
    user_id      UUID,
    title        TEXT NOT NULL DEFAULT '',
    body         TEXT NOT NULL DEFAULT '',
    payload      JSONB NOT NULL,
    send_ok      BOOLEAN NOT NULL DEFAULT true,
    error        TEXT,
    category     TEXT NOT NULL DEFAULT '',
    group_key    TEXT NOT NULL DEFAULT '',
    status       TEXT NOT NULL DEFAULT 'active',  -- active / archived / deleted
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    read_at      TIMESTAMPTZ,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 收件箱查询索引（active only）
CREATE INDEX IF NOT EXISTS idx_notification_log_inbox
    ON notification_log (company_id, user_id, created_at DESC)
    WHERE status = 'active';

-- 按类别筛选索引
CREATE INDEX IF NOT EXISTS idx_notification_log_category
    ON notification_log (company_id, user_id, category, created_at DESC)
    WHERE status = 'active';

-- 分组聚合索引
CREATE INDEX IF NOT EXISTS idx_notification_log_group
    ON notification_log (company_id, user_id, group_key, created_at DESC)
    WHERE group_key != '' AND status = 'active';

-- 未读计数索引
CREATE INDEX IF NOT EXISTS idx_notification_log_unread
    ON notification_log (company_id, user_id)
    WHERE read_at IS NULL AND status = 'active';
```

**字段设计说明**：
- `send_ok` (bool)：投递结果。true=成功，false=失败。替代原 `send_status` TEXT，因为只有两态
- `status` (text enum)：用户侧生命周期状态。`active` → `archived` → `deleted`（单向递进）。替代原 `archived_at`/`deleted_at` 两个时间戳字段，更简洁，索引条件也更干净
- `updated_at`：记录最后状态变更时间（已读/归档/删除/恢复），统一替代原 `archived_at`/`deleted_at` 的时间审计功能
- `read_at`：已读时间，与 status 正交（归档的通知可以是未读的）

### 9.2 API 端点

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/notifications` | 列表查询（含分组） |
| GET | `/notifications/unread-count` | 未读数（已有） |
| PATCH | `/notifications/{id}/read` | 标记已读（已有） |
| POST | `/notifications/read-all` | 全部已读（已有） |
| POST | `/notifications/{id}/archive` | 归档单条 |
| POST | `/notifications/archive-all` | 归档当前筛选视图中的所有通知（受 category/status 参数影响） |
| POST | `/notifications/{id}/unarchive` | 取消归档（恢复） |
| POST | `/notifications/{id}/delete` | 软删除单条 |
| POST | `/notifications/{id}/undelete` | 撤销软删除（恢复 status='active'） |
| GET | `/notifications/stream` | SSE 实时推送（已有） |

> **路由注册顺序**：chi 在同级路由中，静态路径优先于参数路径匹配，无需担心 `/read-all` 或 `/archive-all` 被 `/{id}` 误捕获。`/{id}/archive`、`/{id}/delete`、`/{id}/undelete`、`/{id}/unarchive` 等子路由与现有 `/{id}/read` 并列注册即可。

> **channel 隔离**：所有用户侧操作（归档/删除/已读）的 SQL 必须加 `AND channel = 'in_app'` 条件，确保不影响投递失败等运维记录。Admin 侧接口不加此限制。

### 9.3 GET /notifications 查询参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `limit` | int | 20 | 每页条数 |
| `cursor` | string | - | 游标（上一页最后一条的 created_at ISO 时间戳），首次请求不传 |
| `category` | string | - | 按类别筛选（可选） |
| `status` | enum | all | `unread` / `read` / `all` |
| `archived` | bool | false | 是否查已归档 |
| `grouped` | bool | true | 是否分组返回 |
| `group_key` | string | - | 指定 group_key 值，拉取组内详情 |

> **分页方案**：使用 cursor 分页（基于 created_at 时间戳），而非 offset。避免新通知插入导致偏移漂移丢消息。响应体包含 `nextCursor` 字段，前端传回即可翻页。

> **多租户隔离**：所有查询必须同时传 company_id + member_id 做联合过滤。company_id 从 session context 获取（不由前端传参），确保索引命中和租户隔离。现有 List handler 已满足此条件。

### 9.4 响应体增强

```json
{
  "id": "uuid",
  "eventType": "budget_alert_reached",
  "channel": "in_app",
  "category": "budget_alert",
  "title": "项目 A 预算使用率达到 90%",
  "body": "当前已用 ¥2,340 / ¥2,600",
  "payload": { "ruleID": "rule-123", "projectID": "proj-456", "threshold": 90 },
  "groupKey": "budget_alert:rule-123:2026-07",
  "groupCount": 3,
  "status": "active",
  "createdAt": "2026-07-27T10:30:00Z",
  "readAt": null,
  "updatedAt": "2026-07-27T10:30:00Z"
}
```

分页响应包装：

```json
{
  "items": [...],
  "nextCursor": "2026-07-26T08:00:00Z",
  "hasMore": true
}
```

**字段说明**：
- `status`：用户侧生命周期状态（`active` / `archived` / `deleted`），前端根据此字段渲染列表视图
- `category`：写入时由 InAppChannel 根据 event_type 调用 `EventCategory()` 计算并持久化到 `category` 列，筛选直接走索引
- `payload`：原始 JSON payload，前端据此拼接跳转 URL
- `groupKey`、`groupCount`：分组标识和组内计数
- `updatedAt`：最后变更时间（已读/归档/恢复等操作触发更新）
- 注意：`send_ok` 和 `error` 字段仅在 Admin 接口返回，用户侧 inbox 接口不暴露投递状态

> **actionUrl 由前端计算**：后端不返回 actionUrl。前端维护一个 `getActionUrl(notification)` 映射函数，根据 event_type + payload 拼接路由。这样前端路由变更不需要动后端。

---

## 10. 分组规则

### 10.1 group_key 值生成

由后端 InAppChannel 写入时根据 event 信息生成：

| 事件类型 | group_key 格式 | 分组效果 |
|----------|-----------|---------|
| budget_alert_reached | `budget:{ruleID}:{periodKey}` | 同规则同周期合并 |
| overrun_blocked | `overrun:{projectID}:{date}` | 同项目同天合并 |
| overdraft_expanded | `overdraft:{projectID}:{date}` | 同项目同天合并 |
| key_expiring_soon | `key_expiry:{keyID}` | 同 Key 多次提醒合并 |
| key_expired | `key_expiry:{keyID}` | 同 Key 合并 |
| security_login_new_device | `""` | 不分组（每次独立） |
| system_maintenance_scheduled | `maintenance:{eventID}` | 同事件合并 |
| usage_weekly_report | `""` | 不分组 |
| sync_threshold_exceeded | `sync:{date}` | 同天合并 |

### 10.2 分组查询 SQL

```sql
-- 分组模式：两步查询保证 cursor 分页稳定性
-- Step 1: 找出每组代表行（最新一条）的 id 和 created_at
WITH group_heads AS (
    SELECT DISTINCT ON (COALESCE(NULLIF(group_key, ''), id::text))
        id, group_key, created_at
    FROM notification_log
    WHERE company_id = $1 AND user_id = $2
      AND channel = 'in_app'
      AND status = 'active'
    ORDER BY COALESCE(NULLIF(group_key, ''), id::text), created_at DESC
),
-- Step 2: 对代表行做 cursor 分页（cursor 作用于代表行的 created_at，而非原始行）
paged AS (
    SELECT id, group_key, created_at
    FROM group_heads
    WHERE ($3::timestamptz IS NULL OR created_at < $3)
    ORDER BY created_at DESC
    LIMIT $4
)
-- Step 3: 关联原表拿完整字段 + group_count
SELECT n.id, n.event_type, n.channel, n.category, n.title, n.body, n.payload, n.send_ok,
       n.group_key, n.status, n.created_at, n.read_at, n.updated_at,
       (SELECT COUNT(*) FROM notification_log sub
        WHERE sub.company_id = $1 AND sub.user_id = $2
          AND sub.channel = 'in_app'
          AND sub.status = 'active'
          AND COALESCE(NULLIF(sub.group_key, ''), sub.id::text) = COALESCE(NULLIF(n.group_key, ''), n.id::text)
       ) as group_count
FROM paged p
JOIN notification_log n ON n.id = p.id
ORDER BY p.created_at DESC;
```

- 空 group_key 的条目视为独立组（用 id 做 fallback key）
- cursor 作用于 **分组代表行** 的 created_at，而非原始行。新通知到来改变某组的代表行不会导致翻页丢组或重复
- 使用 cursor 分页：`$3` 为上一页最后一条代表行的 created_at，首次请求传 NULL
- 响应中 `nextCursor` = 本页最后一条代表行的 created_at
- `user_id` 列实际存 member_id（见 9.1 身份标识约定）

> **性能注意**：当通知量超过 10 万级别，DISTINCT ON 全扫描代价大。升级路径：引入物化视图或分组摘要表，定期刷新。当前阶段按单用户几百~几千条通知的规模，此方案足够。

---

## 11. 通知生命周期状态机

```
                    ┌─────────────┐
        新建 ──────▶│   active     │
                    │  (收件箱)    │
                    └──────┬──────┘
                           │
              ┌────────────┼────────────┐
              │            │            │
              ▼            ▼            ▼
         标记已读     点击归档       点击删除
         (read_at     (status=       (status=
          =now)       'archived')    'deleted')
              │            │            │
              │            ▼            ▼
              │    ┌────────────┐  ┌────────────────┐
              │    │  archived   │  │    deleted      │
              │    │  (已归档)   │  │ (不可见,5s撤销) │
              │    └─────┬──────┘  └───────┬────────┘
              │          │                 │
              │    ┌─────┼─────┐     ┌─────┼─────┐
              │    │           │     │           │
              │    ▼           ▼     ▼           ▼
              │  恢复        删除   撤销        终态
              │  (status=   (status= (status=   (数据保留
              │  'active')  'deleted') 'active') 前端不可见)
              │    │           │      │
              │    ▼           ▼      ▼
              │  回到收件箱   终态  回到收件箱
              │
              ▼
        active(已读)
        read_at != NULL
```

说明：
- 生命周期由 `status` 字段控制：`active` → `archived` → `deleted`
- `read_at` 与 `status` 正交：归档的通知可以是未读的
- 所有状态变更同时更新 `updated_at`
- "删除"均为软删除（status='deleted'），数据永久保留在数据库
- 用户侧查询加 `WHERE status != 'deleted'` 或 `status = 'active'` 过滤
- Admin 侧 AdminLog 接口不加此过滤，可查看完整记录
- 删除/归档操作限定 `AND channel = 'in_app'`，不影响投递失败记录
- Undelete 恢复统一回 `active` 状态（丢失"曾归档"信息，可接受的 tradeoff）

---

## 12. 实现计划

### Phase 1 — 基础功能（4-5 天）

- [x] Schema 变更：加 `category`、`group_key`、`status`（active/archived/deleted）、`send_ok`（bool）、`updated_at` 字段 + 索引
- [x] 修改 `NotificationRepository` 接口：List 改 cursor 分页签名，新增 Archive/Delete/Undelete 方法
- [x] 更新现有 API：MarkAllRead / GetUnreadCount 的 WHERE 条件使用 `status = 'active'`
- [x] 后端 API：归档/取消归档/软删除/撤销删除端点
- [x] 后端 API：List 增加 category/status/archived/grouped 参数 + cursor 分页
- [x] 后端 API：archive-all 接受 category/status 参数，归档当前筛选视图
- [x] 后端 API：响应增加 category/groupKey/groupCount/payload/updatedAt 字段
- [x] InAppChannel：写入时生成 group_key 值 + 计算 category 入库
- [x] SSEEvent 结构体扩展：增加 groupKey/category/payload（json.RawMessage）字段
- [ ] 前端路由：新增 /notifications 路由
- [ ] 前端 API 层：补充新端点调用
- [ ] 前端：实现 `getActionUrl(notification)` 映射函数

### Phase 2 — 通知中心页面（3 天）

- [ ] 通知中心页面骨架（Tabs + Toolbar + List）
- [ ] 通知列表项组件（未读/已读两种状态 + 分组 badge）
- [ ] 归档/删除操作 + undo toast（软删除 + 撤销）
- [ ] 空状态组件
- [ ] 筛选 dropdown（类别 + 状态）
- [ ] 无限滚动加载（cursor 分页 + useInfiniteQuery）

### Phase 3 — Popover 优化（1 天）

- [ ] Popover 加类别图标 + 颜色
- [ ] Popover 加分组 badge（×3）
- [ ] Popover 加 "查看全部通知" 入口
- [ ] Popover hover 快捷归档
- [ ] Toast 推送可点击跳转

### Phase 4 — 体验打磨（1 天）

- [ ] 删除/归档动画（slide out）
- [ ] 响应式适配（移动端 sheet、操作按钮始终显示）
- [ ] Dark mode 调优

---

## 13. 设计决策记录

| 决策 | 选择 | 理由 |
|------|------|------|
| 归档 vs 只有已读 | 做归档 | 参考 GitHub Done 模式，"处理完毕"语义比"已读"更明确 |
| 删除策略 | status='deleted' 软删除，不做硬删 | 表混合 in_app 收件箱 + 投递失败记录，软删避免误伤运维数据。行业惯例（GitHub/Linear/阿里云）均不真删通知 |
| 生命周期字段 | 单 `status` 枚举（active/archived/deleted）+ `updated_at` | 替代 `archived_at`/`deleted_at` 双时间戳。状态单向递进不叠加，索引条件从 `archived_at IS NULL AND deleted_at IS NULL` 简化为 `status = 'active'`，更清晰 |
| 投递状态字段 | `send_ok` BOOLEAN | 替代 `send_status` TEXT。只有 sent/failed 两态，bool 更简洁。不存在 pending 中间态（只在投递完成后写入） |
| 字段名 | `group_key` | 避免 SQL 保留字 `group` 引号转义的维护负担，所有 SQL/struct tag 无需特殊处理 |
| category 存储 | 入库持久化（写入时计算） | 筛选 `WHERE category = $x` 直接走索引，避免 event_type IN (...) 反向映射的复杂度 |
| Popover 内做归档/删除 | 只做归档 | 保持 Popover 轻量，重操作引导去独立页 |
| 分页方案 | cursor（基于 created_at） | offset 在新通知插入时偏移漂移丢消息，cursor 无此问题 |
| 分组 SQL | 两步查询（DISTINCT ON 找代表行 → cursor 分页 → JOIN 拿详情） | cursor 作用于代表行保证翻页稳定性。PG 原生支持，无需额外表。升级路径：10 万级别考虑物化视图 |
| 单条删除确认 | Undo toast（立即软删 + 5s 可撤销） | 比 dialog 轻量，比纯前端延迟请求更可靠 |
| 分组展示 | 扁平列表 + badge 显示组内数量 | 通知量不大，折叠/展开交互 ROI 低。类别 dropdown 筛选满足查看同类通知的需求 |
| actionUrl 计算 | 前端 `getActionUrl()` 函数，后端只传 event_type + payload | 前端路由变更不需动后端，职责分离更清晰 |
| 通知中心入口 | 不放侧边栏 nav | 铃铛是全局入口，独立路由做深度管理 |
| 多租户查询 | company_id + member_id 联合过滤 | 确保索引命中 + 租户数据隔离，company_id 从 session context 取。`user_id` 列实际存 member_id |
| SSE 事件结构 | 扩展 groupKey/category/payload（json.RawMessage） | 前端需要 groupKey 做防刷屏去重，payload 用 RawMessage 避免二次 marshal |
| 审计与删除关系 | 软删 + channel 隔离 | `notification_log` 混存 in_app 和投递失败记录，删除 API 限定 `channel='in_app'`，admin 侧不受影响 |
