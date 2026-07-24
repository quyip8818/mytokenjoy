# NewAPI User Quota 同步策略设计

**状态：** 已实现（v1 — PreCreditFunc 模式）  
**实际实现差异：** 设计初稿描述 PostCreditFunc，最终采用 PreCreditFunc（先同步再本地提交），见 `lot/consume.go`。

## 1. 现状分析

### 1.1 架构关系

```
用户 → NewAPI（网关/代理）→ 上游 LLM
                ↓ consume_log
       TJ Backend（入账/计费）
```

- **用户请求不经过 TJ Backend**，直接打到 NewAPI
- NewAPI 用自身的 `user.remain_quota` 做额度预检（先扣后调用）
- TJ Backend 事后从 NewAPI 拉取 consume_log，执行本地入账（FIFO lot 消耗 + 预算投影）

### 1.2 NewAPI 额度模型

| 层级 | 对象 | 当前代码策略 |
|------|------|-------------|
| Token（密钥） | `remain_quota` | 创建时 `UnlimitedQuota: true`，不限额 |
| User（钱包用户） | `remain_quota` | 仅在公司创建/bootstrap 时设置一次 |

关键点：NewAPI 的额度检查作用在 **User** 级别。所有挂载在该 user 下的 token 共享同一个 `user.remain_quota`。

### 1.3 当前问题

正式公司充值后，NewAPI 端 user quota 不会增加。本地 wallet 有钱但 NewAPI 端已耗尽，请求被 NewAPI 直接拒绝。

---

## 2. 设计目标

- NewAPI `user.remain_quota` 始终 >= 本地 `wallet_remain_quota`，保证有钱时 NewAPI 不拦截
- 额度同步是**宽松方向**的（允许 NewAPI 端多于本地），精确扣减在本地 FIFO lot 完成
- NewAPI quota 是"粗粒度闸门"，本地 lot 系统是"精确计量"

---

## 3. 核心设计

### 3.1 Ceiling 模式

不做双向精确同步，把 NewAPI user quota 当作"信用额度天花板"：
- NewAPI 实时扣（每个请求即扣），本地入账异步（consume_log → ingest）
- 同步方向：**只加不减**——本地充值 → 追加 NewAPI quota
- 消费时不反向通知（NewAPI 已自动扣了）

### 3.2 环境分支规则

| 环境 | Lot 类型 | NewAPI User Quota 策略 |
|------|----------|----------------------|
| **非 Prod** | **所有类型（含 Mock）** | `CreditFromLot` 后 `add_quota(deltaQuota)` |
| **Prod** | Paid / Gift / Adjust | `CreditFromLot` 后 `add_quota(deltaQuota)` |
| **Prod** | Mock | **不同步**（创建时已一次性给大额） |
| 任何 | Overdraft | 不同步 |

### 3.3 非 Prod 同步 Mock 的原因

- 非 prod 中 mock lot 是唯一额度来源
- test-model 请求同样扣 `user.remain_quota`
- 不同步则开发者无法正常测试

### 3.4 Prod 排除 Mock 的原因

- Prod trial/demo 创建时已 `add_quota(500000*500000)`
- Trial 期间 test-model 消耗极低，不可能用完

---

## 4. 实现方案

### 4.1 挂载点：`CreditFromLot` PreCreditFunc

同步逻辑通过 `PreCreditFunc` variadic 参数注入 `CreditFromLot`。
PreCreditFunc 在本地事务 **之前** 执行——NewAPI quota 先加，本地再提交。

设计理由（"先加后提交"）：
- PreCreditFunc 成功 + 本地 tx 失败 → NewAPI 多了额度，宽松闸门不影响正确性
- PreCreditFunc 失败 → 本地 tx 不执行，用户看到"充值失败"重试即可
- 反过来（先提交后同步）→ 本地有余额但 NewAPI 拒绝请求，用户体验极差

```go
// domain/billing/lot/consume.go
func CreditFromLot(ctx, st, order, lot, delta, beforeCommit ...PreCreditFunc) error {
    if len(beforeCommit) > 0 && beforeCommit[0] != nil {
        if err := beforeCommit[0](ctx, lotRow); err != nil {
            return err
        }
    }
    return st.WithTx(...)
}
```

