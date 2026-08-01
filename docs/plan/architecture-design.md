# TokenJoy 平台架构设计（早期概览，已被更精确的文档取代）

> **状态：早期目标态设计草案，多处结构性描述已与实际实现不符，仅保留核心判断（一份代码/一个 binary/`SUPPORT_SAAS` 切换、platform_admin 为模型定价 SOT）作为历史参考。**
>
> 权威文档请参阅：[apps/docs/Backend-架构.md](../../apps/docs/Backend-架构.md)（实际分层/目录结构）、[apps/docs/Local-SaaS-架构.md](../../apps/docs/Local-SaaS-架构.md)（Local↔SaaS 协作机制）。
>
> 与实际代码的主要偏差（供追溯参考，不代表当前或计划中的实现）：
> - 本文档设想的独立 `internal/platform/` 子树不存在；platform 相关代码实际扁平放在 `internal/http/handler/platform/`
> - 本文档设想的"Usage Reporter"（local 定期上报用量到 saas）机制不存在；实际总 key 消耗直接记在 SaaS wallet 上，天然可见，无需额外上报
> - "Impersonate"（platform_admin 切换到任意 company context）未实现
> - 数据库表设计（`wallets`、`wallet_transactions`、`budgets`、`recharge_records`、`notifications`、`audit_logs`）与实际 schema 表名不符，实际用的是 `companies`、`company_recharge_orders`、`company_recharge_lots`、`org_nodes`、`platform_keys`、`usage_ledger`、`usage_buckets`、`budget_consumed`、`approval_requests`、`notification_log`、`operation_logs` 等
> - 第 14 节"迁移路线"的 Phase 1-4 均未按本文档设想的路径执行，实际架构通过不同路径达成了等价目标

---

## 1. 一句话概述

TokenJoy 是一个 LLM API 管控平台。客户通过 TokenJoy 的 Gateway 调用 LLM，按用量计费。一份代码、一份 schema、一个 binary，通过 `SUPPORT_SAAS` 环境变量切换部署模式。

---

## 2. 商业模型

```
我们：
  - 拥有一个 LLM 渠道（接各大模型供应商）
  - 通过 platform 管理模型目录和定价
  - 客户按我们的定价使用我们的模型

客户：
  - 通过 TokenJoy Gateway 调用 LLM（我们的渠道）
  - 也可以添加自己的渠道（自带 key，自管，我们不管）
  - 用量按定价扣费
```

NewAPI 是内部 LLM 路由引擎，**不对外暴露**。客户只看到 TokenJoy 的 `/v1/*` Gateway 接口。

---

## 3. 两种部署模式

```env
SUPPORT_SAAS=true|false
```

| SUPPORT_SAAS | 部署者 | 一句话 |
|:---:|--------|--------|
| `true` | 我们 | 全功能：客户功能 + 平台管理（按账户角色区分可见内容）+ 多租户注册 |
| `false` | 客户（私有化） | 单租户客户功能，从 saas 同步模型/定价，向 saas 上报用量 |

**`SUPPORT_SAAS=true` 时 platform 管理是内置的**——不是独立模式，而是同进程中受权限保护的一组路由。platform_admin 登录后能看到管理菜单，普通用户看不到。

---

## 4. 模式能力矩阵

| 能力 | SaaS (true) | Local (false) |
|------|:----:|:-----:|
| 客户功能 `/api/*` | ✅ | ✅ |
| 平台管理 `/api/platform/*`（需 platform_admin 角色） | ✅ | ❌ |
| 模型/定价 CRUD（SOT） | ✅（platform_admin） | ❌ |
| Company 创建、充值、调账 | ✅（platform_admin） | ❌ |
| Impersonate 任何 company | ✅（platform_admin） | ❌ |
| 全局运营统计 | ✅（platform_admin） | ❌ |
| 多租户自助注册 | ✅ | ❌ |
| 客户自管渠道（自带 key） | ✅ | ✅ |
| LLM Gateway `/v1/*` | ✅ | ✅ |
| Ingest（扣费） | ✅ | ✅ |
| Catalog Sync（从 saas 拉） | ❌ | ✅ |
| 首次启动向 saas 注册 companyID | ❌ | ✅ |
| 上报用量到 saas | ❌ | ✅ |

