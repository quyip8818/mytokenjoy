# 集成压测：多公司多用户同步请求 — 超额、通知、封锁

## 0. 架构决策

### 为什么选 Integration Test

| 考量 | 结论 |
|------|------|
| 超额/封锁依赖完整 pipeline（Ingest → Ledger → BudgetConsumed → Overrun Job → DisableKey） | 需要跨 domain 集成验证 |
| 现有 unit test 已覆盖各 domain 内部正确性 | integration 补的是端到端 + 并发场景 |
| CI 日常不需跑（耗时长、依赖 PG） | 通过 build tag 隔离 |
| 一个 test function 覆盖多场景（table-driven subtest） | 减少 fixture 重建开销 |

### Build Tag 隔离

```go
//go:build testhook && integration
```

- 日常 `make test`（`-tags=testhook`）**完全跳过**
- 只有 `make test-integration` 才运行

### 文件位置

```
apps/backend/tests/integration/stress/
    helpers_test.go               ← 环境构建、并发工具、断言
    overrun_blocking_test.go      ← 主测试（subtests + benchmarks）
```

---

## 1. 测试目标

| 机制 | 期望行为 |
|------|----------|
| 预算耗尽检测 | `BudgetExhausted` 在各轴正确触发 |
| Key 封锁 | 超额后 Key 被 `DisablePlatformKey`，Gateway precheck 拒绝 |
| 通知发送 | `overrun_blocked`、`budget_alert_reached`、`overdraft_expanded` 正确投递 |
| 幂等性 | 同一 consume log 重复 ingest 不产生重复扣减或多次封锁 |
| 并发安全 | 多 goroutine ingest 同一 Key 不导致数据错乱 |

## 2. 测试环境

- 模型：`local-test-model`（`DevCallTypeLocalTest`，modelID=1）
- DEPLOY_ENV=`local`
- 数据库：PostgreSQL（testutil 标准 test store，隔离 schema）
- Job Runner：River（嵌入式，手动 `RunOnce` 驱动）
- NewAPI 后端：`StubAdminClient` mock
- 通知：`recordingNotifier`（内存捕获，无 IO）

## 3. 已实现场景

### 3.1 KeyBudgetExhaustion — Key 级封锁 ✅

对应文档场景 A。

```
PK1 budget=100
→ Ingest → consumed=95
→ Ingest → consumed > 100
→ RunOnce (overrun job)
→ PK1 disabled
→ Notification: overrun_blocked, scope=platformKey
```

### 3.2 MemberBudgetIsolation — 成员级封锁 ✅

对应文档场景 B（部分）。

```
M1 personalBudget=50, consumed=48
→ Ingest → consumed > 50
→ RunOnce → M1 keys disabled
→ Notification: scope=member
```

### 3.3 ConcurrentIngestRace — 并发竞态安全 ✅

对应文档场景 D。

```
PK1 budget=100, consumed=90
→ 10 goroutine 各 ingest ~5 points
→ RunOnce → PK1 disabled
→ consumed > 100, notification ≥ 1
→ go test -race 无告警
```

### 3.4 DeptAlertThresholds_NoBlock — 部门级仅通知 ✅

对应文档场景 E。

```
D1 budget=1 (极低), AlertRule thresholds=[50,80,100]
→ Ingest → overrun_blocked(notifyOnly=true)
→ Key 仍 active
```

### 3.5 IdempotentReplay — 幂等重放 ✅

对应文档场景 J。

```
→ Ingest logID → PK1 disabled
→ 再次 Ingest 相同 logID
→ consumed 不增，通知不重发
```

### 3.6 GatewayBlocksAfterOverrun — 端到端链路 ✅

覆盖 Gateway precheck 层。

```
→ Precheck PASS (before overrun)
→ Ingest + overrun
→ Precheck FAIL: ErrBudgetExhausted
```

### 3.7 Benchmark: Ingest Throughput ✅

