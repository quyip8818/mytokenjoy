# Local 版首次启动公司注册

> **状态：已实现** — 见 commit `feat: local setup — runtime CompanyID resolution + setup UI`

> local（selfhosted）版第一次启动时，引导用户向 SaaS 平台注册公司，获取 companyId 并持久化。后续启动直接使用该 companyId，不再依赖硬编码的 `LOCAL_COMPANY_ID`。
>
> 实现差异：
> - 无 setup token（不需要额外保护，setup server 一次性运行）
> - 无离线模式（SaaS 不可达时直接报错）
> - `StoreBootstrap.SchemaPrepared` 拆分为 `SkipSchema` + `SkipSeed`

---

## 现状

### 后端

- `SUPPORT_SAAS=false` 时，bootstrap 使用 `LOCAL_COMPANY_ID`（硬编码 `00000000-0000-7000-8000-000000000002`）创建一个 type=selfhosted 的公司。
- `seed.Init` 在 `store/postgres.New` 中调用（比 `app.New` 更早），`ApplyBootstrap` 使用 `appCfg.LocalCompanyID` 直接插入公司行。
- 已有 `CatalogSync`：local 通过 `CATALOG_SYNC_URL` 从 SaaS 平台拉取模型目录（无认证，公开只读 API）。

### 前端

- `IS_SAAS` 通过 Vite 构建时环境变量 `VITE_SUPPORT_SAAS` 决定。
- `index.html` 是纯静态文件，后端不渲染前端 HTML。

### 问题

1. `LOCAL_COMPANY_ID` 是编译时静态 UUID，多个独立部署之间会冲突。
2. Local 版不向 SaaS 注册，SaaS 侧无法追踪 selfhosted 部署。
3. 首次启动时 DB 为空，bootstrap 直接写入虚拟公司，跳过了用户确认。

---

## 目标

local 版首次启动时：
1. 前端展示"系统初始化"引导页（公司名称、行业、规模、管理员信息）
2. 后端向 SaaS 平台注册公司，获得真实 `companyId`
3. 将 companyId 持久化到 `system_settings`
4. **然后**执行完整 bootstrap + 创建管理员
5. 正常启动应用

**不做的事**：
- 不做 SaaS 账户关联
- 不做公司信息后续编辑
- 不改动 SaaS 版任何现有流程

---

## 核心设计：先 Setup 再完整启动

**删除 `LocalCompanyID` 概念**。Config 不再有静态的 `LOCAL_COMPANY_ID` 环境变量。companyId 完全在运行时 resolve（从 system_settings 读取或 setup 流程产出）。

Setup 在 bootstrap 之前完成。App 启动时 companyId 已确定，bootstrap / 中间件 / 路由无需任何条件分支。

```
main()
  → config.Load()（不再有 LOCAL_COMPANY_ID 字段）
  → openPool + applySchema（DDL only）
  → resolveCompanyID(pool, cfg):
      - SaaS 模式 → 直接用固定 DemoCompanyID（不走 setup）
      - Local 模式：
        1. system_settings.setup_company_id（已 setup）
        2. 没有 → 未初始化

  ├─ 有 companyId → cfg.CompanyID = id（运行时字段，非 env）
  │                → seed.Init（完整 ApplyBootstrap）
  │                → app.New → 正常服务
  │
  └─ 无 companyId → 起 setupServer（mini HTTP，复用已有 pool）
                   → 用户填表 → 调 SaaS 拿 companyId
                   → 写 system_settings.setup_company_id
                   → 创建管理员 user（只写 users 表 + bcrypt hash）
                   → 关闭 setupServer
                   → cfg.CompanyID = newId
                   → seed.Init + app.New → 正常服务
```

**注意**：现有 `postgres.New()` 内部同时做 `applySchema` + `seed.Init`。需要拆开：
- `applySchema` 提前执行（setupServer 需要写 system_settings / users 表，表必须存在）
- `seed.Init` 延后到 companyId 确定后执行
- 通过 `cfg.StoreBootstrap.SchemaPrepared = true` 告诉后续 `postgres.New` 跳过重复 schema apply

**优点**：
- 不拆 bootstrap — `ApplyBootstrap` 保持原样，一揽子执行
- 不存在"app 启动了但公司没建"的中间态
- 去除硬编码 UUID，彻底消除部署冲突

### `LocalCompanyID` 删除影响

