# Backend 代码质量评估报告

> **范围**: 除安全性和测试质量外的全面扫描  
> **日期**: 2026-07-07  
> **覆盖维度**: 架构设计、错误处理、数据完整性、并发安全、API 设计、可观测性、配置与运维

---

## 一、架构与设计

### 好的方面

- Handler 层非常薄，只做 HTTP 关注点（decode → call service → write response）
- DI 是手动构造器注入，清晰无框架
- Store 接口采用 facade + 细粒度 sub-repo 组合，ISP 基本满足

### 问题

| 严重度     | 问题                                                                                                        | 位置                                                            |
| ---------- | ----------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------- |
| **Medium** | `domain/relay/gateway_service.go` 是完整的 HTTP handler（含 `httputil.ReverseProxy`），违反分层             | `internal/domain/relay/gateway_service.go`                      |
| **Low**    | `domain/usage/ingest_outcome.go` 在 domain 层引用 `net/http` 状态码常量                                     | `internal/domain/usage/ingest_outcome.go:5`                     |
| **Low**    | Domain 层直接依赖 `infra/notification`、`infra/permission`、`integration/newapi` 等具体包，而非自身定义接口 | `domain/org/core/deps.go`, `domain/budget/`, `domain/relay/` 等 |
| **Low**    | `OrgRepository` 有 19 个方法，覆盖 member/role/integration/sync/field mapping 多个子域                      | `internal/store/store.go`                                       |

### 建议

1. 将 `gateway_service.go` 的反向代理逻辑移到 `http/handler/relay/` 或独立 gateway 包
2. Domain 对外部依赖定义本地接口（如 `domain/relay/TokenManager` interface），由 app 层注入实现
3. 拆分 `OrgRepository` 为 `MemberRepository` + `IntegrationRepository`

---

## 二、错误处理

### 好的方面

- 统一的 `DomainError{Status, Message, RetryAfter}` 类型体系
- 标准化的 sentinel 工厂：`domain.NotFound()`, `domain.Validation()` 等
- `WriteError` 安全处理：DomainError 返回业务信息，其他错误一律 500 "Internal server error"

### 问题

| 严重度     | 问题                                                                                       | 位置                                  |
| ---------- | ------------------------------------------------------------------------------------------ | ------------------------------------- |
| **Medium** | `billing/service.go` 中 `UpdateRechargeStatus` 失败被静默吞掉，订单可能卡在错误状态        | `billing/service.go:142,147,155`      |
| **Medium** | `rebalanceAxis` 错误被丢弃，budget 分配可能静默失败                                        | `billing/service.go:168`              |
| **Medium** | `types/credential.go` 12 处校验错误用 bare `fmt.Errorf`，到达 handler 会变成 500 而非 400  | `internal/domain/types/credential.go` |
| **Low**    | 79 处 `fmt.Errorf` 未使用 `%w`，断裂了 error chain（尤其 `relay/precheck.go` 全部 10+ 处） | 分散各处                              |
| **Low**    | `WriteError` 对非 DomainError 的 500 不做任何日志记录，问题难以排查                        | `httputil/write.go:38`                |
| **Low**    | `WriteJSON(w, status, nil, err)` 中显式传入的 status code 在 err 非 nil 时被忽略           | `handler/auth/handler.go:51,84`       |
| **Info**   | Domain/handler 层不处理 `context.Canceled`，客户端断连后逻辑仍运行到底                     | 系统性                                |

### 建议

1. `billing/service.go` 的关键状态更新失败应返回 error 或至少 log.Error
2. `credential.go` 的校验错误改用 `domain.Validation(msg)`
3. `WriteError` 增加对 500 的 `slog.Error` 记录（含 request ID）
4. 统一要求 `fmt.Errorf` 包含 `%w`（可加 linter 规则 `errorlint`）

---

## 三、数据完整性

### 好的方面