> **渠道说明**：我们自己的渠道是基础设施配置（运维层面），不在 UI 里管理。客户可以在客户功能里添加自己的渠道，走自己的 key，我们不介入。
> 
> **SaaS 不需要上报用量**：platform_admin 和客户共享同一个 DB，直接查询即可看到所有 company 开销。

---

## 5. 部署拓扑

### 5.1 SaaS 模式（SUPPORT_SAAS=true，我们的环境）

```
┌─────────────────────────────────────────────────────────────────────┐
│  我们的环境                                                          │
│                                                                     │
│  ┌──────────────────────────────────────────────────────────┐       │
│  │  apps/frontend（一个 SPA）                                 │       │
│  │  - 普通用户登录 → 客户功能                                  │       │
│  │  - platform_admin 登录 → 客户功能 + 平台管理菜单            │       │
│  └────────────────────────────┬─────────────────────────────┘       │
│                               │                                     │
│                               │ /api/* + /api/platform/* + /v1/*    │
│                               ▼                                     │
│  ┌─────────────────────────────────────────────────────────────┐    │
│  │            TokenJoy Server（同进程，两组路由）                 │    │
│  │                                                             │    │
│  │  SUPPORT_SAAS=true 时注册：                                  │    │
│  │  ┌─ platform 路由（仅 platform_admin 可访问）────────────┐   │    │
│  │  │  /api/platform/models/*      模型 CRUD               │   │    │
│  │  │  /api/platform/pricing/*     定价管理                 │   │    │
│  │  │  /api/platform/companies/*   company 管理/充值        │   │    │
│  │  │  /api/platform/sync/*        Catalog API (供 local)   │   │    │
│  │  │  /api/platform/impersonate   切换到任何 company        │   │    │
│  │  │  /api/platform/dashboard     全局统计                 │   │    │
│  │  └──────────────────────────────────────────────────────┘   │    │
│  │                                                             │    │
│  │  ┌─ 客户路由（所有登录用户）────────────────────────────┐   │    │
│  │  │  /api/auth/*          登录、注册                      │   │    │
│  │  │  /api/billing/*       钱包、充值                      │   │    │
│  │  │  /api/budget/*        预算                           │   │    │
│  │  │  /api/keys/*          API Key                        │   │    │
│  │  │  /api/models/*        模型列表（读）                   │   │    │
│  │  │  /api/channels/*      客户自管渠道                    │   │    │
│  │  │  /api/org/*           组织架构                        │   │    │
│  │  │  /api/dashboard/*     客户统计                        │   │    │
│  │  │  /api/approval/*      审批                           │   │    │
│  │  │  /v1/*                LLM Gateway                    │   │    │
│  │  └──────────────────────────────────────────────────────┘   │    │
│  │                                                             │    │
│  │  Ingest Worker（poll NewAPI logs → 扣费）                    │    │
│  └──────────────────────────────┬──────────────────────────────┘    │
│                                 │                                   │
│  ┌──────────────────────────────┴──────────────────────────────┐    │
│  │                  PostgreSQL（共享）                            │    │
│  └─────────────────────────────────────────────────────────────┘    │
│                                                                     │
│  ┌─────────────────────────────────────────────────────────────┐    │
│  │         NewAPI (内部 LLM 路由引擎，不对外暴露)                 │    │
│  └─────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────┘
```

> platform 和 saas 路由在同一个进程里注册。权限隔离靠 `RequirePlatformAdmin` 中间件，不靠部署拆分。

### 5.2 Local 模式（SUPPORT_SAAS=false，客户私有化环境）

