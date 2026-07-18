# 统一 Platform Operator 身份 — Super Company 方案

> **目标**：消除独立的 `platform_operators` 体系，operator 变为 `TokenJoyCompanyID`（super company）下的普通 member，共享统一的 session/refresh/权限机制。  
> **收益**：删代码 > 加代码；一套身份、一套 session、一套 audit。  
> **前置依赖**：token-refresh 方案已落地。  
> **前提**：系统未上线，破坏性变更可接受，无存量数据迁移需求。

---

## 1. 现状问题

| 当前 | 问题 |
| --- | --- |
| `platform_operators` 独立表 | 维护两套用户体系（users/members vs operators） |
| `PLATFORM_SESSION_SECRET` 独立 secret | 多一个配置项，多一个泄露面 |
| `tokenjoy_platform_session` 独立 cookie | 两套 cookie 管理逻辑 |
| `PlatformAuth` 独立 middleware | 两套鉴权路径 |
| `credentials.AuthenticatePlatform` | 两套认证逻辑 |
| Platform operator 无 refresh token | 体验不一致 |
| Platform 操作无法复用现有权限系统（roles/permissions） | 权限管控是 hardcoded |
| `PlatformOperatorFromContext` 存 `uuid.UUID` 断言 `string` | 现有 audit trail 中 operatorID 实际为 nil（已知 bug） |

---

## 2. 目标架构

```
TokenJoyCompanyID (super company)
  └── members (原 platform_operators)
        ├── role: platform_admin（绑定 platform:manage 权限）
        └── 使用与普通租户完全相同的 session + refresh + 权限机制
```

Platform operator **不再是独立实体**，而是 super company 下拥有 `platform_admin` 角色的 member。

---

## 3. Schema 变更

```sql
-- members 表：department_id 改为 nullable
-- 原: department_id UUID NOT NULL
-- 新:
department_id UUID,
FOREIGN KEY (company_id, department_id) REFERENCES org_nodes (company_id, id) ON DELETE RESTRICT
```

> PostgreSQL 复合 FK 中任一列为 NULL 时自动不检查引用，无需额外处理。  
> Platform member 没有部门归属，`department_id = NULL` 比造假 org_node 更诚实。

同时删除 `platform_operators` 表定义。

---

## 4. Bootstrap 改造

`BootstrapPlatformIfNeeded` 改为在 super company 下创建 user + member + 赋角色：

```go
func (s *service) BootstrapPlatformIfNeeded(ctx context.Context) error {
    if s.cfg.PlatformBootstrapEmail == "" || s.cfg.PlatformBootstrapPassword == "" {
        return nil
    }
    // 幂等：检查 super company 下是否已有此 email 的 member
    existing, _, _ := s.store.Org().MemberByEmail(ctx, s.cfg.TokenJoyCompanyID, s.cfg.PlatformBootstrapEmail)
    if existing != nil {
        return nil
    }
    // 创建 user（幂等）
    hash, err := bcrypt.GenerateFromPassword([]byte(s.cfg.PlatformBootstrapPassword), bcrypt.DefaultCost)
    if err != nil {
        return err
    }
    userID := uuid.New()
    // ... INSERT INTO users ON CONFLICT (email) DO NOTHING, 取回实际 user ID
    // 创建 member（department_id = NULL）
    memberID := uuid.New()
    // ... INSERT INTO members (id, company_id, user_id, name, department_id, status)
    //     VALUES (memberID, TokenJoyCompanyID, userID, email, NULL, 'active')
    // 确保 platform_admin 角色存在并赋予
    // ... INSERT INTO roles ... ON CONFLICT DO NOTHING
    // ... INSERT INTO member_roles ...
    return nil
}
```

`platform_admin` 角色在 bootstrap 中幂等创建，不依赖手动 seed。

---

## 5. 代码变更

### 5.1 删除项

