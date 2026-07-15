# Backend · v1 → Ingest 链路优化（性能 + Lag）

> **定位**：**性能**（Gateway / Ingest 热路径延迟、吞吐、DB 往返、锁等待）与 **Lag**（`gateway_soft_*` 投影窗，§10）分章写清；二者不少改动双赢。  
> **范围**：Gateway `/v1` 预检 + Ingest 入账 + 抢同一公司锁的背景任务（投影、reconcile）。  
> **非本文**：产品预算规则细则、完整监控体系建设（执法 SLA 指标定义见 [架构终态设计.md](./架构终态设计.md) §14；入账机制见 [Backend-Ingest架构.md](./Backend-Ingest架构.md)）。

---

## 1. 性能结论（先看这个）

整条链路里，**真正值得优化的性能瓶颈**就三类：

| 段 | 现在慢在哪 | 优化后大概能怎样 |
| --- | --- | --- |
| **Gateway 预检** | 大 body 全量读内存；预检 SQL 白名单扫两遍；无用列传输 | 预检 P99 再降几 ms～几十 ms（大 body 场景更明显） |
| **Ingest 入账** | Worker poll 默认 1s；此前每笔重复查 Models/Nodes（**已合并为单次**）；单线程顺序处理 | 入账吞吐可提升；单条处理更快 |
| **锁与背景任务** | ingest / 投影 / 改预算共用公司 advisory 锁，互相堵 | 高峰 ingest TPS 上限被锁抬高；拆锁或加快投影可释放吞吐 |

**已经做对的（别为了性能拆掉）**  
预检只 1 次 PG、不 JOIN `budget_consumed`、不调 NewAPI Admin——这是低延迟的基础。

**别用这些「假优化」**  
预检时现场算预算、调 NewAPI quota、入账事务里同步跑投影——都会**更慢**。

**建议动手顺序（按性价比）**  
最小集 **已落地**：`I1` · `G2+G3` · `I2` · `I5 拆锁` · `I6 并行` · `LISTEN/NOTIFY`。后续：`I3 缓存` → `G1 流式读 body`。

**Lag 与性能分开看**  
企业钱包（`wallet_remain`）和预算管控剩余（`combined_key_remain`）均在 ingest 事务里同步更新，预检读 PG **无投影 lag**。Key / 成员 / 项目轴的 `budget_consumed` 由 Projector 异步写入，用于 overrun 阻断和百分比预警（非预检热路径）。

---

## 2. 性能基线：每段花多少时间

### 2.1 Gateway 预检（不含 LLM 上游）

一次 `/v1` 请求在 TokenJoy 侧同步做的活：

```
读 body → SHA-256 Key → 1× Postgres → 内存判断 → [可选 1× Redis] → 反代 NewAPI
```

| 步骤 | 典型耗时量级 | 主要消耗 |
| --- | --- | --- |
| 读 body（最大 4MB） | 0～数十 ms | 磁盘/网络 IO、内存分配、GC |
| SHA-256 | < 1 ms | CPU |
| `LoadPrecheckContext` | 1～5 ms（视 PG 负载） | 1 次 RTT + SQL CPU |
| `Evaluate` | < 0.1 ms | CPU |
| Redis GET | 0.5～2 ms | 1 次 RTT |
| 反代 | **秒级** | 上游 LLM（不在本文优化范围） |

**目标**：把 TokenJoy 预检压在 **< 20ms P99**（PG 同 AZ、1k RPS 量级，压测校准）。

### 2.2 Ingest 单笔入账

```
读 log 行 → 主库多次查询（mapping、models×2、nodes×2…）→ 1× 事务（锁 + lot + ledger + 入队 job）
```

| 步骤 | 典型耗时量级 | 主要消耗 |
| --- | --- | --- |
| 等 Worker poll | **0～5s 默认** | 配置地板，不是 CPU |
| `GetConsumeLogByID` | 1～3 ms | log DB RTT + 行宽 |
| 事务外读主库 | 5～20 ms | **多次 RTT**（重复查表） |
| `WithTx` 入账 | 5～30 ms | 公司锁等待 + lot FIFO + INSERT |
| **合计（不含排队）** | 约 15～50 ms/条 | 读放大 + 锁 |

