# 成员邀请与注册流程设计

## 1. 背景

`CreateMember` 创建成员后，只写入一条 status=active 的记录，没有触发邀请动作。用户（member）无法收到通知、无法设置密码登录系统。

需要一个完整的「创建 → 邀请 → 注册」流程，支持多渠道分发邀请，并追踪注册来源。

---

## 2. 核心概念

| 概念 | 说明 |
|------|------|
| **创建来源 (source)** | member 通过什么方式进入系统：manual / csv / feishu / dingtalk / wecom / invited |
| **邀请渠道 (inviteChannel)** | 邀请链接通过什么渠道分发：sms / email / admin_link |
| **注册渠道 (registrationChannel)** | 用户最终通过哪个渠道完成注册（嵌入加密 token 中） |

---

## 3. 创建来源 (source)

`Member.Source` 字段，枚举值：

| 值 | 说明 |
|----|------|
| `manual` | 管理员手动添加 |
| `csv` | CSV 批量导入 |
| `feishu` | 飞书同步导入 |
| `dingtalk` | 钉钉同步导入 |
| `wecom` | 企业微信同步导入 |
| `invited` | 通过邀请码直接创建（AcceptInvite 流程） |

---

## 4. 邀请渠道

### 4.1 短信 (sms)

- 触发：创建 member 时提供了 phone
- 行为：发送短信，含注册链接 `{BASE_URL}/invite/accept?code={encrypted_token}`

### 4.2 邮件 (email)

- 触发：创建 member 时提供了 email
- 行为：发送邮件，含注册链接

### 4.3 管理员链接 (admin_link)

- 触发：管理员在 UI 点击"获取邀请链接"
- 行为：生成链接，复制到剪贴板，管理员自行分发

### 4.4 核心原则

**一个 member 只有一条 invite 记录**。DB 中 `company_invites` 只存一条记录（一个 `invite_code`）。不同渠道分发的是同一条 invite 的不同加密 token——token 内嵌不同的 `ch` 字段来区分渠道。

- DB 不存 channel，channel 仅存在于加密 token 密文中
- 同一个 invite 可同时通过短信、邮件、管理员链接分发
- 用户注册时解密 token 即可得知注册渠道

### 4.5 CreateMember 流程

```
CreateMember(input):
  1. 创建 member 记录（status = pending）
  2. 创建一条 invite 记录（生成 invite_code）
  3. if input.phone:
       encrypt({code, ch=sms}) → 发送短信
  4. if input.email:
       encrypt({code, ch=email}) → 发送邮件
  5. if !input.phone && !input.email:
       等待管理员手动获取链接
```

phone 和 email 可以同时存在，两个渠道都发送，使用同一条 invite 记录。

BatchImport 对每行执行同样的逻辑：创建 member(pending) + 创建 invite + 按 phone/email 发送。

---

## 5. 加密 Token

### 5.1 目的

邀请链接中需要携带渠道信息且不可被用户篡改。如果用明文 query param，用户可以伪造来源。

### 5.2 结构

AES-GCM 对称加密，密钥为 `INVITE_SECRET` 环境变量。

**明文 payload：**

```json
{
  "code": "invite_code_hex_string",
  "ch": "sms|email|admin_link",
  "exp": 1700000000
}
```

| 字段 | 说明 |
|------|------|
| `code` | `company_invites.invite_code`，查 DB 用 |
| `ch` | 渠道，写入 `member.registrationChannel` |
| `exp` | 过期时间戳（冗余校验，DB 中也有 `expires_at`） |

**加密：**

```
plaintext = json.Marshal(payload)
nonce = random(12 bytes)
ciphertext = AES-GCM-Seal(key, nonce, plaintext)
token = base64url(nonce + ciphertext)
```

**解密：**

```
raw = base64url.Decode(token)
nonce = raw[:12]
ciphertext = raw[12:]
plaintext = AES-GCM-Open(key, nonce, ciphertext)
payload = json.Unmarshal(plaintext)
```

### 5.3 安全性

- AES-GCM 认证加密，篡改即失败
- channel 不可伪造
- Token 过期双重校验（token.exp + DB expires_at）
- 同一条 invite 可生成多个不同 channel 的 token，底层 invite_code 唯一

---

## 6. 数据模型

### 6.1 `company_invites` 表

保持现有结构不变，新增一列：`id, company_id, email, phone, user_id, role, invite_code, expires_at, accepted_at, accepted_meta, created_at`

- `accepted_meta`：JSONB，accept 时写入，记录注册时的客户端信息：`{"ip": "...", "ua": "..."}`

每个 member 只有一条 invite 记录。渠道信息通过加密 token 携带，不持久化。

### 6.2 Member 新增字段

