# 个人账户管理页面 — 产品需求文档

## 1. 背景与目标

当前系统缺少统一的个人账户管理入口。用户登录后无法自主修改密码、更换联系方式、查看登录记录。本页面旨在：

- 让用户自助完成日常账户维护操作，降低管理员/客服压力
- 提供安全可控的凭证变更流程（验证码二次确认）
- 为后续功能（头像、通知偏好、API Token）预留扩展位

## 2. 现状分析

| 能力 | 后端 | 前端 | 备注 |
|------|------|------|------|
| 设置/修改密码 | ✅ `/auth/set-password` | ❌ 无入口 | 仅 SMS 登录后场景可触发 |
| 重置密码（忘记） | ✅ `/auth/reset-password` | ❌ 无入口 | 需验证码 |
| 更换手机号 | ✅ `UserRepo.UpdatePhone` | ❌ 无 | 仅 DB 层有接口 |
| 更换邮箱 | ✅ `UserRepo.UpdateEmail` | ❌ 无 | 仅 DB 层有接口 |
| 查看会话 | ✅ `SessionRepository` | ❌ 无 | — |
| 编辑个人信息（姓名等） | 由管理员在成员管理操作 | ❌ 无自助 | — |

结论：后端基础能力已就绪，缺乏面向用户的 HTTP 端点和前端页面。

## 3. 用户故事

| # | 角色 | 故事 | 验收标准 |
|---|------|------|----------|
| US-1 | 普通成员 | 我想修改登录密码 | 输入旧密码 + 新密码确认后立即生效；错误提示清晰 |
| US-2 | 普通成员 | 我想更换绑定手机号 | 先验证新手机验证码，再替换；旧号无需验证（已登录态视为本人） |
| US-3 | 普通成员 | 我想更换绑定邮箱 | 向新邮箱发送验证码，验证通过后替换 |
| US-4 | 普通成员 | 我想查看当前绑定信息 | 脱敏显示手机号和邮箱（`138****1234`、`q**@xx.com`） |
| US-5 | 普通成员 | 我想查看我所在的企业列表 | 显示我加入的所有企业（名称、角色、状态） |
| US-6 | 普通成员 | 我想登出所有设备 | 一键吊销除当前外的所有活跃会话 |
| US-7 | 普通成员 | 我想查看最近登录活动 | 列表展示最近 N 条登录记录（时间、IP、设备类型） |

## 4. 页面结构

```
/account
├── 基本信息（Profile Summary）
│   ├── 姓名（只读，由管理员维护）
│   ├── 手机号（脱敏 + 修改按钮）
│   ├── 邮箱（脱敏 + 修改按钮）
│   └── 所属企业 & 角色（只读列表）
│
├── 安全设置（Security）
│   ├── 修改密码
│   ├── 登出所有设备
│   └── 最近登录活动
│
└── （预留）偏好设置（Preferences）
    ├── 通知渠道（短信/邮件/站内）
    └── 语言/时区
```

## 5. 交互流程

### 5.1 修改密码

```
用户点击「修改密码」
  → 弹出 Modal
  → 输入：当前密码、新密码、确认新密码
  → 前端校验：新密码 ≥ 8 位、两次一致
  → 调用 POST /me/change-password { oldPassword, newPassword }
  → 后端验证旧密码 → 更新 → 204
  → 提示成功，关闭 Modal
```

注意：与现有 `/auth/set-password` 区分——`set-password` 是无旧密码场景（SMS 首次设密）；`change-password` 需验证旧密码。

### 5.2 更换手机号

```
用户点击「修改」
  → 弹出 Modal
  → 输入新手机号
  → 调用 POST /auth/verify-code/send { phone, scene: "bind_phone" }
  → 输入验证码
  → 调用 POST /me/change-phone { phone, code }
  → 后端验证码校验 → 更新 user.phone → 同步更新 member.phone → 204
  → 提示成功，刷新页面信息
```

### 5.3 更换邮箱

```
用户点击「修改」
  → 弹出 Modal
  → 输入新邮箱
  → 调用 POST /auth/verify-code/send { email, scene: "bind_email" }
  → 输入验证码
  → 调用 POST /me/change-email { email, code }
  → 后端验证码校验 → 更新 user.email → 同步更新 member.email → 204
  → 提示成功，刷新页面信息
```

### 5.4 登出所有设备

```
用户点击「登出所有设备」
  → 二次确认 Dialog
  → 调用 POST /me/revoke-sessions
  → 后端吊销当前用户所有 session（保留当前）
  → 提示成功
```

## 6. API 设计

