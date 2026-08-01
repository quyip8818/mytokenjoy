# 后端问题清单（已验证）

> **来源**：多轮代码审查累积  
> **图例**：🔴 严重 · 🟠 中等 · 🟡 低优 · ✅ 已修复（留档）

---

## 一、已修复项（本轮验证已关闭）

| 原编号    | 问题                               | 修复方式                                                 |
| --------- | ---------------------------------- | ---------------------------------------------------------- |
| REVIEW#7  | LRU Cache `touch()` O(n) 线性扫描  | `container/list` + map，O(1)                             |
| REVIEW#12 | 无 HTTP 层 rate limiting           | `RateLimitTenant` + `RateLimitLoginPaths` 中间件         |
| REVIEW#15 | 缺乏结构化 access log              | `access_log.go`（method/path/status/latency/company_id） |
| REVIEW#6  | CORS 缺少 `Access-Control-Max-Age` | 已设置 86400                                             |
| SEC-04    | Per-tenant API rate limiting       | 同 REVIEW#12                                             |
| PERF-02   | LRU O(n) touch                     | 同 REVIEW#7                                              |
| PERF-01   | Authz 每请求查 revision            | revisionCache TTL 5s                                     |
| PERF-04   | pgxpool 默认连接池                 | `DB_MAX_CONNS`/`DB_MIN_CONNS` 可配                       |
| BUG-01    | Company ID 生成竞态条件            | UUID v7 迁移，事务内生成，无竞态                          |
| BUG-02    | Keys 域 Entity ID 用 `UnixMilli()` | 统一为 `uuid.Must(uuid.NewV7())`，无碰撞风险              |
| BUG-07    | Budget Rejection 路径无事务        | 审批流程整体重构为 `approval.Engine`，`Reject` 已用 `txRunner` 包裹事务（原 `budget/approvals.go` 已不存在） |
| DEBT-08   | Approval 补偿逻辑吞掉原始 sync 错误 | `approval/engine.go` 的 `Approve` 分别记录 sync 错误与 compensate 错误两条日志，最终仍返回原始 sync error |
| DEBT-10   | `ListApprovals`/`ApprovalBudgetCheck` 全量加载 | `keys.Service` 接口已移除这两个方法，审批查询统一走 `store.ApprovalRepository.List`（SQL `WHERE`+`LIMIT/OFFSET`+`COUNT(*)`） |
| DEBT-01   | `buildServiceRegistry` 用 `panic` 处理运行时错误 | 全仓库已无 `panic(` 用法；`registry.go` 现仅为 `ServiceRegistry` struct 定义，装配逻辑分散在 `compose_domain*.go`，统一走 `error` 返回 |
| DEBT-11   | Ingest Worker 无显式 stop 确认      | 入账任务已迁移到 River `IngestWorker`；`backgroundWorkers.stop()` 调 `river.Stop(ctx)`，走 River 内建 graceful shutdown |

---

## 二、🔴 严重（数据一致性 / 安全）

### BUG-03 — `CreatePlatformKey` 无事务、无补偿

**文件**：`internal/domain/keys/platform_key_create.go`  
**现状**：

1. 事务外 `LoadBudgetContext` 读取 → 验证预算
2. `SetPlatformKeys` 写入 DB（无锁、无事务）
3. `syncPlatformKeyCreate` 调外部 API
4. 外部 API 失败 → DB 已有脏数据（status=active 但外部不存在），**无补偿**

**对比**：审批流程创建 Key（`approval_handler.go` 的 `KeyApprovalHandler`）走 `OnApprovedTx`（内含 `AcquireBudgetLock`）+ `PostApprove` + `Compensate` 三段式，有完整补偿链；但 `CreatePlatformKey` 这条**直接创建**路径（非审批）仍缺失同等保护。

---

### BUG-04 — 审批预算校验：并发竞态已解决，但静默 clamp 掩盖了充足性校验缺失

**文件**：`internal/domain/keys/approval_handler.go`（`KeyApprovalHandler.OnApprovedTx`）  
**现状**：`OnApprovedTx` 已加 `tx.Budget().AcquireBudgetLock(ctx)`，原 TOCTOU 并发竞态已通过锁序列化解决。但扣减部门预留池时：

```go
newReserved := reserved - personalBudgetAdded
if newReserved < 0 {
    newReserved = 0
}
```

预留池不足时**静默 clamp 到 0**，不返回错误也不阻止本次审批通过——`PreApprove` 阶段的无锁校验（`reservedPool >= requestedBudget`）可能因并发场景已经过期，锁内应重新校验并在不足时报错，而非静默吞掉差额。

