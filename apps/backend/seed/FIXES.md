# Seed 模块 — 修复清单

> 范围：**仅** `BOOTSTRAP_MODE=demo`（及依赖 demo 契约的测试）。与生产租户数据无关的项放此文件；NewAPI / Gateway 集成见仓库根 `docs/`。

---

## 边界（避免误归因）

| 数据 | Seed 是否写入 | 实际负责方 |
|------|---------------|------------|
| 公司 / 组织 / 模型 catalog / platform_keys 行 | ✅ `apply/` | — |
| `platform_key_mappings` | ❌ | `newapisync.BootstrapUnsyncedPlatformKeys` + Worker |
| NewAPI Token `group` / `model_limits` | ❌ | `newapisync.TrySyncCreate` |
| `companies.newapi_wallet_user_id`（demo 公司 id=2） | ❌ | SaaS `CreateCompany`；demo 见 [FIX-SEED-004](#fix-seed-004) |
| NewAPI `UserUsableGroups` | ❌ | 集成层 / 本地脚本，见 [Backend-NewAPI集成修复.md](../../../docs/Backend-NewAPI集成修复.md) |

---

## 模型 ID（SSOT）

定义：`contract/ids.go`、`contract.ModelTypeToID`。

| model_id | type | 说明 |
|----------|------|------|
| **1** | `local-test-model` | 仅 dev / demo catalog |
| **100–107** | `gpt-4o` … `qwen-plus` | demo 生产模型 |
| **≥ 100** | 客户自建 | schema `IDENTITY START WITH 100`（非 seed） |

客户新建模型不走 seed；见 schema，不在此维护。

---

## FIX-SEED-001 — `plk-5` 有效白名单为空

| | |
|--|--|
| **现象** | Sync 后 NewAPI `model_limits` 异常；业务上无法选模型 |
| **根因** | `data/platform_keys.json`：`plk-5` 白名单 `[100,103]`；`proj-4` 归属 `dept-5`（`snapshot/budget.go`）；`dept-5` 路由仅 `[101,104]`（`snapshot/org_node.go`）→ `EffectiveWhitelistIDs` 交集 **∅** |
| **修复** | 将 `plk-5.modelWhitelist` 改为与 dept-5 有交集，例如 `[101, 104]` |
| **文件** | `data/platform_keys.json` |

---

## FIX-SEED-002 — 旧库 model_id 与契约不一致

| | |
|--|--|
| **现象** | Gateway `model not allowed`；白名单数字与 catalog 对不上 |
| **根因** | 历史 demo：`local-test-model` 曾为 id **9**，现为 **1**；生产模型曾为 1–8，现为 **100–107** |
| **修复** | `pnpm docker:reset` 或清空库后 `BOOTSTRAP_MODE=demo` 重新引导 |
| **验证** | `tests/seed/contract/`；catalog 中 `local-test-model` 的 `model_id = 1` |

---

## FIX-SEED-003 — 改 seed 后 NewAPI token 未更新

| | |
|--|--|
| **现象** | Postgres 白名单已含 model **1**，NewAPI 仍 `no access to model local-test-model` |
| **根因** | Seed 只改 TokenJoy 行；`model_limits` 在 **sync 时**计算，已 synced 的 mapping 不会自动重算 |
| **修复（运维）** | 对相关 Key **Rotate** 或 disable → enable，触发 `SyncUpdatePlatformKey` |
| **修复（工程）** | 见集成文档 FIX-INT-003 |

---

## FIX-SEED-004 — Demo 公司无 `newapi_wallet_user_id`

| | |
|--|--|
| **现象** | Platform Token 挂在 NewAPI `user_id=1`（root）；钱包 sync / quota cap 与 SaaS 开户不一致 |
| **根因** | `apply/seed_core.go` 的 `insertSeedCompany` 不写 `newapi_wallet_user_id` |
| **是否导致分组错误** | **否** |
| **修复选项** | A) 接受 demo 用 root 钱包（文档说明） B) `runtime` 或 local-only bootstrap 为 company 2 创建 NewAPI 用户并 `UpdateNewAPIWalletUserID` |

---

## Demo Platform Key 对齐表

部门解析与线上一致：`departmentIDForPlatformKey`（member 部门优先，否则 project `OwnerDepartmentID`）。

| Key | 部门 | 白名单 ID | 有效 ID（∩部门路由） | 备注 |
|-----|------|-----------|----------------------|------|
| plk-1 | dept-3 | 100,103,1 | 100,103,1 | 模拟消耗推荐 |
| plk-2 | dept-3 | 100,104,1 | 100,104,1 | |
| plk-1b | dept-3 | 101,104 | 101,104 | 无 model 1 |
| plk-bg-1 | dept-3 | 100,103 | 100,103 | |
| plk-5 | dept-5 | 103,100 | **∅** | **FIX-SEED-001** |
| plk-4 | dept-4 | — | — | disabled，不 bootstrap |

部门路由摘要（`snapshot/org_node.go`）：

| 部门 | 允许 model_id |
|------|----------------|
| dept-3 | 100, 103, 104, 1 |
| dept-5 | 101, 104 |
| dept-4 | 101, 103, 104 |

NewAPI `group` 由集成层计算（`dept-` + 部门 id），**不是** seed 字段；见 [Backend-NewAPI集成修复.md](../../../docs/Backend-NewAPI集成修复.md)。

---

## 变更检查

修改 seed 后：

1. 跑 `tests/seed/contract/`
2. 若改 `platform_keys.json` 或 `org_node.go`：对照上表检查有效白名单
3. 本地需 NewAPI 反映新白名单时：Rotate 相关 Key 或 `docker:reset`
4. 模型 ID 变更： bump `tests/testutil/pg/template.go` 的 `testTemplateVersion`（若动 schema）
