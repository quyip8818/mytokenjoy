# ADR: NewAPI Wallet User Quota 同步策略

> 状态：草案  
> 日期：2026-07-20  
> 关联：docs/plan/unified-quota-implementation-guide.md

---

## 背景

统一 quota 迁移删除了旧的 `wallet_sync` 双向 reconcile 机制。Token 已设为 `unlimited_quota=true`，但 NewAPI 仍在 **user-level** 检查 quota——当 user quota ≤ 0 时拒绝所有请求（`insufficient_user_quota`）。

NewAPI 的限额分两层：

| 层级 | 检查时机 | 我们的策略 |
|------|---------|-----------|
| User quota | 每次 API 调用 | 需要保持 ≥ tokenjoy wallet_remain |
| Token quota | 每次 API 调用 | `unlimited_quota=true`，不检查 |

tokenjoy Gateway precheck 是业务限额控制面（combined_key_remain + wallet_remain），NewAPI user quota 是**物理止损层**——防止绕过 Gateway 直连 NewAPI 时的无限消费。

---

## 设计

### 核心原则

```
tokenjoy wallet_remain = SSOT（单一事实源）
NewAPI user quota = 物理止损镜像（≥ wallet_remain 即可）
```

### 同步方向：单向、增量、事务后

```
充值事务 (PG)                          NewAPI Admin API
─────────────────                     ─────────────────
CreditFromLot (tx) ──commit──►  topUpNewAPIQuota(companyID, quotaGranted)
                                         │
                                         ▼
                                  POST /api/user/topup
                                  { user_id, quota: +quotaGranted }
```

- **单向**：tokenjoy → NewAPI，永不反向读取
- **增量**：每次充值只 TopUp 充入的 delta，不重算全量
- **事务后**：TopUp 在 PG 事务 commit 后执行，失败不回滚充值

### 触发点

所有充值路径的公共出口：`lot_confirm.go` 中每个 `CreditFromLot` 成功后。

```go
// lot_confirm.go
func (s *service) confirmPaidRecharge(...) error {
    ...
    if err := billinglot.CreditFromLot(ctx, s.store, order, lot, lot.QuotaGranted); err != nil {
        return err
    }
    s.topUpNewAPIQuota(ctx, companyID, lot.QuotaGranted) // post-commit, best-effort
    return nil
}
```

### 失败容错

| 场景 | 影响 | 恢复 |
|------|------|------|
| TopUp HTTP 超时 | tokenjoy 侧正常，NewAPI user quota 不足 | 下次充值补上；bootstrap 启动补齐 |
| TopUp 返回错误 | 同上 | 同上 |
| NewAPI 宕机 | Gateway precheck 正常（读 PG），NewAPI 转发失败 | NewAPI 恢复后 bootstrap 补齐 |

**不一致窗口**：从 TopUp 失败到下次补齐之间。此时 Gateway precheck 放行但 NewAPI 拒绝。概率极低（NewAPI 99.9%+ 可用），影响可接受——和 ingest 延迟导致的短暂超发是同类风险。

### Bootstrap 补齐

应用启动时（`provision/bootstrap.go`），对已配置的 wallet company：

```go
currentQuota := client.GetUserQuota(ctx, walletCompanyID)
if currentQuota < co.WalletRemain {
    client.TopUp(ctx, { CompanyID: walletCompanyID, Quota: co.WalletRemain - currentQuota })
}
```

语义：**保证 NewAPI user quota ≥ tokenjoy wallet_remain**。不用 MaxInt32 硬编码。

### 不需要做的事

| 原 wallet_sync 功能 | 新方案 | 原因 |
|---------------------|--------|------|
| 定时 ReconcileWalletDrift | 删除 | bootstrap 启动补齐已覆盖 |
| 反向读 NewAPI quota → 修正 tokenjoy | 删除 | tokenjoy 是 SSOT |
| River debounce job | 删除 | 直接 HTTP 调用，无需异步 |
| 精确同步 model_price → quota 换算 | 删除 | wallet 和 token 解耦；token unlimited |

---

## 改动清单