**目标**：单条处理 **< 50ms P95**（不含 poll 等待）；稳态 **> 200 条/s/实例**（多租户并行后更高）。

### 2.3 谁在和 ingest 抢吞吐

同一 `company_id` 上，`companies FOR UPDATE` 行锁互斥（Ingest 事务通过 `ConsumeLots` 获取）：

- Ingest 入账事务（已移除 advisory lock，改用行锁）
- 管理面改预算 / 项目（仍使用 `AcquireBudgetLock` advisory lock，与 ingest 不互斥）

`budget.Projector` 使用独立的 advisory lock，**不再阻塞 ingest**。

---

## 3. Gateway 预检：性能优化项

---

### G1 · 全量读 body，只为解析一个 `model` 字段

**慢在哪**  
`readAndRestoreBody` 把最多 **4MB** 全部读入 `[]byte`，再 `json.Unmarshal` 整个 body。预检只需要顶层 `model` 字符串。

**性能影响**  
- 大上下文请求：预检阶段 IO + 分配 + GC 明显变长。  
- 高 QPS：内存带宽和 GC 成为瓶颈，拖慢**所有**请求的预检 P99。

**怎么改**  
- 流式/限量读：只读前 8～32KB 解析 `model`；或 `json.Decoder` 读到 `model` 就停。  
- `/v1/models` 跳过 body 解析。

**预期收益**  
- 小 body（< 4KB）：几乎无感。  
- 大 body（100KB+）：预检 body 阶段耗时可降 **50%～90%**。

**代码**  
`domain/gateway/gateway_service.go` · `readAndRestoreBody` · `parseRequestModel`

**优先级**：中高（大 context 租户必做）

---

### G2 · 预检 SQL 对 `model_allowlist` 扫了两遍（**已落地**：CTE 单次扫描）

**慢在哪**  
`gateway_precheck_repo.go` 一条 SQL 里两个相关子查询：  
`EXISTS (...)` + `array_agg(...)`，都扫同一把 Key 的白名单。

**性能影响**  
- 单次查询 CPU 和 buffer 读翻倍。  
- Gateway QPS 高时，这条 SQL 成为 **PG 热点**，预检 RTT 随负载上升。

**怎么改**  
用 CTE 一次扫描，同时产出 `has_allowlist` 和 `allowlist_types[]`。

**预期收益**  
预检 SQL 耗时常见降 **10%～30%**（白名单越长越明显）。

**代码**  
`store/postgres/gateway_precheck_repo.go` · `loadPrecheckContextSQL`

**优先级**：**高**（改动小、收益稳）

---

### G3 · SELECT 了预检用不到的列（**已落地**：去掉 `newapi_wallet_user_id`、`gateway_soft_at`）

**慢在哪**  
例如 `newapi_wallet_user_id` 被查出但 `Evaluate` 不用。

**性能影响**  
每请求多传几十字节～几百字节；乘 QPS 后增加 PG→应用网络与解码开销。

**怎么改**  
删掉热路径无用列。

**预期收益**  
微量，但和 G2 同 PR 顺手做。

**优先级**：**高**

---

### G4 · 增大 pending 批大小 & 缩短 poll（Gateway 无直接关系，列在这里避免漏）

（此项实际在 Ingest，见 I1/I7）

---

### G5 · 可选：预检结果进程内短缓存（未实现，进阶）

**慢在哪**  
每次请求都打 1 次 PG，即使同一 Key 在 100ms 内连打 10 次。

**性能影响**  
极高 QPS、同一 Key 重复请求时，PG QPS = Gateway QPS。

**怎么改**  
按 `key_hash` 缓存 `PrecheckContext` 几十～几百 ms（TTL 极短），Key 状态变更时 bump version 失效。  
**注意**：多实例需极短 TTL 或配合 Redis；实现复杂，v1 可选。

**预期收益**  
热点 Key 场景 PG 读 QPS 可降一个数量级。

