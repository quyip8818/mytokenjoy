# 全局模型 Catalog（消除 per-company 冗余）

> 当前 models 表已通过 `globalCompanyID` 实现全局 catalog 语义（SMS 模型写入全局 company，各租户查询时 UNION 全局+自身）。
> 本文规划将这一隐式约定显式化，简化 SMS 同步路径。
> 依赖 `sms-model-sync-v2.md` 基础设施完成后再实施。

---

## 1. 现状

### 当前实现

```
models 表 (company_id, provider, type) UNIQUE
  ├─ company_id = globalCompanyID → SMS 同步写入的全局模型
  └─ company_id = 租户 ID       → 租户手动添加的自定义模型 (source='manual')

查询层：WHERE company_id = $globalCompanyID OR company_id = $tenantID
  → DedupeEffective() 租户模型优先覆盖同 key 的全局模型
```

### 问题

| 问题 | 说明 |
|------|------|
| 隐式约定 | "globalCompanyID" 概念分散在 models_catalog.go、models_repo_crud.go 中，无文档说明 |
| SMS 同步仍循环 company | `sms-model-sync-v2.md` §4.2 设计了 per-company DELETE+INSERT，但全局写入只需一次 |
| source 语义模糊 | `source='sms'` 和 `company_id=globalCompanyID` 重复表达了"来自 SMS"这一事实 |

---

## 2. 目标

将 SMS 同步从 per-company 循环写入改为**全局写入一次**，与现有查询层（已支持 globalCompanyID UNION）对齐。

```
SMS 同步 Execute():
  1. DELETE FROM models WHERE company_id = $globalCompanyID AND source = 'sms'
  2. INSERT INTO models (...) VALUES (...) -- company_id = globalCompanyID
  → 完成。不循环 company。
```

### 不新建表

现有 `models` 表 + `globalCompanyID` 约定已经是全局 catalog。不需要新建 `model_catalog` 表。

### 不新建 company_model_allowlist

- 当前所有 company 看到相同的模型列表——这是设计意图，SMS 统一管控
- 细粒度控制由现有 `model_allowlist` 表（owner_type: org_node / platform_key）覆盖
- 未来如需 company 级差异化，复用 `model_allowlist` 加 `owner_type='company'` 即可

---

## 3. 变更清单

### 3.1 SMS 同步 ReplaceModels 简化

```go
// 不再接收 companyID 参数，直接写全局
func (t *target) ReplaceModels(ctx context.Context, models []sms.CatalogModel) error {
    _, err := t.db.Exec(ctx, `
        DELETE FROM models WHERE company_id = $1 AND source = 'sms'
    `, t.globalCompanyID)
    if err != nil {
        return err
    }
    for _, m := range models {
        _, err = t.db.Exec(ctx, `
            INSERT INTO models (company_id, provider, type, name, source, sms_synced_at, updated_at)
            VALUES ($1, $2, $3, $4, 'sms', NOW(), NOW())
        `, t.globalCompanyID, m.Provider, m.ModelID, m.DisplayName)
        if err != nil {
            return err
        }
    }
    return nil
}
```

### 3.2 Execute() 去掉 company 循环

```go
// Before:
companies, _ := store.ListActiveCompanyIDs(ctx)
for _, companyID := range companies {
    _ = target.ReplaceModels(ctx, companyID, catalog.Models)
}

// After:
_ = target.ReplaceModels(ctx, catalog.Models)  // 全局一次
_ = target.ReplaceModelRatios(ctx, catalog.Models)  // NewAPI 定价，全局一次
```

### 3.3 SyncTarget 接口调整

```go
type SyncTarget interface {
    ReplaceChannels(ctx context.Context, channels []sms.CatalogChannel) error
    ReplaceModels(ctx context.Context, models []sms.CatalogModel) error      // 去掉 companyID 参数
    ReplaceModelRatios(ctx context.Context, models []sms.CatalogModel) error
}
```

### 3.4 查询层——无变化

现有 `pgModelsRepo.Models()` 已经 `WHERE company_id = $globalCompanyID OR company_id = $tenantID`，无需改动。

### 3.5 model_allowlist 表——无变化

FK 指向 `models(model_id)`，模型从 per-company 迁移到 globalCompanyID 后 model_id 不变（重建即可），allowlist 数据不受影响。

---

## 4. 迁移步骤

| # | 事项 |
|---|------|
| 1 | 确认 `sms-model-sync-v2.md` 基础设施就绪（River Job + system_settings） |
| 2 | 修改 `ReplaceModels` 实现：去掉 companyID 参数，写入 globalCompanyID |
| 3 | Execute() 删除 company 循环 |
| 4 | 清理旧数据：`DELETE FROM models WHERE source = 'sms' AND company_id != $globalCompanyID`（项目未上线，可直接清） |
| 5 | 删除 `store.ListActiveCompanyIDs` 调用（如无其他消费方） |
| 6 | 验证：租户查询仍能看到全局模型 + 自己的 manual 模型 |

---

## 5. 对现有系统的影响

| 模块 | 影响 |
|------|------|
| 计费 (billing) | 无——按 model type 字符串匹配 |
| 预算 (budget) | 无——按 model type 聚合 |
| API 路由 (NewAPI) | 无——model_ratio 全局写入 NewAPI，和 models 表无关 |
| 前端模型选择器 | 无——API 返回结果不变 |
| org_node / platform_key allowlist | 无——FK 指向 model_id 不变 |
| 网关 precheck | 无——走 ModelsRepo 查询，已支持 globalCompanyID |

---

## 6. 前置条件

- `sms-model-sync-v2.md` 实施完成
- 项目未上线，可直接重建 schema

---

## 7. 不做什么

- 不新建 `model_catalog` 表——现有 models 表 + globalCompanyID 已满足
- 不新建 `company_model_allowlist`——现有 `model_allowlist`（org_node / platform_key）已覆盖细粒度控制，未来如需 company 级可加 owner_type
- 不做 per-company 定价差异——定价全局由 SMS 管控
- 不做模型版本管理——YAGNI
- 不改 NewAPI model_ratio 机制