现有代码中 `cfg.LocalCompanyID` 的使用点（全部改为 `cfg.CompanyID`）：

| 位置 | 用途 | 改法 |
|------|------|------|
| `config/config.go` PlatformConfig | env 字段定义 | 删除字段，改为 `CompanyID uuid.UUID`（运行时赋值，不从 env 读取） |
| `config/validate.go` | 校验非 Nil | 删除相关校验（运行时 resolve 保证） |
| `seed/bootstrap/bootstrap.go` | `appCfg.LocalCompanyID` | 改为 `appCfg.CompanyID` |
| `http/middleware/company_resolve.go` | non-SaaS 隐式租户 | 改为 `cfg.CompanyID` |
| `domain/company/service.go` | 保护不可修改 | 改为 `cfg.CompanyID` |
| `app/dev_bootstrap.go` | demo wallet | 改为 `cfg.CompanyID` |
| `integration/newapisync` | demo wallet 判断 | 改为 `cfg.CompanyID` |
| `handler/dev/handler.go` | dev API 上下文 | 改为传入的 companyID |
| `seed/contract/ids.go` | 测试固定值 | 保留为测试常量 `contract.DefaultCompanyID`，测试中通过 `TestConfig()` 赋给 `cfg.CompanyID` |
| 测试文件 | `cfg.LocalCompanyID = ...` | 改为 `cfg.CompanyID = ...` |

---

## 架构

```
┌─────────────────────────────────────────────┐
│  Local 前端 (首次访问 setupServer)            │
│  /setup → SetupPage                         │
│  填写: 公司名 / 行业 / 规模 / 管理员邮箱+密码  │
└──────────────────┬──────────────────────────┘
                   │ POST /api/setup/init
                   ▼
┌─────────────────────────────────────────────┐
│  setupServer (mini HTTP, bootstrap 前)       │
│  1. 调 SaaS API 注册公司 → 拿 companyId      │
│  2. 写 system_settings.setup_company_id      │
│  3. 写 users 表（管理员 bcrypt hash）         │
│  4. 返回成功，关闭自己                        │
└──────────────────┬──────────────────────────┘
                   │ POST /api/platform/register-local
                   ▼
┌─────────────────────────────────────────────┐
│  SaaS 后端 (SaaS only 公共端点)              │
│  创建 type=selfhosted 公司记录               │
│  返回 { companyId }                          │
└─────────────────────────────────────────────┘
                   ▼
┌─────────────────────────────────────────────┐
│  setupServer 关闭后                           │
│  → seed.Init（用 companyId 完整 bootstrap）   │
│  → app.New → 正常服务                        │
└─────────────────────────────────────────────┘
```

---

## API 设计

### SaaS 侧：`POST /api/platform/register-local`（新增，SaaS only）

不走 platform admin 认证链，用注册密钥防滥用。

```
POST /api/platform/register-local
Header: X-Registration-Secret: <shared_secret>
```

请求体：
```json
{
  "name": "极光科技",
  "industry": "互联网",
  "size": "50-200",
  "idempotencyKey": "uuid-v7"
}
```

响应：
```json
{
  "companyId": "uuid-v7"
}
```

实现：
- 在 platform handler 的**公开路由段**（与 `/sync/versions` 同级）注册
- 验证 `X-Registration-Secret` 与 `cfg.LocalRegistrationSecret` 匹配
- `idempotencyKey` 去重：相同 key 返回已创建的 companyId
- 仅插入 `companies` 表（type=selfhosted），**不创建 wallet、不发邮件**——wallet 由 local 侧 bootstrap 时通过 `provisionCompany` 创建
- 返回 companyId

### Local 侧：`POST /api/setup/init`（setupServer 端点）

只存在于 setupServer（bootstrap 前的临时 HTTP），正常启动后不可访问。

请求体：
```json
{
  "companyName": "极光科技",
  "industry": "互联网",
  "size": "50-200",
  "adminEmail": "admin@aurora.com",
  "adminPassword": "secure-password-123",
  "adminName": "张三",
  "setupToken": "xxx"
}
```

响应：
```json
{
  "companyId": "uuid-v7",
  "status": "ok"
}
```

