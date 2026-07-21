# Seed 模块 — 修复清单

> 范围：**仅** `BOOTSTRAP_MODE=demo`（及依赖 demo 契约的测试）。与生产租户数据无关的项放此文件；NewAPI / Gateway 集成见仓库根 `docs/`。

---

## 边界（避免误归因）

| 数据                                                | Seed 是否写入 | 实际负责方                                                                                                                                      |
| --------------------------------------------------- | ------------- | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| 公司 / 组织 / 模型 catalog / platform_keys 行       | ✅ `apply/`   | —                                                                                                                                               |
| `platform_key_mappings` + `key_hash` 落地           | ❌            | 本地：`dev-bootstrap`（`provision.Bootstrap`）；生产：用户 Create → `SyncPlatformKeyCreate`                                                     |
| NewAPI Token `group` / secret                       | ❌            | `TrySyncCreate` / `SyncUpdatePlatformKey`；River 仅重试 outbox（`create_key`、`upsert_channel` 等）                                        |
| `companies.newapi_wallet_company_id`                | ❌            | SaaS `CreateCompany`；demo 见 [FIX-SEED-004](#fix-seed-004)                                                                                    |
| NewAPI `UserUsableGroups`                           | ❌            | `EnsureGroup`（Create 路径）+ 本地脚本保险                                                                                                      |

**Reseed 规则：** 修改 seed 数据后须 **`pnpm docker:reset`**（清卷）；非空库不会自动覆盖 seed。仅 `make dev-bootstrap` 只补 NewAPI sync，不 reseed。

---

## ID 体系（UUID v7）

所有实体 ID 统一为 PostgreSQL `UUID` 类型 + Go `uuid.UUID`。Seed 常量集中在 `seed/contract/ids.go`。

**SSOT**：`contract.ModelTypeToID`（`map[string]uuid.UUID`）。

| 变量名 | type | 说明 |
| --- | --- | --- |
| `IDModelLocalTest` | `dev-local-test` | 仅 dev / demo catalog |
| `IDModel1` ~ `IDModel10` | `deepseek-v4` … `gpt-4o` | demo 生产模型 |

客户新建模型由 DB `DEFAULT gen_random_uuid()` 生成 ID，不走 seed。

---

## 历史修复（全部已完成）

### FIX-SEED-001 — platform key 有效白名单为空 ✅

已通过 UUID 迁移统一修复。seed 中 platform key 白名单现直接引用 `contract.IDModelX` UUID 常量。

### FIX-SEED-002 — model_id 与契约不一致 ✅

已解决。model_id 现为 UUID，seed 和代码共享 `contract/ids.go` 常量，不再有数字 ID 不匹配问题。

### FIX-SEED-003 — 改 seed 后 NewAPI token 未更新

**仍然适用**：Seed 只改 Postgres 行；NewAPI token status/group 在 sync 时推送。已 synced 的 mapping 不会自动重算。  
**修复**：Rotate 相关 Key 或 `docker:reset`。bootstrap reconcile 会对齐 status + group。

### FIX-SEED-004 — Demo 公司无 `newapi_wallet_company_id` ✅

已落地。bootstrap 调用 `CreateUser` + `UpdateNewAPIWalletCompanyID`；wallet username 由 `company.WalletUsername(companyID)` 确定性生成（UUID 去横杠，32 字符）。

---

## Demo Platform Key 对齐表

部门解析与线上一致：`departmentIDForPlatformKey`（member 部门优先，否则 project `OwnerDepartmentID`）。

所有 Key ID、部门 ID、模型 ID 现为 UUID，具体值见 `seed/contract/ids.go`。  
白名单有效性 = Key 白名单 ∩ 部门路由允许列表。

NewAPI `group` 由集成层计算，**不是** seed 字段。

---

## 变更检查

修改 seed 后：

1. 跑 `tests/seed/contract/`
2. 若改 platform key 或 org node：检查有效白名单交集
3. 本地需 NewAPI 反映新白名单时：Rotate 相关 Key 或 `docker:reset`
4. schema 变更：bump `tests/testutil/pg/template.go` 的 `testTemplateVersion`
