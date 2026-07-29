# Platform 模型管理

## 背景

TokenJoy 支持两种部署模式：

- **SaaS 模式**：多租户，TokenJoy 公司作为平台运营方管理全局模型目录和定价
- **Local 模式**：单租户私有化部署，通过 Catalog API 从 SaaS 拉取模型和定价

在 SaaS 模式下，platform admin（`平台管理员` 角色，属于 TokenJoy 公司）是模型目录的唯一 SOT（Source of Truth）。所有租户看到的"内置模型"列表和价格都来自 platform admin 的配置。

在 Local 模式下，不存在 platform admin。模型目录通过后台 `catalogsync` worker 自动从 SaaS 的 Catalog API 同步。

---

## 两种模式的差异

| | SaaS | Local |
|---|---|---|
| 谁管理模型 | platform admin（UI 操作） | 自动同步（catalogsync worker） |
| 模型数据来源 | platform admin 手动 CRUD | 从 SaaS Catalog API 拉取 |
| 定价来源 | platform admin 设定 → `model_pricing` 表 | 跟随 SaaS 同步到本地 `model_pricing` 表 |
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

好处：batch 编辑时不会每改一个就触发同步，攒一批一起发布。

---

## API 详解

### Catalog API（公开只读，无需登录）

供 Local 的 catalogsync worker 调用。注册在 `/api/platform/` 路由下但在鉴权中间件外面。

```
GET /api/platform/sync/versions
响应: { "models": 3 }
```

返回当前发布版本号。Worker 用它和本地版本比较，相同则跳过。

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
      "inputPrice": 2.5,
      "outputPrice": 10.0,
      "capabilities": ["chat", "vision"],
      "maxContext": 128000
    }
  ]
}
```

返回完整模型列表 + 实时价格。价格从 `model_pricing` 表读取后拼入。

### Platform Admin API（需要登录 + `platform:manage` 权限）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/platform/models` | 模型列表（含实时价格） |
| POST | `/api/platform/models` | 创建模型（目录 + 定价一起写） |
| PUT | `/api/platform/models/:id` | 更新模型属性（名称、provider、状态等） |
| DELETE | `/api/platform/models/:id` | 删除模型 |
| PUT | `/api/platform/models/:id/pricing` | 仅更新定价（不改目录信息） |
| POST | `/api/platform/catalog/publish` | 发布（版本号 +1） |

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

### system_settings 表

`catalog.models_version` key 存储当前发布版本号。`Increment` 方法使用原子 SQL：

```sql
INSERT INTO system_settings (key, value) VALUES ($1, '1')
ON CONFLICT (key) DO UPDATE SET value = (system_settings.value::int + 1)::text
RETURNING value::int
```

---

## Local 模式：catalogsync worker

### 配置

```env
CATALOG_SYNC_ENABLED=true
CATALOG_SYNC_URL=https://app.tokenjoy.com
CATALOG_SYNC_INTERVAL_SEC=300
```

SaaS 模式下这三个变量不配置（worker 不启动）。

### 同步流程

```
每 CATALOG_SYNC_INTERVAL_SEC 秒执行一次：

1. GET <CATALOG_SYNC_URL>/api/platform/sync/versions
   → 拿到远端 models version

2. 和本地 system_settings['catalog.models_version'] 比较
   → 相同则跳过，不同则继续

3. GET <CATALOG_SYNC_URL>/api/platform/sync/catalog/models
   → 拿到完整 CatalogModel[]

4. SyncFromPlatform: upsert 到本地 models 表
   → source='platform', company_id=TokenJoyCompanyID
   → 不在列表中的旧 model 被标记 inactive

5. 逐条写入本地 model_pricing 表 (ON CONFLICT skip)
   + best-effort UpsertModelRatio 到本地 NewAPI（gateway 预扣缓存）

6. 更新本地 system_settings['catalog.models_version'] = 远端 version
```

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
| 路由注册（中间件分层） | `internal/http/handler/platform/handler.go` |
| Pricing domain service | `internal/domain/pricing/service.go` |
| Catalog sync worker | `internal/worker/catalogsync/execute.go` |
| Catalog sync HTTP client | `internal/integration/catalogsync/` |
| SyncFromPlatform（models 表 upsert） | `internal/store/postgres/models_repo_crud.go` |
| ModelPricing repo | `internal/store/postgres/model_pricing_repo.go` |
| SystemSettings.Increment | `internal/store/postgres/system_settings_repo.go` |
| 配置字段 | `internal/config/config.go` → PlatformConfig |
| River periodic job | `internal/infra/jobs/kinds_catalogsync.go` |
| River worker adapter | `internal/infra/river/workers/catalog_sync.go` |
| **前端** | |
| Feature module | `apps/frontend/src/features/platform/models/` |
| API 定义 | `apps/frontend/src/api/platform.ts` |
| 路由注册 | `apps/frontend/src/router/routes/platform.ts` |
| Nav 路由配置 | `apps/frontend/src/config/routes.ts` |
