# 部门 Consumed 展示方案（已实现）

## 状态：✅ 已完成

Phase 1 + Phase 2 均已实现并合入 main。

## 实现细节

### Phase 1：当前月份 read-time enrichment

`GetTree` 返回前，从 ledger 批量查当月各部门消耗，填入 tree：

```go
// domain/budget/tree_query.go
func (s *service) GetTree(ctx context.Context) ([]types.BudgetNode, error) {
    s.MaybeRotatePeriod(ctx)
    tree, err := common.LoadBudgetTree(ctx, s.store.Org().Nodes())
    if err != nil { return nil, err }
    periodKey := pkgbudget.OpenSnapshotKey(pkgbudget.PeriodMonthly, s.clk).String()
    enrichTreeConsumed(ctx, s.store.Ledger(), tree, periodKey)
    return tree, nil
}
```

SQL（`SumCostAllDepartments`，单次查询，无 N+1）：

```sql
SELECT department_id, COALESCE(SUM(cost), 0)
FROM usage_ledger
WHERE company_id = $1 AND period_key = $2 AND event_type = 'call_settled'
GROUP BY department_id
```

然后 post-order 递归，把子节点 consumed 累加到父节点（`tree_consumed.go`）。

### Phase 2：历史月份 snapshot 固化 consumed

月初 rotation 存档时，`buildSnapshotFromTx` 接受 `periodKey` 参数并做 enrichment：

```go
// domain/budget/rotation.go
func (s *service) buildSnapshotFromTx(ctx, tx, periodKey) (*SnapshotPayload, error) {
    tree := LoadBudgetTree(...)
    enrichTreeConsumed(ctx, tx.Ledger(), tree, periodKey)
    ...
}
```

查看历史月份时直接读 snapshot，无需再次聚合 ledger。

## 未改数据库

所有改动纯代码层面：
- `store/ledger_repo.go`：新增 `SumCostAllDepartments` 接口方法
- `store/postgres/ledger_repo.go`：实现（单 SQL GROUP BY）
- `domain/budget/service.go`：Store 接口加 `Ledger()`
- `domain/budget/tree_consumed.go`：新文件（enrichTreeConsumed + 递归累加）
- `domain/budget/tree_query.go`：GetTree 调用 enrichment
- `domain/budget/rotation.go`：buildSnapshotFromTx 传入 periodKey

## 性能

- 触发频率：仅用户打开预算管理页时（低频管理端操作）
- SQL 走 `(company_id, period_key)` 索引，结果集几十行，毫秒级
- Phase 2 只在月初 rotation 执行一次
