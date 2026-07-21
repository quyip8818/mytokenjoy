# TokenJoy 后端架构评审

> 约束前提：单实例部署，无需考虑多副本。

## 1. 系统边界

```
SDK 请求 (sk-xxx)                          浏览器 (session cookie)
       │                                         │
       ▼                                         ▼
┌────────────────────────────────────────────────────────────────┐
│                   TokenJoy Backend (:8080)                      │
│                                                                │
│  /v1/*   Gateway: rateLimit → precheck → reverseProxy          │
│  /api/*  管理 API: companyResolve → rateLimit → authz → handler │
│  River   后台 Job 队列 (PostgreSQL backed)                       │
│  Worker  pricingSync (goroutine)                               │
└────────────────────────────────────────────────────────────────┘
              │ proxy                     │ admin API
              ▼                           ▼
┌────────────────────────────────────────────────────────────────┐
│                     NewAPI (:3000)                              │
│  /v1/*  channel routing → upstream LLM                         │
│  Admin: Token/Channel/User/Group/Option CRUD                   │
│  SOT: ModelRatio / CompletionRatio (定价)                       │
│  产出: consume_log (用量原始记录)                                 │
└────────────────────────────────────────────────────────────────┘
```

**核心分工**: TJ 做管控 (鉴权/限额/计费/预算)，NewAPI 做执行 (routing/upstream 调用/用量产出/定价存储)。

## 2. 分层结构

| 层 | 包 | 职责 |
|----|-----|------|
| 入口 | `cmd/server` | config → app.New → http.Server + graceful shutdown |
| 组合根 | `internal/app` | DI 装配: infra → domain services → registry → router + workers |
| Domain | `internal/domain/*` | 业务逻辑，每个 domain 定义自己的 narrow Store interface |
| HTTP | `internal/http` | chi router + handler + middleware，只依赖 domain Service interface |
| Adapter | `internal/adapter` | domain port → infra 桥接 (enqueuer/budget ops/etc.) |
| Identity | `internal/identity` | authn (session/credential) + authz (permission check) |
| Infra | `internal/infra` | River/Redis/notification channels/scheduler |
| Integration | `internal/integration` | 外部 HTTP client (newapi/platform/datasource) |
| Store | `internal/store` | Repository 接口定义 + postgres 实现 |
| Pkg | `internal/pkg` | 纯函数工具 (budget calc/clock/model catalog/unit conversion) |

**依赖方向**: domain 只依赖 interface (Store/Port)，不依赖 integration/infra 实现。

## 3. 关键数据流

### 3.1 Gateway 热路径

```
请求 → parseBearerSecret → hashKey
  → Redis token-bucket rate limit (fail-open)
  → readBody (max 4MB) → parseModel
  → PrecheckCache.Get(keyHash):
      LRU hit → 直接拿 PrecheckContextRow
      miss → PG JOIN (platform_keys + company + allowlist)
  → EvaluateAt:
      company active? wallet > 0? key active? key not expired? model in allowlist?
  → budgetRemainCheck: Redis GET combined_key_remain (fail-open)
  → httputil.ReverseProxy → NewAPI
```

### 3.2 Ingest (用量计费)

```
consume_log (NewAPI) → TJ ingest job:
  1. logStore.GetConsumeLog → 拿到 tokenID + quota
  2. PlatformKeyMapping 反查 → companyID + departmentID + memberID
  3. BuildCallSettledEntry: Amount = Raw.Quota (直接透传，不做价格换算)
  4. TX (company row lock):
     - idempotency check
     - ConsumeLotsLocked (从 lot 扣减 quota)
     - InsertSegments (ledger)
     - IncrementConsumedBatch (budget_consumed 多轴更新)
     - DecrementBatch (combined_key_remain)
     - EnqueueAfterIngest (side-effect jobs in TX)
  5. Post-commit (best-effort):
     - Redis refresh combined_key_remain
     - budget alert check
     - overdraft notification
```

### 3.3 Key 同步

```
keys.CreatePlatformKey
  → newapisync.SyncCreatePlatformKey
    → upsertPendingMapping (PG: status=pending)
    → River job (newapi_sync / create_key)
      → TrySyncCreate:
         EnsureGroup → CreateToken → persistSecret → status=synced
         [失败: RollbackFailedCreate 补偿]
```

