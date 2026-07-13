# Seed（`BOOTSTRAP_MODE=demo` / `minimal`）

Demo / minimal 引导数据，**不**包含生产运行时逻辑。空库按 `BOOTSTRAP_MODE` 写入；非空库不覆盖。

| 目录 | 职责 |
|------|------|
| `contract/` | 固定 ID、模型 catalog 映射（SSOT） |
| `snapshot/` | 预算树、模型、组织路由等快照 |
| `data/` | JSON（platform keys、usage ledger 等） |
| `apply/` | 写入 Postgres（`ApplyTables`） |
| `runtime/` | 启动后追加（usage、充值 lot 等） |

**与生产的边界：** Seed 只写 TokenJoy Postgres。不写 `platform_key_mappings`、不创建 NewAPI Token/Channel、不配置 NewAPI `UserUsableGroups`。

**已知问题与修复：** [FIXES.md](./FIXES.md)