流程：
1. 校验 `setupToken`（启动时生成，输出到 stdout）
2. 向 SaaS `POST /api/platform/register-local` 注册公司，拿到 companyId（离线模式本地生成 UUID v7）
3. 写 `system_settings: setup_company_id = companyId`
4. 写 `system_settings: setup_company_name = companyName`
5. 写 `users` 表：管理员 email + bcrypt hash + name（只写 user 行，member/roles 由后续 bootstrap 处理）
6. 返回成功 → setupServer 自行关闭

### Local 侧：`GET /api/setup/status`（setupServer 端点）

前端判断 setupServer 是否就绪。

```json
{ "ready": true, "setupTokenHint": "前 4 位..." }
```

---

## `provisionCompany` 改造

`CreateCompanyRequest` 新增 `CompanyID` 字段：

```go
type CreateCompanyRequest struct {
    CompanyID    uuid.UUID // 可选；非 Nil 时使用外部 ID（仅 local setup 场景）
    UserID       uuid.UUID
    Name         string
    Industry     string
    Size         string
    Type         string
    InviteEmail  string
    MemberAlias  string
    MemberAvatar string
}
```

`provisionCompany` 中：

```go
companyID := req.CompanyID
if companyID == uuid.Nil {
    companyID = uuid.Must(uuid.NewV7())
}
```

**信任边界在 handler 层**：SaaS platform handler 永远不填 `CompanyID`；bootstrap 时通过 `cfg.CompanyID` 传入 setup 拿到的 ID。

---

## Bootstrap 改造（最小改动）

`ApplyBootstrap` 保持不变。唯一改动是 `insertCompanies` 中 local company 的创建需要使用 setup 拿到的 companyId（通过 `cfg.CompanyID` 传入）。

因为 setup 在 bootstrap 之前完成，`cfg.CompanyID` 在进入 `seed.Init` 时已经是真实 ID。**不需要拆分 bootstrap**。

管理员创建：setup 阶段只写了 `users` 表（email + hash）。Bootstrap 完成后，`BootstrapPlatformIfNeeded`（或类似逻辑）读取 setup 写入的 user，创建对应的 member + role。

改造点：
- `BootstrapPlatformIfNeeded` 改名或新增 `BootstrapLocalAdminIfNeeded`
- 从 `system_settings` 读 admin email → 查 users 表 → 创建 member（role=super_admin, company=cfg.CompanyID）

---

## 配置

### SaaS 侧新增

```
LOCAL_REGISTRATION_SECRET=<hex-32-bytes>    # 空值 = 不开放注册
```

### Local 侧新增

```
SAAS_PLATFORM_URL=https://platform.tokenjoy.com
SAAS_REGISTRATION_SECRET=<same-secret>
SETUP_OFFLINE_MODE=false                    # true = 跳过远程注册，本地生成 UUID
```

### Local 侧删除

```
LOCAL_COMPANY_ID  → 彻底删除此环境变量和 config 字段
COMPANY_NAME      → 删除 validate 校验（改为从 system_settings.setup_company_name 读取）
```

### Config 结构体变更

```go
type PlatformConfig struct {
    SupportSaas       bool      `env:"SUPPORT_SAAS" envDefault:"false"`
    CompanyID         uuid.UUID // 运行时赋值，不从 env 读取
    TokenJoyCompanyID uuid.UUID `env:"TOKENJOY_COMPANY_ID" envDefault:"00000000-0000-7000-8000-000000000001"`
    // ... 其他字段 ...

    // Setup / registration (local only)
    SaasPlatformURL        string `env:"SAAS_PLATFORM_URL"`
    SaasRegistrationSecret string `env:"SAAS_REGISTRATION_SECRET"`
    SetupOfflineMode       bool   `env:"SETUP_OFFLINE_MODE" envDefault:"false"`
}
```

**SaaS 模式下 CompanyID**：SaaS 版 bootstrap 也用一个 companyID 创建 demo 公司（用于种子数据/开发测试）。SaaS 模式下 `resolveCompanyID` 直接返回固定常量 `DemoCompanyID`（`00000000-0000-7000-8000-000000000002`），不走 setup。只有 local 模式才需要 setup。

```go
func resolveCompanyID(pool *pgxpool.Pool, cfg config.Config) uuid.UUID {
    if cfg.SupportSaas {
        return DemoCompanyID // 固定常量，SaaS 不走 setup
    }
    // local 模式：从 system_settings 读取
    id := readSystemSetting(pool, "setup_company_id")
    if id != uuid.Nil { return id }
    return uuid.Nil // 未初始化
}
```

