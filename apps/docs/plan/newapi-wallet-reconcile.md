# NewAPI User Wallet 对齐方案

> ⚠️ **本方案已废弃。** 实际实现采用了更简单的"实时 override"方案，见 [Backend-计费模式.md](../Backend-计费模式.md) §4.4。
>
> 原方案设计了 dirty 标记 + periodic job 定期对齐。最终选择了每次 wallet 变更后直接 `set_quota(mode=override)` 覆盖 NewAPI，无需额外的 worker 或 schema 变更。

---

## 1. 现状

### 当前同步点

| 时机 | 操作 | 代码位置 |
|------|------|---------|
| 充值（PlatformRecharge/Gift/Adjust/SelfRecharge） | `ManageUser(walletUserID, "add_quota", lot.QuotaGranted)` | `billing/service.go → syncQuotaToNewAPI` |
| Company 创建（trial/demo） | `ManageUser(userID, "add_quota", 500000*500000)` | `company/service_create.go` |
| Bootstrap（local） | `ManageUser(userID, "add_quota", 500000*500000)` | `newapisync/provision/bootstrap.go` |

**问题：** 只有"加"没有"对齐"。如果引入合同价格（TokenJoy 入账用合同价，NewAPI 扣全局价），两侧会逐渐偏离。

### wallet 余额公式

```
NewAPI wallet = 历次充值累计 - NewAPI 按全局价扣的消费总量
TokenJoy lot  = 历次充值累计 - TokenJoy 按合同价扣的消费总量

偏差 = Σ (全局价 - 合同价) × 每笔消费的 tokens
```

---

## 2. 方案

两条路径协作：

1. **充值时正常同步**（保持现有逻辑，不改）
2. **定期对齐**（新增 periodic job，把 NewAPI wallet 重置为 TokenJoy SOT）

```
充值 → add_quota（立即，保证用户充完马上能用）
定期 → override wallet = TokenJoy wallet_remain_quota（纠偏）
```

---

## 3. 定期对齐 Worker

### 3.1 Job 定义

```go
// internal/infra/jobs/kinds_wallet_reconcile.go
type WalletReconcileArgs struct{}

func (WalletReconcileArgs) Kind() string { return "wallet_reconcile" }

func (WalletReconcileArgs) InsertOpts() river.InsertOpts {
    return river.InsertOpts{
        Queue:       config.RiverQueueDefault,
        MaxAttempts: 3,
    }
}
```

注册为 periodic job（和 IngestReconcile、CatalogSync 同级）：

```go
// infra/river/periodic/jobs.go
river.NewPeriodicJob(
    river.PeriodicInterval(5 * time.Minute),  // 每 5 分钟
    func() (river.JobArgs, *river.InsertOpts) {
        return jobs.WalletReconcileArgs{}, nil
    },
    nil,
)
```

### 3.2 Worker 实现

```go
// internal/infra/river/workers/wallet_reconcile.go
type WalletReconcileWorker struct {
    river.WorkerDefaults[jobs.WalletReconcileArgs]
    store       store.Store
    quotaSyncer billing.QuotaSyncer
    logger      *slog.Logger
}

func (w *WalletReconcileWorker) Work(ctx context.Context, _ *river.Job[jobs.WalletReconcileArgs]) error {
    // 1. 查找需要对齐的 company（有消费活动的）
    companies, err := w.store.Company().ListDirtyWalletCompanies(ctx)
    if err != nil {
        return err
    }
    
    for _, co := range companies {
        walletUserID, ok := store.ConfiguredNewAPIWalletCompanyID(&co)
        if !ok {
            continue
        }
        
        // 2. 用 TokenJoy SOT（wallet_remain_quota）覆盖 NewAPI wallet
        if err := w.quotaSyncer.ManageUser(ctx, walletUserID, "set_quota", co.WalletRemainQuota); err != nil {
            w.logger.Warn("wallet reconcile failed", "company_id", co.ID, "error", err)
            continue // best-effort，不阻塞其他 company
        }
        
        // 3. 清除 dirty 标记
        _ = w.store.Company().ClearWalletDirty(ctx, co.ID)
    }
    return nil
}
```

### 3.3 ManageUser action

当前 `ManageUser` 支持 `"add_quota"`（增量）。需要确认 NewAPI `/api/user/manage` 是否支持直接设值。

从代码看：
```go
body["mode"] = "add"  // 当前写死 add
```

NewAPI 的 manage API 通常支持 `mode: "override"`。改为：

```go
func (c *Client) ManageUser(ctx context.Context, userID int64, action string, value int64) error {
    body := map[string]any{
        "id":     userID,
        "action": action,
        "value":  value,
    }
    switch action {
    case "add_quota":
        body["mode"] = "add"
    case "set_quota":
        body["mode"] = "override"  // 新增：直接覆盖
        body["action"] = "add_quota"
    }
    return c.do(ctx, "POST", "/api/user/manage", body, nil)
}
```

---

## 4. 如何区分"有消费的 company"

### 方案：dirty 标记

在 `companies` 表加一个字段：

```sql
ALTER TABLE companies ADD COLUMN wallet_dirty BOOLEAN NOT NULL DEFAULT FALSE;
```

