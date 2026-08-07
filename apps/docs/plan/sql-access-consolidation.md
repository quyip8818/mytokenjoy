# SQL 访问收口 Plan

## 问题

Raw SQL 散布在多个层，schema 变更时容易漏改：

| 位置 | 例子 | 问题 |
|------|------|------|
| `store/postgres/*.go` | 所有 repo 实现 | ✅ 正规层，无问题 |
| `seed/apply/*.go` | 初始化 demo 数据 | ⚠️ 直接写 SQL，但属于独立模块，可接受 |
| `seed/bootstrap/*.go` | 首次启动必需数据 | ⚠️ 同上 |
| `app/setup_server.go` | boot sync + admin 创建 | ❌ 绕过 repo，用 raw pgxpool 直接写表 |

这次 currencies 改造就踩了坑：`setup_server.go` 里的 `syncCurrenciesFromSaaS` 用了 `ON CONFLICT (currency)` 而其他地方都改了 `ON CONFLICT (id)`，导致运行时报错。

---

## 目标

1. **所有业务表写操作只通过 `store/` repo 层**（single source of SQL）
2. seed 层例外允许 raw SQL（它本身是 DB 初始化工具，和 repo 同级别）
3. `app/` 层不再直接写业务表

---

## 方案

### Step 1：`setup_server.go` 的 SQL 收口到 repo

当前 `setup_server.go` 用 `*pgxpool.Pool` 直接执行 SQL 是因为 setup 阶段 store 实例还没完全构建。

**改法：** 让 `RunSetupServer` 接收一个 `store.Store`（或至少让 boot sync 函数接收 repo 接口）。

```go
// Before:
func syncCurrenciesFromSaaS(ctx context.Context, pool *pgxpool.Pool, cfg config.Config) error {
    pool.Exec(ctx, `INSERT INTO currencies ...`)
}

// After:
func syncCurrenciesFromSaaS(ctx context.Context, st store.Store, cfg config.Config) error {
    // 解析 HTTP 响应后调用 repo 方法
    st.Billing().InsertCurrencyFromSync(ctx, row)
}
```

涉及的函数：
- `syncCurrenciesFromSaaS` → 改调 `st.Billing().InsertCurrencyFromSync`
- `createAdminUserTx` → 改调 `st.Users().Create` + `st.Members().Create` + `st.MemberRoles().Assign`（这些 repo 方法已存在）

**前置条件：** `RunSetupServer` 的调用方需要在 setup 前就构建 store 实例。看 `compose_*.go` 确认可行性——store 在 pool 创建后立即可构建，不依赖 setup 结果。

### Step 2：加 lint 规则防止回归

在 `.kiro/steering/` 加规则：

```
禁止在 internal/app/ 层直接写 INSERT/UPDATE/DELETE SQL。
业务表写操作必须通过 store/ repo 层。
seed/ 层例外（它是 DB 初始化工具）。
```

### Step 3：seed 层保持现状

seed 的 `TableWriter` 接口（`Exec(sql, args)`）是有意设计——seed 需要写多种表且只在启动时跑一次。把 seed 改成调 repo 方法会引入大量 store 依赖且收益不大。

seed 的 SQL 风险可接受：
- 跑在应用启动最前面，报错立即 fatal
- 表结构变了 seed 也要同步改（这是正常维护）

---

## 影响文件

| 文件 | 改动 |
|------|------|
| `app/setup_server.go` | `syncCurrenciesFromSaaS` 改调 repo；`createAdminUserTx` 改调 repo |
| `app/compose_*.go` | 调整 `RunSetupServer` 签名，传入 store |
| `.kiro/steering/` | 加 SQL 访问规则 |

---

## 工作量

半天。核心是 `setup_server.go` 两个函数改调 repo（4-5 个方法调用替代 raw SQL），加调整 compose 层传参。

---

## 不做的事

- ❌ 不改 seed 层（保持 raw SQL，属于 DB 工具）
- ❌ 不引入 ORM/code gen（当前 repo 模式足够清晰）
- ❌ 不引入 SQL 文件分离（如 `.sql` 模板文件），Go 内联 SQL 在 repo 层足够好
