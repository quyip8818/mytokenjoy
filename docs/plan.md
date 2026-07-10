# TokenJoy 工程待办（Plan）

> **定位**：上线前 fix / 功能点 / 联调与发布门禁的 **唯一 backlog**。完成即删除条目，不留 `[x]`。  
> **关联**：[Roadmap.md](./Roadmap.md)（产品差距）· [Frontend.md](./Frontend.md)（架构与契约）· [Backend.md](./Backend.md) · [NewAPI-集成状态与缺口.md](./NewAPI-集成状态与缺口.md)（现状与可优化点）

---

## 维护约定

1. 新待办只写入本文；禁止再开 `*-计划.md` / `*-下一步.md`
2. 产品级 ❌ 能力（钉钉、OIDC、真实支付等）只维护在 [Roadmap.md](./Roadmap.md)
3. 架构现状见 Backend* / NewAPI 文档；本文不写「已关闭」历史

---

## §1 NewAPI / Gateway（上线前）

架构现状与可优化点见 [NewAPI-集成状态与缺口.md](./NewAPI-集成状态与缺口.md)。

### Fix

- [ ] **`gate-verify` 覆盖 Backend Gateway** — `gate-verify.sh` 增加：seed Key → `POST ${API_URL}/v1/chat/completions`（Bearer sk-）→ 期望 200 或业务 4xx（非 502）
- [ ] **demo Gateway 组合校验** — 只开 `NEWAPI_GATEWAY_ENABLED`、不开 `NEW_API_ENABLED` 时启动失败或明确拒绝挂载
- [ ] **Rebalance / Overrun NewAPI 关闭禁 noop** — `!Enabled()` 时勿静默 `return nil`；入队或处理失败可观测
- [ ] **`noopWalletService`** — NewAPI 关闭时 `AvailableQuota` 恒 0 导致预检误拒；明确行为或禁止该组合下开 Gateway
- [ ] **通知 `NOTIFY_WEBHOOK_URL` 失败可观测** — HTTP 失败勿对调用方一律 `return nil`
- [ ] **Update 严格 Remote-first（可选）** — `UpdatePlatformKey` 现为 DB-first + Sync + 回滚；若上线要求铁律一致则改为先 Remote
- [ ] **`processOrgSync` 多企业** — 勿固定 `DefaultCompanyID`（SaaS）
- [ ] **`host.docker.internal` 文档/compose** — Linux 联调可复现（`extra_hosts` 或改 webhook URL）
- [ ] **`ingest_notify_total` 幂等** — 仅首次 ledger 插入计数（或改名并文档化）
- [ ] **`GET /internal/metrics/ingest` 鉴权** — 生产加 secret 或仅 bind localhost

### 联调签字（真实栈）

前提：`pnpm start:newapi` + 生产契约 env（`DEPLOY_ENV=production` 或等价三联开）。

- [ ] Gateway `/v1/chat/completions` → 200 + ledger 入账
- [ ] Toggle off → 立刻 403；on → 恢复
- [ ] Revoke → 403；DB revoked + NewAPI token gone
- [ ] Rotate → 旧 sk- 403，新 sk- 200，`newapi_key_id` 不变
- [ ] E2E `keys-self-service.spec.ts` · `rotates an existing Key` 稳定通过

```bash
cd apps/backend && go test -tags=testhook \
  ./tests/handler/gateway/... \
  ./tests/domain/newapisync/... \
  ./tests/domain/keys/... -count=1

pnpm -F @tokenjoy/frontend test:e2e -- keys-self-service
```

### 联调检查清单

**入账（方案 B）**

- [ ] `LOG_DATABASE_URL` 指向 `logs` 库；init 已建 `newapi` / `backend` schema
- [ ] `NEW_API_WEBHOOK_SECRET` 与 NewAPI `MANAGEMENT_WEBHOOK_SECRET` 一致
- [ ] NewAPI `LOG_SQL_DSN` → `logs`；patch 镜像已 build（非纯上游镜像）
- [ ] Backend Worker 已启动（webhook 入队后的异步入账 / reconcile / failure retry 依赖 Worker）
- [ ] webhook 返回 `accepted` 后，Worker 一轮内 ledger 可见；`GET /api/internal/metrics/ingest` 可查看 `ingest_reconcile_gaps`、`ingest_jobs_pending`

**NewAPI / 管理面**

- [ ] `NEW_API_ENABLED=true` + `NEW_API_BASE_URL` + `NEW_API_ADMIN_TOKEN`
- [ ] 若开 Gateway：`NEWAPI_GATEWAY_ENABLED=true` 且 `NEW_API_ENABLED=true`
- [ ] 以真实 POST `{log_id}` 与 ledger 为准，不以 compose 仅配 URL 当作已接通

