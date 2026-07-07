# TokenJoy 工程待办（Plan）

> **最后对齐**：2026-07-07  
> **定位**：工程 backlog **唯一入口**；完成项打 `[x]` 或删除，不另开计划文档。  
> **关联**：[Roadmap.md](./Roadmap.md)（产品差距）· [下一步工作清单.md](./下一步工作清单.md)（发布门禁）· [NewAPI-集成状态与缺口.md](./NewAPI-集成状态与缺口.md)（联调架构）

---

## 维护约定

1. 新待办只写入本文对应章节，不新建 `*-计划.md` / `*-下一步.md`
2. 产品级 ❌ 能力（钉钉、OIDC、真实支付等）只维护在 [Roadmap.md](./Roadmap.md)
3. Phase 3 性能规模化（>500 keys / P99>300ms）触发条件见 [下一步工作清单.md](./下一步工作清单.md) §8

---

## §1 NewAPI / Relay

架构现状见 [NewAPI-集成状态与缺口.md](./NewAPI-集成状态与缺口.md) §1–2。

### P0 — 生产功能或资金风险

- [ ] **Gateway `/v1` path 剥离** — `domain/relay/gateway_service.go`；客户端 `/v1/chat/completions` 可能被转成 `/chat/completions` → NewAPI 404
- [ ] **充值跳过 TopUp 仍标 `topped_up`** — `domain/billing/service.go` `topUpAndFinish`；`NEW_API_ENABLED=false` 时 DB 已充值、NewAPI 钱包未增加
- [ ] **Relay 关闭时 Key 同步静默跳过** — `domain/relay/lifecycle_ops.go`；DB 有 Key、NewAPI 无 token，无告警

### P1 — 误配、SaaS、可观测

- [ ] `NEW_API_PUBLIC_URL` 未使用 — 配置冗余，对外 Relay URL 无法与此对齐
- [ ] `RELAY_GATEWAY_ENABLED` 无组合校验 — 只开 Gateway 不开 NewAPI → 路由不挂载，仅 log
- [ ] `wireGatewayService` 失败静默 — `registry.go` 吞错，`relayGateway == nil`
- [ ] Rebalance / Overrun 在 Relay 关闭时空转 — ingest 仍入队，Worker 调用时 `return nil`
- [ ] `noopWalletService` 余额恒 0 — Gateway 预检 403，`GetWallet` 不区分「未配置」
- [ ] 通知 `NOTIFY_WEBHOOK_URL` 失败静默 — HTTP 失败仍 `return nil`，调用方无感知
- [ ] `processOrgSync` 固定 `DefaultCompanyID` — SaaS 多企业 org 同步范围受限
- [ ] `host.docker.internal` 跨平台 — Linux 非 Docker Desktop 时常不可用
- [ ] `gate-verify` 不测 Backend Gateway — 验证通过 ≠ Gateway 可用

### P2–P3 — 清理与可观测

- [ ] `NEW_API_PUBLIC_URL` 落地或删除
- [ ] Rebalance/Overrun、钱包 noop、通知 webhook 失败可观测性补强
- [ ] `gate-verify` 增加 Backend `/v1` Gateway 步骤
- [ ] `ingest_notify_total` 幂等重复也 +1 — 仅首次 ledger 插入时计数，或改名并文档化
- [ ] `GET /internal/metrics/ingest` 无鉴权 — 生产可加 webhook secret 或仅 bind localhost
- [ ] `OutboxKindRebuildAbilities` 等死常量 — 清理或接 Worker 分支

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

- [ ] `pnpm start`：默认无 NewAPI，入账靠测试 mock / memory LogStore
- [ ] `pnpm start:relay`：完整栈；Backend 需配置 `LOG_DATABASE_URL` 与 webhook secret

---

## §2 Backend fake API → 真实现

MSW 已移除；以下接口为保留 UI 而补充的临时实现（代码内 `// TODO(real):`）。

| API | 现状 | 目标 |
| --- | --- | --- |
| `GET/PUT /budget/approvals` | 内存 fake（按 company 隔离，seed 5 条） | 预算审批工作流持久化 |
| `GET/PUT/GET test /org/data-source/field-mappings` | 内存 fake（按 company + platform） | 同步引擎 + DB 持久化 |
| `GET /billing/recharge-records` | 半真（`company_recharge_orders` + invoice/method overlay fake） | 支付渠道、发票系统 |
| `GET /billing/wallet` 的 `totalConsumed` / `totalRequests` | 半真（usage 聚合） | 统一账单域（`billing/service.go`） |
| `GET /me/dashboard` | fake BFF（usage 按 memberId 聚合 + keys quota 代理） | 独立成员分析域（`member/dashboard.go`） |

**刻意保留占位（需 UI 决策或真后端）**

- 钱包发票 Tab / 兑换码：UI 为 disabled 或空态
- 预算树 per-node overrun 可编辑列：overrun 展示列仍在；memberQuota 列已移除