**优先级**：低（先有压测证明 PG 是瓶颈再做）

---

**Gateway 性能反模式（会更慢）**  
- 预检 JOIN `budget_consumed` 现场聚合  
- 预检调 NewAPI Admin 读 quota  
- 预检跑 `LoadBudgetContext` 全量组织树  

---

## 4. Ingest 入账：性能优化项

---

### I1 · Worker poll 间隔（已落地：默认 1s）

**慢在哪**  
`WORKER_POLL_INTERVAL_SEC` 默认现为 **1**（历史默认 5）。若环境仍配 5，任务入队后平均多等 **2.5s** 才被认领。  
这不是 CPU 慢，是**故意 sleep**，但对「入账有多快」影响最大。

**性能影响**  
- 单条入账端到端延迟下限 ≈ poll 间隔 + 处理时间。  
- 突发流量时 pending 堆积，尾延迟线性恶化。

**怎么改**  
1. **零代码**：生产设 `WORKER_POLL_INTERVAL_SEC=1`（或 2）。  
2. **进阶**：log 库 `NOTIFY` + Worker `LISTEN`，有活立刻醒，空闲再 sleep。

**预期收益**  
- 配置改为 1：排队等待 P50 从 ~2.5s 降到 ~0.5s。  
- NOTIFY：有负载时接近 **即时认领**（仍受 DB 事务耗时限制）。

**代码**  
`infra/ingest/worker.go` · `config/config.go`

**优先级**：**最高**（先改配置）

---

### I2 · 每笔 ingest 重复查 `Models()` 和 `Org().Nodes()`（**已落地**：`LoadEntryBuildSnapshot` 各读一次）

**慢在哪**  
`LoadEntryBuildInput`：`Models()` 调 2 次，`Org().Nodes()` 调 2 次；`IngestRaw` 里账期再读 `Nodes()`。  
同一公司、同一秒内几十笔账，**重复拉同一份全表数据**。

**性能影响**  
- 每笔多 **2～4 次** 主库 RTT（每次 1～5ms）。  
- 50 条/s → 多 100～200 次/s 无意义查询，吃满连接池和 PG CPU。

**怎么改**  
`IngestRaw` 开头各读一次，指针传给 `LoadEntryBuildInput`、`OccurrenceDepartmentPeriod`、`resolveBillingAllowedIds`。

**预期收益**  
单笔事务外读库时间降 **30%～50%**； ingest CPU 明显下降。

**代码**  
`domain/usage/ingest.go` · `domain/usage/entry.go`

**优先级**：**高**

---

### I3 · 模型目录 + 组织树无公司级缓存

**慢在哪**  
做完 I2 后，每笔仍各读 1 次全量 Models / Nodes。这两张「配置表」变更极少，读取极频。

**性能影响**  
单租户 burst（例如 100 ingest/s）时，主库 **100 次/s** 返回相同 JSON 体量行集。

**怎么改**  
进程内 `company_id → (models, nodes, revision)`，TTL 30s～5min；管理面写操作 bump revision 失效。

**预期收益**  
稳态 ingest 读 QPS 可降 **10× 以上**（视流量模型）。

**注意**  
TTL 与失效策略要压测；别为了性能用过期缓存算错账（revision 号比纯 TTL 稳）。

**优先级**：中高（I2 之后做）

---

### I4 · 读 log 行把 `content` 大字段也拉回来

**慢在哪**  
`GetConsumeLogByID` SELECT 整行，含可能很大的 `content` JSON。

**性能影响**  
- log DB → Backend 网络传输变长。  
- Go 反序列化 / 分配变慢。  
- Reconcile 补洞时 N 次单条读，放大更明显。

**怎么改**  
Ingest 专用查询只 SELECT 计费字段（token、quota、model、timestamp…）；`content` 若仅 audit 用则延后或另表。

**预期收益**  
单行 log 读耗时随 `content` 体积线性下降；reconcile 批处理收益更大。

**优先级**：中

---

### I5 · ingest 与投影共用 advisory 锁（**已落地**：拆锁）