```
┌── local 实例（客户环境）─────────────────────────────────┐
│                                                           │
│  ┌─────────────────┐                                      │
│  │ 客户前端         │                                      │
│  └────────┬────────┘                                      │
│           │ /api/*  +  /v1/*                              │
│           ▼                                               │
│  ┌───────────────────────────────────────────────────┐    │
│  │     TokenJoy Server (SUPPORT_SAAS=false)            │    │
│  │                                                   │    │
│  │  /api/*             客户功能（单 company）          │    │
│  │  /v1/*              LLM Gateway                   │    │
│  │  Catalog Sync       从 platform 拉模型/定价        │    │
│  │  Ingest Worker      从 NewAPI logs 扣费            │    │
│  │  Usage Reporter     上报用量到 platform            │    │
│  └──────────────────────┬────────────────────────────┘    │
│                         │                                 │
│  ┌──────────────────────┴──────────────────────────┐      │
│  │         PostgreSQL（本地，同一份 schema）          │      │
│  └─────────────────────────────────────────────────┘      │
│                                                           │
│  ┌─────────────────────────────────────────────────┐      │
│  │    NewAPI (本地，内部路由引擎，不对外暴露)         │      │
│  └─────────────────────────────────────────────────┘      │
└───────────────────────────────────────────────────────────┘
         ↕ Catalog Sync（拉取）+ Usage Report（上报）
         ↕
┌── saas 实例 ─────────────────────────────┐
│  /api/platform/sync/*   Catalog API       │
│  /api/platform/usage/*  用量接收          │
└───────────────────────────────────────────┘
```

---

## 6. 数据流

### 6.1 模型/定价写入

```
SaaS 模式（SUPPORT_SAAS=true）：
  platform_admin ──→ /api/platform/models/* ──→ models 表 + NewAPI model_ratio
  普通用户 ──→ /api/models/*（只读）

local 模式：
  Catalog Sync 从 saas /api/platform/sync/* 拉取 → 写入本地 models 表 + 本地 NewAPI ratio
```

**核心原则：platform_admin 是唯一写入者（SOT）。普通用户只读。local 拉取副本。**

### 6.2 LLM 调用链路

```
客户 SDK ──→ TokenJoy /v1/* (Gateway)
                    │
                    ├─ budget precheck（余额/预算校验）
                    │
                    ▼
              NewAPI（内部，路由到 LLM 供应商）
                    │
                    ▼
              NewAPI logs 表
                    │
                    ▼
              Ingest Worker ──→ 匹配 company/key ──→ 扣 wallet
```

客户看到的只有 `/v1/*`，不知道 NewAPI 的存在。

### 6.3 渠道

```
我们的渠道（唯一）：
  - 基础设施配置，运维直接配 NewAPI channels
  - platform 不管、客户不可见
  - 客户用我们的 model 时，流量走这个渠道

客户自管渠道（可选）：
  - 客户在 /api/channels/* 自己添加（带自己的 API key）
  - 对应 NewAPI 里的 channel 记录
  - 我们不干预、不管理、不负责
  - 客户可以用自己渠道的 model，但定价/扣费规则由 models 表决定
```

### 6.4 定价模型

```
platform 管理员设置：
  ├─ 全局默认售价（所有 company 共享的基准）
  └─ per-company 差异化售价（覆盖特定 company）

NewAPI 映射：
  - Group = Company（每个 company 一个 group）
  - 全局定价 → model_ratio
  - per-company 定价 → group ratio 覆盖
  - platform_key 创建时 → NewAPI token 关联到对应 group
```

### 6.5 Local 生命周期

```
1. 首次启动：向 saas 注册 → 获得 companyID → 本地持久化
2. 定期同步：Catalog Sync 拉取最新模型/定价
3. 持续上报：Usage Reporter 定期将用量汇总发送到 saas
4. saas 侧：platform_admin 可查看该 company 的总体开销，防止 local 被滥用
```

---

## 7. 功能模块清单