| 文件 | 变更 |
|------|------|
| `domain/billing/service.go` | `service` struct 加 `adminClient adminport.Port`；`NewService` 接收 port |
| `domain/billing/wallet_topup.go`（新） | `topUpNewAPIQuota(ctx, companyID, delta)` 实现 |
| `domain/billing/lot_confirm.go` | 4 个充值路径调用 `s.topUpNewAPIQuota` |
| `newapisync/provision/bootstrap.go` | 已有 user 时 TopUp 到 `wallet_remain`（替换 MaxInt32 hacky 逻辑） |
| `adapter/billing.go`（如需要） | wire NewService 时传入 adminport.Port |
| 测试 | 验证 TopUp 被调用 + delta 正确 |

---

## 副作用分析

### 1. 重复 TopUp

NewAPI TopUp 是 **additive**（`user.quota += delta`）。如果同一笔充值因为 retry 导致 TopUp 两次：
- tokenjoy 侧不受影响（CreditFromLot 有 idempotency 保护）
- NewAPI 侧 user quota 多加了 delta

后果：NewAPI user quota > tokenjoy wallet_remain。**这没问题**——因为 Gateway precheck（读 PG wallet_remain）在 NewAPI 之前拦截，真正的限额由 tokenjoy 控制。多余的 NewAPI quota 只是"宽松了一点"，不造成安全风险。

### 2. 并发充值

多个充值并发执行各自 TopUp 各自的 delta。NewAPI TopUp 在其服务端是原子的（`UPDATE users SET quota = quota + $1`）。无竞态。

### 3. 消费路径

NewAPI 消费时**自动扣减** user quota（这是 NewAPI 内部行为）。tokenjoy ingest 独立地从 lot 扣减 wallet_remain。两者各算各的，不需要交叉同步。

**关键不变量**：`NewAPI user quota ≥ tokenjoy wallet_remain` 在正常运行时始终成立（因为 NewAPI 消费扣减的 quota 和 tokenjoy ingest 扣减的 wallet_remain 源自同一个 consume_log.quota，数值相等）。

### 4. Overdraft

Overdraft lot 扩展时（`ExpandOverdraftLot`），tokenjoy wallet_remain 保持 0（overdraft 借贷不增加真实余额），所以 **不需要 TopUp**。NewAPI 侧的 user quota 此时可能已经是 0 或负值——这导致 NewAPI 拒绝新请求，和 tokenjoy Gateway precheck 的 "wallet_remain ≤ 0 → 拒绝" 行为一致。不需要额外处理。

### 5. Trial 充值（MockLot）

`SeedTrialCredit` 创建 mock lot 时也走 `CreditFromLot`，加上 TopUp 后 NewAPI user quota 也会增加。Trial → Standard 升级时 `ExpireMockLots` 减少 wallet_remain，但不减少 NewAPI user quota。这没问题——多余 quota 在正常消费中自然扣减；且下次 bootstrap 会修正。

### 6. AdminPort 依赖

billing service 增加了对 `adminport.Port` 的依赖。在 NewAPI 未配置（self-hosted/testing 环境）时 port 为 nil：
- `topUpNewAPIQuota` 内部 nil check → skip
- 不影响纯 tokenjoy 记账功能

---

## 验证场景

| 场景 | 预期 |
|------|------|
| 充值 ¥50 (QPU=500000) | TopUp delta=25,000,000 |
| Gift 1,000,000 quota | TopUp delta=1,000,000 |
| 充值失败（CreditFromLot 回滚） | TopUp 不执行 |
| TopUp HTTP 失败 | 充值正常，log warning |
| Bootstrap 启动 user quota=0, wallet_remain=50M | TopUp delta=50,000,000 |
| Bootstrap 启动 user quota=100M, wallet_remain=50M | 不 TopUp（quota 已 ≥ remain） |
| Overdraft 扩展 | 不 TopUp（wallet_remain 不变） |
| 并发 3 笔充值 | 3 次独立 TopUp，互不干扰 |

---

## 决策

采用**单向增量 TopUp**方案。相比旧 wallet_sync：
- 删除了双向 reconcile 的全部复杂性
- 删除了 River job + debounce + drift 检测
- 保留了 NewAPI user quota 作为物理止损的安全属性
- 失败容错由 bootstrap 启动补齐覆盖，无需定时任务
