# Rebalance 简化方案

> **状态：已实施**（2026-07-16）

## 背景

当前 `rebalanceKey()` 在计算每个 key 的 `newRemain` 时，除了基于月度限额计算 `allocated`，还会额外用 `walletAvailable()` 做一次 `min` 约束：

```go
newRemain := allocated
if walletID > 0 {
    walletUnits := walletAvailable(...)
    if walletUnits < newRemain {
        newRemain = walletUnits
    }
}
```

这导致充值后必须对全公司所有 key 触发 Rebalance，因为 `walletUnits` 变了。

但实际上 **Gateway precheck 已经独立检查 wallet remain**（从 PG 实时读取），所以 per-key 级别再做 wallet min 是多余的。`GatewayChainRemain` 注释也明确写道：

> wallet_remain is checked independently in the precheck path (real-time from PG).
> This chain only evaluates budget-control constraints: key, member, project.

## 问题

1. **充值触发不必要的 Rebalance** — `afterRecharge()` 同时 enqueue `WalletSync` + `RebalanceCompany`。充值不改变任何 key 的月度限额或消耗，Rebalance 是空转。
2. **wallet min 压缩 key 额度** — 当钱包暂时紧张时，key 的 `RemainQuota` 被压低，但 Gateway 已经有钱包硬约束，双重限制没有意义。
3. **`walletAvailable()` 有 O(N) 遍历** — 每个 key rebalance 时重新查所有公司 mapping，计算 `used`，对于 key 多的公司是不必要的性能开销。
4. **`NewAPIKeyRemainQuota` 字段语义混淆** — 它存的是 `min(月度限额剩余, 钱包可用)` 而非纯粹的月度预算剩余，增加理解成本。

## 方案

### 变更 1: 移除 `rebalanceKey()` 中的 wallet min 约束

**删除代码：**
- `RebalanceService.walletAvailable()` 整个方法
- `RebalanceService.newAPIWalletUserID()` 方法（仅被 walletAvailable 间接使用）
- `rebalanceKey()` 中 `walletID > 0` 整个 if 块

**变更后 `rebalanceKey()` 核心逻辑：**

```go
func (s *RebalanceService) rebalanceKey(ctx context.Context, mapping store.PlatformKeyMapping) error {
    // ... 加载 budgetCtx, key, token, models, departments, rules ...

    remainPoint, err := pkgbudget.ComputeRemainForMapping(...)
    if err != nil {
        return err
    }
    allocated := newapiunits.ToNewAPIUnits(remainPoint, models, effectiveIDs)

    // 直接使用 allocated，不再与 wallet 做 min
    if allocated == token.RemainQuota {
        return nil
    }

    req := adminport.UpdateTokenInput{
        ID:          token.ID,
        RemainQuota: &allocated,
    }
    updated, err := s.client.UpdateToken(ctx, req)
    if err != nil {
        return err
    }
    if err := s.store.PlatformKeyMappings().UpdateMappingNewAPIKeyRemainQuota(
        ctx, mapping.PlatformKeyID, updated.RemainQuota,
    ); err != nil {
        return err
    }
    return RefreshPlatformKeyCombined(ctx, s.store, mapping.PlatformKeyID, s.cfg.Clock(), nil)
}
```

### 变更 2: 充值不再触发 Rebalance

**修改 `afterRecharge()`：**

```go
func (s *service) afterRecharge(ctx context.Context, companyID int64) error {
    // 只做 WalletSync，不再 Rebalance
    return s.enqueuer.InsertWalletSync(ctx, companyID)
}
```

**删除：**
- `afterRecharge()` 中 `GetByID`、`ConfiguredNewAPIWalletUserID`、`WithContext`、`InsertRebalanceCompany` 相关代码

### 变更 3: 清理 `RebalanceStore` 接口