```go
// 注册渠道：用户最终通过哪种方式完成注册
// 值: "sms" | "email" | "admin_link" | "" (未注册)
RegistrationChannel string `json:"registrationChannel,omitempty"`
```

```typescript
registrationChannel?: 'sms' | 'email' | 'admin_link'
```

### 6.3 status 变化

| 操作 | status |
|------|--------|
| CreateMember（manual/csv） | `pending` |
| AcceptInvite 完成注册 | `active` |
| 飞书/钉钉/企微同步导入 | `active`（已有外部认证） |

---

## 7. API

### 7.1 创建成员

```
POST /org/members
```

变更：status 改为 `pending`，自动创建 invite 并按 phone/email 发送通知。

### 7.2 获取管理员邀请链接

```
POST /org/members/{id}/invite-link
```

权限：SuperAdmin / OrgAdmin
限制：member.status 必须为 pending，已 active 的成员不可获取

响应：
```json
{ "inviteLink": "https://app.tokenjoy.com/invite/accept?code=<encrypted_token>" }
```

逻辑：查找该 member 的 invite 记录（若不存在则创建）→ encrypt({code, ch=admin_link}) → 拼接 URL 返回。

### 7.3 接受邀请

```
POST /auth/accept-invite
Body: { "inviteCode": "<encrypted_token>", "name": "...", "password": "..." }
```

逻辑：
1. 解密 token → code + channel
2. 用 code 查 invite，验证未过期、未使用
3. 创建/关联 user，设置密码
4. member.status → active，member.registrationChannel → channel
5. 标记 invite accepted，写入 accepted_meta（ip/ua）

### 7.4 批量发送邀请

```
POST /org/members/batch-invite
Body: { "ids": ["uuid1", ...] }
```

对 pending 成员重新发送邀请：更新 invite 的 `expires_at`（续期），重新生成加密 token 并按 phone/email 发送。

---

## 8. 前端 UI

### 8.1 成员表格新增列

| 列 | 字段 | 展示 |
|----|------|------|
| 创建来源 | `source` | badge：手动 / CSV / 飞书 / 钉钉 / 企微 |
| 注册方式 | `registrationChannel` | 短信 / 邮件 / 管理员链接 / 未注册 |

### 8.2 操作菜单

- **重新发送邀请**：status=pending 时显示 → 调用 batch-invite → toast
- **获取邀请链接**：status=pending 且 SuperAdmin/OrgAdmin 可见 → 调用 invite-link API → 复制到剪贴板 → toast

### 8.3 创建成员后 toast

- 有 phone/email："成员已添加，邀请已发送至 {phone/email}"
- 无联系方式："成员已添加，请手动获取邀请链接"

---

## 9. 流程图

```
管理员创建成员
    │
    ├─ 创建 invite 记录（invite_code）
    │
    ├─ 有手机号 → encrypt({code, ch=sms}) → 发短信
    │
    ├─ 有邮箱 → encrypt({code, ch=email}) → 发邮件
    │
    └─ 都没有 → pending，等管理员操作
                  │
                  └─ "获取邀请链接" → encrypt({code, ch=admin_link}) → 复制

用户打开链接 /invite/accept?code=<token>
    │
    ├─ 解密 → invite_code + channel
    ├─ 查 DB 验证 invite
    ├─ 设置姓名 + 密码
    ├─ member.status → active
    ├─ member.registrationChannel → channel
    └─ 签发 session → 跳转首页
```

---

## 10. 安全

1. Token 不可伪造（AES-GCM，密钥仅服务端持有）
2. Token 过期（默认 7 天）
3. Token 一次性（accepted_at 标记后不可重复使用）
4. 渠道不可篡改（channel 在密文中）
5. 获取链接需 SuperAdmin/OrgAdmin 权限
6. 密钥轮转：`INVITE_SECRET` 支持多密钥，解密时逐一尝试

---

## 11. 配置

| 环境变量 | 说明 | 默认值 |
|----------|------|--------|
| `INVITE_SECRET` | AES-256 密钥（32 bytes hex） | 必填 |
| `INVITE_EXPIRE_HOURS` | 有效时间 | 168（7天） |
| `INVITE_BASE_URL` | 链接前缀 | 从请求 Host 推断 |

---

## 12. 实现顺序

1. `pkg/invitetoken`：加解密工具
2. Member 加 `registrationChannel` 字段
3. 改造 `CreateMember`：status=pending + 创建 invite + 按渠道加密发送
4. 改造 `AcceptInvite`：解密 token → code 查 invite → 写 registrationChannel
5. `GetMemberInviteLink` API（复用已有 invite，生成 ch=admin_link token）
6. 前端：表格新增列
7. 前端：dropdown "获取邀请链接" + "重新发送邀请"
8. 前端：创建成员 toast 优化