| 模块 | 删除内容 |
| --- | --- |
| `store/platform_repo.go` | `PlatformOperator` struct, `PlatformRepository` interface |
| `store/postgres/platform_repo.go` | 整个文件 |
| `store/store.go` | `Platform() PlatformRepository` 方法 |
| `identity/credentials/service.go` | `AuthenticatePlatform` 方法 |
| `identity/httpx/token.go` | `PlatformSessionCookie`, `ResolvePlatformSessionToken`, `SetPlatformSessionCookie`, `ClearPlatformSessionCookie`, `ParsePlatformToken` |
| `identity/httpx/context.go` | `WithPlatformOperator`, `PlatformOperatorFromContext` |
| `http/middleware/platform_auth.go` | 整个文件 |
| `http/deps/public.go` | `PlatformSessionToken` 字段 |
| `http/deps/deps.go` | `PlatformSessionToken` 字段 |
| `config/config.go` | `PlatformSessionSecret` 字段 |
| `config/validate.go` | `PlatformSessionSecret` 校验 |
| `app/compose_http.go` | `platformToken` 创建逻辑 |
| `store/postgres/schema.sql` | `platform_operators` 表定义 |

### 5.2 新增/修改项

| 模块 | 变更 |
| --- | --- |
| `store/postgres/schema.sql` | `members.department_id` 去掉 `NOT NULL` |
| `infra/permission` | 新增 `platform:manage` 权限常量 |
| `identity/credentials/service.go` | `BootstrapPlatformIfNeeded` 重写（见 Section 4） |
| `http/middleware/require_platform.go` | 新建 `RequirePlatformAdmin` middleware |
| `http/middleware/company_resolve.go` | 移除 `/api/platform/` 的 skip 逻辑 |
| `http/handler/platform/handler.go` | Login 改用 `AuthenticateMember` + `issueTokenPair`；其他 handler 从 `SessionContext.Member.ID` 取操作者 ID |
| `app/compose_http.go` | 删除 `platformToken`；platform handler 复用 `memberToken` |

---

## 6. RequirePlatformAdmin middleware

```go
func RequirePlatformAdmin(tokenJoyCompanyID uuid.UUID) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            session, ok := httpx.SessionFromContext(r.Context())
            if !ok || session.CompanyID != tokenJoyCompanyID {
                httputil.WriteStatus(w, http.StatusForbidden, httputil.MsgForbidden)
                return
            }
            if !authz.HasAny(session.Permissions, "platform:manage") {
                httputil.WriteStatus(w, http.StatusForbidden, httputil.MsgForbidden)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

必须检查 `platform:manage` 权限，不能只靠 companyID。理由：
- 成本为零（`RequireSession` 已 resolve permissions 到 context）
- 防御性编程：即使 super company 下误加了非 admin member，也不会越权

---

## 7. Platform handler 改造

### 7.1 Login

```go
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
    var body loginBody
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
        httputil.WriteStatus(w, http.StatusBadRequest, "Bad request")
        return
    }
    member, err := h.credentials.AuthenticateMember(r.Context(), h.tokenJoyCompanyID, body.Email, body.Password)
    if err != nil {
        httputil.WriteJSON(w, http.StatusUnauthorized, nil, err)
        return
    }
    if _, err := h.issueTokenPair(w, r, member.CompanyID, member.ID, member.UserID); err != nil {
        httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
        return
    }
    // 清除可能残留的旧 platform cookie
    httpx.ClearPlatformSessionCookie(w)
    httputil.WriteJSON(w, http.StatusOK, map[string]string{"memberId": member.ID.String()}, nil)
}
```

### 7.2 路由注册

```go
func (h *Handler) RegisterRoutes(r chi.Router) {
    r.Post("/auth/login", h.Login)
    r.Group(func(r chi.Router) {
        r.Use(httpmiddleware.RequireSession(h.protected))
        r.Use(httpmiddleware.RequirePlatformAdmin(h.tokenJoyCompanyID))
        r.Get("/companies", h.ListCompanies)
        r.Post("/companies", h.CreateCompany)
        r.Patch("/companies/{id}", h.UpdateCompany)
        r.Post("/companies/{id}/recharge", h.RechargeCompany)
        r.Post("/companies/{id}/gift", h.GiftCompany)
        r.Post("/companies/{id}/adjust", h.AdjustCompany)
        r.Get("/channels", h.ListChannels)
        r.Post("/channels", h.CreateChannel)
    })
}
```

### 7.3 Audit trail 改造

所有使用 `PlatformOperatorFromContext` 的 handler 改为从 session context 取操作者：

```go
// 原:
operatorIDStr, _ := httpmiddleware.PlatformOperatorFromContext(r.Context())
operatorID, _ := uuid.Parse(operatorIDStr)