- Usage ingest 的 Insert + Apply + enqueue 全部在单个事务内
- Counter 更新用 `SET consumed = consumed + $3` 保证原子性
- Worker 的 outbox 模式保证最终一致

### 问题

| 严重度     | 问题                                                                                                               | 位置                                          |
| ---------- | ------------------------------------------------------------------------------------------------------------------ | --------------------------------------------- |
| **Medium** | `budget/service.go` 的 `UpdateNode`/`UpdateMemberQuota` 执行 read-validate-write 无事务保护，PG 下存在 lost update | `internal/domain/budget/service.go:53-86,114` |
| **Medium** | Memory store `WithTx` 在 snapshot 后释放锁再执行 fn，并发 WithTx 回滚可能覆盖已提交数据                            | `internal/store/memory/tx.go:44-53`           |
| **Low**    | `ComputeRemainQuotaCNY` 读取 budget 后再 UpdateToken，存在 TOCTOU 间隙（rebalance worker 最终修正）                | `internal/domain/relay/quota.go`              |
| **Low**    | Relay lifecycle 状态转换无前置状态校验，pending 状态的 mapping 可被并发 update                                     | `internal/domain/relay/lifecycle_ops.go`      |

### 建议

1. `budget/service.go` 的多步写操作包裹在 `WithTx` 中
2. Memory store `WithTx` 应持锁到 fn 执行完毕（或改用 COW 语义）
3. Relay lifecycle 增加状态守卫：`if mapping.Status != synced { return ErrInvalidState }`

---

## 四、并发与性能

### 好的方面

- Worker goroutine 正确响应 context cancellation，无泄漏风险
- `ingestmetrics` 对高频计数器用 `atomic.Int64`，低频字段用 RWMutex
- 单线程 worker tick 避免了 intra-worker race

### 问题

| 严重度   | 问题                                                                                           | 位置                                      |
| -------- | ---------------------------------------------------------------------------------------------- | ----------------------------------------- |
| **Low**  | Memory store 单把 RWMutex 保护所有 company 数据，任何写操作阻塞全部读                          | `internal/store/memory/store.go:14`       |
| **Low**  | WalletService 缓存过期时存在 thundering herd：多个 goroutine 同时 miss 后都请求外部服务        | `internal/domain/company/wallet.go:51-69` |
| **Low**  | LRU cache 的 `touch` 方法在锁内做 O(n) 线性扫描                                                | `internal/identity/authz/cache.go:67`     |
| **Info** | 无 PG store 的 `SELECT FOR UPDATE`，依赖原子增量（对 counter 足够，对 read-modify-write 不够） | 系统性                                    |

### 建议

1. Memory store 如需支撑高并发测试，可改为 per-repo 或 per-company 锁
2. WalletService 加 `singleflight.Group` 去重并发外部请求
3. LRU cache 考虑切换到 `container/list` 实现 O(1) touch

---

## 五、API 设计

### 好的方面

- 统一的 `/api/{domain}/{resource}` 命名
- 一致的分页类型 `PageResult[T]{Items, Total, Page, PageSize}`
- 错误响应结构统一 `{"message": "..."}`

### 问题

| 严重度     | 问题                                                                                                   | 位置                                   |
| ---------- | ------------------------------------------------------------------------------------------------------ | -------------------------------------- |
| **Medium** | 多个列表接口无分页（keys/provider, keys/platform, billing/recharge-records），数据量增长后会 OOM       | 各 handler                             |
| **Low**    | 无标准响应信封，客户端只能靠 HTTP status 区分成功/失败，不利于统一 error handling                      | `internal/http/response/json.go`       |
| **Low**    | 幂等性只有 billing recharge 实现，其他写接口（create member/role/key）无防重                           | 系统性                                 |
| **Low**    | Platform handler 绕过 `httputil.DecodeJSON`，错误信息不一致（"Bad request" vs "Invalid request body"） | `handler/platform/handler.go:44,72,99` |
| **Info**   | Billing handler 路由注册风格与其他 handler 不同（flat vs `r.Route`）                                   | `handler/billing/handler.go:28-35`     |

