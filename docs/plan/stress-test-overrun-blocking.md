# 压力测试计划：多公司多用户同步请求 — 超额、通知、封锁

## 1. 测试目标

验证在多公司、多用户并发同步请求（Gateway `/v1/chat/completions`）场景下，以下机制仍能正确运作：

| 机制 | 期望行为 |
|------|----------|
| 预算耗尽检测 | `BudgetExhausted` 在各轴（platformKey / member / project / department）正确触发 |
| Key 封锁 | 超额后平台 Key 被 `DisablePlatformKey`，后续请求返回 403 |
| 通知发送 | `overrun_blocked`、`budget_alert_reached`、`overdraft_expanded` 事件在正确时机投递 |
| 幂等性 | 同一 consume log 被重复 ingest 不产生重复扣减或多次封锁 |
| 公司隔离 | 公司 A 的超额不影响公司 B 的 Key 状态 |
| 并发安全 | 多 goroutine 同时 ingest 同一 PlatformKey 不导致负数余额或 panic |

## 2. 测试环境

- 模型使用 `local-test-model`（`DevCallTypeLocalTest`，modelID < 100）
- DEPLOY_ENV=`local`（允许 dev 模型通过 Gateway allowlist 检查）
- 数据库：SQLite（testutil 标准 test store）+ River job runner
- NewAPI 后端：`httptest.Server` mock（返回固定 200 + usage payload）
- 通知：`InMemoryNotifier`（断言内容无需真实投递）
- 每家公司使用独立 seed（`testutil.NewTestStore` per company）

## 3. 测试场景矩阵

### 场景 A：单公司单用户 — Key 预算超额封锁

| 步骤 | 动作 | 预期 |
|------|------|------|
| 1 | 创建公司 C1，部门 D1，成员 M1，PlatformKey PK1(budget=100) | - |
| 2 | Ingest 90 points（mock log） | PK1 status=active，combined_key_remain≈10 |
| 3 | Ingest 15 points（超额） | `ShouldEnqueueOverrun` → true |
| 4 | River runner 执行 overrun job | PK1 status=disabled |
| 5 | Gateway 请求 PK1 | 返回 403 `budget exhausted` |
| 6 | 断言 notifier 收到 `overrun_blocked` 事件，payload 含 scope=platformKey |

### 场景 B：单公司多用户 — Member 个人预算超额

| 步骤 | 动作 | 预期 |
|------|------|------|
| 1 | 公司 C1，成员 M1(personalBudget=50), M2(personalBudget=200) | - |
| 2 | M1 的 Key 消耗 55 points | M1 所有 Key disabled，M2 不受影响 |
| 3 | Gateway 请求 M2 的 Key | 200 OK |
| 4 | Gateway 请求 M1 的 Key | 403 |
| 5 | 通知只发给 M1 scope=member |

### 场景 C：多公司并发 — 隔离性

| 步骤 | 动作 | 预期 |
|------|------|------|
| 1 | 创建 C1(wallet=1000, key budget=50) 和 C2(wallet=1000, key budget=50) | - |
| 2 | 并发 ingest：C1 消耗 60（超额），C2 消耗 30 | - |
| 3 | 断言 C1 PK disabled，C2 PK active | 公司隔离 |
| 4 | 通知列表分离，C2 无 overrun 通知 |

### 场景 D：高并发 ingest 同一 Key — 竞态安全

| 步骤 | 动作 | 预期 |
|------|------|------|
| 1 | PK1 budget=100，当前 consumed=90 | - |
| 2 | 10 goroutine 各 ingest 5 points（总计 50） | - |
| 3 | 最终 consumed=140，remain≤0 | - |
| 4 | PK1 disabled 恰好一次（幂等锁保证） | - |
| 5 | overrun 通知恰好一条 |

### 场景 E：部门预算阈值告警 + 超额（仅通知不封锁）

