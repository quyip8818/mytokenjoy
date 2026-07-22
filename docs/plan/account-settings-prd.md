# 个人账户管理页面 — 产品需求文档

## 1. 背景与目标

当前系统缺少统一的个人账户管理入口。用户登录后无法自主修改密码、更换联系方式、查看登录记录。本页面旨在：

- 让用户自助完成日常账户维护操作，降低管理员/客服压力
- 提供安全可控的凭证变更流程（验证码二次确认）
- 为后续功能（头像、通知偏好、API Token）预留扩展位

## 2. 现状分析

| 能力 | 后端 | 前端 | 备注 |
|------|------|------|------|
| 设置密码（首次） | ✅ `/auth/set-password` | ❌ 无入口 | SMS 登录后场景可触发 |
| 修改密码（需旧密码） | ❌ 无端点 | ❌ 无 | 需新增 |
| 更换手机号 | ✅ `UserRepo.UpdatePhone` | ❌ 无 | 仅 DB 层有方法 |
| 更换邮箱 | ✅ `UserRepo.UpdateEmail` | ❌ 无 | 仅 DB 层有方法 |
| 吊销所有会话 | ⚠️ `SessionRepo.RevokeAllByUser` | ❌ 无 | 现有方法不支持排除当前 session，需新增 `RevokeAllByUserExcept` |
| 查看会话/登录记录 | ✅ sessions 表含 ip/user_agent | ❌ 无 | 需新增查询端点 |
| 查看所属企业 | ✅ `UserRepo.ListMemberCompanies` | ❌ 无 | — |
| 编辑个人信息（姓名等） | 由管理员在成员管理操作 | ❌ 无自助 | — |

结论：后端基础能力大部分已就绪，缺乏面向用户的 HTTP 端点和前端页面。

## 3. 用户故事

| # | 角色 | 故事 | 验收标准 |
|---|------|------|----------|
| US-1 | 普通成员 | 我想设置或修改登录密码 | 有旧密码时需验证旧密码；无密码时直接设置；新密码 ≥ 8 位；错误提示清晰 |
| US-2 | 普通成员 | 我想更换绑定手机号 | 向新手机发验证码，验证通过后替换；旧号无需验证（登录态视为本人） |
| US-3 | 普通成员 | 我想更换绑定邮箱 | 向新邮箱发验证码，验证通过后替换 |
| US-4 | 普通成员 | 我想查看当前绑定信息 | 脱敏显示手机号和邮箱（`138****1234`、`q**@xx.com`） |
| US-5 | 普通成员 | 我想查看我所在的企业列表 | 显示我加入的所有企业（名称、角色），标记当前企业 |
| US-6 | 普通成员 | 我想登出所有设备 | 一键吊销除当前外的所有活跃会话 |
| US-7 | 普通成员 | 我想查看最近登录活动 | 列表展示最近 N 条登录记录（时间、IP、设备类型） |

## 4. 页面结构

```
/account
├── 基本信息（Profile Summary）
│   ├── 姓名（只读，由管理员维护）
│   ├── 手机号（脱敏 + 修改按钮）
│   ├── 邮箱（脱敏 + 修改按钮）
│   └── 所属企业 & 角色（只读列表，标记当前）
│
├── 安全设置（Security）
│   ├── 设置/修改密码
│   ├── 登出所有设备
│   └── 最近登录活动
│
└── （预留）偏好设置（Preferences）
    ├── 通知渠道（短信/邮件/站内）
    └── 语言/时区
```

## 5. 交互流程

### 5.1 设置/修改密码

前端根据 `hasPassword` 字段决定展示哪种表单：

**已有密码（hasPassword = true）：**
```
用户点击「修改密码」
  → 弹出 Modal
  → 输入：当前密码、新密码、确认新密码
  → 前端校验：新密码 ≥ 8 位、两次一致
  → 调用 POST /me/change-password { oldPassword, newPassword }
  → 后端验证旧密码 → 更新 → 204
  → 提示成功，关闭 Modal
```

**未设密码（hasPassword = false）：**
```
用户点击「设置密码」
  → 弹出 Modal
  → 输入：新密码、确认新密码
  → 调用 POST /auth/set-password { password }（复用现有端点）
  → 204
  → 提示成功，刷新 hasPassword 状态
```