测试通过 `cfg.CompanyID = contract.DefaultCompanyID` 赋值，行为不变。

---

## 持久化

| key | value | 写入时机 |
|-----|-------|----------|
| `setup_company_id` | uuid string | setup 完成时 |
| `setup_company_name` | string | setup 完成时 |
| `setup_admin_email` | string | setup 完成时 |
| `setup_idempotency_key` | uuid string | setup 调 SaaS 前 |

**companyId 解析优先级**（main 启动时）：
- SaaS 模式 → 固定 `DemoCompanyID` 常量（不走 setup）
- Local 模式：
  1. `system_settings.setup_company_id`（已初始化）
  2. 没有 → 未初始化 → 起 setupServer

测试：`TestConfig()` 直接赋 `cfg.CompanyID = contract.DefaultCompanyID`，不走 resolve。

---

## setupServer 实现

```go
// cmd/server/main.go 或 app/setup_server.go
func runSetupServer(ctx context.Context, pool *pgxpool.Pool, cfg config.Config) (uuid.UUID, error) {
    setupToken := generateSetupToken() // 随机 hex，输出到 stdout
    log.Printf("Setup token: %s", setupToken)

    r := chi.NewRouter()
    r.Use(middleware.RealIP, middleware.Logger)

    // 静态文件（前端 /setup 页面）
    r.Handle("/*", http.FileServer(http.Dir(cfg.StaticDir)))

    // Setup API
    r.Get("/api/setup/status", handleSetupStatus(setupToken))
    r.Post("/api/setup/init", handleSetupInit(pool, cfg, setupToken))

    srv := &http.Server{Addr: ":" + cfg.Port, Handler: r}
    // 阻塞直到 setup 完成（handler 内部 signal）
    // setup 完成后 graceful shutdown
    ...
    return companyID, nil
}
```

setupServer 只依赖：
- `*pgxpool.Pool`（写 system_settings + users）
- `cfg`（SaaS URL / secret / port）
- 一个 HTTP client（调 SaaS register-local）

不依赖 `companySvc`、`assembleRegistry` 或任何 domain service。

---

## 前端

### 初始化检测

前端 local 模式启动时请求 `GET /api/setup/status`：
- 如果 setupServer 在运行 → `{ ready: true }` → 显示 `/setup` 页面
- 如果正常 app 在运行（已初始化）→ 404 → 正常进入登录页

### 路由

| 路径 | 说明 |
|------|------|
| `/setup` | 初始化引导页，仅 local 模式，setupServer 运行时可见 |

### 流程

1. 用户访问 local 实例 → 前端 `GET /api/setup/status`
2. 收到 `{ ready: true }` → 渲染 setup 表单
3. 用户填写公司信息 + 管理员信息 + setupToken → `POST /api/setup/init`
4. 成功 → 提示"初始化完成，即将重启"→ 等几秒后刷新页面 → 进入正常登录页

### UI 复用

复用 SaaS 版 `auth-card.tsx` 中 `create_company` 的表单字段（公司名/行业/规模），
额外增加管理员邮箱+密码+setupToken 字段。

### 代码组织

```
features/setup/
  hooks/use-setup-page.ts
  components/setup-form.tsx
  index.ts
routes/setup.tsx
api/setup.ts
```

---

## 后端代码组织

```
app/
  setup_server.go      — setupServer 逻辑（mini HTTP + setup handler）
  resolve_company.go   — resolveCompanyID（env → system_settings → 未初始化）

handler/platform/
  register_local.go    — SaaS 侧 register-local 端点
```

### router 注册（SaaS 侧）

在 `platform/handler.go` 的 `RegisterRoutes` 公开路由段：
```go
r.Post("/register-local", h.RegisterLocal)
```

### main 入口改造

```go
func main() {
    cfg := config.Load()       // 不再有 LocalCompanyID 字段
    pool := openPool(cfg)
    applySchema(pool, cfg)     // DDL — 从 postgres.New 中提取出来

    companyID := resolveCompanyID(pool, cfg) // SaaS→固定ID; local→读 system_settings
    if companyID == uuid.Nil {
        // 未初始化 → 起 setupServer，阻塞等用户完成
        id, err := runSetupServer(ctx, pool, cfg)
        if err != nil { log.Fatal(err) }
        companyID = id
    }

    cfg.CompanyID = companyID             // 运行时赋值
    cfg.StoreBootstrap.SchemaPrepared = true  // 告诉 postgres.New 跳过 schema
    // seed.Init 在 postgres.New 内部执行（有 companyID 了）
    // 正常启动：app.New → 正常服务
    ...
}
```