| 步骤 | 动作 | 预期 |
|------|------|------|
| 1 | D1 budget=1000，AlertRule threshold=[50,80,100] | - |
| 2 | Ingest 510 points（>50%） | `budget_alert_reached` threshold=50 |
| 3 | Ingest 310 points（>80%） | `budget_alert_reached` threshold=80 |
| 4 | Ingest 200 points（>100%） | `overrun_blocked` + notifyOnly=true |
| 5 | Key 仍然 active（部门级仅通知，不封锁 Key） |

### 场景 F：项目预算超额 — 项目级 Key 封锁

| 步骤 | 动作 | 预期 |
|------|------|------|
| 1 | Project P1(budget=200)，PK1 scope=project 绑定 P1 | - |
| 2 | Ingest 210 points | overrun job 封锁 scope=project Key |
| 3 | 同项目其他 project_member Key 也被封锁 |
| 4 | 通知 payload 含 scope=project, projectId |

### 场景 G：项目内成员预算超额 — 精细封锁

| 步骤 | 动作 | 预期 |
|------|------|------|
| 1 | Project P1，memberBudget[M1]=80，PK1 scope=project_member | - |
| 2 | M1 在 P1 消耗 85 points | - |
| 3 | 仅 M1 的 project_member Key 被封锁，P1 项目 Key 不受影响 |
| 4 | 通知 scope=project_member |

### 场景 H：Wallet 余额不足 — Gateway 直接拒绝

| 步骤 | 动作 | 预期 |
|------|------|------|
| 1 | C1 wallet remain < minEstimatePoint | - |
| 2 | Gateway 请求 | 返回 "insufficient wallet points" |
| 3 | 无 ingest 产生 |

### 场景 I：Overdraft 扩展通知

| 步骤 | 动作 | 预期 |
|------|------|------|
| 1 | 公司 C1 有 lot 余量接近耗尽 | - |
| 2 | Ingest 触发 overdraft lot 消费 | - |
| 3 | 收到 `overdraft_expanded` 通知 |

### 场景 J：幂等 ingest — 重复不触发二次封锁

| 步骤 | 动作 | 预期 |
|------|------|------|
| 1 | Ingest logID=5001 → PK1 disabled | - |
| 2 | 再次 Ingest logID=5001 | 幂等检查跳过，consumed 不增，通知不重发 |

## 4. 实现架构

```
apps/backend/tests/domain/budget/stress_overrun_test.go   ← 主测试文件
apps/backend/tests/testutil/stress/                       ← 压测辅助工具
    companies.go      → 批量创建多公司场景
    concurrent.go     → 并发 ingest helper
    assertions.go     → 通知/Key状态/consumed 断言
```

### 关键辅助函数

```go
// BuildMultiCompanyStress 创建 N 家公司，每家含部门、成员、PlatformKey。
// 返回 []CompanyFixture，各自有独立 context 和 store 切面。
func BuildMultiCompanyStress(t *testing.T, n int, opts StressOpts) []CompanyFixture

// ConcurrentIngest 并发执行 N 次 ingest，等全部完成后返回错误列表。
func ConcurrentIngest(t *testing.T, ingest *usage.IngestService, logIDs []int64) []error

// AssertKeyStatus 断言 PlatformKey 最终状态。
func AssertKeyStatus(t *testing.T, st store.Store, keyID, expectedStatus string)

// AssertNotifications 断言 InMemoryNotifier 收到指定事件。
func AssertNotifications(t *testing.T, notifier *mock.InMemoryNotifier, expected []ExpectedNotification)
```

## 5. 测试模型选择

使用 `local-test-model`（`modelcatalog.DevCallTypeLocalTest`）：

- ModelID = 1（< ProdCatalogModelIDStart=100），属 dev catalog
- Gateway 在 `DEPLOY_ENV=local` 时免 allowlist 检查
- Ingest pipeline 不依赖真实 LLM 后端
- Mock NewAPI Server 返回固定 usage：

```go
func mockNewAPIHandler(inputTokens, outputTokens int) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        resp := map[string]any{
            "id":    "chatcmpl-stress",
            "model": "local-test-model",
            "usage": map[string]int{
                "prompt_tokens":     inputTokens,
                "completion_tokens": outputTokens,
                "total_tokens":      inputTokens + outputTokens,
            },
        }
        json.NewEncoder(w).Encode(resp)
    }
}
```