### 5.2 更换手机号

```
用户点击「修改」
  → 弹出 Modal
  → 输入新手机号
  → 调用 POST /auth/verify-code/send { phone, purpose: "bind" }
  → 输入验证码
  → 调用 POST /me/change-phone { phone, code }
  → 后端在同一请求内完成：验证码校验 → 检查唯一性 → 更新 user.phone → 204
  → 提示成功，刷新页面信息
```

> **实现说明**：`change-phone` 是一次性 atomic 操作——verify + update 在同一请求中完成，无需前端先调 verify 再调 update。

### 5.3 更换邮箱

```
用户点击「修改」
  → 弹出 Modal
  → 输入新邮箱
  → 调用 POST /auth/verify-code/send { email, purpose: "bind" }
  → 输入验证码
  → 调用 POST /me/change-email { email, code }
  → 后端在同一请求内完成：验证码校验 → 检查唯一性 → 更新 user.email → 204
  → 提示成功，刷新页面信息
```

> **实现说明**：同 change-phone，`change-email` 也是 atomic 操作（verify + update 单请求完成）。

### 5.4 登出所有设备

```
用户点击「登出所有设备」
  → 二次确认 Dialog
  → 调用 POST /me/revoke-sessions
  → 后端吊销当前用户所有 session（排除当前 session）
  → 提示成功
```

### 5.5 最近登录活动

```
进入页面自动加载
  → 调用 GET /me/login-activity?limit=20&offset=0
  → 展示列表：时间、IP 地址、设备/浏览器（前端格式化原始 user_agent）
  → 当前会话标记「当前」
  → 支持翻页加载更多
```

## 6. API 设计

### 新增端点（挂在 `/me` 路由组下）

从 session claims 取 `UserID` 操作用户数据。`/me` 下的端点使用 UserID 级别鉴权，不走 company-scoped permission check（因为 profile、企业列表等数据天然跨 company）。

| Method | Path | 功能 | Request Body | Response |
|--------|------|------|-------------|----------|
| GET | `/me/profile` | 获取个人信息 | — | 见下方 |
| POST | `/me/change-password` | 修改密码（需旧密码） | `{ oldPassword, newPassword }` | 204 |
| POST | `/me/change-phone` | 更换手机号 | `{ phone, code }` | 204 |
| POST | `/me/change-email` | 更换邮箱 | `{ email, code }` | 204 |
| POST | `/me/revoke-sessions` | 登出其他设备 | — | 204 |
| GET | `/me/login-activity` | 最近登录活动 | `?limit=20&offset=0` | `{ items: [...], total }` |

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

- `phone`/`email` 在 API 层脱敏返回，前端不做脱敏逻辑
- `current` 通过 session claims 中的 `companyId` 判断

### GET /me/login-activity 响应示例

```json
{
  "items": [
    {
      "time": "2026-07-20T10:30:00Z",
      "ip": "203.0.113.42",
      "userAgent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) ...",
      "current": true
    },
    {
      "time": "2026-07-18T08:15:00Z",
      "ip": "198.51.100.7",
      "userAgent": "Mozilla/5.0 (iPhone; CPU iPhone OS 18_0 like Mac OS X) ...",
      "current": false
    }
  ],
  "total": 42
}
```

- 参数：`limit`（默认 20，上限 100）、`offset`（默认 0）
- 数据源：直接查 `sessions` 表（已含 `ip`、`user_agent`、`created_at`），`current` 通过 session ID 匹配
- `userAgent`：P0 后端返回原始 user_agent 字符串，前端做基础格式化展示；P1 后端增加解析后的 `device` 字段（如 "Chrome 126 / macOS"）

### 错误响应约定

| 状态码 | 场景 |
|--------|------|
| 400 | 参数校验失败、验证码错误、旧密码错误（`code: "wrong_password"`） |
| 409 | 手机号/邮箱已被其他账户绑定 |
| 429 | 验证码发送频率限制 |

## 7. 安全约束

