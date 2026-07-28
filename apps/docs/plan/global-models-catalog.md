# 全局模型 Catalog

> `models` 表通过 `globalCompanyID`（= `tokenJoyCompanyID`）实现全局 catalog。
> 查询层 UNION 全局+租户；租户通过 ToggleModel 创建覆盖行。
> SMS 同步全局写入一次（`company_id = globalCompanyID`），不再 per-company 循环。

---

## 1. 数据模型

```sql
CREATE TABLE IF NOT EXISTS models (
    model_id     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id   UUID NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
    provider     TEXT NOT NULL,
    type         TEXT NOT NULL,
    name         TEXT NOT NULL,
    description  TEXT NOT NULL DEFAULT '',
    endpoint     TEXT,
    api_key      TEXT,
    endpoint_model_name TEXT,
    max_context  INT NOT NULL DEFAULT 0,
    max_tokens   INT NOT NULL DEFAULT 0,
    active       BOOLEAN NOT NULL DEFAULT TRUE,
    capabilities TEXT[] NOT NULL DEFAULT '{}',
    source       TEXT NOT NULL DEFAULT 'manual',  -- 'sms' | 'manual' | 'seed'
    sms_synced_at TIMESTAMPTZ,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (company_id, provider, type)
);
```

### 关键字段

| 字段 | 说明 |
|------|------|
| `active` | `true` = UI 可见可选；`false` (inactive) = UI 不展示，但已配置的调用链继续工作 |
| `source` | `'sms'` = SMS 同步写入；`'manual'` = 租户手动创建/覆盖；`'seed'` = 初始化数据 |
| `sms_synced_at` | 最近一次 SMS 同步触及此行的时间戳，用于 staleness 判断 |

---

## 2. 查询层

```go
// models_repo_crud.go — Models()
SELECT ... FROM models
WHERE company_id = $globalCompanyID OR company_id = $tenantCompanyID
ORDER BY CASE WHEN company_id = $globalCompanyID THEN 0 ELSE 1 END, model_id

// DedupeEffective(): 以 provider+type 为 key，后出现的覆盖前面的
// → 租户行覆盖全局行
```

Repo 层 `Models()` 返回全量（含 inactive），不做 active 过滤。

---

## 3. 展示层 vs 执行层

| 层 | 位置 | 行为 | 用途 |
|----|------|------|------|
| 展示层 | service `ListModels` / `ListModelsWithPricing` | `FilterVisible` + `FilterActive` | 模型选择器、可选列表 API |
| 执行层 | gateway precheck / `ModelByType` / `ModelByProviderType` | 不过滤 active，只要行存在即返回 | 调用放行、已有配置 enrich |

展示层过滤放在 **service 层**（不在 repo SQL），因为 `Models()` 还被 `ListRoutingRules`、`ResolveRouting`、`ValidateWritableIDs` 消费——这些需要看到 inactive 模型以 enrich 已有配置。

---

## 4. inactive 的语义

| 场景 | 行为 |
|------|------|
| 模型选择器（前端可选列表） | 不展示 inactive 模型 |
| 已配置的 allowlist / org_nodes 路由规则 | 继续生效，不中断已有配置 |
| gateway precheck（调用时校验） | 只看 allowlist，不看模型 active 状态 |
| ToggleModel（tenant 覆盖） | 全局 inactive 的模型不能被 tenant 重新 activate（入口封死） |
| routing rule enrich / 编辑 | inactive 模型仍可展示名字、仍可保留在配置中 |

---

## 5. SMS 同步

### SyncTarget 接口

```go
type SyncTarget interface {
    ReplaceChannels(ctx context.Context, channels []sms.CatalogChannel) error
    SyncModels(ctx context.Context, models []sms.CatalogModel) error
    ReplaceModelRatios(ctx context.Context, models []sms.CatalogModel) error
}
```

### Target 实现

```go
type Target struct {
    port            adminport.Port
    store           store.Store
    globalCompanyID uuid.UUID
}

func (t *Target) SyncModels(ctx context.Context, models []sms.CatalogModel) error {
    infos := make([]types.ModelInfo, 0, len(models))
    for _, m := range models {
        infos = append(infos, types.ModelInfo{
            CompanyID: t.globalCompanyID,
            Provider:  m.Provider,
            Type:      m.ModelID,
            Name:      m.DisplayName,
            Source:    "sms",
            Active:    true,
        })
    }
    return t.store.Models().SyncFromSMS(ctx, t.globalCompanyID, infos)
}
```

