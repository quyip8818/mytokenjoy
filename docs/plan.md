# TokenJoy 工程待办（Plan）

> **最后对齐**：2026-07-07  
> **定位**：工程 backlog **唯一入口**；完成项打 `[x]` 或删除，不另开计划文档。  
> **关联**：[Roadmap.md](./Roadmap.md)（产品差距）· [Frontend.md](./Frontend.md)（架构与契约）· [Backend.md](./Backend.md) §2.5（Keys 约束）· [NewAPI-集成状态与缺口.md](./NewAPI-集成状态与缺口.md)（联调架构）

---

## 维护约定

1. 新待办只写入本文，不新建 `*-计划.md` / `*-下一步.md`
2. 产品级 ❌ 能力（钉钉、OIDC、真实支付等）只维护在 [Roadmap.md](./Roadmap.md)
3. 架构约定与领域数据模型见 [Frontend.md](./Frontend.md) §2、§5.0.1

---

## §1 NewAPI / Relay

架构现状见 [NewAPI-集成状态与缺口.md](./NewAPI-集成状态与缺口.md) §1–2。

### P0 — 生产功能或资金风险

- [ ] **Gateway `/v1` path 剥离** — `domain/relay/gateway_service.go`；客户端 `/v1/chat/completions` 可能被转成 `/chat/completions` → NewAPI 404
- [ ] **充值跳过 TopUp 仍标 `topped_up`** — `domain/billing/service.go` `topUpAndFinish`；`NEW_API_ENABLED=false` 时 DB 已充值、NewAPI 钱包未增加
- [ ] **Relay 关闭时 Key 同步静默跳过** — `domain/relay/lifecycle_ops.go`；DB 有 Key、NewAPI 无 token，无告警

### P1 — 误配、SaaS、可观测

- [ ] `RELAY_GATEWAY_ENABLED` 无组合校验 — 只开 Gateway 不开 NewAPI → 路由不挂载，仅 log
- [ ] `wireGatewayService` 失败静默 — `registry.go` 吞错，`relayGateway == nil`
- [ ] Rebalance / Overrun 在 Relay 关闭时空转 — ingest 仍入队，Worker 调用时 `return nil`
- [ ] `noopWalletService` 余额恒 0 — Gateway 预检 403，`GetWallet` 不区分「未配置」
- [ ] 通知 `NOTIFY_WEBHOOK_URL` 失败静默 — HTTP 失败仍 `return nil`，调用方无感知
- [ ] `processOrgSync` 固定 `DefaultCompanyID` — SaaS 多企业 org 同步范围受限
- [ ] `host.docker.internal` 跨平台 — Linux 非 Docker Desktop 时常不可用
- [ ] `gate-verify` 不测 Backend Gateway — 验证通过 ≠ Gateway 可用

### P2–P3 — 清理与可观测

- [ ] Rebalance/Overrun、钱包 noop、通知 webhook 失败可观测性补强
- [ ] `gate-verify` 增加 Backend `/v1` Gateway 步骤
- [ ] `ingest_notify_total` 幂等重复也 +1 — 仅首次 ledger 插入时计数，或改名并文档化
- [ ] `GET /internal/metrics/ingest` 无鉴权 — 生产可加 webhook secret 或仅 bind localhost
- [x] `OutboxKindRebuildAbilities` / `OutboxKindRevokeToken` 死常量 — 已删（revoke 走同步 `SyncRevokePlatformKey`）

### 联调检查清单

**入账（方案 B）**

- [ ] `LOG_DATABASE_URL` 指向 `logs` 库；init 已建 `newapi` / `backend` schema
- [ ] `NEW_API_WEBHOOK_SECRET` 与 NewAPI `MANAGEMENT_WEBHOOK_SECRET` 一致
- [ ] NewAPI `LOG_SQL_DSN` → `logs`；patch 镜像已 build（非纯上游镜像）
- [ ] Backend Worker 已启动（reconcile / failure retry 依赖 Worker）
- [ ] `GET /api/internal/metrics/ingest` 可查看 `ingest_reconcile_gaps`、`ingest_failures_pending`

