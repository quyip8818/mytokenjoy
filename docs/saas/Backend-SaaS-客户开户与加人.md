# SaaS 客户开户与加人（实现文档）

| | |
| --- | --- |
| 状态 | **部分已落地；缺口按 PR 切片实施** |
| 日期 | 2026-07-14 |
| 关联 | [Backend.md](./Backend.md) §2.4 · [Backend-NewAPI-多租户钥匙代建.md](./Backend-NewAPI-多租户钥匙代建.md) · [权限管理.md](./权限管理.md) · [Frontend.md](./Frontend.md) §5.2 / §5.9 · [Roadmap.md](./Roadmap.md) |

---

## 0. 架构符合性（先读）

**方向正确，核心租户/NewAPI 链路已对齐；不需要重做开户事务或钥匙模型。**

现状是 **「平台运营开户 + 超管 accept-invite」**，不是完整的 **「客户自助注册 Company + 企内邀请加人」**。差距主要在 **产品入口与企业内 Invite 写通**，不在 Postgres↔NewAPI 边界。

| 子系统 | 判定 |
| --- | --- |
| CreateCompany + NewAPI 钱包用户 + 回滚 | ✅ 符合；自助开户应 **复用** 本 service，勿另起一套 |
| SaaS `company_id ≥ 1_000_000`、双 Session | ✅ 符合 |
| 超管 invite 写入 + `POST /auth/accept-invite` | ✅ 后端符合；❌ 前端无页/无 API 封装 |
| 企业内 InviteMember | ❌ `501 NotImplemented`；BatchInvite 假计数 |
| 公开自助注册 Company | ❌ 无；仅 `POST /api/platform/companies`（平台 operator） |
| NewAPI 单 Admin 代建 / 客户不碰 Admin | ✅ 符合目标态；勿改为每企 Admin AT |

**要改：** Invite 写通、accept-invite 前端、契约字段统一、（可选）自助开户公开 API + 限流、平台控制台 UI。  
**不必改：** `service_create` 内 CreateUser/钱包语义、一企一钱包、Admin 代建 Token、平台/企业 Session 分离。

---

## 1. 目标语义（SSOT）

```text
开 Company（平台代开 或 客户自助，同一 domain）
  → Postgres company + NewAPI wallet user W
  → InviteEmail 模式: 生成 invite / UserID 模式: 注册人即超管
加人（企内）
  → company_invites → accept-invite(userID) → member + 企业 JWT
发 Key
  → CreateToken(user_id=W)（已有 as-built）
```

| 不变量 | |
| --- | --- |
| S1 | 企业 API 必须绑定 `company_id`；禁止跨企 |
| S2 | 客户永不持有 `NEW_API_ADMIN_TOKEN` |
| S3 | 客户不登录 NewAPI；钱包密码不入库 |
| S4 | 该企 PlatformKey → Token.`user_id` == wallet |
| S5 | 平台 JWT ≠ 企业 JWT（双 Secret） |
| S6 | 开户 CreateUser 失败必须整单 ROLLBACK |

客户感知：Company / Member / PlatformKey。不感知 NewAPI `users`。

---

## 2. 现状索引

| 能力 | 路径 | 状态 |
| --- | --- | --- |
| 平台开户 | `POST /api/platform/companies` → `company/service_create.go` | ✅（InviteEmail 模式） |
| 钱包用户 | `provisionCompany` 内 `CreateUser` → `newapi_wallet_company_id` | ✅ |
| 超管 invite 行 | CreateCompany InviteEmail 模式写 `company_invites`（7 天） | ✅ |
| accept-invite | `POST /api/auth/accept-invite` → `service_invite.go` | ✅ API + 双路径（已登录/未登录） |
| 已登录接受邀请 | `GET /api/auth/invites/pending` + `POST /auth/accept-invite` | ✅ |
| 注册流程 | `POST /auth/register/init` + `/accept` + `/company` | ✅ API（SaaS only + RegistrationEnabled） |
| 企内 InviteMember | `org/structure/member_batch.go` → **NotImplemented** | ❌ |
| BatchInvite | 返回假 `sent`，不写 invite | ❌ |
| 手动/导入加人 | `CreateMember` / batch-import，立即 active、无密码 | ✅ 组织面；非密码自助 |
| 前端 platform / invite | Frontend.md：未接入 | ❌ |

契约：后端 body 字段为 `inviteCode`（已统一）。

---

## 3. Gap → 实施项

| ID | 项 | 优先级 | 说明 |
| --- | --- | --- | --- |
| A | 契约收口：`inviteCode` | P0 | Frontend.md / 类型与后端对齐 |
| B | accept-invite 前端 | P0 | `authApi.acceptInvite` + `/invite/accept` 路由页 |
| C | InviteMember 写通 | P0 | 写 `company_invites`；复用 AcceptInvite；测 domain + handler |
| D | BatchInvite 真实化或降级 | P1 | 真写 invite **或** 明确 501/隐藏 UI；禁止假成功 |
| E | 自助开户 API（可选） | P1 | 新公开/半公开端点 → **调用现有** `CreateCompany`；限流 + slug/email 校验；**禁止**无鉴权暴露 `/platform/companies` |
| F | 平台控制台前端 | P2 | platform login + 开户/列企；展示超管 invite 链接 |
| G | 发信 | P2 | invite 邮件可后置；先返回 code/链接 |