### SyncFromSMS（diff-deactivate + upsert）

```go
// Step 1: UPSERT — 标记 active=true，刷新 sms_synced_at=NOW()
// Step 2: DEACTIVATE — active=true AND source='sms' AND sms_synced_at < NOW()-10s
//         → 本次未被 touch 的行 = 已下架，标记 inactive
```

用 `sms_synced_at` 时间戳做 staleness 判断，避免 type 集合匹配的边界问题。

### Execute 流程

```go
func (e *SMSSyncExecutor) syncModels(ctx context.Context) (int, error) {
    resp, err := e.client.FetchModels(ctx)
    if err != nil { return 0, err }
    if err := e.target.SyncModels(ctx, resp.Data); err != nil {
        return 0, fmt.Errorf("sync models: %w", err)
    }
    if err := e.target.ReplaceModelRatios(ctx, resp.Data); err != nil {
        return 0, fmt.Errorf("replace model ratios: %w", err)
    }
    return resp.Version, nil
}
```

不再循环 company，不再调 `listActiveCompanyIDs`。

---

## 6. 租户覆盖（ToggleModel）

```go
func (s *service) ToggleModel(ctx context.Context, id uuid.UUID, enabled bool) error {
    model := store.Models().ModelByID(ctx, id)

    if model.CompanyID == TokenJoyCompanyID {
        // 全局模型路径
        if enabled && !model.Active {
            return Validation("model has been delisted and cannot be enabled")
        }
        // 创建/更新 tenant 覆盖行
        ...
    }

    // 租户覆盖行路径
    if enabled {
        globalModel := store.Models().GlobalModelByProviderType(ctx, model.Provider, model.Type)
        if globalModel != nil && !globalModel.Active {
            return Validation("model has been delisted and cannot be enabled")
        }
    }
    model.Active = enabled
    store.Models().UpdateModel(ctx, model)
}
```

两条路径都校验全局行 active 状态，防止 tenant 重新激活已下架模型。

---

## 7. Gateway Precheck

```sql
-- allowlist_types 不过滤 active 状态，只看 allowlist 里有什么模型
COALESCE(
    array_agg(DISTINCT mdl.type ORDER BY mdl.type) FILTER (WHERE mdl.type IS NOT NULL),
    '{}'
) AS allowlist_types
```

Gateway 只看 allowlist 成员关系，不看模型 active 状态。已配置的 inactive 模型继续可调用。

---

## 8. model_allowlist

```sql
FOREIGN KEY (model_id) REFERENCES models (model_id) ON DELETE RESTRICT
```

- FK 使用 `ON DELETE RESTRICT`（模型永不物理删除，作为安全网）
- 已配置 allowlist 条目指向 inactive 模型时：展示层不在选择器展示，执行层正常放行
- `IsCallTypeAllowed` 不过滤 active 状态，与 gateway precheck 对齐

---

## 9. 对现有系统的影响

| 模块 | 影响 |
|------|------|
| 计费 (billing) | 无——按 model type 字符串匹配 |
| 预算 (budget) | 无——按 model type 聚合 |
| NewAPI 路由 | 无——model_ratio 全局写入 NewAPI，与 models 表无关 |
| 前端模型选择器 | API 仅返回 active=true 的模型，inactive 模型自动消失 |
| model_allowlist | FK 不断裂，已配置条目保留；gateway 只查 allowlist 放行 |
| org_nodes 路由规则 | default_model_id / fallback_model_id FK 不断裂，继续生效 |
| 网关 precheck | 只查 allowlist，不看 active 状态 |
| ToggleModel | 全局 inactive 模型拒绝创建 active 覆盖行 |
| 新 company 注册 | 注册即可见全局模型（查询 UNION globalCompanyID） |

---

## 10. 不做什么

- 不新建 `model_catalog` 表——现有 models + globalCompanyID 已满足
- 不新建 `company_model_allowlist`——ToggleModel + model_allowlist 已覆盖所有场景
- 不做 per-company 定价差异——定价全局由 SMS 管控
- 不做模型版本管理
- 不改 channel 同步（已是全局的 diff-delete）
- 不物理删除模型行——所有下架操作均为 `active=false`
- 不级联清理 allowlist / org_nodes 中引用 inactive 模型的条目——已配置继续工作