**Relay / 管理面（与入账独立）**

- [ ] `NEW_API_ENABLED=true` + `NEW_API_BASE_URL` + `NEW_API_ADMIN_TOKEN`（Key 同步、充值 TopUp）
- [ ] 若开 Gateway：`RELAY_GATEWAY_ENABLED=true` 且 **P0 path 问题已修**
- [ ] 不以 `settle_webhook.sh` 或 compose 里仅配置 URL 作为「notify 已接通」依据 — 以真实 POST `{log_id}` 与 ledger 为准

**本地**

- [ ] `pnpm start`：默认无 NewAPI，入账靠测试 mock / PostgreSQL `newapi.logs`（`testutil.SeedConsumeLog`）
- [ ] `pnpm start:relay`：完整栈；Backend 需配置 `LOG_DATABASE_URL` 与 webhook secret

---

## §2 Backend fake API → 真实现

MSW 已移除；以下接口仍有 `// TODO(real):` 或半真实现。

| API                                                        | 现状                                                            | 目标                                    |
| ---------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------- |
| `GET/PUT /budget/approvals`                                | [x] 已持久化（`budget_approvals` 表）                           | —                                       |
| `GET/PUT/GET test /org/data-source/field-mappings`         | [x] 已持久化（`org_integration.field_mappings` JSONB）            | —                                       |
| `GET /billing/recharge-records`                            | 半真（`company_recharge_orders`；invoice/method 部分字段待真渠道） | 支付渠道、发票系统                      |
| `GET /billing/wallet` 的 `totalConsumed` / `totalRequests` | [x] billing 域经 `usage.Reader` 聚合（`billing/wallet_stats.go`） | —                                       |
| `GET /me/dashboard`                                        | [x] 独立 `memberanalytics` 域（`usage.Reader` + keys quota）       | —                                       |

**刻意保留占位（需 UI 决策或真后端）**

- 钱包发票 Tab / 兑换码：UI 为 disabled 或空态
- 预算树 per-node overrun 可编辑列：overrun 展示列仍在；memberQuota 列已移除

---

## §3 Keys 域兼容清理

架构约束见 [Backend.md](./Backend.md) §2.5。

### P0 — Platform Key Rotate

- [ ] NewAPI Admin 提供 token rotate 或等价端点
- [ ] `relay/interface.go` 新增 `SyncRotatePlatformKey(ctx, platformKeyID) (fullKey string, err error)`
- [ ] `platform_key_actions.go` 替换 HTTP 501；复用 `updatePlatformKeyFullKey`
- [ ] 前端 `key-rotate-confirm` 成功路径恢复 `key-reveal`
- [ ] HTTP：`POST /api/keys/platform/{id}/rotate` → 200 + PlatformKey（含新 `fullKey`）；Relay 关闭 → 503

**验收：** rotate 成功更新 DB `full_key` / `key_prefix`；旧 secret 网关侧失效；不存在 key → 404

### P1 — 审批通过 + Relay 同步跨事务一致性

- [ ] 采用 outbox / `provisioning` 状态（方案 B：与 `OutboxKindCreateToken` 一致）
- [ ] `ApproveApproval` 与 `syncPlatformKeyCreate` 失败态可解释、可重试；不得静默成功

**验收：** Relay 失败时审批与 key 状态可解释；重试成功无需重新审批

### P2 — Workflow 错误展示统一

- [ ] `features/workflow/workflows/**` 内固定文案 `catch` 改为 `workflowErrorMessage(err, fallback)`
- [ ] 已接入勿重复改：`key-form`、`approval-review`、`model-create/edit`、`provider-key-form`、`reject-reason`、`whitelist-config`

### P3 — 种子数据契约