---

## 安全

- `X-Registration-Secret` 防止任意人调 SaaS 注册端点
- setupServer 的 `setupToken`（一次性，输出到 stdout）防止公网抢注
- setupServer 只在未初始化时运行，初始化后 `/api/setup/*` 不存在（正常 app 不挂这些路由）
- 管理员密码 bcrypt 存储

---

## 幂等与故障恢复

1. Setup handler 调 SaaS 前先写 `setup_idempotency_key` 到 system_settings
2. SaaS `register-local` 用 idempotencyKey 去重（相同 key 返回相同 companyId）
3. Setup handler 写 `setup_company_id` 后才返回成功
4. 如果中途崩溃：重启 → resolveCompanyID 未找到 → 再次起 setupServer → 用户重试 → idempotencyKey 保证不重复创建

---

## 离线部署

`SETUP_OFFLINE_MODE=true`：
- 跳过远程注册，本地生成 UUID v7
- 其他流程不变
- TokenJoyCompanyID 仅作为角色锚点

---

## 开发环境重置

```bash
pnpm reset local test
```

效果：清除 `system_settings` 中 setup_* 相关 key + 清空 companies/members 等数据。
下次启动重新进入 setupServer。

---

## 文件变更预估

| 操作 | 路径 | 说明 |
|------|------|------|
| 新建 | `app/setup_server.go` | setupServer + setup handler |
| 新建 | `app/resolve_company.go` | resolveCompanyID 逻辑 |
| 新建 | `handler/platform/register_local.go` | SaaS 侧公开注册端点 |
| 修改 | `handler/platform/handler.go` | 公开路由段加 register-local |
| 修改 | `config/config.go` | 删除 `LocalCompanyID` env 字段，改为运行时 `CompanyID`；加 SaasPlatformURL / SaasRegistrationSecret / SetupOfflineMode |
| 修改 | `config/validate.go` | 删除 LocalCompanyID 相关校验 |
| 修改 | `store/postgres/postgres.go` | 提取 applySchema 为可独立调用的公开函数；支持 SchemaPrepared 跳过 |
| 修改 | `seed/bootstrap/bootstrap.go` | `appCfg.LocalCompanyID` → `appCfg.CompanyID` |
| 修改 | `seed/bootstrap/companies.go` | 同上 |
| 修改 | `http/middleware/company_resolve.go` | `cfg.LocalCompanyID` → `cfg.CompanyID` |
| 修改 | `domain/company/service.go` | CreateCompanyRequest 加 CompanyID 字段；保护判断改用 `cfg.CompanyID` |
| 修改 | `domain/company/service_create.go` | provisionCompany 支持外部 ID |
| 修改 | `handler/dev/handler.go` | `localCompanyID` → 从 cfg.CompanyID 获取 |
| 修改 | `app/dev_bootstrap.go` | `cfg.LocalCompanyID` → `cfg.CompanyID` |
| 修改 | `integration/newapisync` | 同上 |
| 修改 | `cmd/server/main.go` | 启动流程：schema → resolve → setup → seed → app |
| 修改 | `identity/credentials/service.go` | 新增 BootstrapLocalAdminIfNeeded |
| 修改 | `seed/contract/ids.go` | 删除 `LocalCompanyID`，保留 `DefaultCompanyID` 供测试用 |
| 修改 | 测试文件（多处） | `cfg.LocalCompanyID` → `cfg.CompanyID` |
| 新建 | `features/setup/` | 前端 setup feature |
| 新建 | `routes/setup.tsx` | 前端 setup 页面 |
| 新建 | `api/setup.ts` | 前端 setup API |

---

## 不做的事

- 不做 SaaS 账户绑定
- 不做 companyId 迁移（新安装才走此流程）
- 不做多次注册（setupServer 只运行一次）
- 不做公司信息编辑
- 不拆分 bootstrap（setup 在 bootstrap 之前完成）
- 彻底删除 `LOCAL_COMPANY_ID` 概念，消除硬编码 UUID 冲突