Update/Revoke/Rotate 类似: 本地状态先变更 → River job 同步到 NewAPI。

### 3.4 定价

- **SOT**: NewAPI 的 ModelRatio / CompletionRatio option
- **写入**: 创建/更新自定义模型 → `UpsertModelRatio` 写入 NewAPI; PricingSync Worker 定时从 Platform 拉取官方定价 → `UpdateOption` 写入 NewAPI
- **读取展示**: `ListModelsWithPricing` 实时查 NewAPI pricing API → `PriceFromRatio` 转换为 元/1M tokens 展示
- **计费单位**: NewAPI consume_log 的 quota 字段 (已按 ratio 换算)，TJ 直接透传不做价格计算
- **本地 models 表**: 只存 provider/type/name/endpoint/capabilities 等元数据，**不存 price**

## 4. Worker 架构

River (PostgreSQL job queue) 注册的 workers:

| Worker | 触发 | 作用 |
|--------|------|------|
| ingest | 每条 consume_log | 用量计费核心 |
| ingest_reconcile | 定时/手动 | 批量补偿未处理的 log |
| overrun | ingest 判断 remain≤0 | 多层预算评估 → 禁用 key |
| rebalance | budget 变更 | 重算 combined_key_remain |
| newapi_sync | key CRUD | TJ→NewAPI token 同步 |
| org_sync | 组织变更 | 远程数据源同步 |
| budget_reconcile | 定时 | 全量重算 budget_consumed |
| dashboard_project | ingest 后 | 仪表盘投影更新 |
| dashboard_reconcile | 定时 | 投影全量修复 |
| watchdog | periodic | scheduler L2 → bulk enqueue |
| notification_delivery | 事件触发 | 多渠道通知分发 |

额外: `pricingsync.Worker` 作为 goroutine 运行，定时从 Platform 拉取定价写入 NewAPI。

## 5. 做得好的地方

1. **Narrow Store interface** — 每个 domain 只声明自己需要的 repository 方法 (如 `models.Store` 只要 `Models()` + `Org()`)，避免 god-interface 耦合。
2. **Port/Adapter 方向正确** — domain 依赖 `adminport.Port` 接口，`integration/newapi.Client` 是实现。adapter 层做胶水。
3. **Ingest 事务设计严谨** — company lock 序列化 + idempotency key + 事务内入队 side-effects，保证 exactly-once 语义。
4. **Gateway fail-open** — Redis 不可达时 rate limit 和 budget check 都 fail-open，不会因缓存故障拒绝正常请求。
5. **Precheck cache 主动失效** — key 变更 → `InvalidateByKeyID`; 公司冻结 → `InvalidateCompany`。不纯靠 TTL。
6. **NewAPI Token 设 unlimited** — TJ 侧管预算，NewAPI 侧不做限额。职责不重叠。
7. **River + PG** — job 和业务数据在同一 DB 内，事务性入队天然 exactly-once，无需分布式事务。

## 6. 可优化点

> 按 重要性 × 实现难度 排列。重要性 = 数据正确性 > 可观测性 > 性能 > 代码清晰度。

### #1 CreateModel 写 pricing 的 fire-and-forget

**重要性**: 高 (数据正确性) | **难度**: 低 (改几行错误处理)

`CreateModel` 写入本地 DB 后，`UpsertModelRatio` 写 NewAPI 失败只 `slog.Warn`，模型照常创建。

**后果**: 模型存在但 NewAPI 侧无定价 → 该模型的 consume_log quota 按默认 ratio 计算 (可能为 0) → 用量不计费。

**修复**: 失败时回滚本地创建或返回 error。`UpdateModel` 已对 pricing 写失败返回 error，`CreateModel` 应保持一致。

---

### #2 缺少 Gateway 可观测性 metrics

**重要性**: 高 (运维必需) | **难度**: 低 (加几个 counter/histogram)

Gateway 热路径只有 `slog.Info` 级别拒绝日志。无 Prometheus/OTEL metrics。出问题时没有数据可查。