### 新增端点（挂在 `/me` 路由组下，需已认证）

| Method | Path | 功能 | Request Body | Response |
|--------|------|------|-------------|----------|
| GET | `/me/profile` | 获取个人信息 | — | `{ phone, email, name, companies[] }` |
| POST | `/me/change-password` | 修改密码（需旧密码） | `{ oldPassword, newPassword }` | 204 |
| POST | `/me/change-phone` | 更换手机号 | `{ phone, code }` | 204 |
| POST | `/me/change-email` | 更换邮箱 | `{ email, code }` | 204 |
| POST | `/me/revoke-sessions` | 登出其他设备 | — | 204 |
| GET | `/me/login-activity` | 最近登录活动 | `?limit=20` | `[{ time, ip, userAgent }]` |

### GET /me/profile 响应示例

```json
{
  "phone": "138****1234",
  "email": "qu**@example.com",
  "name": "张三",
  "hasPassword": true,
  "companies": [
    {
      "companyId": "uuid",
      "companyName": "TokenJoy Inc.",
      "role": "admin",
      "current": true
    }
  ]
}
```

注意：`phone`/`email` 在 API 层脱敏返回，前端不做脱敏逻辑。

## 7. 安全约束

| 场景 | 防护措施 |
|------|---------|
| 修改密码 | 必须验证旧密码；新密码 ≥ 8 位 |
| 更换手机/邮箱 | 必须验证新手机/邮箱的验证码（复用现有 verify-code 服务） |
| 验证码频率 | 复用现有限流：同号码 60s 间隔、5 分钟有效、连续错误锁定 |
| 登出设备 | 仅吊销非当前 session |
| 信息展示 | API 层脱敏，防止中间人获取完整号码 |
| CSRF | 现有 JWT Cookie + SameSite 已覆盖 |

## 8. 前端实现方案

| 层 | 位置 | 职责 |
|----|------|------|
| 路由 | `routes/account/page.tsx` | 页面入口，组合 features 组件 |
| Feature | `features/account/` | hooks、components、index.ts |
| API | `api/account.ts` | HTTP 请求封装 |
| 共享组件 | 复用现有 Dialog、Form、Input | — |

页面入口从 Sidebar / Header 用户头像下拉菜单进入。

## 9. 后端实现方案

- 在现有 `handler/me/` 下新增端点
- `change-password` 逻辑：验旧密码 → bcrypt 新密码 → `UserRepo.UpdatePassword`
- `change-phone` / `change-email`：调 `verifyCode.Verify` → `UserRepo.UpdatePhone/Email` → 同步 `MemberRepo.UpdatePhone/Email`
- `revoke-sessions`：`SessionRepo.DeleteByUserIDExcept(currentSessionID)`
- `login-activity`：需在 `sessions` 表增加 `ip`、`user_agent` 字段（或新建 `login_logs` 表）

## 10. 数据变更

无需 migration（项目未上线），直接修改 schema.sql：

1. `sessions` 表增加 `ip TEXT`、`user_agent TEXT` 字段
2. 或新建 `user_login_logs` 表：

```sql
CREATE TABLE IF NOT EXISTS user_login_logs (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id),
    ip         TEXT NOT NULL DEFAULT '',
    user_agent TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_login_logs_user ON user_login_logs(user_id, created_at DESC);
```

## 11. 分期计划

### P0（MVP，建议首期交付）

- [x] 查看个人基本信息（脱敏手机/邮箱/姓名/企业）
- [x] 修改密码（需旧密码）
- [x] 更换手机号（验证码流程）
- [x] 更换邮箱（验证码流程）

### P1（第二期）

- [ ] 登出所有设备
- [ ] 最近登录活动

### P2（远期预留）

- [ ] 头像上传
- [ ] 通知偏好设置
- [ ] 个人 API Token 管理
- [ ] 账户注销

## 12. 开放问题

| # | 问题 | 建议 |
|---|------|------|
| 1 | 更换手机/邮箱是否需要管理员审批？ | 建议不需要——本人登录态 + 验证码已足够；管理员可在成员管理看到变更 |
| 2 | 同一手机/邮箱被其他用户占用时如何处理？ | 后端返回冲突错误，前端提示「该手机号已被其他账户绑定」 |
| 3 | 是否允许同时清空手机和邮箱？ | 不允许——至少保留一个可用的登录凭证 |
| 4 | 姓名是否允许自助修改？ | 当前由管理员维护（Member 属于 Company），建议保持只读 |
| 5 | 是否需要操作日志/审计？ | P1 再考虑，当前先做功能 |
