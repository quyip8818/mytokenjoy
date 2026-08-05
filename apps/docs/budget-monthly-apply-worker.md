# P0: 月初 Apply Worker — 预配 Snapshot 生效机制

## 背景

用户可在"下月"视图中预配 budget（部门额度、成员额度、项目额度）。预配数据存储在 `budget_snapshot` 表中，期间键为下月的 `YYYY-MM`。

月初到来时，需要将 snapshot 中的预配数值 **apply 到 live 表**，使其对实际扣费生效。

## 数据流

```
下月 snapshot (budget_snapshot.snapshot JSONB)
    │
    ▼  月初 cron (ApplyBudgetSnapshot worker)
    │
    ├── org_nodes.budget / reserved_pool        ← snapshot.Tree[].Budget / ReservedPool
    ├── org_members.personal_budget             ← snapshot.Members[].PersonalBudget
    └── budget_projects.budget                  ← snapshot.Projects[].Budget
```

consumed 不需要重置 — `budget_consumed` 表按 `period_key` 分区，新月自动从 0 开始。

## 执行时机

嵌入现有 watchdog → scheduler → worker 流程：

```go
// scheduler/due.go
func (s *Service) tenantDue(...) {
    currentMonth := pkgbudget.OpenSnapshotKey(PeriodMonthly, s.clk).String()
    if monthDue(tbs, currentMonth) {
        work.NeedsMonthRebalance = true
        work.NeedsBudgetApply = true  // ← 新增
    }
}
```

执行顺序：
1. `ArchivePreviousPeriod` — 存档上月 live 数据（已实现）
2. **`ApplyBudgetSnapshot`** — 如果当月有 snapshot，apply 到 live
3. `Rebalance` — 基于新 budget 重算 combined_key_remain

## 实现伪代码

```go
func (s *service) ApplyBudgetSnapshot(ctx context.Context) error {
    period := pkgbudget.SnapshotKey(PeriodMonthly, clock.NowUTC(s.clk))
    snap, err := s.store.BudgetSnapshot().Get(ctx, period)
    if err != nil { return err }
    if snap == nil { return nil } // 没有预配，budget 保持不变

    var payload SnapshotPayload
    json.Unmarshal(snap.Snapshot, &payload)

    return s.store.WithTx(ctx, func(tx store.Store) error {
        for _, node := range payload.Tree {
            tx.Org().Nodes().UpdateBudget(ctx, node.ID, node.Budget, node.ReservedPool)
        }
        for _, member := range payload.Members {
            tx.Org().UpdateMemberPersonalBudget(ctx, member.MemberID, member.PersonalBudget)
        }
        for _, project := range payload.Projects {
            tx.Budget().UpdateProjectBudget(ctx, project.ID, project.Budget)
        }
        return nil
    })
}
```

## 幂等性

- 重复 apply 同一个 snapshot 是安全的（SET 语义，非 INCREMENT）
- 可选：apply 后在 snapshot 上标记 `applied_at`，避免重复执行时的多余 DB 写入
- `TenantBackgroundState.LastRebalancedPeriod` 已经保证每月只触发一次

## 注意事项

- Apply 必须在 Rebalance 之前执行（Rebalance 需要读到最新的 budget 来计算 remain）
- Snapshot 中的 `consumed` 字段在 apply 时忽略（只 apply budget 配置）
- Apply 只处理 snapshot 中存在的实体；如果成员/项目在预配后被删除，对应 UPDATE 会 no-op（rows affected = 0），不报错

## Store 接口扩展（需新增）

```go
// store.OrgNodeRepository
UpdateBudget(ctx context.Context, nodeID uuid.UUID, budget float64, reservedPool *float64) error

// store.OrgRepository (或 MemberRepository)
UpdateMemberPersonalBudget(ctx context.Context, memberID uuid.UUID, personalBudget float64) error

// store.BudgetRepository
UpdateProjectBudget(ctx context.Context, projectID uuid.UUID, budget float64) error
```

## DueWork 扩展

```go
type DueWork struct {
    // ...existing fields...
    NeedsBudgetApply bool
}
```

## 预估工作量

- 新增 3 个 store 方法（简单 UPDATE）
- 新增 `ApplyBudgetSnapshot` domain 方法
- 修改 scheduler/due.go + 新增 worker（或嵌入 RebalanceWorker）
- 约 100-150 行代码