---

## 4. PR 切片（按序）

### PR-A — 契约与本文

- 统一 `inviteCode`（Frontend.md § auth）
- 本文为实施 SSOT；`Backend.md` 已链本文

### PR-B — accept-invite 前端

| 交付 | |
| --- | --- |
| API | `apps/frontend/src/api/auth.ts`：`acceptInvite({ inviteCode, password, name? })` |
| 路由 | `/invite/accept`（公开）；成功后 `refreshSession` / 进企业面（对齐 `use-login-page`） |
| 不改 | Backend AcceptInvite 行为 |

### PR-C — 企业内 InviteMember（主缺口）

| 交付 | |
| --- | --- |
| Domain | 实现 `InviteMember`：校验权限/邮箱 → 写 `company_invites`（role=member 或请求角色）→ 返回 `inviteCode`（及可选过期时间） |
| HTTP | 现有 org members invite 路由改为成功语义；错误码沿用 domain |
| Accept | **复用** `AcceptInvite`；不新建激活 API |
| 测 | `tests/domain/org` + handler：invite → accept → login |
| 非目标 | 本 PR 可不发邮件 |

`BatchInvite`：本 PR 可仍 501，或在 PR-D 一并真写多条 invite。

### PR-D — BatchInvite

- 对每个目标写 invite **或** UI 去掉批量邀请直至实现
- 禁止「返回 sent>0 但库中无行」

### PR-E — 自助开户（产品确认后）

```text
POST /api/auth/register-company   # 示例路径；限流
  → company.Service.CreateCompany（同今日逻辑）
  → 201 + inviteCode（超管）
```

| 护栏 | |
| --- | --- |
| 鉴权 | 公开或验证码/邮箱验证；**不是**平台 Cookie |
| 滥用 | 限流、slug 唯一、邮箱唯一策略 |
| NewAPI | 仍单 Admin AT；失败 ROLLBACK |
| 平台路径 | `POST /platform/companies` **保留** 给运营 |

### PR-F — 平台前端

- `platformApi` + `/platform/login` + 开户表单；展示 invite 链接给超管邮件/复制

---

## 5. 安全（实施约束）

| 做 | 不做 |
| --- | --- |
| 自助开户用独立公开路由 + 限流 | 把 CreateCompany 挂到无鉴权的 `/platform/*` |
| 企业 invite 仅本企 RBAC（`org:members`） | 客户持有 NewAPI Admin |
| accept-invite 一次性 code | 第二套 CreateUser-per-member |
| 保持双 Session Secret | NewAPI Console 对企业开放 |

NewAPI Admin by-id 能力仍仅服务 Backend；SaaS 客户增多时 **不要** 给多名真人开 NewAPI Admin 账号（见钥匙代建文）。

---

## 6. 验收

| # | 条件 |
| --- | --- |
| 1 | 平台开户仍：company + wallet + 超管 invite；CreateUser 失败无脏 company |
| 2 | 企内 InviteMember → DB 有 invite 行 → accept-invite 得 member 可 login |
| 3 | 前端 `/invite/accept` 可用；字段为 `inviteCode` |
| 4 | BatchInvite 无假成功（真写或不可达） |
| 5 | （若做 E）自助开户与平台开户共用 CreateCompany；限流生效 |
| 6 | 无企业 NewAPI 密码/AT 入库 |

---

## 7. 代码索引

```text
开户     domain/company/service_create.go       — CreateCompany (双模式: UserID / InviteEmail)
                                                  provisionCompany (内部 helper)
                                                  addMember (内部 helper, 幂等)
         http/handler/platform/handler.go       — POST /platform/companies (InviteEmail 模式)
邀请激活 domain/company/service_invite.go       — AcceptInvite (接收 UserID)
                                                  PendingInvitesForUser (batch 查)
         http/handler/auth/handler.go           — POST /auth/accept-invite (双路径)
                                                  GET /auth/invites/pending
注册流程 http/handler/register/handler.go       — POST /auth/register/init, /accept, /company
         identity/registertoken/token.go        — 短期注册 session JWT
Mode     http/middleware/mode_guard.go          — RequireSaaS / RequireLocal
企内邀请 domain/org/structure/member_batch.go   ← 待实现
         http/handler/org/member.go
钥匙     docs/Backend-NewAPI-多租户钥匙代建.md
```

---

## 8. 决策记录

| 日期 | 决策 |
| --- | --- |
| 2026-07-14 | 现架构方向正确；不重做开户/NewAPI 边界 |
| 2026-07-14 | 主缺口：InviteMember + accept-invite 前端；自助开户为可选新入口复用 CreateCompany |
| 2026-07-14 | 本文升格为实施文档 |
| 2026-07-18 | CreateCompany 重构为双模式（UserID / InviteEmail）；AcceptInvite 接收 UserID（handler 层负责 User 创建）|
| 2026-07-18 | 注册流程实现：register/init + accept + company 三端点；RequireSaaS + RegistrationEnabled 守卫 |
| 2026-07-18 | PendingInvitesForUser 新增（支持 email/phone/userID 多标识匹配）|
