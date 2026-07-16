# 后端问题清单（已验证）

> **快照日期**：2026-07-16  
> **来源**：`docs/REVIEW.md` + 多租户审计 + 工程收口 + 本轮代码审查  
> **图例**：🔴 严重 · 🟠 中等 · 🟡 低优 · ✅ 已修复（留档）

---

## 一、已修复项（本轮验证已关闭）

| 原编号 | 问题 | 修复方式 |
|--------|------|----------|
| REVIEW#7 | LRU Cache `touch()` O(n) 线性扫描 | `container/list` + map，O(1) |
| REVIEW#12 | 无 HTTP 层 rate limiting | `RateLimitTenant` + `RateLimitLoginPaths` 中间件 |
| REVIEW#15 | 缺乏结构化 access log | `access_log.go`（method/path/status/latency/company_id） |
| REVIEW#6 | CORS 缺少 `Access-Control-Max-Age` | 已设置 86400 |
| SEC-04 | Per-tenant API rate limiting | 同 REVIEW#12 |
| PERF-02 | LRU O(n) touch | 同 REVIEW#7 |
| PERF-01 | Authz 每请求查 revision | revisionCache TTL 5s |
| PERF-04 | pgxpool 默认连接池 | `DB_MAX_CONNS`/`DB_MIN_CONNS` 可配 |

---

## 二、🔴 严重（数据一致性 / 安全）

### BUG-01 — Company ID 生成竞态条件

**文件**：`internal/domain/company/service_create.go`  
**现状**：事务外 `List()` 全量读取 → 手动 `max(ID)+1`。两个并发 `CreateCompany` 得到同一 `nextID`，导致主键冲突。

```go
companies, err := s.store.Company().List(ctx)
var nextID int64 = 1
for _, t := range companies {
    if t.ID >= nextID { nextID = t.ID + 1 }
}
// ... 之后才进入 WithTx
```

**建议**：在事务内使用 PostgreSQL `SELECT MAX(id)+1 ... FOR UPDATE` 或 SERIAL。

---

### BUG-02 — Keys 域 Entity ID 使用 `UnixMilli()` 无随机后缀

**文件**：`internal/domain/keys/platform_key_create.go`、`approval.go`  
**现状**：`fmt.Sprintf("plk-%d", time.Now().UnixMilli())`，同毫秒并发 → ID 碰撞。  
**对比**：`budget/service.go` 的 `generateBudgetID` 已加 `rand.Read` 4 字节。

**建议**：统一使用 `generateBudgetID` 模式或 ULID。

---

### BUG-03 — `CreatePlatformKey` 无事务、无补偿

**文件**：`internal/domain/keys/platform_key_create.go`  
**现状**：

1. 事务外 `LoadBudgetContext` 读取 → 验证预算
2. `SetPlatformKeys` 写入 DB（无锁、无事务）
3. `syncPlatformKeyCreate` 调外部 API
4. 外部 API 失败 → DB 已有脏数据（status=active 但外部不存在），**无补偿**

**对比**：`ApproveApproval` 有 `compensateFailedKeyApproval`。

---

### BUG-04 — `ApproveApproval` 预算检查 TOCTOU

**文件**：`internal/domain/keys/approval.go`  
**现状**：`reservedPool` 在事务外读取并校验，事务内未重新验证。两个并发审批可能同时通过但实际余额不足。

---

### BUG-05 — `UpdatePlatformKey` / `TogglePlatformKey` 无事务

**文件**：`platform_key_update.go`、`platform_key_actions.go`  
**现状**：读-验证-改-写-调外部 API 全程无事务/无锁。并发更新同一 key → lost-update 或状态不一致（外部成功但本地写失败）。

---

### BUG-06 — Auth handler 绕过 body 大小限制

**文件**：`internal/http/handler/auth/handler.go`  
**现状**：`Login` 和 `AcceptInvite` 直接 `json.NewDecoder(r.Body).Decode()`，未用 `httputil.DecodeJSON`（限制 1MB）。攻击者可发送大 payload 消耗内存。

---

### BUG-07 — Budget Rejection 路径无事务

**文件**：`internal/domain/budget/approvals.go`  
**现状**：Rejection 分支不在事务中。并发 approve + reject 同一条可能导致状态不确定。

```go
} else {
    items, err := s.store.Budget().BudgetApprovals(ctx)
    // ... 检查 status == "pending"
    s.store.Budget().UpdateBudgetApproval(ctx, id, input.Status, ...)
}
```

---

## 三、🟠 中等（Hacky / 技术债）

### DEBT-01 — `buildServiceRegistry` 使用 `panic` 处理运行时错误

**文件**：`internal/app/registry.go`  
**现状**：Gateway wire 或 Identity wire 失败 → `panic(err)`，绕过 graceful shutdown。  
**建议**：改为返回 `(ServiceRegistry, error)`。

---

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

### DEBT-04 — `SimulateDelay` 散布在 domain 方法中

**文件**：`keys/*.go`、`budget/projects.go`、`budget/approvals.go`  
**现状**：每个 Create/Update/Approve 操作内嵌 300–500ms `delayer.Wait`。生产 `SimulateDelay=false` 时为 no-op，但代码噪音大、新人易误解。  
**建议**：如需 demo 延迟，在 HTTP 中间件统一处理。

---

### DEBT-05 — Gateway 所有 precheck 失败统一返回 403