```
TokenJoy Server（一个 binary）
│
├── Platform 管理 /api/platform/*（SUPPORT_SAAS=true，需 platform_admin 角色）
│   ├── Auth（平台管理员登录）
│   ├── Models（模型目录 CRUD、上下线）
│   ├── Pricing（全局定价 + per-company 定价 → 同步 NewAPI ratio）
│   ├── Companies（创建/充值/调账/状态/开销统计）
│   ├── Catalog API（/api/platform/sync/* — 供 local 拉取）
│   ├── Usage 接收（/api/platform/usage/* — 接收 local 上报）
│   ├── Impersonate（切换到任何 company context）
│   └── Dashboard（全局运营统计）
│
├── 客户功能 /api/*（saas + local）
│   ├── Auth（登录、invite-accept）
│   ├── Register（自助注册，仅 SUPPORT_SAAS=true）
│   ├── Company（成员管理）
│   ├── Org（组织架构）
│   ├── Billing（钱包、充值记录）
│   ├── Budget（项目/成员预算管控）
│   ├── Keys（Platform Key CRUD + NewAPI token 同步）
│   ├── Models（模型列表 + 定价展示 — 读 models 表）
│   ├── Channels（客户自管渠道 — 自带 key 接供应商）
│   ├── Dashboard（客户用量统计）
│   ├── Approval（审批流）
│   ├── Notification（通知）
│   └── Audit（操作日志）
│
├── Gateway /v1/*（saas + local，对外暴露的 LLM 接口）
│   └── 接收客户请求 → budget precheck → 转发给内部 NewAPI
│
├── Ingest Worker（saas + local，poll NewAPI logs → 扣费）
│
├── Catalog Sync Worker（仅 local）
│   └── 定期从 saas 的 /api/platform/sync/* 拉取模型/定价
│
└── Usage Reporter（仅 local）
    └── 定期上报用量到 saas
```

---

## 8. 后端代码结构

```
apps/backend/
├── cmd/server/main.go
├── internal/
│   ├── app/                        # 启动组装
│   ├── config/                     # SUPPORT_SAAS 等配置
│   │
│   ├── domain/                     # 业务逻辑（客户侧）
│   │   ├── company/                # 公司、成员
│   │   ├── billing/                # 钱包、充值、扣费
│   │   ├── budget/                 # 预算管控
│   │   ├── models/                 # 模型列表（客户视角，读 models 表）
│   │   ├── keys/                   # Platform Key + NewAPI token 同步
│   │   ├── org/                    # 组织架构
│   │   ├── usage/                  # 用量统计
│   │   ├── approval/               # 审批流
│   │   ├── notification/           # 通知
│   │   ├── audit/                  # 操作日志
│   │   ├── dashboard/              # 客户统计
│   │   ├── gateway/                # LLM gateway precheck
│   │   ├── adminport/              # NewAPI admin 操作抽象接口
│   │   ├── grants/                 # 权限
│   │   ├── types/                  # 共享类型
│   │   └── port/                   # 通用 port 接口
│   │
│   ├── platform/                   # 平台管理域（SUPPORT_SAAS=true，platform_admin 专属）
│   │   ├── domain/
│   │   │   ├── pricing/            # 模型目录 + 定价 + NewAPI ratio 推送
│   │   │   ├── company/            # 运营视角：company 管理/充值/调账
│   │   │   ├── catalog/            # Catalog API（供 local 拉取）
│   │   │   └── dashboard/          # 全局运营统计
│   │   ├── handler/                # /api/platform/* 路由
│   │   └── store/                  # platform 专属查询
│   │
│   ├── http/
│   │   ├── router.go               # 路由注册总入口
│   │   ├── handler/                # /api/* 客户路由
│   │   │   ├── auth/
│   │   │   ├── billing/
│   │   │   ├── budget/
│   │   │   ├── keys/
│   │   │   ├── models/
│   │   │   ├── channels/           # 客户自管渠道
│   │   │   ├── org/
│   │   │   ├── dashboard/
│   │   │   ├── approval/
│   │   │   ├── notification/
│   │   │   ├── audit/
│   │   │   ├── ingest/
│   │   │   └── dev/
│   │   └── middleware/
│   │
│   ├── identity/                   # 认证、session、JWT
│   │
│   ├── integration/
│   │   ├── newapi/                 # NewAPI admin HTTP client（内部）
│   │   └── newapisync/             # PlatformKey ↔ NewAPI token 双向同步
│   │
│   ├── worker/
│   │   ├── ingest/                 # Ingest worker（poll logs → 扣费）
│   │   ├── catalogsync/            # Catalog sync（仅 local）
│   │   └── usagereport/            # Usage reporter（仅 local → 上报 saas）
│   │
│   ├── store/                      # 数据库 repository
│   ├── infra/                      # 基础设施（邮件、Redis 等）
│   └── pkg/                        # 内部工具包
```

---