**改动**  
Ingest 不再调用 `AcquireBudgetLock`。`ConsumeLots` 内部使用 `companies FOR UPDATE` 行锁保护钱包写入。  
Projector 保留 `AcquireBudgetLock` advisory lock（仅保护 `budget_consumed` 投影游标）。  
两者不再互斥。

**收益**  
锁竞争严重时，ingest 吞吐可提升 **数倍**。

**代码**  
`domain/usage/ingest.go` · `domain/billing/lot/consume.go`

**优先级**：**已完成**

---

### I6 · pending 处理单线程顺序执行（**已落地**：并行 goroutine pool）

**改动**  
Claim 后按 `company_id` 分组；组间 worker pool（8 goroutine）并行，组内串行（避免同公司行锁冲突）。配合 LISTEN/NOTIFY 唤醒（5s fallback poll）。

**收益**  
多租户混合流量：实例总吞吐接近 **线性扩展**（直到 PG 或连接池打满）。

**代码**  
`infra/ingest/worker.go` · `infra/ingest/group.go`

**优先级**：**已完成**

---

### I7 · 每批只处理 20 条 pending

**慢在哪**  
`INGEST_JOB_BATCH_SIZE` 默认 **20**。高负载时 Worker 频繁醒来又很快睡下，调度开销占比高。

**怎么改**  
提高到 50～100（配合 I6 并行），观察 log DB `SKIP LOCKED` 与主库连接池。

**预期收益**  
高 QPS 时单位时间处理条数上升；poll 次数减少。

**优先级**：中（与 I1 一起调参）

---

### I8 · ingest 事务内工作能否再瘦（边界项）

**现状**  
事务内：幂等检查 + `ConsumeLots`（可能扫多个 lot）+ ledger INSERT + 2× River `InsertTx`。

**可探查**  
- lot FIFO 是否热点行锁竞争（`fifo_head_lot_id`）。  
- ledger INSERT 是否可用 prepared statement / 批量 insert（若改为批处理 ingest）。  

**优先级**：低（先做完 I1～I6）

---

## 5. Reconcile：补洞路径的性能（非主路径，但实现很重）

主路径性能靠 webhook + pending；reconcile 只在漏单时跑。当前实现**偏慢**，会占 log DB / 主库资源，间接拖慢整体。

---

### R1 · N+1 读 log：先 list id，再每条 `GetConsumeLogByID`

**慢在哪**  
500 个 id → **501 次** log DB 往返。

**怎么改**  
`WHERE id = ANY($1::bigint[])` 一次取齐。

**预期收益**  
单轮 reconcile log 读耗时降 **~50%**。

---

### R2 · 每成功一行写一次 cursor

**慢在哪**  
500 条成功 = **500 次** UPSERT cursor。

**怎么改**  
批末写一次 cursor。

**预期收益**  
log DB 写 QPS 大幅下降，单轮 reconcile 总时长缩短。

---

### R3 · 全局单实例 reconcile

**慢在哪**  
`scheduler_locks` 只允许一个实例跑，补洞吞吐不随扩容增加。

**性能建议**  
主路径把 pending 做到足够快，别让 reconcile 常态化跑满；R1/R2 足够 v1。

---

## 6. 背景任务：只谈「占 CPU / 占锁 / 拖慢 ingest」

不写一致性，只写**这些任务如何拖 ingest 的后腿**。

---

### P1 · 投影批处理太重，长时间占着公司锁

**慢在哪**  
`budget_projector.go` 每批最多 500 条 ledger：读 cursor → 累加 consumed → 重算 touched keys 的 `gateway_soft_*` → 可能 `LoadBudgetContext` 全量加载。

**对 ingest 的性能影响**  
投影事务持有 `AcquireBudgetLock` 越久，**同公司 ingest 等锁越久**（见 I5）。

**怎么改**  
- 确保 refresh 路径**不**重复 `LoadBudgetContext`（只用批内已算数据，终态 C5）。  
- 单批耗时过长时缩小 `batchSize` 换更短锁持有（吞吐与锁等待的权衡）。  
- 只更新本批 touched 的 key 行，避免全表扫。