## 6. 并发参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| 公司数 | 5 | 同时活跃公司 |
| 每公司用户数 | 3 | 每家公司的成员（各有独立 PlatformKey） |
| 每用户并发请求 | 10 | goroutine 数 |
| 单请求消耗 | 5 points | 模拟 token cost |
| Key 预算 | 30 points | 刚好在 ~6 次后超额 |
| 部门预算 | 500 points | 触发 alert threshold |
| 成员个人预算 | 25 points | 部分成员先超额 |

## 7. 验证要点

### 7.1 正确性断言

- [ ] 超额后 Key status 变为 disabled/revoked
- [ ] Gateway precheck 对 disabled Key 返回 `ErrBudgetExhausted`
- [ ] 通知 payload 中 consumed > budget
- [ ] 幂等：相同 idempotencyKey 不产生重复 ledger entry
- [ ] Alert dedup key 防止相同周期同阈值重复告警

### 7.2 隔离性断言

- [ ] 公司 A 超额封锁，公司 B 同 model 请求仍 200
- [ ] 成员 M1 超额，同部门 M2 Key 不受影响
- [ ] Project P1 超额，不影响同部门非项目 Key

### 7.3 并发安全断言

- [ ] budget_consumed 最终值 = sum(all ingested amounts)
- [ ] DisablePlatformKey 调用幂等（多次调用不报错）
- [ ] 无 data race（`go test -race`）
- [ ] 无 deadlock（AcquireBudgetLock 在高并发下不卡死）

## 8. 运行方式

```bash
# 完整压力测试（含 -race 检测）
cd apps/backend
go test -race -timeout 120s -run TestStress ./tests/domain/budget/

# 仅特定场景
go test -race -run TestStressMultiCompanyIsolation ./tests/domain/budget/
go test -race -run TestStressConcurrentIngestSameKey ./tests/domain/budget/

# 调节并发参数（通过环境变量）
STRESS_COMPANIES=10 STRESS_USERS_PER_CO=5 STRESS_CONCURRENCY=20 \
  go test -race -timeout 300s -run TestStress ./tests/domain/budget/
```

## 9. 预期输出示例

```
=== RUN   TestStressOverrunBlocking
=== RUN   TestStressOverrunBlocking/ScenarioA_SingleKey
    --- PASS: key PK1 disabled after 105/100 consumed
=== RUN   TestStressOverrunBlocking/ScenarioB_MultiMember
    --- PASS: M1 keys disabled, M2 keys active
=== RUN   TestStressOverrunBlocking/ScenarioC_MultiCompany
    --- PASS: 5 companies isolated, only 2 exceeded → disabled
=== RUN   TestStressOverrunBlocking/ScenarioD_Concurrent
    --- PASS: 10 goroutines, final consumed=140, key disabled once, 1 notification
=== RUN   TestStressOverrunBlocking/ScenarioE_DeptAlert
    --- PASS: 3 threshold alerts + overrun notify-only, key still active
--- PASS: TestStressOverrunBlocking (4.21s)
```

## 10. 风险与限制

| 风险 | 缓解 |
|------|------|
| SQLite 并发写锁限制 | 每公司独立 store instance，或使用 WAL 模式 |
| River job runner 单线程 | 测试中手动调用 `runner.RunOnce` 确保 overrun job 被处理 |
| 真实场景有 Redis 缓存 | 压测跳过 `CombinedKeyCache`（设为 noop），仅验证 PG 层正确性 |
| Wallet 余额来自远端 | Mock `NewAPIWalletReader` 返回固定值 |

## 11. 后续扩展

- 增加 Gateway 层 E2E 压测（HTTP 级别 vegeta/k6）
- 加入 Redis CombinedKeyCache 一致性验证（PG vs Redis 最终一致）
- 大规模场景：100 公司 × 50 用户 × 100 并发，验证无内存泄漏
- 模拟网络延迟下 NewAPI DisablePlatformKey 调用失败后的重试逻辑
## 10. QPS 瓶颈分析与性能基准

