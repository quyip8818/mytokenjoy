# 成员状态

## 状态

| 状态 | 值 | 说明 | 可登录 |
|------|------|------|--------|
| 待激活 | `pending` | 已邀请，未完成注册 | 否 |
| 已启用 | `active` | 正常使用 | 是 |
| 已禁用 | `disabled` | 管理员禁用或移除 | 否 |

## 转换规则

```
pending ──（注册激活）──▶ active ◀──▶ disabled
   │                        │
   ▼                        ▼
 硬删                    disabled
```

| 从 | 到 | 方式 |
|----|------|------|
| pending | active | 用户完成注册（AcceptInvite） |
| pending | 硬删 | 管理员删除，物理移除记录 |
| active | disabled | 管理员禁用/移除 |
| disabled | active | 管理员启用 |

**禁止**：管理员手动将 pending 改为 active。

## 删除

| 状态 | 行为 |
|------|------|
| pending | 硬删，物理移除 member + invite 记录 |
| active / disabled | 软删，状态改为 `disabled`，API Key 失效，记录保留 |

## 管理员操作

| 状态 | 操作 |
|------|------|
| pending | 获取邀请链接、重新发送邀请、删除 |
| active | 停用、删除 |
| disabled | 启用 |

## 列表接口

`GET /org/members?status=active,pending`

`status` 支持逗号分隔，不传则返回全部。前端默认传 `active,pending`。