**预期收益**  
投影单批耗时降 → 锁等待降 → **ingest TPS 上升**。

**代码**  
`domain/budget/budget_projector.go` · `domain/budget/gateway_summary.go`

**优先级**：中高

---

### P2 · River 冷链 job 太碎，调度开销大

**慢在哪**  
每笔 ingest 入队 `wallet_sync`；投影再入队 `rebalance` / `overrun`。同公司短时间多个 River job 抢 worker 线程。

**性能影响**  
River 调度、序列化、claim 开销；worker 线程忙于碎 job，**投影 self-chain 排队变长** → 间接延长锁竞争。

**怎么改**  
合并为 `platform_sync`（wallet + rebalance 一次跑完）。

**预期收益**  
River 队列深度下降；背景 CPU 下降；投影追平更快 → ingest 少等锁。

**优先级**：低～中（v1 后可做）

---

## 7. 不推荐做的「优化」（会更慢或更脆）

| 做法 | 为什么 |
| --- | --- |
| 预检 JOIN `budget_consumed` 现场算剩余 | SQL 变重，RTT 上升 |
| 预检 HTTP 调 NewAPI | 多一跳网络，毫秒变百毫秒 |
| ingest 事务内同步写 `budget_consumed` / `gateway_soft_*` | ingest 事务时间暴增，吞吐暴跌 |
| 去掉幂等 / 去掉公司锁 | 可能快一点点，并发错账 |
| 微服务拆 Gateway 和 Ingest | 多一跳，运维成本涨 |

---

## 8. 怎么验证性能（压测指标，不是监控项目）

改之前先记 baseline，改之后对比同一压测脚本：

| 指标 | 怎么量 | 优化项 |
| --- | --- | --- |
| Gateway 预检 P99 | wrk/hey 打 `/v1/models`，不含上游 | G1 G2 G3 |
| Gateway 预检 P99（大 body） | 100KB+ JSON body 同上 | G1 |
| ingest 单条处理时间 | log→ledger 时间戳差（不含 poll） | I2 I3 I4 I5 |
| ingest 端到端延迟 | webhook accept → ledger commit | I1 I6 I7 |
| ingest 吞吐 | 条/秒/实例（稳态） | I1 I2 I3 I5 I6 I7 |
| PG 连接池等待 | pool acquire wait | I2 I3 I6 |
| advisory lock 等待时间 | ingest 事务内计时 | I5 P1 |

**验收目标（压测环境校准，非生产承诺）**

| 场景 | 目标 |
| --- | --- |
| Gateway 预检 P99（无上游） | < 20ms @ 1k RPS |
| ingest 处理时间 P95（不含 poll） | < 50ms |
| ingest 端到端 P95（`POLL_INTERVAL=1`） | < 500ms |
| 单实例 ingest 稳态吞吐 | > 200 条/s（多租户并行） |

---

## 9. 实施顺序（纯性能性价比）

> **Lag**：§10 与下列多项重叠（尤其 I1、P1、I6、I5）；lag 专项顺序见 §10.7。

### 第一批：立刻见效，几乎无风险

1. **I1** — `WORKER_POLL_INTERVAL_SEC=1`  
2. **G2 + G3** — 预检 SQL 瘦身  
3. **I2** — Models/Nodes 只读一次  

### 第二批：提吞吐

4. **I3** — 公司级配置缓存  
5. **I6 + I7** — pending 并行 + 批大小  
6. **I4** — log 窄列查询  
7. **R1 + R2** — reconcile 批量化（减 DB 争抢）  

### 第三批：架构向

8. **G1** — 流式读 model（大 body 场景）  
9. **P1** — 投影单批减负（为 ingest 解锁）  
10. **I5** — 评估拆 advisory 锁  
11. **P2** — 合并 River 冷链  

---

## 10. Lag 优化（已大幅缩减）

> **架构改进后**：`wallet_remain` 和 `combined_key_remain` 均在 Ingest 事务内原子更新。Gateway 预检读到的数据**无投影 lag**。  
> 剩余的 lag 仅影响 `budget_consumed`（用于 overrun 阻断和百分比预警），由 Projector 异步维护。