- [ ] `platform_keys.json` 删除 `memberName`、`budgetGroupName`、`appName` 等不入库字段

---

## §4 前端架构收尾

约定见 [Frontend.md](./Frontend.md) §2。**不恢复 MSW**。

### 迁移债务（`check-conventions` 目标态）

- [x] 删除 `routes/*/hooks/` 副本（canonical 在 `features/*/hooks/`）
- [x] 迁移 `components/{budget,org,keys}/` → `features/{domain}/components/`
- [x] 删除 orphan 页：`routes/budget/overview.tsx`、`allocation.tsx`、`routes/billing/`
- [x] 删除 `tests/routes/`（canonical 在 `tests/features/`）
- [x] 删除未接线 budget/org workflow（页面已用内联组件；保留 keys/models 活跃链 + `member-search`）

### 工程优化（非阻断）

- [ ] 预算默认账期去硬编码 `2026-06`（`lib/demo-clock.ts`、`use-budget-page.ts`）
- [ ] Workflow 按域动态 `import()`，减小首屏包体
- [ ] Zod response schema 试点 → OpenAPI/orval 生成类型
- [ ] `@tanstack/react-virtual` 大表格按需引入（行数 >500）
- [ ] `eslint-plugin-boundaries` 部分替代 `check-conventions.ts`
- [ ] Workflow 统一 `onSubmit` 错误与 toast；步骤级 Zod + react-hook-form

### E2E 扩展

细节见 [superpowers/plans/2026-07-07-regenerate-e2e-tests.md](./superpowers/plans/2026-07-07-regenerate-e2e-tests.md)。

- [ ] Task 1–8：Auth、Admin 导航、Dashboard、成员工作台、Org、Budget、Keys、Models/Audit/Wallet
- [ ] 优先：预算审批、Key 申请、组织同步 happy path

### UI 抛光（不阻断发布）

- [ ] Workflow 面板 header/footer 间距（`workflow-panel-chrome.tsx`）
- [ ] 表单 Label 统一 `text-xs text-muted-foreground`（`workflow-form-field.tsx`）

### 模型目录最优改造（全量）

完整方案见 **[Backend-模型目录最优改造计划.md](./Backend-模型目录最优改造计划.md)**（5 阶段 PR：modelcatalog 基础包 → 路由/白名单 modelId 化 → Keys/precheck/计费 → 前端 → SaaS/文档）。

- [x] Phase 0：`internal/pkg/modelcatalog` + enabled 过滤
- [x] Phase 1：路由 / allowlist API → `modelId[]`（Breaking）
- [x] Phase 2：PlatformKey 白名单、precheck、计费闭环
- [x] Phase 3：前端 model-create 双字段、picker/routing/keys 对齐
- [x] Phase 4：SaaS ID 分配器、启动校验、`ModelUsage.callType`、文档同步（`pnpm verify` 待后续执行）

---

## §5 发布与验收

**发布顺序：** 产品模型手工验收 → 生产 DDL → 前后端同发 → UI 像素验收 → E2E。

| 门禁                                           | 级别     |
| ---------------------------------------------- | -------- |
| 产品模型手工验收（6 项）                       | **阻断** |
| Handler / Feature 单测复跑                     | **阻断** |
| `models` 四列迁移（**仅早期生产库**；新库 `schema.sql` 已含列） | 建议     |
| 前后端同发                                     | **阻断** |
| UI 像素验收                                    | 建议     |
| E2E（keys / models / audit / wallet / member） | 建议     |

**回滚：** DDL 仅 additive、不回滚；应用须前后端成对回滚。

### 产品模型手工验收（阻断）

