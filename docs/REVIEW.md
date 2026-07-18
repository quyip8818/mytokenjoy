# Backend 代码审查报告

## 总体评价

整体架构设计良好，分层清晰（cmd → app → domain → store），依赖注入通过组合函数实现，代码风格一致。以下列出需要关注的问题，按严重程度排列。

---

## 严重问题 (Bugs / 潜在数据损坏)

### 1. Company ID 生成存在竞态条件 ✅ 已修复

> **已通过 UUID v7 迁移修复（2026-07-18）。** Company ID 现为事务内 `uuid.Must(uuid.NewV7())` 生成。

---

### 2. Entity ID 使用 `time.Now().UnixMilli()` 缺少随机后缀 ✅ 已修复

> **已通过 UUID v7 迁移修复（2026-07-18）。** 所有实体 ID 统一为 `uuid.Must(uuid.NewV7())`。

---

### 3. Approval 审批操作的补偿逻辑吞掉原始错误

**文件**: `internal/domain/keys/approval.go`

```go
_, err = s.syncPlatformKeyCreate(ctx, created, applicant.DepartmentID)
if err != nil {
    if compErr := s.compensateFailedKeyApproval(...); compErr != nil {
        return compErr  // 原始 sync 错误被丢弃
    }
}
return err
```

当补偿失败时，调用方只看到补偿错误，丢失了根因（sync 失败原因）。补偿成功时会正确返回原始 `err`。

**建议**: 使用 `errors.Join(err, compErr)` 或 `fmt.Errorf("sync failed: %w; compensate: %v", err, compErr)` 保留完整上下文。

---

### 4. `buildServiceRegistry` 中 panic 用于运行时错误

**文件**: `internal/app/registry.go`

```go
gw, err := wireGatewayService(cfg, i)
if err != nil {
    panic(fmt.Errorf("wire gateway service: %w", err))
}
authzSvc, credSvc, memberToken, platformToken, err := wireIdentity(cfg, i.store)
if err != nil {
    panic(err)
}
```

`buildServiceRegistry` 在 `newApp()` 调用路径上。如果配置变化导致 gateway 或 identity wire 失败（例如运行时密钥轮转），panic 会绕过 graceful shutdown。

**建议**: 让 `buildServiceRegistry` 返回 `(ServiceRegistry, error)`，错误向上传播。

---

## Hacky / 技术债

### 5. `Void` 响应返回 JSON `null` 而非 204 No Content

**文件**: `internal/http/response/json.go`

```go
func Void(w http.ResponseWriter) {
    JSON(w, http.StatusOK, nil) // 响应体是 "null\n"
}
```

对无返回值的写操作（Delete、Toggle 等），返回 `200 null` 不符合 REST 惯例。

**建议**: 改为 `w.WriteHeader(http.StatusNoContent)` 或返回 `{}`。

---

### 6. CORS 中间件缺少 `Access-Control-Max-Age`

**文件**: `internal/http/middleware/cors.go`

没有设置 `Access-Control-Max-Age`，导致浏览器每次跨域请求都发送 OPTIONS 预检，增加延迟。

**建议**: 对 OPTIONS 响应添加 `Access-Control-Max-Age: 86400` (或适当值)。

---

### 7. LRU Cache `touch()` 使用线性扫描

**文件**: `internal/identity/authz/cache.go`

```go
func (c *LRUCache) touch(key cacheKey) {
    for i, existing := range c.order {  // O(n) 扫描
        if existing == key {
            c.order = append(c.order[:i], c.order[i+1:]...)
            break
        }
    }
    c.order = append(c.order, key)
}
```

`maxSize` 默认 4096，每次 `Get`/`Put` 都做 O(n) 线性扫描 + 切片移动。在高并发下性能差。

**建议**: 使用 `container/list` 双向链表 + map 实现 O(1) LRU，或直接用 `hashicorp/golang-lru`。

---

### 8. `keys.ListApprovals` 全量加载后在内存过滤

**文件**: `internal/domain/keys/service.go`

```go
func (s *service) ListApprovals(ctx context.Context, tab, memberID string) ([]types.KeyApproval, error) {
    items, err := s.store.Keys().Approvals(ctx) // 加载全部 approvals
    // ... 内存 filter
}
```

同样 `ApprovalBudgetCheck` 也全量加载所有 approvals 只为找一条记录。数据量增长后会成瓶颈。

**建议**: 在 store 层添加 `ApprovalByID(ctx, id)` 和支持 status/member 筛选的查询方法。

---

### 9. `DecodeJSON` 传 nil ResponseWriter 给 MaxBytesReader

**文件**: `internal/http/httputil/decode.go`

```go
r.Body = http.MaxBytesReader(nil, r.Body, maxRequestBodySize)
```

根据 Go 源码，当 `w` 为 nil 时 MaxBytesReader 不会在超限时关闭连接（无法设置 `Connection: close` header）。客户端持续发送大 body 时，server 仍需读完或等连接超时。

**建议**: 在中间件层限制请求体大小，或传入实际的 `http.ResponseWriter`。

---

### 10. `SimulateDelay` / `delayer.Wait` 在生产代码路径中

**文件**: `internal/domain/keys/platform_key_create.go`, `approval.go`

```go
if err := s.delayer.Wait(ctx, 500*time.Millisecond); err != nil {
    return types.PlatformKey{}, err
}
```

每个 key 创建/审批操作都有 500ms 的人为延迟（即使 `SimulateDelay=false` 时 delayer 是 no-op）。这个模式散布在多个 service 方法中，阅读代码时容易误解。