### 10.1 系统瓶颈点分析

请求处理分两条路径，瓶颈不同：

| 路径 | 操作 | 瓶颈 | 理论上限 |
|------|------|------|----------|
| **Gateway Precheck**（读路径） | 1× PG 查询（LoadPrecheckContext） + 1× Redis GET（CombinedKeyRemain） | PG pool 50 conn + Redis | **2000–5000 QPS/实例** |
| **Ingest Pipeline**（写路径） | TX: LockForUpdate(company) → Ledger Insert → BudgetConsumed UPSERT → CombinedKey Decrement | **PG 行锁 per company** | **200–500 QPS/公司** |

关键瓶颈解释：

```
Gateway Precheck (读):
  ┌─────────────────────────────────────────┐
  │ 1. PrecheckCache (内存) ─ 命中 → 0ms    │  ← 热路径无 IO
  │ 2. Cache miss → PG LoadPrecheckContext  │  ← 单查询 ~1–3ms
  │ 3. Redis GET combined_key_remain        │  ← ~0.5ms
  └─────────────────────────────────────────┘
  总延迟: 命中缓存 < 1ms, 未命中 ~3ms
  瓶颈: PG 连接池 (默认 50 MaxConns)
  理论: 50 conn ÷ 3ms = ~16,000 QPS (实际考虑其他查询竞争 → 2000–5000)

Ingest Pipeline (写):
  ┌─────────────────────────────────────────┐
  │ BEGIN TX                                │
  │ 1. SELECT ... FOR UPDATE (company row)  │  ← 同公司串行
  │ 2. 幂等检查                             │
  │ 3. Lot 消费                             │
  │ 4. Ledger INSERT                        │
  │ 5. budget_consumed UPSERT batch         │
  │ 6. combined_key_remain Decrement        │
  │ 7. Enqueue overrun job                  │
  │ COMMIT                                  │
  └─────────────────────────────────────────┘
  总延迟: ~5–20ms/tx
  瓶颈: pg_advisory_xact_lock(100, companyID) — 同公司完全串行
  理论: 1000ms ÷ 10ms = ~100 QPS/公司 (单公司)
  跨公司完全并行 → N公司 × 100 = N×100 QPS
```

### 10.2 实际 QPS 估算

| 场景 | 预估 QPS | 限制因素 |
|------|----------|----------|
| 单公司 Ingest | 80–150 | company row lock 串行 |
| 10 公司并发 Ingest | 800–1500 | PG pool 50 conn 成为瓶颈 |
| 50 公司并发 Ingest | 2000–3000 | PG CPU + WAL write |
| Gateway Precheck (缓存命中) | 10,000+ | CPU + 内存 |
| Gateway Precheck (缓存未命中) | 2000–5000 | PG pool |
| Rate Limiter 上限 | 30 req/s/key | 配置 `RATE_LIMIT_V1_RATE=30` |

**关键结论：**
- Gateway 读路径不是瓶颈（有 PrecheckCache + Redis）
- Ingest 写路径瓶颈是 **per-company row lock**，但跨公司完全并行
- 真实系统中 Rate Limiter 限制单 Key 30 QPS，这才是业务层上限
- Overrun Job 是异步的（River queue），不阻塞 ingest 本身

### 10.3 集成测试中的性能基准

在 integration test 中加入一个 benchmark subtest，输出 QPS 基线：