---

## §3 Keys 域兼容清理

规格细节见 [archive/清理兼容与死代码-下一步.md](./archive/清理兼容与死代码-下一步.md)。

### P0 — Platform Key Rotate

- [ ] NewAPI Admin 提供 token rotate 或等价端点
- [ ] `relay/interface.go` 新增 `SyncRotatePlatformKey`
- [ ] `platform_key_actions.go` 替换 HTTP 501
- [ ] 前端 `key-rotate-confirm` 成功路径恢复 `key-reveal`

### P1 — 审批通过 + Relay 同步跨事务一致性

- [ ] 设计 `provisioning` / outbox 重试状态（避免审批已通过但 `full_key` 为空且静默成功）
- [ ] `ApproveApproval` 与 `syncPlatformKeyCreate` 失败态可解释、可重试

### P2 — Workflow 错误展示统一

- [ ] `features/workflow/workflows/**` 内固定文案 `catch` 改为 `workflowErrorMessage(err, fallback)`
- [ ] 覆盖：`member-form`、`budget-group-form`、`role-form`、`import-preview` 等（已接入项勿重复改）

### P3 — 种子数据契约

- [ ] `platform_keys.json` 删除 `memberName`、`budgetGroupName`、`appName` 等不入库字段

---

## §4 前端架构收尾

架构指南见 [前端架构优化与模块化建议.md](./前端架构优化与模块化建议.md)。**不恢复 MSW**（Vitest 用 `createMockApis`，E2E/dev 用真 backend）。

### 迁移债务（`check-conventions` 目标态）

- [ ] 删除 `routes/*/hooks/` 副本（canonical 在 `features/*/hooks/`；注册路由已从 `@/features/*` 导入）
- [ ] 迁移 `components/{budget,org,keys}/` → `features/{domain}/components/`
- [ ] 删除 orphan 页：`routes/budget/overview.tsx`、`allocation.tsx`
- [ ] 迁移 `tests/routes/` → `tests/features/`（当前 6 个遗留测试文件）

### 工程优化（非阻断）

- [ ] 预算默认账期去硬编码 `2026-06`（`lib/demo-clock.ts`、`use-budget-page.ts`）
- [ ] Workflow 按域动态 `import()`，减小首屏包体
- [ ] Zod response schema 试点 → OpenAPI/orval 生成类型（长期）
- [ ] 大表格页按需引入 `@tanstack/react-virtual`（行数 >500）

### E2E 扩展

实施细节见 [superpowers/plans/2026-07-07-regenerate-e2e-tests.md](./superpowers/plans/2026-07-07-regenerate-e2e-tests.md)。

- [ ] Task 1：Auth & Session E2E 重写
- [ ] Task 2：Admin 路由导航覆盖
- [ ] Task 3：Dashboard 交互
- [ ] Task 4：成员工作台
- [ ] Task 5：Org 域（structure / roles / data-source）
- [ ] Task 6：Budget 域
- [ ] Task 7：Keys 域（platform / approval / mine）
- [ ] Task 8：Models / Audit / Wallet

**关键路径（优先于全量 Task）**

- [ ] 预算审批 happy path
- [ ] Key 申请 happy path
- [ ] 组织同步 happy path

---

## §5 发布与验收

发布门禁详见 [下一步工作清单.md](./下一步工作清单.md) §5–7。Phase 3 量化触发（>500 keys / P99>300ms）不在本表重复。

### 产品模型手工验收（阻断发布）

- [ ] 平台 Key：部门树点击，列表与 `departmentId` 筛选一致
- [ ] 平台 Key：成员 / 项目 Tab 切换正确，数据不串
- [ ] 审批四 Tab：pending / approved / rejected / all 正确；角标仅 pending
- [ ] 模型列表：内置 / 自定义 Tab；custom 显示 `endpoint`
- [ ] Postgres 重启后 custom 模型 `endpoint` 持久化仍在
- [ ] 改名同步：改成员名 / 预算组名后，平台 Key 列表展示名即时更新（enrich）

### 自动化与 E2E（建议）

- [ ] 发布前复跑：`go test ./tests/handler/keys/... ./tests/handler/models/...` + 前端 keys/models feature 测试
- [ ] E2E：`pnpm -F @tokenjoy/frontend test:e2e -- keys models audit wallet member`

### 权限手工 QA

见 [权限管理.md](./权限管理.md) §12.3：

- [ ] 首屏仅 1 次 `GET /api/session`
- [ ] 角色变更后 nav 更新（无需 F5）
- [ ] 多 Tab revision 同步（broadcast / revision 头）

---

## §6 长期产品差距

钉钉/企微、IM 审批通知、预算阈值 Worker、OIDC、真实支付、`/platform/*` 前端、热存归档等见 [Roadmap.md](./Roadmap.md)。不在本文维护细节。