移除 `Company()` 和 `Models()` 中仅被 `walletAvailable` 使用的部分依赖（如果 `rebalanceKey` 仍需要 `Models()` 做 `ToNewAPIUnits` 换算，则保留）。

实际上 `rebalanceKey` 本身就需要 `Models()` 和 `Company()`（用于 `ComputeRemainForMapping`），所以接口不需要变动。

### 变更 4: 简化 `billing.JobEnqueuer` 接口（可选）

如果 `InsertRebalanceCompany` 只在 `afterRecharge` 使用，可以从 `billing.JobEnqueuer` 接口中移除：

```go
type JobEnqueuer interface {
    InsertWalletSync(ctx context.Context, companyID int64) error
    // InsertRebalanceCompany 已删除
}
```

## 涉及文件

| 文件 | 变更 |
|------|------|
| `internal/domain/budget/rebalance.go` | 删除 `walletAvailable()`、`newAPIWalletUserID()`，简化 `rebalanceKey()` |
| `internal/domain/billing/lot_confirm.go` | 简化 `afterRecharge()`，移除 Rebalance 触发 |
| `internal/domain/billing/ports.go` | 从 `JobEnqueuer` 移除 `InsertRebalanceCompany` |
| `internal/adapter/billing.go` | 删除 `InsertRebalanceCompany` 实现 |
| `tests/domain/budget/rebalance_test.go` | 更新测试，不再验证 wallet cap |
| `tests/worker/processors_test.go` | 适配接口变化 |

## Rebalance 保留的触发场景

移除充值触发后，Rebalance 仍在以下场景工作（这些都是合理的）：

| 触发场景 | 原因 |
|---------|------|
| 月度周期开始 (`EnsureMonthRebalance`) | 新周期 consumed 归零，remain 需要重算 |
| Budget reconcile | consumed 修正后 remain 变化 |
| Approval 通过 | 成员预算配置变化 |
| Project 删除 | 关联 key 的预算归属变化 |
| NewAPI sync 完成 | key mapping 建立后首次设置 remain |

## 收益

1. **充值链路减少一个异步 job** — 少了 Rebalance 对全公司 key 的遍历 + N 次外部 API 调用
2. **语义清晰** — `NewAPIKeyRemainQuota` 纯粹反映月度预算剩余，不混入 wallet 约束
3. **减少无效 Redis 更新** — 充值后不再刷新所有 key 的 combined_key_remain（值没变）
4. **删除 ~60 行复杂逻辑** — `walletAvailable` 的 O(N) 遍历和跨 key 分配计算

## 风险评估

**低风险**。Gateway 已有独立的 wallet remain 硬约束，移除 per-key wallet min 不会导致超额消费。唯一的行为差异是：当钱包余额不足时，以前某些 key 的 `RemainQuota`（NewAPI 侧）会被压低，现在不会了——但 Gateway precheck 仍然会拦截。


## 实施结果（2026-07-16）

方案已完整落地，额外清理超出原计划范围：

| 变更 | 说明 |
|------|------|
| `walletAvailable()` / `newAPIWalletUserID()` | 已删除 |
| `afterRecharge()` → Rebalance | 已移除，只保留 `InsertWalletSync` |
| `billing.JobEnqueuer.InsertRebalanceCompany` | 接口和实现已删除 |
| `NewAPIKeyRemainQuota` 字段 | 从 struct、接口、PG repo、schema.sql **彻底删除** |
| `UpdateMappingNewAPIKeyRemainQuota` 方法 | 已删除 |
| `newapisync` `capRemainUnits()` | 已删除，内联为 `newapiunits.ToNewAPIUnits` |
| `syncdeps.Deps.Wallet` 字段 | 已删除（newapisync 不再依赖 wallet service） |
| `newapisync.New()` 签名 | 去掉 `wallet` 参数 |
| `platform_key_mappings.newapi_key_remain_quota` 列 | schema.sql 已移除 |

净删约 250 行，新增约 45 行。`go build ./...` 和 `go vet` 通过。