```go
t.Run("Benchmark_IngestThroughput", func(t *testing.T) {
    if testing.Short() {
        t.Skip("skip benchmark in short mode")
    }
    
    companies := 5
    ingestsPerCompany := 100
    env := buildStressEnv(t, StressEnvOpts{Companies: companies, ...})
    
    start := time.Now()
    var wg sync.WaitGroup
    for i := 0; i < companies; i++ {
        wg.Add(1)
        go func(coIdx int) {
            defer wg.Done()
            for j := 0; j < ingestsPerCompany; j++ {
                logID := seedConsumeLog(t, env.Companies[coIdx], j)
                _ = env.Companies[coIdx].Ingest.IngestByLogID(ctx, logID, "webhook")
            }
        }(i)
    }
    wg.Wait()
    elapsed := time.Since(start)
    
    totalOps := companies * ingestsPerCompany
    qps := float64(totalOps) / elapsed.Seconds()
    t.Logf("Ingest throughput: %d ops in %v = %.0f QPS (%d companies parallel)",
        totalOps, elapsed, qps, companies)
    
    // 基线断言：确保没有严重性能退化
    if qps < 50 {
        t.Errorf("QPS too low: %.0f (expected >= 50 for %d companies)", qps, companies)
    }
})

t.Run("Benchmark_GatewayPrecheck", func(t *testing.T) {
    if testing.Short() {
        t.Skip("skip benchmark in short mode")
    }
    
    env := buildStressEnv(t, StressEnvOpts{Companies: 1, MembersPerCo: 1, ...})
    fx := env.Companies[0]
    keyHash := fx.Members[0].KeyHash
    
    // Warm cache
    _ = fx.Precheck.Run(ctx, keyHash, "local-test-model", gateway.PrecheckOpts{})
    
    iterations := 1000
    start := time.Now()
    for i := 0; i < iterations; i++ {
        _ = fx.Precheck.Run(ctx, keyHash, "local-test-model", gateway.PrecheckOpts{})
    }
    elapsed := time.Since(start)
    
    qps := float64(iterations) / elapsed.Seconds()
    t.Logf("Gateway precheck (cached): %d ops in %v = %.0f QPS", iterations, elapsed, qps)
    
    if qps < 5000 {
        t.Errorf("Precheck QPS too low: %.0f (expected >= 5000 with warm cache)", qps)
    }
})
```

### 10.4 能否实现 & 建议

**结论：可以实现，但需要分层。**

| 层级 | 测什么 | 怎么测 | 在哪跑 |
|------|--------|--------|--------|
| Unit benchmark | 单函数延迟 | `go test -bench` | CI（快） |
| Integration stress | 多公司并发正确性 + QPS 基线 | 本次计划 | `make test-integration` |
| HTTP load test | 真实 Gateway QPS | k6 / vegeta → httptest.Server | 独立脚本，手动触发 |
| Production load | 真实环境瓶颈 | k6 against staging | 部署后 |

**对于当前 integration test 的建议：**

1. **加 benchmark subtest**（如上），输出 QPS 数字作为 regression guard
2. **不要 hard-fail 在 QPS 上**（CI 机器性能差异大），改用 `t.Logf` 输出 + 宽松下限
3. **真正想压 QPS 上限**，用 k6 打 HTTP 层更合适（后续扩展）
4. **Overrun job 延迟**也要度量：从 ingest 完成到 key 被 disable 的时间（取决于 River worker poll interval）

### 10.5 Rate Limit 配置与实际业务上限

当前配置（`config.go`）：
```
RATE_LIMIT_V1_RATE  = 30  (每秒 token 补充)
RATE_LIMIT_V1_BURST = 60  (令牌桶容量)
```

这意味着：
- 单个 PlatformKey **最多 30 QPS 持续**，burst 60
- 10 个 Key → 300 QPS
- 100 个 Key → 3000 QPS

**业务层 QPS 上限 = Key 数量 × 30**，远低于系统底层能力。集成测试应在 Rate Limiter 之下测试（直接调用 Ingest/Precheck），QPS 基线用来检测性能退化而非测极限。

## 11. 后续扩展

- Gateway HTTP 层 E2E：k6 脚本打 httptest.Server，验证 rate limit + precheck + proxy 全链路
- Redis CombinedKeyCache 一致性验证：PG 写入后断言 Redis 缓存同步
- 大规模场景：100 公司 × 50 成员 × 100 并发（nightly only）
- 网络故障注入：mock DisablePlatformKey 返回 error，验证重试逻辑
- Overrun job 延迟 SLO：从 ingest 完成到 key disable 的 P99 < 500ms