### 建议

1. 所有列表接口加分页，对未分页的接口至少加 `LIMIT 1000` 硬上限
2. 写操作考虑通用幂等中间件（基于 `Idempotency-Key` header）
3. 统一使用 `httputil.DecodeJSON`

---

## 六、可观测性

### 好的方面

- 结构化日志 `slog` + JSON 输出
- Request ID 中间件生成并写入响应头
- Audit trail 存在（虽范围有限）

### 问题

| 严重度     | 问题                                                                                    | 位置                               |
| ---------- | --------------------------------------------------------------------------------------- | ---------------------------------- |
| **Medium** | Request ID 从未被下游使用——无 getter 函数、不出现在日志中、不转发给 relay，本质是死代码 | `middleware/requestid.go`          |
| **Medium** | 无请求级 access log（method/path/status/duration），问题排查困难                        | 系统性                             |
| **Medium** | 健康检查 `/healthz` 永远返回 200，不验证 DB/relay 连通性，k8s readiness probe 失效      | `handler/health.go`                |
| **Low**    | 无 Prometheus metrics 端点，无法对接标准监控栈                                          | 系统性                             |
| **Low**    | Audit trail 仅覆盖 platform 操作，租户内用户操作（key 创建、成员变更）无审计            | `domain/company/platform_audit.go` |
| **Info**   | 无 OpenTelemetry / 分布式 tracing                                                       | 系统性                             |

### 建议

1. 导出 `RequestIDFromContext()` 并在 slog 中间件中注入
2. 增加 access log 中间件（method, path, status, duration, request_id）
3. `/healthz` 验证 DB `Ping()` + relay 健康（可选 `/readyz` 分离）
4. 后续引入 Prometheus client 暴露关键指标（request count/latency, error rate, budget usage）

---

## 七、配置与运维

### 好的方面

- 优雅关停完整：signal → cancel worker → close DB → drain HTTP（10s timeout）
- 配置校验根据功能开关动态要求必填项
- `ReadHeaderTimeout: 5s` 防 slowloris

### 问题

| 严重度   | 问题                                                                      | 位置                        |
| -------- | ------------------------------------------------------------------------- | --------------------------- |
| **Low**  | 数值型配置（`WorkerPollIntervalSec`、PORT）无范围校验，0 或负值会导致异常 | `internal/config/config.go` |
| **Info** | 无 schema migration 工具集成（当前依赖内存 store，PG 只用于集成测试）     | 系统性                      |

---

## 优先级排序（Top 10 建议改进项）

| #   | 维度       | 建议                                             | 影响               |
| --- | ---------- | ------------------------------------------------ | ------------------ |
| 1   | 数据完整性 | `budget/service.go` read-modify-write 加事务     | 防止并发丢失更新   |
| 2   | 错误处理   | `credential.go` 校验错误改用 `domain.Validation` | 修复误返 500       |
| 3   | 可观测性   | 增加 access log 中间件 + request ID 注入         | 基本排障能力       |
| 4   | 错误处理   | `billing/service.go` 关键错误不再静默丢弃        | 防止订单状态不一致 |
| 5   | API 设计   | 无分页列表接口加硬上限                           | 防止 OOM           |
| 6   | 可观测性   | `/healthz` 验证 DB 连通性                        | 有效的存活检测     |
| 7   | 架构       | `gateway_service.go` 移出 domain 层              | 恢复分层清晰       |
| 8   | 错误处理   | `WriteError` 对 500 记录 slog.Error              | 未知错误可追踪     |
| 9   | 并发       | WalletService 加 singleflight                    | 防缓存击穿         |
| 10  | 数据完整性 | Relay lifecycle 增加状态守卫                     | 防状态机混乱       |
