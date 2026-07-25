# 计划：统一 Quota / 货币字段命名

## 目标

消除 `DisplayAmount`、`DisplayCost` 等历史命名歧义，建立清晰的命名规范：

- **Quota 维度**的字段带 `Quota` 前缀/后缀
- **货币维度**的字段直接用 `Cost`（花费）或 `Budget`（额度）

## 命名规范

| 维度                         | 命名规则                                           | 示例                                                |
| ---------------------------- | -------------------------------------------------- | --------------------------------------------------- |
| Quota（整数，钱包层）        | 带 `Quota` 前缀/后缀                               | `QuotaAmount`, `WalletRemainQuota`, `QuotaConsumed` |
| 货币（float64，预算/花费层） | `Cost`（花费）/ `Budget`（额度）/ `Remain`（剩余） | `Cost`, `Budget`, `Consumed`                        |
| 币种                         | `BillingCurrency`                                  | `BillingCurrency`                                   |

## 重命名清单

### 1. Ledger 层：`Amount` → `QuotaAmount`

UsageLedgerEntry.Amount（int64，quota 维度）改为 QuotaAmount，消除与"货币金额"的歧义。

| 位置                                     | 变更                                                     |
| ---------------------------------------- | -------------------------------------------------------- |
| `domain/types/usage_ledger.go`           | struct field `Amount int64` → `QuotaAmount int64`        |
| DB column `usage_ledger.amount`          | → `quota_amount`                                         |
| `store/postgres/ledger_repo.go`          | SELECT/scan 列名                                         |
| `store/postgres/ledger_repo_write.go`    | INSERT 列名                                              |
| `domain/usage/entry_build.go`            | `Amount: input.Raw.Quota` → `QuotaAmount:`               |
| `domain/billing/lot/ledger.go`           | `entry.Amount = seg.Quota` → `entry.QuotaAmount =`       |
| `domain/dashboard/bucket_from_ledger.go` | `QuotaConsumed: entry.Amount` → `entry.QuotaAmount`      |
| `domain/usage/ingest.go`                 | `entry.Amount` → `entry.QuotaAmount`（lot consume 传参） |
| `domain/budget/consumed_attrib.go`       | 注释更新                                                 |
| `seed/snapshot/audit.go`                 | `Amount: seedQuota(...)` → `QuotaAmount:`                |

### 2. Ledger 层：`DisplayAmount` → `Cost`

UsageLedgerEntry.DisplayAmount（float64，每 segment 的货币花费）改为 Cost。

| 位置                                                      | 变更                                                                |
| --------------------------------------------------------- | ------------------------------------------------------------------- |
| `domain/types/usage_ledger.go`                            | struct field `DisplayAmount float64` → `Cost float64`               |
| DB column `usage_ledger.display_amount`                   | → `cost`                                                            |
| `store/postgres/ledger_repo.go`                           | SELECT/scan 列名                                                    |
| `store/postgres/ledger_repo_write.go`                     | INSERT 列名                                                         |
| `store/postgres/schema.sql`                               | 列定义                                                              |
| `domain/billing/lot/ledger.go`                            | `entry.DisplayAmount = seg.DisplayAmount` → `entry.Cost = seg.Cost` |
| `domain/usage/ingest.go`                                  | `spend += seg.DisplayAmount` → `seg.Cost`                           |
| `domain/usage/ledger_audit.go`                            | `entry.DisplayAmount` → `entry.Cost`                                |
| `domain/dashboard/bucket_from_ledger.go`                  | `entry.DisplayAmount` → `entry.Cost`                                |
| `domain/budget/budget_reconcile.go`                       | `entry.DisplayAmount` → `entry.Cost`                                |
| `domain/budget/consumed_attrib.go`                        | 注释                                                                |
| `adapter/usage_billing.go`                                | `seg.DisplayAmount` → `seg.Cost`                                    |
| `seed/snapshot/audit.go`                                  | `DisplayAmount: row.Cost` → `Cost: row.Cost`                        |
| `store/postgres/ledger_repo.go` (`SumAmountByDepartment`) | `SUM(display_amount)` → `SUM(cost)`                                 |

### 3. Lot Segment：`DisplayAmount` → `Cost`

| 位置                            | 变更                                             |
| ------------------------------- | ------------------------------------------------ |
| `domain/billing/lot/consume.go` | `Segment.DisplayAmount float64` → `Cost float64` |
| `domain/usage/ports.go`         | `LotSegment.DisplayAmount` → `Cost`              |
| `adapter/usage_billing.go`      | mapping 处                                       |

### 4. Dashboard Bucket：`DisplayCost` → `Cost`

