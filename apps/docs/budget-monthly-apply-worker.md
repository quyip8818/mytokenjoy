# 月度预算轮转（as-built）

## 月切时序

```text
月初时刻
  │
  ├─ Watchdog (hourly) 检测 monthDue(lastRebalancedPeriod != current)
  │     └─ 只对"还没切"的 company 入队 Rebalance(company axis)
  │
  └─ RebalanceWorker(company axis)
        ├─ budget.RotatePeriod (事务内，幂等)
        │     ├─ 1. Archive 上月 live → budget_snapshot (如未存在)
        │     ├─ 2. Apply 当月预配 snapshot → live tables (如存在)
        │     └─ 3. SetLastRebalancedPeriod = current
        └─ ProcessAxis (基于最新 budget + 新 consumed=0 重算 combined_key_remain)
```

关键：**RotatePeriod 在 ProcessAxis 之前**，确保 ProcessAxis 读到的 budget 是 apply 后的值。

`MaybeRotatePeriod`（GetTree 内 lazy 调用）保留作为 fallback——进程刚起来、watchdog 还没跑到时前端可以自愈。它额外会入队一个 Rebalance job 来刷新 remain。

---

## consumed 月度重置机制

`budget_consumed` 按 `(company_id, axis_kind, axis_id, period_key)` 为主键。

**重置不需要任何 job**——新月 `period_key` 变了（如 `2026-07` → `2026-08`），查询当月 consumed 时拿到的就是 0（行不存在 = 0）。所有人的预算到下个月自动又能用了。

```text
7 月：budget_consumed WHERE period_key = '2026-07' → consumed = 850
8 月：budget_consumed WHERE period_key = '2026-08' → 行不存在 → consumed = 0
                                                     ↑ 预算自动恢复
```

靠 `OpenSnapshotKey(PeriodMonthly, clock)` 在读路径上算出当前 period_key 来实现。

月切后第一次 Rebalance 读到新 period_key 的 consumed = 0，加上最新 budget，重算 `combined_key_remain`。Gateway 预检读到新的 remain，额度恢复生效。

---

## Watchdog 月切判断优化（待实施）

当前实现逐个遍历所有 active company 再逐一判断 `lastRebalancedPeriod != current`。

优化方向：用一条 bulk query 只取出"还没切的"：

```sql
SELECT company_id FROM tenant_background_state
WHERE last_rebalanced_period != $current
```

只对这些 company 入队 Rebalance，跳过其余。当前 company 数量少时无必要，规模增长后可作为性能优化。

---

## 无需处理的事项

| 事项 | 原因 |
|------|------|
| consumed "重置" job | 不需要。period_key 按月分片，新月查询自然 = 0 |
| combined_key_remain 重置 | 不需要。Rebalance 基于新 consumed(=0) + 新 budget 重算 |
| ledger 历史清理 | P2，数据量增长慢，后续再议 |

---

## 代码索引

```text
domain/budget/rotation.go             — RotatePeriod, MaybeRotatePeriod, rotatePeriod, applySnapshot
domain/budget/service.go              — Service interface
domain/budget/snapshot.go             — CopyPeriod, GetTreeForPeriod, buildCurrentSnapshot
infra/river/workers/rebalance.go      — RebalanceWorker (RotatePeriod → ProcessAxis)
infra/scheduler/due.go                — monthDue, DueWork
store/postgres/budget_consumed_repo.go — IncrementConsumed, ListConsumed
store/postgres/budget_snapshot_repo.go — Upsert, Get, Delete
```