// 新:
session, _ := httpx.SessionFromContext(r.Context())
operatorID := session.Member.ID
```

---

## 8. CompanyResolve middleware 调整

移除 `/api/platform/` 的 skip 逻辑：

```go
// 原:
if strings.HasPrefix(r.URL.Path, "/api/platform/") {
    next.ServeHTTP(w, r)
    return
}

// 新: 删除此分支
```

改造后 platform 路由上的 JWT claims 包含 `CompanyID = TokenJoyCompanyID`，`CompanyResolve` 会自然 resolve 到 super company context。好处：
- 日志、metrics 中能正确标注 company
- 与其他路由行为一致

---

## 9. 配置变更

| 变量 | 变化 |
| --- | --- |
| `PLATFORM_SESSION_SECRET` | **删除** |
| `PLATFORM_BOOTSTRAP_EMAIL` | 保留，含义变为"super company 首个 admin 的邮箱" |
| `PLATFORM_BOOTSTRAP_PASSWORD` | 保留，含义不变 |
| `TOKENJOY_COMPANY_ID` | 保留，作为 super company 标识 |

---

## 10. 安全考量

| 关注点 | 方案 |
| --- | --- |
| Super company member 泄露 cookie 能否访问租户数据？ | 否 — `RequireSession` middleware 仅允许访问自己 company_id 的数据 |
| 租户 member 能否访问 platform API？ | 否 — `RequirePlatformAdmin` 检查 `companyID == TokenJoyCompanyID` + `platform:manage` 权限 |
| 统一 secret 是否降低安全性？ | 否 — 原来两个 secret 只是隔离签发方，JWT 验证已通过 company_id 区分 |
| Super company 能否被误删？ | 不允许 — company 删除逻辑加 guard（与 LocalCompanyID 相同处理） |
| Super company 下非 admin member 能否越权？ | 否 — `RequirePlatformAdmin` 强制检查 `platform:manage` 权限 |

---

## 11. 不做什么

| 排除 | 理由 |
| --- | --- |
| 数据迁移脚本 | 系统未上线，无存量数据 |
| Super company 下搞组织架构/部门 | YAGNI，department_id = NULL |
| Platform admin 的细粒度权限拆分 | YAGNI，初期所有 platform admin 等权 |
| 改 platform API 路径 | `/platform/*` 保持不变，仅鉴权机制变化 |
| Platform 前端 app | 当前无需求 |
| 向后兼容/分阶段部署 | 系统未上线，一步到位 |

---

## 12. 工作量

| 内容 | 估时 |
| --- | --- |
| Schema 变更（department_id nullable + 删 platform_operators 表） | 0.25d |
| Bootstrap 改造（创建 user + member + role） | 0.5d |
| Platform handler 登录改造 + RequirePlatformAdmin middleware | 0.5d |
| CompanyResolve 调整 + handler audit trail 改造 | 0.25d |
| 删除 PlatformSession* 全套代码 + 配置清理 | 0.5d |
| 测试修复（credentials_test、middleware_test 等） | 0.5d |

**总计约 2.5 个工作日**。

---

## 13. 删除清单（checklist）

- [ ] `store/platform_repo.go`
- [ ] `store/postgres/platform_repo.go`
- [ ] `http/middleware/platform_auth.go`
- [ ] `identity/httpx/token.go` 中 Platform* 相关常量和函数
- [ ] `identity/httpx/context.go` 中 `WithPlatformOperator`、`PlatformOperatorFromContext`
- [ ] `config.PlatformSessionSecret`
- [ ] `deps.PlatformSessionToken`
- [ ] `compose_http.go` 中 `platformToken` 创建
- [ ] `schema.sql` 中 `platform_operators` 表
- [ ] `credentials.AuthenticatePlatform`
- [ ] `company_resolve.go` 中 `/api/platform/` skip 分支
- [ ] 相关测试文件更新
