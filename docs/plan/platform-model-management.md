# Platform 模型管理实施方案

> 将模型/定价管理从 SMS 迁入 apps platform，platform_admin 成为唯一 SOT。

---

## 现状

```
SMS (SOT) ──Catalog API──→ apps smssync worker ──→ models 表 + NewAPI ratio
```

- SMS 管理模型目录 + 设定售价 → 推送 SMS-NewAPI ratio
- SMS 暴露 Catalog API（OAuth2 守卫，分区版本号 + 增量同步）
- apps smssync worker 定期从 SMS 拉取 → 写 models 表（source='sms'）+ 推 apps-NewAPI ratio
- 价格存储在 NewAPI（ModelRatio/CompletionRatio），models 表不存价格
- channel 管理也通过 SMS 同步，但 platform 只有一个 channel（tokenjoy）

## 目标态

```
SaaS 模式:
  platform_admin ──→ models 表 (目录) + NewAPI ratio (价格)
  platform_admin ──→ 点击「发布」bump version
  Catalog API (只读，无认证，/api/platform/sync/) ──暴露给 local──→

Local 模式:
  catalogsync worker ──从 saas Catalog API 拉──→ local models 表 + local NewAPI ratio
```

**核心约束：**
- models 表只存目录信息，价格只在 NewAPI
- schema 不变（现有字段已足够）
- Catalog API 只读无认证，注册在 /platform 路由组下但在鉴权中间件外
- 保留租户隔离——platform 全局模型用 `CompanyID = TokenJoyCompanyID`，客户侧 guard 不动
- 不做 channel 同步——platform 只有一个固定 channel（tokenjoy），local 启动时 seed

---

## 实施步骤

### Step 1: Platform 模型管理（后端）

**目标**：platform_admin 能 CRUD 模型 + 设定价格 + 发布。

1. platform handler（`internal/http/handler/platform/`）直接组合 `store.Models()` + `adminport.Port`，不扩展 `models.Service` interface：

```go
// handler/platform/models.go — 新文件
// platform admin 操作 CompanyID=TokenJoyCompanyID 的全局模型
// 直接调用 store + adminport，不走 models.Service（后者有租户 guard）
```

现有 `models.Service` 保持不变——它是客户侧 API 的 domain 层，`requireTenantModel` 拒绝修改全局模型的逻辑继续生效。

2. 路由注册（中间件分层）：

```go
func (h *Handler) RegisterRoutes(r chi.Router) {
    // 公开只读 — Catalog sync（无鉴权）
    r.Get("/sync/versions", h.CatalogVersions)
    r.Get("/sync/catalog/models", h.CatalogModels)

    // 需要鉴权 — platform admin
    r.Post("/auth/login", h.Login)
    r.Group(func(r chi.Router) {
        r.Use(httpmiddleware.RequireSession(...))
        r.Use(httpmiddleware.RequirePlatformAdmin(...))
        // 现有
        r.Get("/companies", h.ListCompanies)
        // ...
        // 新增模型管理
        r.Get("/models", h.ListModels)
        r.Post("/models", h.CreateModel)
        r.Put("/models/{id}", h.UpdateModel)
        r.Delete("/models/{id}", h.DeleteModel)
        r.Put("/models/{id}/pricing", h.SetModelPricing)
        r.Post("/catalog/publish", h.PublishCatalog)
    })
}
```

最终路径：
```
公开只读:
  GET  /api/platform/sync/versions
  GET  /api/platform/sync/catalog/models

需要鉴权:
  GET    /api/platform/models
  POST   /api/platform/models
  PUT    /api/platform/models/:id
  DELETE /api/platform/models/:id
  PUT    /api/platform/models/:id/pricing
  POST   /api/platform/catalog/publish
```

3. Catalog API 响应格式：

```
GET /api/platform/sync/versions → { "models": N }
GET /api/platform/sync/catalog/models → { "version": N, "data": [...CatalogModel] }
```

