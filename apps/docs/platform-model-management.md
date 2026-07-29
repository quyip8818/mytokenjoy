# Platform 模型管理

## 背景

TokenJoy 支持两种部署模式：

- **SaaS 模式**：多租户，TokenJoy 公司作为平台运营方管理全局模型目录和定价
- **Local 模式**：单租户私有化部署，通过 Catalog API 从 SaaS 拉取模型和定价

在 SaaS 模式下，platform admin（`平台管理员` 角色，属于 TokenJoy 公司）是模型目录的唯一 SOT（Source of Truth）。所有租户看到的"内置模型"列表和价格都来自 platform admin 的配置。

在 Local 模式下，不存在 platform admin。模型目录和定价通过后台 `catalogsync` worker 自动从 SaaS 的 Catalog API 同步。

---

## 两种模式的差异

| | SaaS | Local |
|---|---|---|
| 谁管理模型 | platform admin（UI 操作） | 自动同步（catalogsync worker） |
| 模型数据来源 | platform admin 手动 CRUD | 从 SaaS Catalog API 拉取 |
| 定价来源 | platform admin 设定 → `model_pricing` 表 | 独立 pricing sync 通道从 SaaS 同步 |
| TokenJoyCompany | 存在，platform admin 属于此公司 | 存在但无人登录 |
| Platform handler | 启用（SaaS only） | 不启用 |
| catalogsync worker | 不启用 | 启用 |
| `PLATFORM_BOOTSTRAP_EMAIL` | 需要配置 | 不配置 |

---

## SaaS 模式：Platform Admin 工作流

### 日常操作

1. 登录 TokenJoy 公司账号（`admin@tokenjoy.me`），nav 中出现"平台管理 → 模型目录"
2. 在模型目录页添加/编辑/删除模型，设置定价
3. 确认无误后点击"发布"按钮
4. 发布后，Local 实例会在下一个同步周期（默认 5 分钟）自动拉取最新数据

### 发布机制

发布不是"把数据推到 local"——而是 bump 一个版本号。Local worker 定期检查版本号，发现变更时主动拉取。

- **模型发布**：点击"发布"按钮 → bump `catalog.models_version`
- **定价变更**：`SetGlobalPrice`/`SetContractPrice` 自动 bump `catalog.pricing_version`（无需手动发布）

好处：模型 batch 编辑时攒一批一起发布；定价变更即时生效（下次 sync 周期内）。

---

## API 详解

### Catalog API（供 Local catalogsync worker 调用）

注册在 `/api/platform/` 路由下。分为公开端点和 sync token 保护端点。

**公开端点（无需鉴权）：**

```
GET /api/platform/sync/versions
响应: { "models": 3, "pricing": 2 }
```

返回当前模型和定价的发布版本号。Worker 分别和本地版本比较，相同则跳过。

```
GET /api/platform/sync/catalog/models
响应: {
  "version": 3,
  "data": [
    {
      "modelId": "gpt-4o",
      "displayName": "GPT-4o",
      "provider": "openai",
      "callType": "chat",
      "capabilities": ["chat", "vision"],
      "maxContext": 128000
    }
  ]
}
```

返回完整模型目录（不含定价）。全局数据，无公司隔离。定价通过独立的 `/sync/catalog/pricing` 端点获取。

**Sync token 保护端点（per-company 隔离）：**

```
GET /api/platform/sync/catalog/pricing
Header: Authorization: Bearer cst_<hex64>
响应: {
  "version": 2,
  "data": [
    { "modelType": "gpt-4o", "inputPrice": 1.0, "outputPrice": 4.0, "isContract": true },
    { "modelType": "claude-3", "inputPrice": 3.0, "outputPrice": 15.0, "isContract": false }
  ]
}
```

返回全局价 + 该公司合同价合并结果。合同价覆盖同 modelType 的全局价。`isContract=true` 表示该价格来自公司专属合同。`version` 字段为当前 `catalog.pricing_version` 值。

鉴权由 `RequireSyncToken` 中间件完成：从 `Authorization: Bearer cst_xxx` 提取 token → SHA-256 hash → 查 `companies.sync_token_hash` → 验公司 status=active → 注入 companyID 到 context。

### 注册端点（Local setup 调用）

```
POST /api/platform/register-local
Header: X-Registration-Secret: <shared_secret>
Body: { "name", "industry", "size", "idempotencyKey" }
响应 201: { "companyId": "uuid", "syncToken": "cst_..." }
响应 409: token 刚签发（60s 内），使用已有 token
```