**建议**: 添加:
- `gateway_requests_total{result=allowed|rejected|rate_limited}`
- `gateway_precheck_latency_seconds` (histogram)
- `gateway_cache_hit_total` / `gateway_cache_miss_total`
- `gateway_upstream_duration_seconds`

---

### #3 models 的 InputPrice/OutputPrice 字段在 struct 中但不在 DB 中

**重要性**: 中 (开发者认知负担) | **难度**: 低 (加注释或拆 DTO)

`types.ModelInfo` 有 `InputPrice`/`OutputPrice` 字段，DB schema 没有对应列。`InsertModel`/`UpdateModel` 的 PG 实现会忽略它们。

**问题**: 新开发者会以为这些字段持久化了。`CreateModel` 中先 assign 再 insert (price 实际丢失)，再写 NewAPI — 如果 NewAPI 也失败 (见 #1)，price 就彻底丢了。

**建议**: 字段加 `// transient: from NewAPI pricing, not persisted` 注释。或拆为 `ModelInfoWithPricing` DTO。

---

### #4 UpdateToken 每次 GET + PUT (2x HTTP)

**重要性**: 中 (延迟翻倍) | **难度**: 中 (需加快照字段 + 维护逻辑)

NewAPI PUT 全量替换。TJ 每次 update 先 GET 当前值 merge 再 PUT。

**影响**: Key 启禁/group 变更延迟 2x。GET 和 PUT 之间有并发更新窗口。

**建议**: 在 `platform_key_mappings` 冗余 Token 快照 (name/group/status/expired_time)，update 时用本地快照 merge，省掉 GET。sync 成功后刷新快照。

---

### #5 Gateway body 全量读取只为 parse model 字段

**重要性**: 中 (热路径性能) | **难度**: 中 (需要 TeeReader/Peek 正确实现)

`readAndRestoreBody` 将 body 全量读入内存 (≤4MB) 再 `json.Unmarshal` 提取 model 字段。

**影响**: 每请求一次 ≤4MB 内存分配+拷贝。典型请求几 KB 无所谓，但 embedding 大 payload 时 GC 压力可观。

**建议**: `bufio.Reader.Peek(1024)` 只读前 1K bytes 手工 parse model 字段 (几乎总在 JSON 开头)，配合 `io.MultiReader` 恢复给 proxy。当前有 `MaxBytesReader(4MB)` 保护，不紧急。

---

### #6 Ingest company lock 粒度

**重要性**: 中 (吞吐瓶颈) | **难度**: 高 (需重构事务边界)

每笔 ingest TX 都 `LockForUpdate(companyID)`，同一公司所有 ingest 严格串行。

**当前状态**: 单实例 + River 并发有限 → 实际冲突率低。

**风险**: 某公司流量暴涨 (batch 跑任务) 时 ingest 积压。

**建议**: 暂不改。遇到瓶颈时考虑 batch 模式 (单次 TX 处理多条 log) 或拆分 lot 消费和 budget_consumed 为两段事务。

---

### #7 PrecheckCache negative cache penetration

**重要性**: 低 (有 rate limit 前置保护) | **难度**: 中

对"key 不存在"缓存 30s。攻击者用大量随机 key hash 打 gateway，每个新 hash 穿透到 PG。

**建议**: 低优先级。rate limit 已经前置挡掉绝大多数。如需加固，可在 PG 查询前加 bloom filter。

---

### #8 Overrun 评估的全局 advisory lock

**重要性**: 低 (overrun 是低频操作) | **难度**: 低

`evaluateOverrun` 用全局 advisory lock 序列化所有 overrun 评估。

**影响**: 大量 key 同时超限时串行处理。

**建议**: 暂不改。如果频繁触发，改为 per-company advisory lock。

## 7. 总结

这是一套设计干净的分层架构。核心优势在于:
- 管控面/执行面职责分离清晰
- Domain 间通过 exported interface 协作，无循环依赖
- Ingest 事务设计保证了计费正确性

**立即可做**: #1 (几行代码修复计费正确性风险) + #2 (Gateway metrics，运维基础设施)
**短期**: #3 (注释/DTO 清理) + #4 (UpdateToken 快照优化)
**等遇到再说**: #5 #6 #7 #8