CatalogModel 格式（扩展 capabilities + maxContext）：

```json
{
  "modelId": "gpt-4o",
  "displayName": "GPT-4o",
  "provider": "openai",
  "callType": "chat",
  "inputPrice": 15.0,
  "outputPrice": 60.0,
  "capabilities": ["chat", "vision"],
  "maxContext": 128000
}
```

> 价格从 NewAPI 实时读取后拼入响应，不存在 models 表。

4. 发布机制（原子 increment）：

```go
// store/system_settings_repo.go — 新增方法
func (r *pgSystemSettingsRepo) Increment(ctx context.Context, key string) (int, error) {
    var val int
    err := r.db.QueryRow(ctx, `
        INSERT INTO system_settings (key, value) VALUES ($1, '1')
        ON CONFLICT (key) DO UPDATE SET value = (system_settings.value::int + 1)::text
        RETURNING value::int
    `, key).Scan(&val)
    return val, err
}
```

```
platform_admin 编辑模型/价格（实时写入 models 表 + NewAPI）
  ↓
点击「发布」→ POST /api/platform/catalog/publish
  ↓
store.SystemSettings().Increment("catalog.models_version")
  ↓
local catalogsync worker 下次检查 versions 时发现变更 → 拉取最新数据
```

### Step 2: catalogsync worker（改造 smssync）

**目标**：local 从 saas platform 拉取模型 + 价格。

1. `worker/smssync/` → rename 为 `worker/catalogsync/`
2. `integration/sms/` → 替换为 `integration/catalogsync/`：
   - URL 从 SMS 改为 saas platform Catalog API
   - 无需认证（Catalog API 公开只读）
   - 数据结构复用 `PartitionVersions` + `CatalogModel`（扩展 capabilities/maxContext）
3. worker 逻辑简化：
   - 只同步 models 分区（删除 channels 分区）
   - 比较版本号 → 拉取 → SyncFromPlatform + ReplaceModelRatios
4. `SyncFromSMS` rename → `SyncFromPlatform`，source 值 `'sms'` → `'platform'`
5. SaaS 模式不启动 catalogsync，Local 模式启动

配置：

```env
# Local 模式
CATALOG_SYNC_ENABLED=true
CATALOG_SYNC_URL=https://app.tokenjoy.com
CATALOG_SYNC_INTERVAL_SEC=300
```

### Step 3: 前端 + 清理

**前端**（`apps/frontend/src/features/platform/models/`）：
- 模型列表页：表格展示（名称、type、provider、输入/输出价格、状态）+ 筛选
- 创建/编辑表单：name、type、provider、capabilities、maxContext、价格
- 价格编辑弹窗：输入价格/输出价格（元/M tokens）
- 「发布」按钮：调用 publish API

**清理**：
- 删除 `integration/sms/`
- 删除 `http/handler/sync/smssync.go`
- 删除 `SMS_SYNC_*` 配置项
- 修改 `internal/app/compose_worker.go`（`buildSMSSyncExecutor` → `buildCatalogSyncExecutor`）
- 修改 `internal/infra/jobs/` 中 SMS sync periodic job → catalog sync
- models 表中 `source` 字段：SaaS platform 写入 `source='platform'`，local catalogsync 写入 `source='platform'`（同源）

---

## 数据流详解

### SaaS: platform_admin 创建模型

```
POST /api/platform/models { type: "gpt-4o", name: "GPT-4o", provider: "openai", inputPrice: 15, outputPrice: 60, capabilities: ["chat","vision"], maxContext: 128000 }
  │
  ├─ 写 models 表: (company_id=TokenJoyCompanyID, provider="openai", type="gpt-4o", name="GPT-4o", source="platform", active=true, capabilities=["chat","vision"], max_context=128000)
  │
  └─ 调 adminport.UpsertModelRatio("gpt-4o", 15, 60)
       → NewAPI: modelRatio = 15/2 = 7.5, completionRatio = 60/15 = 4
```

