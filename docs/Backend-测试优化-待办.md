# Backend 测试优化待办

> **对齐**：2026-07-10  
> **定位**：测试效率与稳定性；以当前磁盘代码为准，不以历史计划文档为准。  
> **近期变更**：`CLOCK_ANCHOR` + `Config.Clock()`；开账/发生双轨 period（见 [Backend-业务时钟与账期.md](./Backend-业务时钟与账期.md)）。  
> **相关**：日常 backlog 见 [plan.md](./plan.md)；配置契约见 [Backend-配置架构.md](./Backend-配置架构.md)。

---

## 1. 当前架构（以代码为准）

```text
TestConfig(opts...)
  CLOCK_ANCHOR=2026-06-19（测试默认；生产禁止）
        │
NewTestStore
  ├─ 非 minimal：clone test_template → SchemaPrepared=true → postgres.New（跳过 DDL/seed）
  │              → resetRuntimeTables（TRUNCATE recharge/buckets/ledger）
  └─ minimal：空 schema → 全量 DDL + minimal seed
        │
可选：NewTestStoreWithDemoRuntime / applyDemoRuntime
  → seed/runtime.ApplyDemo（buckets + recharge + ledger）
        │
NewTestApp / NewTestRouter
  → 当前实现：建 store 后**总会** applyDemoRuntime（见缺口）
```

| 路径 | 职责 |
| --- | --- |
| `tests/testutil/pg/template.go` + `clone.go` | `test_template` + 每测 `test_<hex>` 克隆 |
| `tests/testutil/store.go` | `NewTestStore` / `PreparedConfig` |
| `tests/testutil/runtime_seed.go` | `resetRuntimeTables` / `applyDemoRuntime` / `NewTestStoreWithDemoRuntime` |
| `tests/testutil/app.go` | HTTP/app fixture |
| `internal/config` | `Clock()` / production 禁止 `CLOCK_ANCHOR` |

**时钟约定（重要）**：budget snapshot / precheck / ingest / overrun 等账期路径统一走 `cfg.Clock()`（见 [Backend-业务时钟与账期.md](./Backend-业务时钟与账期.md)）。  
**不要再引入** `SnapshotAnchor` / `DemoToday` / 进程级全局锚点；demo 可复现日期只通过 **`CLOCK_ANCHOR` → `Config.Clock()`**。

---

## 2. 已完成且仍在

| 项 | 说明 |
| --- | --- |
| Schema 模板克隆 | `test_template` + clone；template version；orphan cleanup（keep 32） |
| 静态 / runtime 分离 | 克隆后 `resetRuntimeTables`；需 usage 数据时 `NewTestStoreWithDemoRuntime` |
| `PreparedConfig` API | 跳过已克隆 schema 上的 DDL/seed（调用点仍不全） |
| 并行与分区 | `TEST_PARALLEL=8`、`TestPartitionMonths=12` |
| 快速子集 | `make test-fast`（仅 `tests/pkg`） |
| DB 清理 | `make test-db-clean` / `cmd/testdbclean` |
| 时钟统一 | `CLOCK_ANCHOR` + `Config.Clock()`（替代已废弃的 demo 全局锚点思路） |

来源：`93b3f12`（测试效率）+ 后续 env/clock 重构。

---

## 3. 文档曾宣称完成、实际未站住（当前缺口）

env 重构（`eb8da61`）与后续改动后，待办里旧的「已完成」断言过期：

| 旧断言 | 现状 |
| --- | --- |
| `NewTestApp` 默认静态 seed、runtime opt-in | **否**：`app.go` 无条件调用 `applyDemoRuntime` |
| Audit 已改用 `NewTestStore` | **否**：仍多用 `NewTestStoreWithDemoRuntime` |
| `tables_test` / `persist_test` 已用 `PreparedConfig` | **否**：克隆 schema 上仍可能二次 DDL/seed |

这三项是**下一步优先落地**的内容，不是「暂缓」。

---

## 4. 下一步（效率 + 稳定性，避免过度工程）