**文件**：`internal/domain/gateway/gateway_service.go`  
**现状**：key 不存在、预算耗尽、模型不允许 → 全返 `403 "request rejected"`。调用者无法区分失败原因。  
**PRD 要求**：401（key 无效）/ 403（模型不允许）/ 429（超限）/ 502（供应商不可用）。

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

### DEBT-08 — Approval 补偿逻辑吞掉原始 sync 错误

**文件**：`internal/domain/keys/approval.go`  
**现状**：

```go
if compErr := s.compensateFailedKeyApproval(...); compErr != nil {
    return compErr  // 原始 sync error 丢失
}
return err
```

补偿失败时调用方只看到补偿错误，丢失 sync 根因。  
**建议**：`fmt.Errorf("sync: %w; compensate: %v", err, compErr)`。

---

### DEBT-09 — Store 层每次调用 new Repo 实例

**文件**：`internal/store/postgres/postgres.go`  
**现状**：`Company()`、`Invite()`、`Billing()` 等每次返回新 `struct{db pool}`，不复用。  
**建议**：启动时缓存（如 `domain` 分组已做）。

---

### DEBT-10 — `keys.ListApprovals` / `ApprovalBudgetCheck` 全量加载

**文件**：`internal/domain/keys/service.go`  
**现状**：加载全部 approvals 后内存过滤/遍历查 ID。数据量增长后性能下降。  
**建议**：Store 层加 `ApprovalByID`、`ApprovalsByStatus` 查询。

---

### DEBT-11 — Ingest Worker 无显式 stop 确认

**文件**：`internal/app/compose_worker.go`  
**现状**：`backgroundWorkers.stop()` 只停 River；Ingest Worker 靠 context cancel 退出，无 WaitGroup 等待。shutdown 期间 in-flight 可能中断。  
**建议**：加 `Worker.Stop()` + `sync.WaitGroup`。

---

### DEBT-12 — Keys 域 `SetPlatformKeys` / `SetApprovals` 全量覆盖

**文件**：`internal/domain/keys/*.go`  
**现状**：全量 load → 内存修改 → 全量写回。并发下可能 lost-update。Budget 域已用 `AcquireBudgetLock` + 事务保护，但 Keys 域无。  
**建议**：短期加 advisory lock + 事务；长期改增量 API（InsertKey、UpdateKeyStatus）。

---

## 四、🟡 低优先级

| 编号 | 文件 | 问题 |
|------|------|------|
| LOW-01 | `keys/handler.go` | 写路由使用 `ReadRoutes` 命名（语义不明） |
| LOW-02 | `keys/handler.go` | `ProviderCreate` 返回 200 而非 201 Created |
| LOW-03 | `rebalance.go` | `active := mappings[:0]` 原地过滤修改输入 slice |
| LOW-04 | `config.go` | `SessionTTLSec=0` 无 validate 保护 |
| LOW-05 | `middleware/cors.go` | 无 origin 匹配时仍设置 `Allow-Methods/Headers`（应条件设置） |
| LOW-06 | `authz/cache.go` | revision 变更时旧条目仅靠 LRU 淘汰，无主动 invalidate |
| LOW-07 | `defaultCompanyRoles` | 生成角色时用 `panic(err)` 而非返回 error |

---

## 五、工程收口（跨域/联调项）

以下来自 `docs/工程收口.md`，与本文重点互补：

| 编号 | 优先级 | 项 | 说明 |
|------|--------|-----|------|
| ENG-01 | P0 | NewAPI/Gateway 联调验收 | `pnpm verify:gate` + 真实 full-stack |
| ENG-02 | P0 | 通知失败不可观测 | `Send()` 失败仅 log，调用方不感知 |
| ENG-03 | P0 | Update Remote-first | `platform_key_update.go` 先写 DB 再 Remote（同 BUG-05） |
| ENG-04 | P2 | `use-budget-allocation-edit.ts` 残留 `reservedDraft` | 前端应删 |
| ENG-05 | P2 | 移出 project roster 未禁用 Key | `UpdateProject` 仅 prune budgets |

---

## 六、PRD 差距（阻塞上线）

来自 `docs/PRD-差距分析.md` P0 项：

| 编号 | 差距 | 关联 |
|------|------|------|
| PRD-P0-1 | Gateway 自定义 `blockMessage` 文案未消费 | US-08 |
| PRD-P0-2 | Anthropic `/v1/messages` 路径未支持 | US-12 |
| PRD-P0-3 | Gateway 超限应返回 HTTP 429（非 403） | US-12 |

---

## 七、待定设计决策

| 编号 | 问题 | 来源 |
|------|------|------|
| DECISION-01 | 预算总额度是否与钱包余额挂钩 | `docs/todos/budget-wallet-limit.md` |
| DECISION-02 | JWT 是否添加 `iss`/`aud` 声明 | 多租户审计 SEC-07 |

---

## 八、修复优先级建议

### P0 — 数据安全（上线前必须修）

BUG-01、BUG-02、BUG-03、BUG-05、BUG-06、PRD-P0-1/2/3

### P1 — 可靠性（版本内修）

BUG-04、BUG-07、DEBT-01、DEBT-05、DEBT-07、DEBT-11、ENG-02

### P2 — 代码质量（迭代优化）

DEBT-02、DEBT-03、DEBT-04、DEBT-08、DEBT-09、DEBT-10、DEBT-12

### P3 — 代码卫生

LOW-01 ~ LOW-07