---

### BUG-05 — `UpdatePlatformKey` / `TogglePlatformKey` 并发保护不足

**文件**：`platform_key_update.go`、`platform_key_actions.go`  
**现状**：

- `UpdatePlatformKey` 经 `persistPlatformKeyWithNewAPISync`：写 DB → 调远程同步 → **远程失败时已有本地回滚**（`platformKeys[idx] = previous` 后重写）。"无补偿"描述已不准确。
- `TogglePlatformKey` 先调 `s.newAPISync.SyncUpdatePlatformKey`（远程），成功后才写本地状态，**远程成功但本地写失败时无任何回滚**。
- 两者读-验证-改-写全程仍**无事务/无 advisory lock**，并发更新同一 key 仍可能 lost-update。

---

### BUG-06 — Auth handler 绕过 body 大小限制

**文件**：`internal/http/handler/auth/handler.go`  
**现状**：`Login`、`AcceptInvite`、`SetPassword`、`ResetPassword` 均直接 `json.NewDecoder(r.Body).Decode()`，未走 `httputil.DecodeJSON`（限制请求体大小）。攻击者可发送大 payload 消耗内存。

---

## 三、🟠 中等（Hacky / 技术债）

### DEBT-02 — `Void` 响应返回 `200 null` 而非 204

**文件**：`internal/http/response/json.go`  
**现状**：`Void()` → `JSON(w, 200, nil)` → body 为 `null\n`。不符合 REST 惯例。  
**建议**：改为 `204 No Content` 或返回 `{}`。

---

### DEBT-03 — `DecodeJSON` 传 `nil` ResponseWriter 给 MaxBytesReader

**文件**：`internal/http/httputil/decode.go`  
**现状**：`http.MaxBytesReader(nil, r.Body, maxRequestBodySize)`。`w=nil` 时超限不会自动关闭连接。  
**建议**：在中间件层限制或传入真实 `w`。

---

### DEBT-04 — `Delayer.Wait` 散布在 domain 方法中

**文件**：`keys/*.go`、`budget/projects.go`、`budget/alerts.go`、`models/service.go` 等  
**现状**：每个 Create/Update/Approve 操作内嵌 300–500ms `s.delayer.Wait(ctx, ...)`。生产 `SimulateDelay=false` 时为 no-op，但代码噪音大、新人易误解。  
**建议**：如需 demo 延迟，在 HTTP 中间件统一处理。

---

### DEBT-05 — Gateway 所有 precheck 失败统一返回 403

**文件**：`internal/domain/gateway/gateway_service.go`  
**现状**：key 不存在、预算耗尽、模型不允许 → 全返 `403 "request rejected"`（限流已正确走独立的 429 路径，与 precheck 403 无关）。调用者无法区分失败原因。  
**PRD 要求**：401（key 无效）/ 403（模型不允许）/ 429（超限）/ 502（供应商不可用）。与 [PRD-差距分析.md](./PRD-差距分析.md) P0-3 为同一问题。

---

### DEBT-06 — Gateway `parseBearerSecret` 大小写敏感

**文件**：`internal/domain/gateway/auth.go`  
**现状**：只匹配 `"Bearer "`（首字母大写）。RFC 7235 规定 scheme 大小写不敏感。  
**建议**：`strings.EqualFold` 或 `strings.ToLower` 前缀匹配。

---

### DEBT-07 — `response.JSON` Encode 错误被忽略

**文件**：`internal/http/response/json.go`  
**现状**：`_ = json.NewEncoder(w).Encode(v)`。若 `v` 含无法序列化字段 → header 已写、body 截断。  
**建议**：先 `json.Marshal` 检查，失败写 500。

---

### DEBT-09 — Store 层部分 Repo 每次调用 new 实例

**文件**：`internal/store/postgres/postgres.go`  
**现状**：`Org()`/`Budget()`/`Keys()`/`Models()`/`Audit()` 已在启动时缓存于 `domainRepos`（`newDomainRepoSet`），复用同一实例；但 `Company()`/`User()`/`Invite()`/`Billing()`/`Notification()`/`Session()` 等仍每次调用返回新 `struct{db pool}`。影响有限（无状态 struct，仅多一次分配），非高优先级。

---

### DEBT-12 — Keys 域 `SetPlatformKeys` 直接 CRUD 路径仍无锁