**本地**

- [ ] `pnpm start`：默认无 NewAPI；入账靠测试 mock / `testutil.SeedConsumeLog`
- [ ] `pnpm start:newapi`：完整栈；Backend 配置 `LOG_DATABASE_URL` 与 webhook secret

---

## §2 半真 API / 占位

| API | 现状 | 上线前目标 |
| --- | --- | --- |
| `GET /billing/recharge-records` | 半真（`company_recharge_orders`；invoice/method 待真渠道） | 明确占位文案，或接支付/发票渠道 |

**刻意保留占位（需产品决策）**

- 钱包发票 Tab / 兑换码：UI disabled 或空态
- 预算树 overrun 展示列保留；memberQuota 列已移除

---

## §3 Keys

架构约束见 [Backend.md](./Backend.md) §2.5。

### 审批通过 + NewAPISync 同步跨事务一致性

- [ ] 采用 outbox / `provisioning` 状态（与 `OutboxKindCreateKey` 一致）
- [ ] `ApproveApproval` 与 `syncPlatformKeyCreate` 失败态可解释、可重试；不得静默成功

**验收：** NewAPI 失败时审批与 key 状态可解释；重试成功无需重新审批

### Workflow 错误展示

- [ ] `features/workflow/workflows/**` 内固定文案 `catch` 改为 `workflowErrorMessage(err, fallback)`
- [ ] 已接入勿重复改：`key-form`、`approval-review`、`model-create/edit`、`provider-key-form`、`reject-reason`、`whitelist-config`

### 种子数据契约

- [ ] `platform_keys.json` 删除 `memberName`、`budgetGroupName`、`appName` 等不入库字段

---

## §4 前端（上线前 / 建议）

约定见 [Frontend.md](./Frontend.md) §2。**不恢复 MSW**。

### 工程优化

- [ ] 预算默认账期去硬编码 `2026-06`（`lib/demo-clock.ts`、`use-budget-page.ts`）
- [ ] Workflow 按域动态 `import()`，减小首屏包体
- [ ] Zod response schema 试点 → OpenAPI/orval 生成类型
- [ ] `@tanstack/react-virtual` 大表格按需引入（行数 >500）
- [ ] `eslint-plugin-boundaries` 部分替代 `check-conventions.ts`
- [ ] Workflow 统一 `onSubmit` 错误与 toast；步骤级 Zod + react-hook-form

### E2E 扩展

- [ ] Auth、Admin 导航、Dashboard、成员工作台、Org、Budget、Keys、Models/Audit/Wallet
- [ ] 优先：预算审批、Key 申请、组织同步 happy path

### UI 抛光（不阻断）

- [ ] Workflow 面板 header/footer 间距（`workflow-panel-chrome.tsx`）
- [ ] 表单 Label 统一 `text-xs text-muted-foreground`（`workflow-form-field.tsx`）

---

## §5 发布与验收

**发布顺序：** 产品模型手工验收 → 生产 DDL（若需）→ 前后端同发 → UI 像素验收 → E2E。

| 门禁 | 级别 |
| --- | --- |
| 产品模型手工验收（6 项） | **阻断** |
| Handler / Feature 单测复跑 | **阻断** |
| `models` 四列迁移（**仅早期生产库**；新库 `schema.sql` 已含） | 建议 |
| 前后端同发 | **阻断** |
| UI 像素验收 | 建议 |
| E2E（keys / models / audit / wallet / member） | 建议 |

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

> 新安装 / wipe 重建走 `schema.sql` 全量 DDL，**无需**执行下列脚本。

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

## §7 性能与权限规模化（触发后立项）

**当前不必立项。** 满足以下**任一**条件时启动：

- `platform_keys` 行数 > **500**
- `GET /keys/platform` P99 > **300ms**

| # | 任务 | 技术方向 |
| --- | --- | --- |
| 1 | 删冗余列 | `DROP member_name, budget_group_name`；repo 停读写 |
| 2 | SQL 筛选 | `ListPlatformKeysFiltered`，JOIN members / budget_groups |
| 3 | 真分页 | `page` / `pageSize` / `total` + SQL `LIMIT/OFFSET` |
| 4 | 列表 RBAC | 非管理员默认 `departmentId=会话部门` |
| 5 | 后端搜索 `q` | 名称/前缀模糊 |
| 6 | `visibility` 运行时 | 与 allowlist、部门路由合并校验 |
| 7 | Models `type` query | 仅当模型数 > 500 |

**可提前立项（上线前建议）：**

- [ ] 部门管理员仅能看本部门 Key（#4）
- [ ] `visibility` 须真正限制模型访问（#6）

**约束：** 不引入平行 enrich API；SQL 筛选与 enrich 同路径；RBAC 在后端强制。