## 9. 数据库

### 9.1 统一 Schema

一份 schema，所有模式共用。

**客户侧表（saas + local）：**

| 表 | 用途 |
|----|------|
| companies | 公司 |
| members | 成员 |
| org_nodes | 组织架构 |
| wallets | 钱包 |
| wallet_transactions | 钱包交易记录 |
| budgets | 预算 |
| platform_keys | API Key |
| platform_key_mappings | Key ↔ NewAPI token 映射 |
| recharge_records | 充值记录 |
| approval_* | 审批 |
| notifications | 通知 |
| audit_logs | 操作日志 |

**模型/定价表（saas 下 platform_admin 写入，普通用户只读，local 由 Catalog Sync 写入副本）：**

| 表 | saas 写入 | saas 读取 | local |
|----|:---:|:---:|:---:|
| models | platform_admin | 所有用户 | Catalog Sync 写入 |
| model_pricing | platform_admin | 所有用户 | Catalog Sync 写入 |

### 9.2 NewAPI 侧（内部，不对外）

| 表 | 写入者 |
|----|--------|
| channels | 运维配置（我们的渠道）+ 客户自管渠道写入 |
| tokens | newapisync（PlatformKey 模块） |
| abilities | newapisync（RebuildAbilities） |
| model_ratio | platform_admin pricing handler / Catalog Sync (local) |
| groups | company 创建时 |
| logs | NewAPI 内部路由（被 Ingest Worker 消费） |

---

## 10. 认证与权限

| 角色 | 路由 | 适用模式 |
|------|------|---------|
| platform_admin | `/api/platform/*` + `/api/*`（只读） | saas |
| company_admin | `/api/*` 本 company | saas + local |
| company_member | `/api/*` 本 company（受限） | saas + local |

- 统一登录入口，登录后按角色看到不同菜单
- platform_admin = 属于 TokenJoyCompanyID 的 member
- `/api/platform/*` 中间件强制校验 platform_admin，local 模式不注册这组路由
- Impersonate：platform_admin 在同一 SPA 内切换 company context（无需跳转）

### platform_admin 对客户数据的权限规则

```
/api/platform/*   → 完整读写（模型、定价、company 管理、充值、调账）
/api/*            → 只读任何 company 的数据，不能写入/修改

具体：
  GET  /api/billing/*     ✅ 可以查看任何 company 的钱包
  POST /api/billing/*     ❌ 不能操作客户的钱包（充值走 /api/platform/companies/*)
  GET  /api/keys/*        ✅ 可以查看任何 company 的 Key
  POST /api/keys/*        ❌ 不能帮客户创建 Key
  GET  /api/org/*         ✅ 可以查看任何 company 的组织架构
  PUT  /api/org/*         ❌ 不能修改客户的组织
```

**实现方式**：platform_admin 访问 `/api/*` 时，中间件注入只读 company context（跳过 companyID 归属校验，但拦截所有写操作）。需要写入客户数据的管理操作（充值、状态变更）必须通过 `/api/platform/companies/*` 专用接口。

---

## 11. 配置

```env
# ─── 模式 ───
SUPPORT_SAAS=true|false

# ─── 通用 ───
PORT=8010
DATABASE_URL=postgres://...
LOG_DATABASE_URL=postgres://...
REDIS_URL=redis://...
DEPLOY_ENV=local|staging|production

# ─── NewAPI（内部组件，所有模式都需要）───
NEW_API_ENABLED=true
NEW_API_BASE_URL=http://newapi:3000       # 内网地址，不对外
NEW_API_ADMIN_USER_ID=1
NEW_API_WEBHOOK_SECRET=xxx
NEW_API_GATEWAY_ENABLED=true

# ─── local 专属 ───
CATALOG_SYNC_URL=https://app.tokenjoy.com
CATALOG_SYNC_COMPANY_ID=uuid
CATALOG_SYNC_API_KEY=secret
CATALOG_SYNC_INTERVAL_SEC=300
```

---

## 12. 启动逻辑（伪代码）