### SaaS: 发布

```
POST /api/platform/catalog/publish
  │
  └─ store.SystemSettings().Increment("catalog.models_version") — 原子 +1
```

### SaaS: Catalog API 响应（供 local 拉取）

```
GET /api/platform/sync/versions
  └─ { "models": system_settings['catalog.models_version'] }

GET /api/platform/sync/catalog/models
  │
  ├─ 读 models 表 (company_id=TokenJoyCompanyID, source='platform', active=true)
  ├─ 读 NewAPI ListModelPricing → 拿到所有 ratio
  └─ 拼成 CatalogModel[] 返回（含 capabilities, maxContext）
```

### Local: catalogsync 拉取

```
catalogsync worker (定期 CATALOG_SYNC_INTERVAL_SEC)
  │
  ├─ GET saas/api/platform/sync/versions → 比较本地 system_settings['catalog.models_version']
  │
  │  版本相同 → skip
  │  版本不同 ↓
  │
  ├─ GET saas/api/platform/sync/catalog/models → CatalogModel[]
  ├─ SyncFromPlatform: upsert local models 表 (source='platform', company_id=TokenJoyCompanyID)
  ├─ ReplaceModelRatios: 逐个调 local adminport.UpsertModelRatio
  └─ 更新本地 system_settings['catalog.models_version'] = remote version
```

### Local: channel 管理

```
local 启动时 seed 一个固定 channel:
  name="tokenjoy", baseURL=<SaaS gateway URL>, models="*"

不做动态 channel 同步。channel 配置通过 env/seed 管理。
```

### 客户读模型列表（任何模式）

```
GET /api/models
  │
  ├─ 读 models 表 (active=true, company 可见范围)
  ├─ 读 NewAPI ListModelPricing → 合并价格
  └─ 返回 ModelInfoWithPricing[]
```

（与现有逻辑完全一致，无需改动客户侧。）

---

## 文件变更清单

| 操作 | 路径 |
|------|------|
| 新建 | `internal/http/handler/platform/models.go` — platform 模型 CRUD + catalog sync 路由 |
| 修改 | `internal/http/handler/platform/handler.go` — RegisterRoutes 中间件分层 + 新路由 |
| 改造 | `internal/worker/smssync/` → `internal/worker/catalogsync/` |
| 新建 | `internal/integration/catalogsync/client.go`（替代 sms client，无认证） |
| 修改 | `internal/store/models_repo.go` — `SyncFromSMS` → `SyncFromPlatform` |
| 修改 | `internal/store/postgres/models_repo_crud.go` — rename 实现 + source='platform' |
| 修改 | `internal/store/postgres/system_settings_repo.go` — 新增 `Increment` 方法 |
| 修改 | `internal/app/compose_worker.go` — `buildSMSSyncExecutor` → `buildCatalogSyncExecutor` |
| 修改 | `internal/infra/jobs/` — SMS sync periodic job → catalog sync |
| 修改 | `internal/http/router.go` — 移除 `synchandler.Mount` 条件块 |
| 修改 | `internal/config/config.go` — `SMS_SYNC_*` → `CATALOG_SYNC_*` |
| 删除 | `internal/integration/sms/` |
| 删除 | `internal/http/handler/sync/smssync.go` |
| 新建 | `apps/frontend/src/features/platform/models/` |

---

## 不做的事

- models 表不加字段（现有 schema 够用）
- models 表不存价格（价格只在 NewAPI）
- 不做 per-company 差异化定价（后续需求再加）
- 不做 channel 同步（platform 只有一个固定 channel tokenjoy，seed 管理）
- Catalog API 不加认证（只读公开，无敏感数据）
- 不扩展 `models.Service` interface（platform handler 直接组合 store + adminport）
- 不做向后兼容（项目没上线）