| 位置                                      | 变更                                      |
| ----------------------------------------- | ----------------------------------------- |
| `domain/types/usage.go`                   | `UsageBucketRow.DisplayCost` → `Cost`     |
| `domain/types/usage.go`                   | `UsageAggregateRow.DisplayCost` → `Cost`  |
| `domain/types/usage.go`                   | `UsageSummaryTotals.DisplayCost` → `Cost` |
| `domain/types/usage.go`                   | `.Spend()` 方法内部改引用（方法名保留）   |
| DB column `usage_buckets.display_cost`    | → `cost`                                  |
| `store/postgres/usage_repo.go`            | SQL 列名（~8 处）                         |
| `store/postgres/usage_aggregate.go`       | 聚合逻辑                                  |
| `store/postgres/schema.sql`               | 列定义                                    |
| `domain/dashboard/bucket_from_ledger.go`  | `DisplayCost:` → `Cost:`                  |
| `domain/dashboard/dashboard_reconcile.go` | `.DisplayCost` 比较                       |
| `seed/runtime/usage.go`                   | `DisplayCost:` → `Cost:`                  |

### 5. 前端工具函数重命名

| 当前                           | 新名                        | 用途                             | 文件                   |
| ------------------------------ | --------------------------- | -------------------------------- | ---------------------- |
| `quotaToDisplay(quota)`        | `quotaToMoney(quota)`       | quota→货币数值（仅钱包/dev）     | `lib/quota-display.ts` |
| `displayToQuota(display)`      | `moneyToQuota(money)`       | 货币→quota（仅钱包/dev）         | `lib/quota-display.ts` |
| `formatDisplayCurrency(quota)` | `formatQuotaAsMoney(quota)` | quota→格式化货币字符串（仅 dev） | `lib/quota-display.ts` |
| `formatMoney(amount)`          | ✅ 不动                     | 货币金额→格式化字符串            | —                      |

调用方（`features/dev/lib/constants.ts` 等）跟随更新。

### 6. 不动的

| 名字                                               | 理由               |
| -------------------------------------------------- | ------------------ |
| `Budget` / `Consumed` / `Remain`                   | 上下文已明确是货币 |
| `combined_key_remain`                              | ADR 决策不改名     |
| `WalletRemainQuota`                                | 已带 Quota 后缀    |
| `QuotaPerUnit` / `QuotaGranted` / `QuotaRemaining` | 已带 Quota 前缀    |
| `BillingCurrency`                                  | 清晰               |
| `TotalCost` / JSON `cost`                          | 已是最终形态       |
| `QuotaConsumed`（bucket）                          | 已带 Quota 前缀    |
| `.Spend()` 方法                                    | 对外 API，语义清晰 |

## 执行步骤

项目未上线，直接改 schema，不需要 migration。

### Step 1：DB Schema

修改 `schema.sql`：

- `usage_ledger.amount` → `quota_amount`
- `usage_ledger.display_amount` → `cost`
- `usage_buckets.display_cost` → `cost`

### Step 2：Go struct + store 层

1. `UsageLedgerEntry.Amount` → `QuotaAmount`
2. `UsageLedgerEntry.DisplayAmount` → `Cost`
3. `Segment.DisplayAmount` → `Cost`（lot/consume.go + usage/ports.go）
4. `UsageBucketRow.DisplayCost` → `Cost`（及 AggregateRow、SummaryTotals）
5. 所有 postgres repo 的 SQL 列名和 scan 对齐
6. `SumAmountByDepartment` SQL 改为 `SUM(cost)`

### Step 3：Domain 逻辑层

跟随 struct rename 更新引用（纯机械替换）：

- `entry_build.go`、`ingest.go`、`lot/ledger.go`
- `bucket_from_ledger.go`、`dashboard_reconcile.go`
- `consumed_attrib.go`、`budget_reconcile.go`
- `adapter/usage_billing.go`、`ledger_audit.go`

### Step 4：前端

1. `lib/quota-display.ts` 重命名 3 个函数
2. `features/dev/lib/constants.ts` 调用方更新

### Step 5：Seed + 测试

- `seed/snapshot/audit.go`、`seed/runtime/usage.go`
- 所有 `_test.go` 中的 `DisplayAmount`、`DisplayCost`、`.Amount` 引用

### Step 6：验证

```bash
cd apps/backend && go build ./...
cd apps/frontend && npx tsc --noEmit
cd apps/backend && go test ./...
```

## 预估影响

| 区域           | 文件数       | 改动处      |
| -------------- | ------------ | ----------- |
| Schema         | 1            | 3 列        |
| Go struct 定义 | 4            | 6 字段      |
| Go store/repo  | 5            | ~15 处 SQL  |
| Go domain 逻辑 | 10           | ~25 处引用  |
| Go seed        | 2            | ~5 处       |
| Go 测试        | ~15          | ~35 处      |
| 前端           | 3            | ~8 处       |
| **合计**       | **~40 文件** | **~100 处** |

全部是机械化 rename，零逻辑变更。一个 commit 完成。