输出单公司串行 QPS 基线。

### 3.8 Benchmark: Gateway Precheck ✅

输出 precheck QPS（含 PG 查询）。

### 3.9 Benchmark: Concurrent Ingest ✅

20 goroutine 并发 ingest，输出有效 QPS。

## 4. 未实现场景（待定）

| 场景 | 原因 | 复杂度 | 建议 |
|------|------|--------|------|
| **C: 多公司隔离** | 需要创建多个独立公司 fixture（各自 schema） | 高 | 价值高但需要重构 `buildStressEnv` 支持多公司 |
| **F: 项目预算封锁** | 需要 project fixture + scope=project key | 中 | 现有 unit test `TestOverrunProjectAxis` 已覆盖 |
| **G: 项目内成员封锁** | 需要 project_member scope key | 中 | 现有 unit test 已覆盖 |
| **H: Wallet 余额不足** | Gateway evaluate 层逻辑，与 ingest 无关 | 低 | `TestGatewayRejectsInsufficientWallet` 可单独加 |
| **I: Overdraft 通知** | 需要 lot 接近耗尽的精确 fixture | 中 | 现有 `ingest_overdraft_test.go` 已覆盖 |

### 决策建议

- **F/G/I**：现有 unit test 已有覆盖，integration 重复验证 ROI 低，**建议不做**
- **C（多公司隔离）**：这是 integration test 唯一不可替代的价值点，**建议做**
- **H**：简单加一个 gateway evaluate 测试即可，**建议做**

## 5. 运行方式

```bash
# 运行集成压测（需要 DATABASE_URL + PG）
cd apps/backend
make test-integration

# 单独场景
go test -tags="testhook,integration" -race \
  -run TestOverrunBlockingIntegration/KeyBudgetExhaustion \
  ./tests/integration/stress/...

# Benchmarks
go test -tags="testhook,integration" -race \
  -run TestBenchmark ./tests/integration/stress/...
```

## 6. QPS 瓶颈分析

### 系统瓶颈

| 路径 | 瓶颈 | 理论上限 |
|------|------|----------|
| Gateway Precheck（读） | PG pool 50 conn, PrecheckCache 内存命中 | 2000–10000+ QPS/实例 |
| Ingest Pipeline（写） | `SELECT ... FOR UPDATE` per-company 串行 | 80–150 QPS/公司 |
| Rate Limiter | `RATE_LIMIT_V1_RATE=30` per key | 30 QPS/key |

### Benchmark 阈值（宽松，防 CI flaky）

| 指标 | 下限 | 说明 |
|------|------|------|
| Ingest QPS（串行） | ≥ 10 | 单公司，含 PG TX + River enqueue |
| Gateway Precheck QPS | ≥ 100 | 含 PG 查询（无 Redis） |
| Concurrent Ingest 成功率 | ≥ 50% | 20 goroutine，锁竞争下容许部分失败 |

## 7. 验证清单

### 正确性 ✅

- [x] 超额后 Key disabled
- [x] Gateway precheck 对 disabled Key 返回错误
- [x] 通知 payload 含 scope 字段
- [x] 幂等：相同 logID 不重复扣减
- [ ] Alert dedup（待 C 场景补充）

### 并发安全 ✅

- [x] `go test -race` 无告警（`-race` flag 已配置）
- [x] `budget_consumed` 最终值合理
- [x] 无 deadlock（120s timeout）

### 隔离性

- [ ] 多公司隔离（场景 C — 待实现）

## 8. 后续扩展

- **场景 C**：多公司隔离测试 — 修改 `buildStressEnv` 支持 N 个独立公司
- **场景 H**：Wallet 余额不足 Gateway 拒绝 — 简单 evaluate 层测试
- Gateway HTTP 层 E2E：k6 / vegeta 打 httptest.Server
- Redis CombinedKeyCache 一致性验证
- Overrun job 延迟 SLO：ingest → key disable P99 < 500ms
