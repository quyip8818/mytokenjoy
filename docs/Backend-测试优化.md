# Backend 测试套件优化分析

> **双目标：** ① 提升 **有效 coverage**（主路径 + 架构契约）；② 提升 **test 速度**（降低 schema clone、分层 CI）。  
> **基线（2026-07 初）：** 615 个 `Test*`、全量墙钟 ~115s。  
> **本 PR 后（2026-07-12 实测）：** **617** `Test*`、墙钟 **~101s**；`make test-fast` **无 PG**。

相关约定见 [Backend.md §5](./Backend.md#5-测试与-seed)。Gateway 拒绝场景 SSOT 见 [Backend-结构优化.md §2.1](./Backend-结构优化.md#21-gateway-rejection_cases-ssot)（`rejection_cases`，可独立 PR）。

---

## 0. 架构约束（不可妥协）

1. **`call_chain_test.go`** — PRD 10.3 校验**顺序**（invalid key 先于 budget），不得并入 rejection matrix。  
2. **Gateway evaluate ↔ precheck** — 可共享 **case 数据**；precheck 每 case 仍须独立 store（污染 case 不可共享 fixture）。  
3. **Handler 删镜像测** — 须先新增 `mutating_contract_test.go`。✅ 已完成。  
4. **跨层重复判定** — HTTP 状态码 / job state / SQL roundtrip 与 domain 业务断言是**不同维度**。

### SSOT 一览

| 维度 | 位置 |
| --- | --- |
| 业务规则 | `tests/domain/<域>/` |
| GET API 形状 | `tests/handler/core/contract_test.go` |
| 写 API smoke | `tests/handler/core/mutating_contract_test.go` |
| Middleware 行为 | `tests/http/middleware/`（`stubs_test.go` + `middleware_test.go`） |
| NewAPISync outbox | `tests/domain/newapisync/outbox_*.go` |
| Gateway 场景 | `tests/testutil/gateway/` |
| 纯函数 | `tests/pkg/`（无 PG） |
| Keys 写 403 | `tests/handler/authz/authz_cases_test.go` |

---

## 1. 现状（本 PR 后）

| 指标 | 数值 |
| --- | ---: |
| `Test*` | **617** |
| 测试包 | **47** |
| 全量墙钟 | **~101s**（`-p 2`） |
| `test-fast` | `./tests/pkg/...`，**无 Postgres** |

---

## 2. 本 PR 变更摘要

### Track S — 速度

| 动作 | 状态 |
| --- | --- |
| `clock_align_test.go` 迁出 `tests/pkg/` → `tests/domain/budget/` | ✅ |
| 删跨层重复测（overrun / list recharge / suspended gateway / reconcile logs） | ✅ |
| `tests/pkg/newapiunits/quota_test.go` SSOT；删 `integration/newapi/quota.go` 薄委托 | ✅ |
| `FloatPtr` → `testutil/budget/ptr.go` | ✅ |

### Track C — Coverage

| 动作 | 状态 |
| --- | --- |
| middleware M1–M7（`TestMiddlewareBehaviors`） | ✅ |
| newapisync outbox N1–N3 | ✅ |
| mutating contract H1–H2；删 `handler/budget/budget_test.go` | ✅ |

---

## 3. 重复 / 待做（PR3 可选）

Gateway precheck / evaluate / handler 对同一拒绝场景有**重复 case 数据**（非功能盲区）。PR3 抽 `tests/testutil/gateway/rejection_cases.go` + handler G2/G3 smoke；与 [Backend-结构优化.md §2.1](./Backend-结构优化.md#21-gateway-rejection_cases-ssot) 同一事项，**可独立 PR**。**完整 PR3 优先级低**（见 team 讨论）。

| 场景 | evaluate | precheck | handler HTTP |
| --- | --- | --- | --- |
| insufficient wallet | ✅ | ✅ | ✅ |
| suspended company | ✅ | ✅ | ❌ |
| model not allowlist | ✅ | ✅ | ❌ |

---

## 4. Coverage 清单

### 4.1 P0 — 本 PR 已关闭

#### A. `http/middleware`

文件：[`tests/http/middleware/`](../apps/backend/tests/http/middleware/) — chi + stub 栈（**不要** `NewApp`）；stub 在 `stubs_test.go`；M7 单独 `NewTestStore`。

| # | Case | Clone |
| --- | --- | ---: |
| M1 | `company_resolve` — 无 tenant → 400 | 0 |
| M2 | platform 路由 skip | 0 |
| M3 | `platform_auth` — 无 token → 401 | 0 |
| M4 | login bypass | 0 |
| M5 | suspended 禁写 | 0 |
| M6 | authz revision header | 0~1 |
| M7 | 篡改 JWT → 401 | 1 |

#### B. `domain/newapisync` outbox

| # | Case | 文件 | Clone |
| --- | --- | --- | ---: |
| N1 | payload JSON roundtrip | `outbox_payloads_test.go` | 0 |
| N2 | `IsPermanentOutboxError` | `outbox_errors_test.go` | 0 |
| N3 | rebalance remain_quota 封顶 | `lifecycle_rebalance_test.go` | 1 |
| N4 | worker 路由（可选） | 扩 `worker/runner_test.go` | 1 |

#### C. `handler/core/mutating_contract_test.go`

| # | Case | 说明 |
| --- | --- | --- |
| H1 | PUT `/api/budget/departments/dept-3` | 替代已删 `budget_test.go` |
| H2 | PUT approval reject（`appr-2`） | 写操作 smoke |
| H3 | POST keys 403 | **不新增** — SSOT 在 `authz_cases_test.go` |

#### D. Gateway handler（PR3 可选）

G2 model not allowed、G3 suspended company — handler 403 映射 smoke；domain 已覆盖业务拒绝。

### 4.2 P1 — 后续

company grants 落库链、memberanalytics 聚合、alert 通知、handler/keys 分页、`internal/infra/permission/*_test.go` 迁入 `tests/` 或白名单。

---

## 5. 速度优化

### 5.1 瓶颈

墙钟 ≈ (CloneSchema 次数 × ~0.34s) + 断言 CPU + PG 争用。

### 5.2-A 本地开发命令

```bash
# 纯 pkg（无 Postgres）
make test-fast

# middleware / outbox 纯单测
go test -tags=testhook ./tests/http/middleware/... ./tests/domain/newapisync/... \
  -run 'TestMiddleware|TestOutbox|TestIsPermanent' -count=1

# 改单域
go test -tags=testhook ./tests/domain/gateway/... -count=1

# handler 契约
go test -tags=testhook ./tests/handler/core/... -run 'Contract|Mutating' -count=1

# 提交前 / CI
make test-unit
```

### 5.2-B Subtest 共享 Store 边界

**可共享：** 只读 GET（`contract_test`）、evaluate 全表（无 store）、纯函数 table。  
**不可共享：** 会改 wallet / key status / company status 的 case；不同 seed opts 的 Gateway rejection。

### 5.2-C CI 分层

| 阶段 | 命令 | 触发 |
| --- | --- | --- |
| Fast gate | `make test-fast` | push / 本地 |
| Full gate | `make test-unit` | PR / main |
| Integration | `pnpm verify:gate` | nightly / release |

---

## 6. 验收标准

| 指标 | 实测 | 验证 |
| --- | ---: | --- |
| `Test*` | **617** | `rg '^func Test'` |
| P0 盲区 | **0** | §4.1 ✅ |
| 全量墙钟 | **~101s** | `time make test-unit-nocache` |
| `test-fast` | 无 PG ✅ | `make test-fast` |
| PRD 10.3 | 保留 | `call_chain_test.go` |

---

## 7. 不建议

| 做法 | 原因 |
| --- | --- |
| 合并 call_chain 进 rejection matrix | 丢失顺序契约 |
| 1 共享 store 跑全部 Gateway 拒绝 case | 状态污染 |
| 未补 mutating 就删 budget handler 测 | PUT 盲区 |
| 追求 90% line coverage | 行为优先 |

---

## 10. 附录

**已完成：** 删重复测、`budget_test.go`、`integration/newapi/quota.go`、`tests/pkg/newapi/`；迁 `clock_align`、`quota_test`、`FloatPtr`；新增 middleware / outbox / mutating。

**PR3 待做（可选）：** `rejection_cases.go`、handler G2/G3 — 详见 **§12**；checklist：[plan.md §9](./plan.md#9-backend-测试pr3)。

| PR | 范围 | 状态 |
| --- | --- | --- |
| **PR1** Track S | test-fast 边界、删重复、newapiunits SSOT | ✅ |
| **PR2** Track C | middleware、outbox N1–N3、mutating H1–H2 | ✅ |
| **PR3** | Gateway rejection_cases + G2/G3（可选） | 待做 |

---

## 12. PR3 实施规格

> 代码实现的唯一技术说明；[plan.md §9](./plan.md#9-backend-测试pr3) 仅 checklist。

### 12.1 约束

- **可做：** evaluate 与 precheck 共享 case **数据**（name + mutator + `GatewayScenarioOpts`）
- **不可做：** 1 个 shared store 跑 wallet=0 / suspended / disabled key
- **不可做：** 合并 `call_chain_test.go`

### 12.2 `tests/testutil/gateway/rejection_cases.go`

```go
type RejectionCase struct {
    Name      string
    MutatePC  func(*domaingateway.PrecheckContext)
    Scenario  GatewayScenarioOpts
    Model     string // default "gpt-4o"
    WantHTTP  int    // handler；0 = 不测
}
```

初始 case：insufficient_wallet、suspended_company、model_not_allowlist、inactive_key、exhausted_soft_remain（仅 evaluate）。

### 12.3 重构

| 文件 | 动作 |
| --- | --- |
| `evaluate_test.go` | `TestEvaluateRejects` 遍历 `RejectionCases()` |
| `precheck_test.go` | 同一 table 子集；推荐去 subtest `Parallel` |
| `handler/gateway/gateway_test.go` | `TestGatewayRejectionHTTPMapping`：G2/G3 各独立 `BuildGatewayScenario` |

### 12.4 Handler G2 / G3

| Case | 期望 |
| --- | --- |
| G2 model not allowed | **403** |
| G3 suspended company | **403** |

G1 已有 `TestGatewayRejectsInsufficientWallet`，可并入同一 `Test*` 的 subtest（可选）。

### 12.5 验收

| 指标 | 目标 |
| --- | --- |
| `Test*` | 617 → ~622–628 |
| 墙钟 | ≤101s（净增 clone ≤3） |
| SSOT | rejection case 数据只在 `rejection_cases.go` |
