# P0：将 Billing Rate 纳入 AuthzSvc LRU 缓存

## 问题

`RequireSession` middleware 是每个认证请求的热路径。当前 `AuthzSvc.GetSessionContext` 中，`ResolveCompanyChargeRate` **在 LRU 缓存之外无条件执行**，每次请求产生 2 次 DB 查询：

```
Company.GetByID(companyID)     → 1 DB
Billing.GetCurrency(currency)  → 1 DB
```

即使 authz LRU 完全命中（member 权限未变），这 2 次查询仍然发生。

### 影响量化

- 假设系统 1000 req/s，每个请求 2 次不必要 DB 查询 = **2000 额外 QPS 到 Postgres**
- billing currency 和 quota_per_unit 是公司级管理配置，变更频率 < 1 次/天
- 当前 authz LRU 缓存命中率在稳态下 > 95%（4096 条目，revision TTL 5s），但 billing 完全不享受缓存红利

---

## 方案

### 核心思路

billing rate（`currency` + `quotaPerUnit`）和 member authz 一样，以 `(companyID, memberID, revision)` 为缓存键。当 revision 不变时，billing 配置不可能变。

将 billing 字段加入 LRU cacheValue，查询逻辑移到 cache miss 分支。

### 代码变更

#### 1. `internal/identity/authz/cache.go`

```go
// 变更前
type cacheValue struct {
    member      types.Member
    permissions []string
    readOnly    bool
}

// 变更后
type cacheValue struct {
    member          types.Member
    permissions     []string
    readOnly        bool
    billingCurrency string
    quotaPerUnit   int64
}
```

对应修改 `Get` 和 `Put` 方法签名：

```go
// 变更后
func (c *LRUCache) Get(companyID uuid.UUID, memberID uuid.UUID, revision int64) (types.Member, []string, bool, string, int64, bool) {
    key := cacheKey{companyID: companyID, memberID: memberID, revision: revision}
    c.mu.Lock()
    defer c.mu.Unlock()
    elem, ok := c.items[key]
    if !ok {
        return types.Member{}, nil, false, "", 0, false
    }
    c.ll.MoveToFront(elem)
    entry := elem.Value.(*lruEntry)
    v := entry.value
    return v.member, append([]string(nil), v.permissions...), v.readOnly, v.billingCurrency, v.quotaPerUnit, true
}

func (c *LRUCache) Put(companyID uuid.UUID, memberID uuid.UUID, revision int64, member types.Member, permissions []string, readOnly bool, billingCurrency string, quotaPerUnit int64) {
    key := cacheKey{companyID: companyID, memberID: memberID, revision: revision}
    c.mu.Lock()
    defer c.mu.Unlock()
    if elem, ok := c.items[key]; ok {
        c.ll.MoveToFront(elem)
        elem.Value.(*lruEntry).value = cacheValue{
            member:          member,
            permissions:     append([]string(nil), permissions...),
            readOnly:        readOnly,
            billingCurrency: billingCurrency,
            quotaPerUnit:   quotaPerUnit,
        }
        return
    }
    if c.ll.Len() >= c.maxSize {
        oldest := c.ll.Back()
        if oldest != nil {
            c.ll.Remove(oldest)
            delete(c.items, oldest.Value.(*lruEntry).key)
        }
    }
    entry := &lruEntry{
        key: key,
        value: cacheValue{
            member:          member,
            permissions:     append([]string(nil), permissions...),
            readOnly:        readOnly,
            billingCurrency: billingCurrency,
            quotaPerUnit:   quotaPerUnit,
        },
    }
    c.items[key] = c.ll.PushFront(entry)
}
```

#### 2. `internal/identity/authz/service.go`

```go
func (s *service) GetSessionContext(ctx context.Context, companyID uuid.UUID, memberID uuid.UUID) (types.SessionContext, error) {
    revision, err := s.revCache.get(ctx, companyID, s.store.Company())
    if err != nil {
        return types.SessionContext{}, err
    }

    companyType := companyTypeFromContext(ctx, companyID, s.store)

    // ← 整体命中：包含 billing，0 DB 查询
    if member, perms, readOnly, currency, ppu, ok := s.cache.Get(companyID, memberID, revision); ok {
        return types.SessionContext{
            CompanyID:       companyID,
            CompanyType:     companyType,
            AuthzRevision:   revision,
            Member:          member,
            Permissions:     perms,
            ReadOnly:        readOnly,
            BillingCurrency: currency,
            QuotaPerUnit:   ppu,
        }, nil
    }

    // ← Cache miss：查 authz + billing，然后一起写入缓存
    authz, err := s.store.Org().GetMemberAuthz(ctx, companyID, memberID)
    if err != nil {
        return types.SessionContext{}, err
    }
    if authz == nil || authz.Member.Status != types.MemberStatusActive {
        return types.SessionContext{}, domain.NewDomainError(404, "Member not found")
    }

    currency, ppu, err := billing.ResolveCompanyChargeRate(ctx, s.store, companyID)
    if err != nil {
        return types.SessionContext{}, err
    }

    permissions := ResolveMemberPermissions(authz.Member, authz.Roles)
    readOnly := IsReadOnlySession(permissions)

    s.cache.Put(companyID, memberID, revision, authz.Member, permissions, readOnly, currency, ppu)

    return types.SessionContext{
        CompanyID:       companyID,
        CompanyType:     companyType,
        AuthzRevision:   revision,
        Member:          authz.Member,
        Permissions:     permissions,
        ReadOnly:        readOnly,
        BillingCurrency: currency,
        QuotaPerUnit:   ppu,
    }, nil
}
```

