# ModelRatio 同步：现状与修复

## 背景

Passthrough 计费的正确性依赖于：**NewAPI 的 ModelRatio 和 TJ 的全局定价保持一致**。

NewAPI `logs.quota` 的计算公式：
```
quota = tokens × model_ratio × quota_per_unit (简化)
```

TJ 通过 `UpsertModelRatio` 把全局价转换为 NewAPI 格式推送：
```go
// pkg/modelcatalog/pricing.go
modelRatio = inputPrice / 2          // 元/1M tokens → ratio per 1K tokens
completionRatio = outputPrice / inputPrice
```

如果 ratio 不一致，用户在 NewAPI gateway 侧的预扣额和 TJ passthrough 入账会有偏差。

---

## 当前已实现的推送路径

| # | 触发场景 | 代码位置 | 方式 |
|---|---------|---------|------|
| 1 | 平台管理员改全局价 | `domain/pricing/service.go` → `SetGlobalPrice()` | 改价后立即 best-effort 推送 |
| 2 | 创建自定义模型（带价格） | `domain/models/service.go` → `CreateModel()` | 建模型后立即推送 |
| 3 | 更新模型价格 | `domain/models/service.go` → `UpdateModel()` | 改价后立即推送 |
| 4 | CatalogSync 拉取全局价（私有化） | `worker/catalogsync/execute.go` → `syncPricing()` | 拉到新价后逐条推送 |
| 5 | 平台管理员通过 HTTP 设模型价格 | `handler/platform/models.go` → 调用 `SetGlobalPrice` | 同路径 1 |

所有路径都是 best-effort：推送失败只 `slog.Warn`，不阻塞业务。

---

## 当前缺失：全量兜底

`pricing/service.go` 中有 `FullSyncToNewAPI()` 方法：

```go
// FullSyncToNewAPI pushes all current global prices to NewAPI (cron job).
func (s *Service) FullSyncToNewAPI(ctx context.Context) error {
    prices, _ := s.store.ModelPricing().CurrentPricesBatch(ctx, s.cfg.TokenJoyCompanyID, time.Now())
    for _, p := range prices {
        _ = s.client.UpsertModelRatio(ctx, p.ModelType, p.InputPrice, p.OutputPrice)
    }
    return nil
}
```

**问题：这个方法没有被任何地方调用。** 注释写着 "cron job" 但从未注册为 River PeriodicJob。

### 影响

- 当前架构（TJ 自己 `CalcQuota` 重算）：影响小，只是 gateway 预扣额有偏差
- Passthrough 架构：**致命**——ratio 不一致 = 入账金额错误，无兜底修复

---

## 修复方案

### 注册为 River PeriodicJob

在 `infra/river/periodic/jobs.go` 中添加：

```go
// 全量对齐 NewAPI ModelRatio（兜底 best-effort 推送失败的场景）
if cfg.IngestEnabled() {
    periodicJobs = append(periodicJobs, river.NewPeriodicJob(
        river.PeriodicInterval(5*time.Minute),
        func() (river.JobArgs, *river.InsertOpts) {
            return jobs.PricingFullSyncArgs{}, nil
        },
        nil,
    ))
}
```

新增 Job Args：

```go
// infra/jobs/kinds_pricing.go
package jobs

import "github.com/riverqueue/river"

type PricingFullSyncArgs struct{}

func (PricingFullSyncArgs) Kind() string { return "pricing_full_sync" }

func (PricingFullSyncArgs) InsertOpts() river.InsertOpts {
    return river.InsertOpts{Queue: config.RiverQueueDefault, UniqueOpts: river.UniqueOpts{ByPeriod: 4 * time.Minute}}
}
```

新增 Worker：

```go
// infra/river/workers/pricing_sync.go
type PricingFullSyncWorker struct {
    river.WorkerDefaults[jobs.PricingFullSyncArgs]
    pricingSvc *domainpricing.Service
}

func (w *PricingFullSyncWorker) Work(ctx context.Context, _ *river.Job[jobs.PricingFullSyncArgs]) error {
    return w.pricingSvc.FullSyncToNewAPI(ctx)
}
```

### 间隔选择：5 分钟

- 全量推送是 read-modify-write（GET options → merge → PUT），开销 O(模型数)
- 模型数通常 < 100，每次 < 200 次 HTTP 调用，可接受
- 5min 意味着最坏情况 ratio 偏差窗口 = 5 分钟

---

## Passthrough 上线后的额外保障

### 入账时记录 raw.Quota

在 `call_detail` 中保留 `raw.Quota` 原始值：

```go
entry.CallDetail.RawQuota = entry.QuotaAmount  // passthrough 前的原始值
entry.QuotaAmount = int64(math.Ceil(float64(entry.QuotaAmount) * discount))
```

这样事后对账可以发现 ratio 不一致导致的偏差。

### 对账告警（可选，后续迭代）

定时抽样：对比 `models.input_price` 转换出的 ratio 和 NewAPI 实际存储的 ModelRatio option 值，不一致时告警。

---

## 验证清单

- [ ] `FullSyncToNewAPI` 注册为 PeriodicJob
- [ ] Worker 注册到 River workersBundle
- [ ] 启动后观察日志确认定时执行
- [ ] 手动 kill 一次 `UpsertModelRatio` 推送，验证 5min 后自动恢复