`billing.service.syncQuotaToNewAPI` 作为 PreCreditFunc 注入：

```go
func (s *service) syncQuotaToNewAPI(ctx context.Context, lot store.RechargeLot) error {
    if lot.LotKind == store.LotKindOverdraft { return nil }
    if s.cfg.IsProductionDeploy() && lot.LotKind == store.LotKindMock { return nil }
    if s.quotaSyncer == nil { return nil }
    walletUserID, ok := company.ResolveNewAPIWalletCompanyID(ctx, s.store.Company())
    if !ok { return nil }
    return s.quotaSyncer.ManageUser(ctx, walletUserID, "add_quota", lot.QuotaGranted)
}
```

### 4.2 失败处理

- 不回滚本地充值（本地 lot 是 source of truth）
- warn log
- 极端情况下运维手动执行 `ManageUser` 补齐

### 4.3 去除 bootstrap 中的重复 add_quota

`provision/bootstrap.go` 和 `service_create.go` 中 trial/demo 的 `ManageUser(add_quota, 500000*500000)` 删除。
改为：在 Prod 环境，`SeedTrialCredit` 创建 mock lot → `CreditFromLot` → 命中 `IsProduction() && LotKindMock` → 不同步。
Prod trial/demo 的 NewAPI 额度保留由公司创建流程中**单独**给一次大额（保留 `service_create.go` 中那行）。

等一下，这有矛盾——如果 Prod mock lot 不同步，但也删了创建时的 `add_quota`，那 trial 公司 NewAPI 就没有额度了。

**修正：Prod 环境保留 `service_create.go` 中 trial/demo 创建时的一次性 `add_quota`。只删 `provision/bootstrap.go` 中的重复（因为 bootstrap 只在非 prod seed 时跑，而非 prod 会走通用同步路径）。**

### 4.4 公司升级（Trial → Standard）

`ExpireMockLots` 成功后，清零 NewAPI quota：

```go
ManageUser(ctx, walletUserID, "add_quota", -currentNewAPIRemain)
```

清零后用户必须充值才能继续使用——这是预期行为（升级 = 试用结束 → 付费开始）。

### 4.5 无历史数据迁移

项目未上线，不存在已有公司数据需要修复。

---

## 5. 边界情况

| 场景 | 处理 |
|------|------|
| 公司无 `newapi_wallet_company_id` | 跳过 |
| gift lot（amount=0 但 quota>0） | 同步 `lot.QuotaGranted` |
| NewAPI 暂时不可用 | warn log + 跳过 |
| 并发充值 | `add_quota` 是增量操作，天然安全 |
| NewAPI remain > 本地 remain | 正常（宽松闸门） |
| bootstrap 非 prod 重复跑 | bootstrap 中删掉独立 add_quota，由通用路径处理 |

---

## 6. 数据流

```
充值流程：
  充值 → CreditFromLot(本地事务) → commit 后 SyncCallback
              ↓                              ↓
     wallet_remain_quota += Δ      ManageUser(add_quota, Δ)

请求流程：
  用户 → NewAPI 扣 user.remain_quota → 转发上游 → consume_log

入账流程：
  consume_log → Ingest → ConsumeLots(FIFO) → wallet_remain_quota -= Δ
                                              (不通知 NewAPI)
```

---

## 7. 不变量

1. `NewAPI user.remain_quota >= 本地 wallet_remain_quota - 未入账消耗`
2. Token 始终 unlimited，额度检查只在 user 层
3. 同步只加不减（升级清零除外）
4. Prod mock lot 不同步；Prod trial/demo 创建时一次性给大额
5. 非 prod 所有 lot 都同步

---

## 8. 实施清单

| 优先级 | 任务 |
|--------|------|
| **P0** | `CreditFromLot` 增加 post-commit SyncCallback 机制 |
| **P0** | billing service 注入 adminport client，实现 `newAPISyncCallback` |
| **P0** | 删除 `provision/bootstrap.go` 中的独立 `add_quota`（非 prod 由通用路径覆盖） |
| **P1** | 升级流程中 `ExpireMockLots` 后清零 NewAPI quota |