**文件**：`internal/domain/keys/platform_key_create.go`、`platform_key_update.go`、`platform_key_actions.go`  
**现状**：全量 load → 内存修改 → 全量写回，**无 advisory lock/事务**。审批流程创建 Key（走 `approval_handler.go` 的 `OnApprovedTx`）已有 `AcquireBudgetLock` 保护；但直接 CRUD（非审批）路径仍暴露并发 lost-update 风险，与 BUG-03/BUG-05 同根因。  
**建议**：短期加 advisory lock + 事务；长期改增量 API（InsertKey、UpdateKeyStatus）。

---

## 四、🟡 低优先级

| 编号   | 文件                  | 问题                                                         |
| ------ | --------------------- | ------------------------------------------------------------ |
| LOW-01 | `keys/handler.go`     | 写路由使用 `ReadRoutes` 命名（语义不明）                     |
| LOW-02 | `keys/handler.go`     | `ProviderCreate` 返回 200 而非 201 Created                   |
| LOW-03 | `rebalance.go`        | `active := mappings[:0]` 原地过滤修改输入 slice              |
| LOW-04 | `config.go`           | `SessionTTLSec=0` 无 validate 保护                           |
| LOW-05 | `middleware/cors.go`  | `Access-Control-Allow-Methods`/`Allow-Headers` **始终无条件设置**（不区分 Origin 是否命中白名单），应仅在 Origin 匹配时设置 |
| LOW-06 | `authz/cache.go`      | revision 变更时旧条目仅靠 LRU 淘汰，无主动 invalidate        |

---

## 五、工程收口（跨域/联调项）

以下为跨域/联调项：

| 编号   | 优先级 | 项                                                   | 说明                                                    |
| ------ | ------ | ---------------------------------------------------- | ------------------------------------------------------- |
| ENG-01 | P0     | CI 引用不存在的脚本，构建实际会失败                   | `.github/workflows/ci.yml` 调用 `pnpm verify`；`package.json` 未定义该脚本（也无 `verify:gate`），CI 会以 "Missing script" 失败。需要么在 `package.json` 补上 `verify`（组合 `lint` + `test` + `build`），要么改 CI 直接调用现有脚本 |
| ENG-02 | P1     | 通知 enqueue 失败分支不可观测                        | `DispatchAsync` 正常路径已走 River queue（内建重试/可观测）；仅当 `InsertNotificationDelivery` 入队失败、fallback 到同步 `ch.Send()` 时，该次失败仅记 `notification_log`，调用方不感知 |
| ENG-03 | P0     | Update Remote-first（Toggle 路径）                   | `TogglePlatformKey` 先调远程再写本地，本地写失败无回滚（同 BUG-05） |
| ENG-04 | P2     | `use-budget-allocation-edit.ts` 已是死代码           | 全仓库无组件引用（实际使用的是 `budget-allocation-dialog.tsx`），应直接删除整个文件，而非仅清理其中的 `reservedDraft` |
| ENG-05 | P2     | 移出 project roster 未禁用 Key                       | `UpdateProject` 变更 `MemberIDs` 时仅调 `pruneMemberBudgets` 裁剪预算，未禁用该成员在项目下的 Platform Key |

---

## 六、PRD 差距（阻塞上线）

来自 `PRD-差距分析.md` P0 项：

| 编号     | 差距                                     | 关联  |
| -------- | ---------------------------------------- | ----- |
| PRD-P0-1 | Gateway 自定义 `blockMessage` 文案未消费 | US-08 |
| PRD-P0-2 | Anthropic `/v1/messages` 路径未支持      | US-12 |
| PRD-P0-3 | Gateway 超限应返回 HTTP 429（非 403）    | US-12 |
| PRD-P0-4 | 预警/阻断事件 `Priority` 硬编码为 normal，实际只投递站内通知 | US-08 |

---

## 七、待定设计决策

| 编号        | 问题                          | 来源                                |
| ----------- | ----------------------------- | ------------------------------------ |
| DECISION-01 | 预算总额度是否与钱包余额挂钩  | plan/未实现与优化方向.md            |
| DECISION-02 | JWT 是否添加 `iss`/`aud` 声明 | 多租户审计                           |

---

## 八、修复优先级建议

### P0 — 数据安全（上线前必须修）

BUG-03、BUG-06、ENG-03、PRD-P0-1/2/3/4

### P1 — 可靠性（版本内修）

BUG-04、BUG-05、DEBT-05、DEBT-07、DEBT-12、ENG-02

### P2 — 代码质量（迭代优化）

DEBT-02、DEBT-03、DEBT-04、DEBT-06、DEBT-09、ENG-04、ENG-05

### P3 — 代码卫生

LOW-01 ~ LOW-06
