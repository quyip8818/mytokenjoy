# Backend 测试优化

PostgreSQL 集成测试隔离与 Seed 选用指南。

**相关：** [Backend-seed.md](./Backend-seed.md) · [Backend.md](./Backend.md) · [Backend-架构.md](./Backend-架构.md)

---

## 1. 测试 Store 隔离

| 机制           | 说明                                                                             |
| -------------- | -------------------------------------------------------------------------------- |
| **实现**       | `tests/testutil.NewTestStore` / `NewTestApp` → `postgres.New`                    |
| **Schema**     | 每测独立 schema（`openTestSchema`），并行安全                                    |
| **Build tag**  | `-tags=testhook` 暴露 `postgres.MainPool` 等测试钩子                             |
| **运行态清理** | 测后 `TRUNCATE usage_buckets, company_recharge_orders`（`clearDemoRuntimeSeed`） |

前置：`pnpm start:postgres` 或本地 `DATABASE_URL` 可用。

```bash
cd apps/backend && make test-unit
```

---

## 2. Seed 选用表

| 场景                    | 用法                                          | 成员规模        | 说明                                                           |
| ----------------------- | --------------------------------------------- | --------------- | -------------------------------------------------------------- |
| **默认集成测试**        | `NewTestStore(t)`                             | ~41 人          | 全量管理面 seed；运行态表测后清空                              |
| **新测仅需锚点**        | `NewTestStore(t, testutil.WithMinimalSeed())` | 8 人（锚点）    | 仅 contract 闭包：锚点成员、`plk-1`、单预算组、无 usage_ledger |
| **场景化数据**          | `orgfix` / `relayfix` / `budgetfix`           | 按需            | **优先**于改 seed；见各 `tests/testutil/*`                     |
| **Demo / `pnpm start`** | `postgres.New` 默认                           | ~41 人 + 运行态 | `IsDemoProfile()` 时写入 usage_buckets、充值单                 |
| **E2E**                 | 全量 demo 库                                  | 同上            | 与本地 `pnpm start` 一致                                       |
| **仅要契约 ID**         | `import seed/contract`                        | 无 DB           | 轻量引用 `IDDept3`、`DemoPassword` 等                          |

### 2.1 API 对照

| 需求        | 包 / 符号                                                |
| ----------- | -------------------------------------------------------- |
| 锚点 ID     | `github.com/tokenjoy/backend/seed/contract`              |
| 全量快照    | `seed.Load(cfg)`                                         |
| 最小快照    | `seed.LoadMinimal(cfg)`                                  |
| 管理面落库  | `postgres.ApplySeedTables`                               |
| Demo 运行态 | `seed/runtime.ApplyUsageBuckets` / `ApplyRechargeOrders` |

### 2.2 何时用 MinimalSeed

适合：纯 CRUD、鉴权、单成员/单 Key 逻辑，不依赖分页列表或大量 filler 成员。

不适合：组织分页、看板序列、预算组多 Key 配额、审计列表丰富度——继续用默认全量 seed。

---

## 3. Ingest 测试约束

| 项         | 约定                                                    |
| ---------- | ------------------------------------------------------- |
| 消费日志   | `testutil.SeedConsumeLog` + `WithIngestEnabled(true)`   |
| Relay 映射 | `relayfix.UpsertMapping` 对齐 `contract.IDPlatformKey1` |
| 日志库     | `LOG_DATABASE_URL` 与主库同 schema（测试默认）          |

---

## 4. 与 Fixture 边界

- **Seed**：空库一次性引导，demo 与测试共享（见 [Backend-seed.md](./Backend-seed.md)）
- **Fixture**：`testutil/*fix` 在已 seed 库上追加场景数据
- **禁止**：在 `tests/` 定义稳定锚点 ID；统一使用 `seed/contract`
