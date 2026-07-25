# ADR: Platform Key 软删除

## 决策

删除 = `UPDATE status = 'deleted'`。废弃 `revoked` 状态和 `RevokePlatformKey` 方法。

## 为什么

硬删除会丢飞行中请求的入账。软删除零成本解决：行还在，ingest 正常写 consumed。

## 状态流转

```
active ←→ disabled
  ↓            ↓
  └──→ deleted ←┘   (终态，不可恢复)
```

## 关键约束

`SetPlatformKeys` 末尾有 prune（`DELETE WHERE id NOT IN ...`）。如果 `PlatformKeys()` 排除了 deleted 行，其他写路径会把 deleted 行物理删掉。

所以：
- **`PlatformKeys()` 继续返回全量（含 deleted）**
- **`DeletePlatformKey` 用独立单行 UPDATE，不走 `SetPlatformKeys`**

## 实现

### store

```go
// 新增，同构于 DisablePlatformKey
func (r *pgKeysRepo) SoftDeletePlatformKey(ctx context.Context, keyID uuid.UUID) error {
    companyID := store.CompanyID(ctx)
    _, err := r.db.Exec(ctx, `
        UPDATE platform_keys SET status = 'deleted', updated_at = NOW()
        WHERE company_id = $1 AND id = $2 AND status != 'deleted'
    `, companyID, keyID)
    return err
}
```

### domain

```go
func (s *service) DeletePlatformKey(ctx context.Context, id uuid.UUID) error {
    if err := s.requireNewAPI(); err != nil { return err }
    key, err := s.store.Keys().PlatformKeyByID(ctx, id)
    if err != nil { return err }
    if key == nil { return domain.NotFound("Not found") }
    if key.Status == "deleted" { return nil } // 幂等
    if err := s.newAPISync.SyncRevokePlatformKey(ctx, id); err != nil { return err }
    if err := s.store.Keys().SoftDeletePlatformKey(ctx, id); err != nil { return err }
    s.cacheInvalidator.InvalidateByKeyID(id)
    return nil
}
```

### 移除

- `RevokePlatformKey` 方法 + `PUT /platform/{id}/revoke` 路由
- `newAPIRevokeKey` helper（无调用方）
- 前端 `platformKeyApi.revoke()`，管理员页面 `onRevoke` → `onDelete`

### 前端

过滤展示：`keys.filter(k => k.status !== 'deleted')`

## 为什么账目准确

| 关注点 | 保证 |
|--------|------|
| 额度释放 | budget 计算只累加 `status == "active"` 的 key，deleted 自动不占池 |
| 飞行中入账 | `PlatformKeyByID` 不过滤 status，行还在，正常入账 |
| 网关拦截 | `KeyStatus != "active"` 即拒绝，deleted 立即生效 |
| prune 安全 | `PlatformKeys()` 返回 deleted 行，其他写路径 prune 不会误删 |

## 故障收敛

唯一风险：远端 revoke 成功但本地 UPDATE 失败 → 用户重试即可（远端 revoke 幂等）。不做补偿。