| 场景 | 防护措施 |
|------|---------|
| 修改密码 | 必须验证旧密码；新密码 ≥ 8 位 |
| 更换手机/邮箱 | 验证新号码/邮箱的验证码（复用现有 verify-code 服务） |
| 验证码频率 | 复用现有限流：同号码 60s 间隔、5 分钟有效、连续错误锁定 |
| 手机/邮箱唯一性 | 后端更新前查重，冲突返回 409 |
| 至少保留一个凭证 | 流程设计为「替换」而非「删除」——用户只能将手机/邮箱换为新值（需验证码），不提供清空操作，天然保证凭证不会丢失 |
| 登出设备 | 排除当前 session |
| 信息展示 | API 层脱敏，防止中间人获取完整号码 |
| CSRF | 现有 JWT Cookie + SameSite 已覆盖 |

## 8. 后端实现要点

- 在现有 `handler/me/` 下新增端点，复用 `ProtectedHandlerBase`
- `change-password`：bcrypt compare 旧密码 → hash 新密码 → `UserRepo.UpdatePassword`
- `change-phone` / `change-email`：
  - 调 `verifyCode.Verify(ctx, channel, address, code)`
  - 查重：`UserRepo.GetByPhone/Email`，非空则 409
  - 更新：`UserRepo.UpdatePhone/Email`
- `revoke-sessions`：从 session claims 取当前 session ID，新增 `SessionRepo.RevokeAllByUserExcept(userID, exceptSessionID)` 方法
- `login-activity`：新增 `SessionRepo.ListByUser(userID, limit)` 查询（含已过期/已吊销的）
- `SendCode` 需新增 `purpose: "bind"` 支持：跳过"用户必须存在"校验（当前仅 `"register"` 跳过）

## 9. 前端实现方案

| 层 | 位置 | 职责 |
|----|------|------|
| 路由 | `routes/account/index.tsx` | 页面入口，组合 features 组件 |
| Feature | `features/account/` | hooks、components、index.ts |
| API | `api/account.ts` | HTTP 请求封装 |
| 共享组件 | 复用现有 Dialog、Form、Input | — |

页面入口从 Header 用户头像下拉菜单进入。

## 10. 数据变更

无需新增表或字段。`sessions` 表已有 `ip`、`user_agent`、`created_at`，可直接支持登录活动查询。

P1 实现 login-activity 时需补充索引（现有 partial index 仅覆盖 `revoked_at IS NULL` 的活跃 session，无法加速含历史记录的查询）：

```sql
CREATE INDEX idx_sessions_user_created ON sessions(user_id, created_at DESC);
```

如果后续需要长期保留登录记录（session 过期删除后仍可查），再考虑独立 `user_login_logs` 表：

```sql
-- 预留方案，当前不实现
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

### P0（MVP，首期交付）

- [ ] GET /me/profile（脱敏手机/邮箱/姓名/企业列表）
- [ ] 设置/修改密码（根据 hasPassword 区分流程）
- [ ] 更换手机号（验证码流程）
- [ ] 更换邮箱（验证码流程）
- [ ] 登出所有设备
- [ ] SendCode 支持 `purpose: "bind"`（后端改动）
- [ ] 前端 /account 页面 + features/account 模块

### P1（第二期）

- [ ] 最近登录活动（GET /me/login-activity）
- [ ] 后端 user_agent 解析：响应增加 `device` 字段（如 "Chrome 126 / macOS"），前端优先展示 `device`

### P2（远期预留）

- [ ] 头像上传
- [ ] 通知偏好设置
- [ ] 个人 API Token 管理
- [ ] 账户注销

## 12. 开放问题

| # | 问题 | 决定 |
|---|------|------|
| 1 | 更换手机/邮箱是否需要管理员审批？ | 不需要——登录态 + 验证码已足够；管理员可在成员管理看到变更 |
| 2 | 同一手机/邮箱被其他用户占用？ | 后端返回 409，前端提示「该手机号/邮箱已被其他账户绑定」 |
| 3 | 是否允许同时清空手机和邮箱？ | 不存在此场景——流程仅支持「替换」（输入新值 + 验证码），不提供「删除」操作 |
| 4 | 姓名是否允许自助修改？ | 保持只读（Member 属于 Company，由管理员维护） |
| 5 | 登录活动是否需要长期保留？ | P0 先查 sessions 表；如果有保留需求，P1 再加独立日志表 |