#### 3. 测试调整

`tests/identity/authz/session_cache_test.go` 中的 `TestGetSessionContextCachesByRevision` 无需修改逻辑——它验证的是"第二次调用不查 OrgRepo"，这个断言依然成立。

可以额外加一个断言：billing 结果一致。

---

## 正确性分析

### billing 数据何时变？

| 触发场景 | 影响字段 | 是否 bump authz_revision |
|----------|----------|------------------------|
| 管理员改公司 billing currency | `companies.billing_currency` | **需要确认** |
| 管理员改币种 quota_per_unit | `currencies.quota_per_unit` | **需要确认** |

**关键前提**：`authz_revision` 只在权限相关变更时 bump（角色变更、成员增删）。如果 billing 配置变更不 bump revision，缓存会返回过期的 billing 数据。

### 两种处理方式

**方式 A（推荐）**：billing 配置变更时也 bump `authz_revision`

- 优点：zero-cost，一行 SQL（`UPDATE companies SET authz_revision = authz_revision + 1 WHERE id = $1`）
- 缺点：会触发所有该公司 member 的 LRU miss，产生一波 DB 查询。但 billing 变更 < 1 次/天，完全可接受
- 副作用：前端会收到 revision 变化 → refetch session。无害（会拿到新 billing rate）

**方式 B**：给 billing 单独加一层短 TTL 缓存

- 优点：不耦合 authz_revision
- 缺点：增加复杂度、新的缓存层、TTL 期间数据不一致

**推荐方式 A**——在 billing 配置变更的 repo 方法中加一行 bump revision。

---

## 需要确认的前置条件

1. **billing 配置变更代码在哪？** 需要找到修改 `companies.billing_currency` 和 `currencies.quota_per_unit` 的路径，确认是否已 bump revision 或需要加。
2. **`companyType` 变更是否 bump revision？** 当前 `companyTypeFromContext` 从 context 取值（CompanyResolve middleware 注入），不走缓存，无问题。但如果未来 companyType 也缓存化，需要同样挂到 revision。

---

## 改进后性能对比

| 场景 | 变更前 DB 查询 | 变更后 DB 查询 |
|------|---------------|---------------|
| 稳态（连续请求，revision 缓存内） | 2 | **0** |
| revision TTL 边界（5s 一次） | 3 | **1** |
| revision 变更（cache miss） | 4 | **3** |
| 冷启动首次请求 | 4 | **3** |

**稳态 QPS 节省**：假设 1000 req/s → 每秒减少 2000 次 Postgres 查询。

---

## 风险与回退

- **风险**：如果 billing 变更场景遗漏了 bump revision，缓存会短暂返回旧 billing rate。影响范围：该公司 member 在 LRU TTL（revision 变前一直有效）内看到旧费率。最差情况是计费精度在一个 LRU 生命周期内偏差。
- **回退**：如果发现问题，把 `ResolveCompanyChargeRate` 调用移回 cache hit 分支（恢复原逻辑），不影响任何数据。
- **监控**：上线后观察 Postgres 连接池 active queries 是否下降、p99 延迟变化。

---

## 实施步骤

1. 确认 billing 变更路径是否已 bump `authz_revision`，不足则补上
2. 修改 `cache.go`：扩展 `cacheValue` + 调整 `Get`/`Put` 签名
3. 修改 `service.go`：将 `ResolveCompanyChargeRate` 移入 cache miss 分支
4. 跑现有测试 `TestGetSessionContextCachesByRevision` 确认不回归
5. 可选：加一个测试验证 billing 数据缓存命中

---

## 文件清单

| 文件 | 变更类型 |
|------|---------|
| `internal/identity/authz/cache.go` | 扩展 cacheValue 结构 + Get/Put 签名 |
| `internal/identity/authz/service.go` | 调整 GetSessionContext 逻辑 |
| billing 变更 repo（待确认） | 加 bump authz_revision |
| `tests/identity/authz/session_cache_test.go` | 可选：加 billing 缓存断言 |
