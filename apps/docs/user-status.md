# 用户（公司成员）状态设计

## 状态定义

| 状态 | 值 | 含义 | 可登录 |
|------|------|------|--------|
| 待激活 | `pending` | 已被邀请但尚未注册/激活 | 否 |
| 已启用 | `active` | 正常使用中 | 是 |
| 已禁用 | `disabled` | 管理员手动禁用（含"移除成员"） | 否 |

> 没有独立"已删除"状态。对已注册用户，"删除/移除" = 将状态改为 `disabled`，记录保留。

## 状态流转

```
邀请创建 ──▶ pending ──▶（用户完成注册）──▶ active ◀──▶ disabled
               │                              │
               ▼                              ▼
          硬删（物理移除）              disabled（软删，记录保留）
```

### 允许的转换

| 当前状态 | 目标 | 触发方式 |
|----------|------|----------|
| pending | active | 用户自己完成注册激活（AcceptInvite） |
| pending | （硬删） | 管理员删除，物理移除邀请和成员记录 |
| active | disabled | 管理员禁用 / 管理员移除成员 |
| disabled | active | 管理员重新启用 |

### 禁止的操作

| 操作 | 规则 | 原因 |
|------|------|------|
| pending → active（管理员手动） | ❌ | 必须由用户自己完成注册激活 |
| pending → disabled | ❌ | 未激活用户无需禁用，直接硬删即可 |
| disabled → （硬删） | ❌ | 已注册用户有使用数据，不允许物理删除 |

## 删除策略

| 当前状态 | 删除行为 | 说明 |
|----------|----------|------|
| pending | **硬删** | 物理移除 member 记录 + 关联邀请，用户从未使用系统 |
| active / disabled | **软删** | 状态改为 `disabled`，关联 API Key 立即失效，消费/审计记录保留 |

## 管理员操作菜单

| 用户状态 | 可用操作 |
|----------|----------|
| pending | 获取邀请链接、重新发送邀请、删除（硬删） |
| active | 禁用、删除（改为 disabled） |
| disabled | 启用、—（不可再删，已是终态） |

## 后端实现要点

1. **状态常量**：`MemberStatusActive = "active"`, `MemberStatusDisabled = "disabled"`, `MemberStatusPending = "pending"`
2. **DeleteMembers**：检查 status —— pending 则物理移除，其他则改 status 为 `disabled` + 禁用关联 Key
3. **UpdateMemberStatus**：拒绝将 `pending` 改为 `active`（返回 400："用户需完成注册后自动激活"）
4. **登录校验**：仅 `active` 状态允许登录
5. **列表展示**：默认不显示 `disabled` 成员，除非筛选"已禁用"