公司创建幂等（同 idempotencyKey 不重复建），每次签发新 sync token 并覆盖旧 hash。60 秒防重窗口防止网络重试覆盖有效 token。

### Platform Admin API（需要登录 + `platform:manage` 权限）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/platform/models` | 模型列表（含实时价格） |
| POST | `/api/platform/models` | 创建模型（目录 + 定价一起写） |
| PUT | `/api/platform/models/:id` | 更新模型属性（名称、provider、状态等） |
| DELETE | `/api/platform/models/:id` | 删除模型 |
| PUT | `/api/platform/models/:id/pricing` | 仅更新定价（不改目录信息） |
| POST | `/api/platform/catalog/publish` | 发布模型目录（models version +1） |
| GET | `/api/platform/pricing` | 全局定价列表 |
| PUT | `/api/platform/pricing` | 设全局定价（自动 bump pricing version） |
| GET | `/api/platform/companies/:id/pricing` | 某公司合同价列表 |
| PUT | `/api/platform/companies/:id/pricing` | 设合同价（自动 bump pricing version） |

创建模型时的请求体：
```json
{
  "type": "gpt-4o",
  "name": "GPT-4o",
  "provider": "openai",
  "inputPrice": 2.5,
  "outputPrice": 10.0,
  "capabilities": ["chat", "vision"],
  "maxContext": 128000
}
```

---

## 数据存储

### models 表

全局模型的 `company_id = TokenJoyCompanyID`，`source = 'platform'`。

**models 表不存价格。** 价格存在 `model_pricing` 表（见 `model-pricing.md`）。

模型同步使用 `catalog_synced_at` 时间戳列做 diff-disable：同步批次内所有 model 设同一个 batch timestamp，批次结束后把 timestamp 更早的 platform model 标记为 inactive。

### model_pricing 表

Append-only 价格时间线。唯一约束 `(company_id, model_type, effective_from)`。

- 全局价：`company_id = TokenJoyCompanyID`
- 合同价：`company_id = <具体公司 ID>`

`CurrentPricesBatch` 取每个 model_type 的最新 effective_from 行。Ingest 优先用合同价，无合同价时 fallback 全局价。

### system_settings 表

| key | 说明 |
|-----|------|
| `catalog.models_version` | 模型目录发布版本号（手动 Publish bump） |
| `catalog.pricing_version` | 定价版本号（SetGlobalPrice/SetContractPrice 自动 bump） |
| `catalog_sync_token` | Local 侧的 sync token 明文（setup 写入） |
| `setup_company_id` | Local 侧注册获得的 companyId（setup 写入） |
| `register_local:<idempotencyKey>` | 注册幂等映射 → companyId |

版本号的 `Increment` 方法使用原子 SQL：

```sql
INSERT INTO system_settings (key, value) VALUES ($1, '1')
ON CONFLICT (key) DO UPDATE SET value = (system_settings.value::int + 1)::text
RETURNING value::int
```

### companies 表（sync 相关列）

```sql
sync_token_hash  CHAR(64)     -- SHA-256(cst_token)，SaaS 侧验证用
token_issued_at  TIMESTAMPTZ  -- 签发时间，60s 防重窗口判断
```

`sync_token_hash` 有 UNIQUE INDEX（WHERE NOT NULL），中间件通过 hash 反查 company。

---

## Local 模式：catalogsync worker

### 配置

```env
CATALOG_SYNC_ENABLED=true
CATALOG_SYNC_URL=https://app.tokenjoy.com
CATALOG_SYNC_INTERVAL_SEC=300
```

SaaS 模式下这三个变量不配置（worker 不启动）。

Sync token 由 setup 流程自动获取并存入 `system_settings` 表（key: `catalog_sync_token`）。Worker 启动时从 system_settings 读取，无需手动配置。

### 同步流程

```
每 CATALOG_SYNC_INTERVAL_SEC 秒执行一次：

1. GET /sync/versions → { models: M, pricing: P }

2. Models sync（独立通道）：
   - 比较本地 catalog.models_version 与远端 models version
   - 相同 → 跳过
   - 不同 → GET /sync/catalog/models
     → SyncFromPlatform: upsert 到本地 models 表
     → 更新本地 catalog.models_version

3. Pricing sync（独立通道）：
   - 比较本地 catalog.pricing_version 与远端 pricing version
   - 相同 → 跳过
   - 不同 → GET /sync/catalog/pricing (Bearer cst_xxx)
     → 遍历返回数据：
       - isContract=false → 写入 model_pricing (company_id=globalCompanyID)
                          + push NewAPI gateway cache
       - isContract=true  → 写入 model_pricing (company_id=localCompanyID)
     → 更新本地 catalog.pricing_version
```