**写入时机：** 每次入账（`IngestRaw`）成功后，标记该 company dirty：

```go
// ingest.go — 事务内，在 step 7 enqueue 之前
_ = st.Company().MarkWalletDirty(ctx, companyID)
```

**清除时机：** 对齐 worker 成功覆盖 NewAPI wallet 后清除。

**效果：**
- 没有消费 → wallet_dirty = false → 对齐 worker 跳过
- 有消费 → wallet_dirty = true → 对齐 worker 处理 → 清除

这样即使有 10000 个 company，只有本周期有消费的那几十个会被处理。

### 查询

```go
func (r *companyRepo) ListDirtyWalletCompanies(ctx context.Context) ([]store.Company, error) {
    rows, err := r.db.Query(ctx, `
        SELECT id, name, ..., wallet_remain_quota, new_api_wallet_company_id
        FROM companies
        WHERE wallet_dirty = TRUE AND new_api_wallet_company_id IS NOT NULL
        LIMIT 100
    `)
    // ...
}
```

---

## 5. 现有充值同步是否可以去除？

**不能完全去除。** 理由：

充值后用户立即使用 → 如果只靠定期对齐（5 分钟后），用户充完钱要等 5 分钟才能调 API。

**保留充值时 add_quota**，作用是"立即可用"。定期对齐的作用是"纠偏"。

| 同步点 | 保留/去除 | 理由 |
|--------|----------|------|
| 充值时 `add_quota` | **保留** | 用户充值后立即可用 |
| Company 创建时给大值 | **保留**（trial/demo）| 初始 quota |
| Bootstrap 给大值 | **保留**（local）| Local 初始化 |
| **新增：定期对齐** | **新增** | 纠偏合同价偏差 |

---

## 6. 无合同价的 company 是否需要对齐？

**不需要频繁对齐。** 如果没有合同价，NewAPI 全局价和 TokenJoy 用同一个 ratio，两侧 wallet 自然保持一致（充值加多少，消费扣多少，完全对等）。

但仍然有极小的偏差可能（浮点精度、并发边界），所以 dirty 标记机制会自动覆盖这些 case——有消费就标 dirty，对齐 worker 统一处理，不区分是否有合同价。

---

## 7. 时序图

```
用户充值 ¥100
  │
  ├─ TokenJoy: lot.QuotaGranted = 50,000,000 quota
  ├─ TokenJoy: company.wallet_remain_quota += 50,000,000
  ├─ NewAPI:   ManageUser("add_quota", 50,000,000)  ← 立即同步
  │
  ... 用户消费（有合同价 = 全局价的 50%）...
  │
  ├─ NewAPI: 按全局价扣 wallet → 扣了 1000 quota
  ├─ TokenJoy: 入账按合同价 → 实际只扣 500 quota
  ├─ TokenJoy: company.wallet_remain_quota = 49,999,500
  ├─ TokenJoy: company.wallet_dirty = true
  │
  ... 5 分钟后定期对齐 ...
  │
  ├─ Worker: 发现 wallet_dirty = true
  ├─ Worker: ManageUser("set_quota", 49,999,500)  ← 覆盖 NewAPI wallet
  ├─ Worker: ClearWalletDirty
  │
  结果：NewAPI wallet = 49,999,500 = TokenJoy SOT
```

---

## 8. 对齐频率选择

| 频率 | 最大偏差窗口 | 适用场景 |
|------|-------------|---------|
| 1 分钟 | 1 分钟消费量 | 高精度需求 |
| 5 分钟 | 5 分钟消费量 | **推荐**（平衡精度和负载） |
| 15 分钟 | 15 分钟消费量 | 低消费场景 |

对于合同价 = 全局价 50% 的极端 case：5 分钟窗口内 NewAPI wallet 多扣了 2 倍，但由于 token 级 UnlimitedQuota，实际影响只是 user wallet 提前见底（安全方向）。

---

## 9. 实施步骤

| # | 改动 | 优先级 |
|---|------|--------|
| 1 | `companies` 表加 `wallet_dirty` 字段 | P0 |
| 2 | `CompanyRepository` 加 `MarkWalletDirty` / `ClearWalletDirty` / `ListDirtyWalletCompanies` | P0 |
| 3 | `ManageUser` 支持 `"set_quota"` action（mode=override） | P0 |
| 4 | `IngestRaw` 事务内标记 `wallet_dirty = true` | P0 |
| 5 | `WalletReconcileWorker` + periodic job 注册 | P0 |
| 6 | 验证 NewAPI `/api/user/manage` 支持 `mode: "override"` | P0 |
| 7 | 保留现有充值 `add_quota` 逻辑不变 | — |

---

## 10. 注意事项

- `set_quota`（override）是幂等的——多次覆盖同一个值无副作用
- dirty 标记在事务内设置，和 ledger 写入原子——不会漏标
- 对齐失败（NewAPI 不可达）→ best-effort，下次 periodic 重试（dirty 不清除）
- 对齐 worker 有 LIMIT 100 防止单次处理太多 company
- 无需区分有无合同价——dirty 机制自动只处理有消费的 company
- 充值同步保留——保证用户充值后立即可用
- TokenJoy `wallet_remain_quota` 是 SOT，NewAPI wallet 只是它的镜像