- [ ] 平台 Key：部门树点击，列表与 `departmentId` 筛选一致
- [ ] 平台 Key：成员 / 项目 Tab 切换正确，数据不串
- [ ] 审批四 Tab：pending / approved / rejected / all 正确；角标仅 pending
- [ ] 模型列表：内置 / 自定义 Tab；custom 显示 `endpoint`
- [ ] Postgres 重启后 custom 模型 `endpoint` 持久化仍在
- [ ] 改名同步：改成员名 / 预算组名后，平台 Key 列表展示名即时更新（enrich）

### 自动化（发布前复跑）

```bash
cd apps/backend && go test ./tests/handler/keys/... ./tests/handler/models/... -count=1
pnpm -F @tokenjoy/frontend test -- tests/features/keys tests/features/models
pnpm -F @tokenjoy/frontend test:e2e -- keys models audit wallet member
```

### 可选补强（非阻断）

- [ ] 改名同步集成测试
- [ ] 重启持久化集成测试（`endpoint` 落库）
- [ ] 成员视角审批 `memberId` 接入 `use-approval-page`

### UI 像素验收（建议）

- 视觉基准 commit `716eeec`；对比 `git diff 716eeec HEAD -- apps/frontend/src/features/<domain>/components/`
- `/keys/mine` 无基准，单独约定

### `models` 四列生产迁移（仅早期生产库）

> 新安装 / wipe 重建走 `schema.sql` 全量 DDL，**无需**执行下列脚本。仅对 `schema.sql` 引入四列之前的存量库执行一次。

```sql
ALTER TABLE models
  ADD COLUMN IF NOT EXISTS model_type   TEXT NOT NULL DEFAULT 'builtin',
  ADD COLUMN IF NOT EXISTS description  TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS visibility   TEXT NOT NULL DEFAULT 'all',
  ADD COLUMN IF NOT EXISTS endpoint     TEXT;

UPDATE models SET model_type = 'custom' WHERE provider = 'custom' AND model_type = 'builtin';
UPDATE models SET model_type = 'builtin' WHERE provider <> 'custom' AND model_type = 'builtin';
UPDATE models SET visibility = 'all' WHERE visibility = '' OR visibility IS NULL;
```

本地：`docker compose down -v` 重建（见 [Backend-存储架构.md](./Backend-存储架构.md)）。

### 权限手工 QA

见 [权限管理.md](./权限管理.md) §12.3：

- [ ] 首屏仅 1 次 `GET /api/session`
- [ ] 角色变更后 nav 更新（无需 F5）
- [ ] 多 Tab revision 同步（broadcast / revision 头）

---

## §6 长期产品差距

钉钉/企微、IM 审批通知、预算阈值 Worker、OIDC、真实支付、`/platform/*` 前端、热存归档等见 [Roadmap.md](./Roadmap.md)。

---

## §7 Phase 3 — 性能与权限规模化

**当前不必立项。** 满足以下**任一**条件时启动：

- `platform_keys` 行数 > **500**
- `GET /keys/platform` P99 > **300ms**

| #   | 任务                | 技术方向                                                           |
| --- | ------------------- | ------------------------------------------------------------------ |
| 1   | 删冗余列            | `DROP member_name, budget_group_name`；repo 停读写                 |
| 2   | SQL 筛选            | `keys_repo.ListPlatformKeysFiltered`，JOIN members / budget_groups |
| 3   | 真分页              | `page` / `pageSize` / `total` + SQL `LIMIT/OFFSET`                 |
| 4   | 列表 RBAC           | 非管理员默认 `departmentId=会话部门`                               |
| 5   | 后端搜索 `q`        | 名称/前缀模糊，替代前端全量 `search`                               |
| 6   | `visibility` 运行时 | 与 `model_allowlist`、部门路由合并校验                             |
| 7   | Models `type` query | 仅当模型数 > 500                                                   |

**可提前立项（不依赖性能触发）：**

- [ ] 上线前部门管理员仅能看本部门 Key（#4）
- [ ] `visibility` 须真正限制模型访问（#6）

**约束：** 不引入平行 enrich API；SQL 筛选与 enrich 同路径；RBAC 在后端强制。