模型和定价各自独立 version，互不影响。改价不需要重新同步模型，模型变更也不触发定价同步。

### Sync Token 生命周期

| 事件 | 行为 |
|------|------|
| Setup 首次运行 | 调 register-local → SaaS 签发 cst_ token → 存入 system_settings |
| Worker 启动 | 从 system_settings 读 token，构造 client |
| Token 丢失 | 运维重跑 setup（等 60s 窗口过），新 token 覆盖旧的 |
| Token 泄露 | 同上——重跑 setup 即 rotate |
| 公司停用 | SaaS 标记 status=inactive，sync 返回 403 |

### Channel 管理

Platform 只有一个 channel（`tokenjoy`），指向 SaaS gateway。不做动态 channel 同步。Local 启动时 seed 固定 channel 即可。

---

## 权限体系

### 后端

`RequirePlatformAdmin` 中间件做双重检查：
1. Session 的 `CompanyID == TokenJoyCompanyID`（确保是 TokenJoy 公司的人）
2. Session 持有 `platform:manage` permission

普通租户即使知道 API 路径也无法访问（CompanyID 不匹配）。

### 前端

路由 `/platform/models` 配置 `requiredPermissions: [PERMISSION.PLATFORM_MANAGE]`。只有持有该权限的用户才能在 nav 中看到"平台管理"分组。

### 角色配置

`平台管理员` 是 preset role（manifest.json），配置为 `["*"]`（全部权限）。通过 `PLATFORM_BOOTSTRAP_EMAIL` + `PLATFORM_BOOTSTRAP_PASSWORD` 环境变量在 bootstrap 阶段自动创建。

---

## 前端页面

路径：`/platform/models`

功能：
- 模型列表表格（名称/type、provider、输入价格、输出价格、状态）
- 操作：编辑定价（弹窗）、启用/禁用切换、删除
- 发布按钮：调用 publish API，bump version

代码组织：
- `features/platform/models/hooks/use-platform-models-page.ts` — 数据 + 操作逻辑
- `features/platform/models/components/platform-models-page-shell.tsx` — UI
- `api/platform.ts` — API 定义
- `router/routes/platform.ts` — TanStack Router 路由注册

---

## 文件索引

| 职责 | 路径 |
|------|------|
| **后端** | |
| Platform handler（CRUD + Catalog API） | `internal/http/handler/platform/models.go` |
| Platform pricing handler | `internal/http/handler/platform/pricing.go` |
| Catalog pricing sync endpoint | `internal/http/handler/platform/catalog_pricing.go` |
| 路由注册（中间件分层） | `internal/http/handler/platform/handler.go` |
| RequireSyncToken 中间件 | `internal/http/middleware/sync_token.go` |
| Pricing domain service | `internal/domain/pricing/service.go` |
| Catalog sync worker（executor） | `internal/worker/catalogsync/execute.go` |
| Catalog sync HTTP client | `internal/integration/catalogsync/client.go` |
| Catalog sync types | `internal/integration/catalogsync/types.go` |
| Worker 组装（compose） | `internal/app/compose_worker.go` |
| SyncFromPlatform（models 表 upsert） | `internal/store/postgres/models_repo_crud.go` |
| ModelPricing repo | `internal/store/postgres/model_pricing_repo.go` |
| Company repo（含 sync token 方法） | `internal/store/postgres/company_repo.go` |
| SystemSettings.Increment | `internal/store/postgres/system_settings_repo.go` |
| 配置字段 | `internal/config/config.go` → PlatformConfig |
| River periodic job | `internal/infra/jobs/kinds_catalogsync.go` |
| River worker adapter | `internal/infra/river/workers/catalog_sync.go` |
| Setup server（token 持久化） | `internal/app/setup_server.go` |
| **测试** | |
| Sync token + pricing 鉴权测试 | `tests/handler/platform/sync_token_test.go` |
| Pricing sync worker 测试 | `tests/worker/catalogsync/pricing_sync_test.go` |
| 模型同步测试 | `tests/domain/models/global_catalog_test.go` |
| **前端** | |
| Feature module | `apps/frontend/src/features/platform/models/` |
| API 定义 | `apps/frontend/src/api/platform.ts` |
| 路由注册 | `apps/frontend/src/router/routes/platform.ts` |
| Nav 路由配置 | `apps/frontend/src/config/routes.ts` |