```go
func New(cfg config.Config) (*App, error) {
    db := connectDB(cfg.DatabaseURL)
    newAPIClient := newapi.NewClient(cfg.NewAPIBaseURL)
    router := chi.NewRouter()

    // 客户功能 + Gateway + Ingest（所有模式）
    mountCustomerRoutes(router, db, newAPIClient, cfg)
    mountGateway(router, cfg)
    startIngestWorker(db, cfg)

    if cfg.SupportSaas {
        // SaaS：platform 管理路由（RequirePlatformAdmin 中间件保护）
        mountPlatformRoutes(router, db, newAPIClient, cfg)
    } else {
        // Local：Catalog Sync + Usage Report
        startCatalogSync(db, newAPIClient, cfg)
        startUsageReporter(cfg)
    }

    return &App{Router: router}, nil
}
```

---

## 13. 前端

一个 SPA（`apps/frontend/`），登录后按角色显示不同内容：

| 角色 | 看到的内容 | 后端 API |
|------|-----------|---------|
| platform_admin | 客户功能 + 平台管理菜单 | `/api/*` + `/api/platform/*` |
| company_admin / member | 仅客户功能 | `/api/*` |

```
apps/frontend/src/
├── features/
│   ├── ...（现有客户功能）
│   └── platform/              # 平台管理页面（lazy load）
│       ├── models/            # 模型目录 CRUD
│       ├── pricing/           # 定价管理
│       ├── companies/         # company 管理/充值
│       ├── dashboard/         # 全局统计
│       └── index.ts
├── routes/
│   ├── ...（客户路由）
│   └── platform/              # /platform/* 前端路由
└── components/
    └── layout/
        └── sidebar            # isPlatformAdmin 时渲染 platform 菜单
```

- platform 路由前端通过 `isPlatformAdmin` 守卫，后端通过 `RequirePlatformAdmin` 中间件双重校验
- local 模式后端不注册 `/api/platform/*`，前端请求直接 404
- impersonate：platform_admin 在同一 SPA 内切换 company context，无需跳转域名

---

## 14. 迁移路线

### Phase 1：配置重构

- [ ] 保持 `SUPPORT_SAAS` bool 不变（`true` = SaaS 全功能，`false` = Local 私有化）
- [ ] 确保 router.go 中 `if cfg.SupportSaas` 注册 platform 路由
- [ ] 确保 `!cfg.SupportSaas` 时启动 Catalog Sync + Usage Reporter

### Phase 2：Platform 管理增强

- [ ] 创建 `internal/platform/` 子树
- [ ] 从现有 `http/handler/platform/` 迁入，扩展模型/定价管理
- [ ] 移除 SMS Sync（`worker/smssync/`、`integration/sms/`），模型数据由 platform_admin 直接管理
- [ ] Catalog API：`/api/platform/sync/*` 供 local 拉取

### Phase 3：Local Catalog Sync

- [ ] `worker/catalogsync/` 实现：从 saas 的 sync API 拉取模型/定价
- [ ] Local 首次启动向 saas 注册 company
- [ ] Usage Reporter：定期上报用量到 saas

### Phase 4：前端 Platform 功能

- [ ] `features/platform/` 目录（lazy load）
- [ ] 模型/定价管理页面
- [ ] Company 管理 + 充值 + 开销统计
- [ ] Impersonate（切换 company context）
- [ ] 路由守卫 + 侧边栏按角色渲染

---

## 15. 核心原则

1. **一份代码、一份 schema、一个 binary** — `SUPPORT_SAAS` bool 切换两种模式
2. **platform_admin 是唯一 SOT 写入者** — 模型/定价只在 SaaS 模式由 platform_admin 写入
3. **local 通过 Catalog Sync 获取副本** — 定期从 saas 拉取，上报用量回 saas
4. **NewAPI 是内部组件** — 不对外暴露，客户只看到 TokenJoy `/v1/*` Gateway
5. **渠道分两层** — 我们的渠道（运维配置）+ 客户自管渠道（客户功能），互不干扰
6. **权限靠后端中间件** — 不靠前端隔离或部署拆分，同一个 SPA 按角色渲染不同菜单
7. **platform_admin 掌控全局** — 所有 company 的开销可见，impersonate 排查问题，防止 local 被滥用
8. **编译产物唯一** — 一个 Docker image，环境变量决定行为