**建议**: 如果只是 demo 用途，考虑用中间件统一处理，而非嵌入每个 domain 方法。

---

### 11. Ingest Worker 没有显式 stop 机制

**文件**: `internal/infra/ingest/worker.go`, `internal/app/compose_worker.go`

`Worker.Start()` 启动 goroutine 靠 context 取消退出。`backgroundWorkers.stop()` 只停 River：

```go
func (b *backgroundWorkers) stop(ctx context.Context) {
    if b.river != nil {
        _ = b.river.Stop(ctx)
    }
    // ingest worker 没有被显式 stop
}
```

虽然 context cancel 会间接停止 ingest worker，但没有等待其退出确认。shutdown 期间可能有 in-flight 作业被中断。

**建议**: 添加 `Worker.Stop()` 方法返回 channel 或使用 `sync.WaitGroup` 确认清理完成。

---

## 架构建议

### 12. 没有 HTTP 层 rate limiting

整个 API 层没有发现任何 rate limit 中间件。`DomainError` 支持 `TooManyRequests` 状态码，但没有实际限速逻辑。

**建议**: 至少对认证端点（`/api/session/login`）加 IP 级别限速，防止暴力破解。

---

### 13. AuthzCache 不按 companyID 隔离失效

`LRUCache` 的 key 包含 `{companyID, memberID, revision}`。但当某个公司的权限变更时（revision 变化），旧 revision 的缓存条目不会被主动驱逐，只能靠 LRU 淘汰。如果缓存满载，攻击者可以让缓存充满无效条目。

**建议**: revision 变更时按 companyID 批量失效旧条目，或设置 TTL。

---

### 14. store 层大量重复的 repo 实例化

**文件**: `internal/store/postgres/postgres.go`

每次调用 `Store.Company()`, `Store.Invite()` 等都 `new*Repo(s.pool)`，每次创建新实例。对比 `domain` 分组（`org`, `budget`, `keys`, `models`, `audit`）则是启动时实例化一次。

**建议**: 统一缓存所有 repo 实例（在构造 `Store` 时创建），避免不必要的内存分配。

---

### 15. 缺乏结构化 request logging 中间件

有 `Recover` 和 `RequestID` 中间件，但没有 access log / request duration 中间件。生产环境调试困难。

**建议**: 添加 structured access log 中间件，记录 method、path、status code、duration、request_id。

---

### 16. `response.JSON` 的 Encode 错误被忽略

**文件**: `internal/http/response/json.go`

```go
_ = json.NewEncoder(w).Encode(v)
```

如果 `v` 中含有无法序列化的字段（如 `chan`、`func`），encode 会静默失败，但 header 已经写入（200），客户端收到截断的 JSON。

**建议**: 先 `json.Marshal(v)` 检查错误，失败时写 500 响应（需在 WriteHeader 之前）。或使用 buffer 模式。

---

### 17. 事务内的 `SetPlatformKeys` / `SetApprovals` 全量覆盖

从代码看，`SetPlatformKeys`、`SetApprovals`、`SetMembers` 等似乎是全量写入模式（先加载所有→修改→全量写回）。这在并发场景下可能产生 lost-update 问题。

**建议**:

- 短期: 确保这些操作在事务内且使用 `SELECT ... FOR UPDATE` 锁定
- 长期: 改为增量操作（InsertApproval, UpdateApprovalStatus 等）

---

## 小问题

| 文件                 | 问题                                                                                                                                                 |
| -------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------- |
| `config.go`          | `SessionTTLSec` int 类型，设为 0 时 issuer 默认 86400，但 config 层没有 validate 这个值                                                              |
| `middleware/cors.go` | 无 origin 匹配时仍设置 `Allow-Methods` 和 `Allow-Headers`（应该只在 origin 匹配时设置）                                                              |
| `response/json.go`   | `Void()` 命名不直观，实际返回 `null` body                                                                                                            |
| `keys/handler.go`    | `ProviderCreate` 返回 200 而非 201 Created                                                                                                           |
| `compose_worker.go`  | `var _ = (*pgxpool.Pool)(nil)` 无用的编译检查残留                                                                                                    |
| `app.go`             | `closers []func()` - cancel 在 stop workers 之前调用，context 已取消后 `bgWorkers.stop(context.Background())` 用了新 context，逻辑正确但顺序暗含假设 |

---

## 做得好的地方

- Domain error → HTTP status 映射干净统一
- WithTx 事务模式简洁，txStore 实现所有 repo 接口确保事务一致性
- 配置验证全面，环境差异处理得当
- Seed 数据分 minimal/demo 模式
- Recover 中间件防止 panic 传播
- Session token 使用 HMAC-SHA256，random session ID 防 replay
- 密码使用 bcrypt，认证失败信息统一（不区分用户不存在和密码错误）
- Gateway body size limit 和 request body limit 都有

---

## 优先级建议

1. **P0 (尽快修)**: Company ID 竞态 (#1)、Entity ID 碰撞 (#2)
2. **P1 (版本内修)**: panic 替换为 error (#4)、rate limiting (#12)、ingest worker graceful stop (#11)
3. **P2 (迭代优化)**: LRU 性能 (#7)、全量查询优化 (#8)、response 模式修正 (#5, #16)
4. **P3 (tech debt)**: logging 中间件 (#15)、CORS 优化 (#6)、repo 实例化 (#14)