### 10.1 当前 Lag 模型

| 执法层 | 数据源 | 有无 lag | 预检怎么用 |
| --- | --- | --- | --- |
| **L0 企业钱包** | `companies.wallet_remain` | **无** | Ingest 同事务更新；预检同条 SQL 读取 |
| **L1 预算管控** | `platform_keys.combined_key_remain` | **无** | Ingest 同事务 DecrementBatch；预检同条 SQL 读取 |
| **L2 Overrun 阻断** | `budget_consumed` → Overrun job | **有**（等 Projector） | 投影批末入队，异步 disable Key |
| **L3 百分比预警** | `budget_consumed` → checkAlertThresholds | **有**（等 Projector） | 投影批末检查并发通知 |

**预检热路径（L0 + L1）已无 lag。** 只有冷路径副作用（L2 overrun、L3 预警）存在 ~1s 的 River job 调度延迟。

### 10.2 已落地的优化

| 项 | 机制 | 效果 |
| --- | --- | --- |
| `combined_key_remain` 移入 Ingest 事务 | 同事务 DecrementBatch | 预检 lag = 0 |
| LISTEN/NOTIFY | `pg_notify` 唤醒 worker | Ingest 排队延迟从 ~500ms 降到 ~50ms |
| 并行处理 | goroutine pool (8) | 多租户吞吐线性提升 |
| 拆 advisory lock | ConsumeLots 用 `companies FOR UPDATE` 行锁 | Projector 不再阻塞 ingest |
| wallet 从 chain 剥离 | `GatewayChainRemain` 不含 wallet | 充钱无需重算 combined_key_remain |

### 10.3 不再需要的优化

以下优化在旧架构下有意义（`gateway_soft_remain` 由 Projector 异步维护），现在已不需要：

| 做法 | 为何不再需要 |
| --- | --- |
| 缩短 `budget_projection` ByPeriod | 预检已不依赖 Projector 产出 |
| 投影单批减负缩短锁 | Ingest 和 Projector 已拆锁，不互斥 |
| 缩小投影 batchSize | 同上，Projector 锁不影响 ingest |
| Redis `budgetcheck` 用于减 lag | `combined_key_remain` 已实时，Redis 仅做多实例一致性加强 |

---

## 11. 代码索引

| 路径 | 性能 / Lag 相关逻辑 |
| --- | --- |
| `domain/gateway/gateway_service.go` | body 读取、反代 |
| `store/postgres/gateway_precheck_repo.go` | 预检 SQL |
| `domain/gateway/evaluate.go` | 内存判断（wallet + soft） |
| `domain/gateway/precheck.go` | PG 预检 + 可选 Redis 二次挡 |
| `infra/ingest/worker.go` | poll 间隔、claim、reconcile |
| `domain/usage/ingest.go` | 入账主路径、事务 |
| `domain/usage/entry.go` | `LoadEntryBuildInput` 重复读 |
| `app/port_usage.go` | ingest 同事务入队 `budget_projection` |
| `store/postgres/log_repo.go` | log 行读取 |
| `domain/budget/budget_projector.go` | 投影批处理、占锁、self-chain |
| `domain/budget/gateway_summary.go` | `ComputeGatewaySummaryUpdates` |
| `pkg/budget/chain.go` | `GatewayChainRemain`（摘要计算公式） |
| `infra/jobs/args.go` | `budget_projection` Unique `ByPeriod` |
| `infra/river/workers/budget_projection.go` | River 投影 worker |
| `store/postgres/projection_progress_repo.go` | `budget_projection_progress` cursor |
| `config/config.go` | `WORKER_POLL_INTERVAL_SEC`、`INGEST_JOB_BATCH_SIZE`、`RIVER_MAX_WORKERS` |

**相关文档**：[Backend-Ingest架构.md](./Backend-Ingest架构.md) · [Backend-离线任务.md](./Backend-离线任务.md) · [架构终态设计.md](./架构终态设计.md)（§12–§14 执法层级与 `projection_lag_seconds`）