按投入产出排序；**不做** Gateway 大重构、不调 `TEST_PARALLEL`、不上 SnapshotAnchor。

### P0 — 效率（约半天）

1. **`NewTestApp` runtime 改回 opt-in**  
   - 文件：`tests/testutil/app.go`  
   - 去掉无条件 `applyDemoRuntime`；默认只靠 clone + 静态 seed。  
   - 需要 ledger/buckets/recharge 的用例显式 `NewTestStoreWithDemoRuntime`，或通过小 helper（例如 `NewTestAppWithDemoRuntime`）注入。  
   - **收益**：HTTP fixture（约数十次 `NewRouter`/`NewApp`）每测少灌一轮 demo runtime；墙钟收益最大。

2. **`PreparedConfig` 补齐调用点**  
   - `tests/seed/tables_test.go`、`tests/store/postgres/persist_test.go`（及相关 `postgres.New` + 已克隆 URL 的路径）。  
   - **收益**：消除克隆后重复 DDL 的超时/假慢。

3. **Audit helper 拆分**  
   - `ListOperations*` → 静态 `NewTestStore`（读 `operation_logs`）。  
   - `ListCalls*` → 保留 `NewTestStoreWithDemoRuntime`。  
   - **收益**：audit 包每测少灌 ledger。

### P1 — 稳定性（与性能正交，建议单独修）

| 项 | 说明 |
| --- | --- |
| `TestWorkerProcessesOverrunQueue` | overrun → disable key 路径不稳定；对齐策略阈值 / enqueue / mapping |
| 并行下偶发 FK / deadlock | 如 `relay_mappings` 外键、`40P01`；优先保证 fixture 绑定 seed 内已有 `platform_key`；重现时单独 `-run` |
| 前端 `auth-session-provider` 失败 | 非 backend 性能专项 |

全绿目标：先单包绿，再 `make test-unit-nocache`；并行 flake 勿用「再 seed 一遍」糊弄。

### P2 — 可选小项（收益小）

| 项 | 说明 |
| --- | --- |
| `tests/config`、`tests/pkg/org/org_nodes_test` 补 `t.Parallel()` | 纯测，省秒级墙钟 |
| 文档同步 | `Backend.md` 仍写 `-parallel 4`（Makefile 已是 8）应改正 |

---

## 5. 明确不做

| 项 | 原因 |
| --- | --- |
| `SnapshotAnchor` / `DemoToday` / 全局时间锚点 | 污染 prod；已由 `CLOCK_ANCHOR` + `Clock()` 覆盖 |
| Gateway 轻量 fixture 体系 | 改动面大，ROI 低 |
| SaaS HTTP 全量下沉 domain | 大 refactor |
| 盲目抬 `TEST_PARALLEL` / `-p` | 易压垮 PG，先修 P0 |
| `time.Sleep(10ms)` → injectable clock | 总墙钟可忽略 |
| Advisory lock 改 try_lock | 当前够用 |
| 前端 Vitest 瘦配置 / 去 tsc / 拆 Provider | 前端全量 ~15s，边际收益低 |

---

## 6. 验证命令

```bash
# 本地 Postgres
pnpm start:postgres

cd apps/backend
make test-db-clean              # 孤儿 schema 过多时
make test-unit-nocache          # 全量、无缓存

# 抽查（P0 落地后）
go test -tags=testhook -count=1 ./tests/seed/... -v
go test -tags=testhook -count=1 ./tests/domain/audit/... -v
go test -tags=testhook -count=1 ./tests/handler/core/... -v
```

**验收标准（P0）**：

- `NewTestApp` / `NewTestRouter` 路径不再默认灌 ledger；依赖 usage 的测仍绿。  
- `TestApplyTablesMatchesSnapshot` / persist 重启测秒级完成、无超时。  
- Audit 包中仅 call-log 测走 DemoRuntime。  
- `make test-unit-nocache` 稳定时间回到 clone 架构后的量级（约数分钟级），且不以注入 demo 时钟污染生产路径。
