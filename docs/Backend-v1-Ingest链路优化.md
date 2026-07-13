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
最小集 **已落地**：`I1` · `G2+G3` · `I2`。后续：`I3 缓存` → `I6 并行 ingest` → `G1 流式读 body` → `I5 拆锁` → `P2 投影减负`。

**Lag 与性能分开看**  
企业钱包（`wallet_remain`）在 ingest 事务里同步更新，预检读 PG **无投影 lag**；Key / 成员 / 项目轴靠 `gateway_soft_remain`，由 `budget_projection` 异步刷新，才有 lag 窗口。§10 只列**不拖慢 Gateway / Ingest 热路径**的缩窗手段（不少与上文性能项同改、双赢）。

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

同一 `company_id` 上，`AcquireBudgetLock` 互斥：

- Ingest 入账事务  
- `budget_projection` 批处理（默认每批 500 条 ledger）  
- 管理面改预算 / 项目  

投影批处理越久，ingest 等锁越久——这是**性能问题**，不是「数据晚几秒展示」的问题。

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

### I5 · ingest 与投影共用 advisory 锁 → ingest 事务被堵

**慢在哪**  
`ingest.go` 与 `budget_projector.go` 事务开头都 `AcquireBudgetLock(company_id)`。  
投影一批 500 条可能要跑 **数百 ms～数秒**，期间同公司 **所有 ingest 事务排队等锁**。

**性能影响**  
- 入账 TPS 上限 ≈ `1 / (ingest_tx_time + lock_wait)`。  
- 高峰时 lock_wait 远大于 ingest 本身，**吞吐断崖**。

**怎么改（需评审）**  
拆锁：ingest 只锁 `companies` 行（`FOR UPDATE`）+ ledger 插入；投影只锁 `budget_projection_progress`。  
或缩小投影批大小 / 优化投影单批耗时（见 P2），先减锁持有时间。

**预期收益**  
锁竞争严重时，ingest 吞吐可提升 **数倍**。

**优先级**：中（先靠 P2 + 压测看 `lock_wait` 占比）

---

### I6 · pending 处理单线程顺序执行

**慢在哪**  
Worker claim 一批（最多 20 条）后，for 循环里**串行** `IngestByLogID`。A 公司慢事务阻塞 B 公司。

**性能影响**  
多租户场景下，实例 ingest 吞吐 = 最慢那条链，无法吃满多核。

**怎么改**  
Claim 后按 `company_id` 分组；组间 worker pool（如 8 goroutine）并行，组内串行（避免同公司锁冲突）。

**预期收益**  
多租户混合流量：实例总吞吐接近 **线性扩展**（直到 PG 或连接池打满）。

**注意**  
调大 `pgxpool` max conns；控制 pool 大小与 goroutine 数匹配。

**优先级**：中高

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

## 10. Lag 优化（不牺牲热路径性能）

> **Lag 指什么**：`usage_ledger` 已入账，但 `gateway_soft_remain` / `budget_consumed` 还没跟上。Gateway 预检读的是**摘要列**，不是现场聚合 consumed。  
> **不影响性能** = 不改 Gateway 预检 SQL 形态、不在 ingest 事务里加投影、不为减 lag 给预检加 JOIN/Admin 调用。

### 10.1 Lag 从哪来（公式）

```
L1 lag ≈ ingest 排队
       + 等 budget_projection 被 River 认领
       + 等 AcquireBudgetLock（与 ingest / 改预算互斥）
       + Projector RunBatch（≤500 条/批）
       + 批末写 gateway_soft_* + self-chain 调度间隙
```

| 执法层 | 数据源 | 有无 lag | 预检怎么用 |
| --- | --- | --- | --- |
| **L0 企业钱包** | `companies.wallet_remain` | **无** | ingest 同事务更新；预检同条 SQL 读取 |
| **L1 多轴预算** | `platform_keys.gateway_soft_*` | **有** | `Evaluate` 读 `soft_remain`；NULL 放行 |
| **L2 硬关 Key** | Overrun job | **有**（≥ L1） | 投影批末入队，异步 disable |

因此：**减 lag = 让 Projector 更快把 ledger 追平到摘要**，而不是让预检变重。

### 10.2 双赢项（减 lag，且热路径更快 / 不变）

这些与 §4–§6 性能项是同一改动，实施时一并收益：

| 项 | 减 lag 的机制 | 对 Gateway / Ingest 热路径 |
| --- | --- | --- |
| **I1** 缩短 ingest poll | ledger 更早落库 → 投影可更早开跑 | 仅 Worker 配置，预检无感 |
| **I2 / I3** 减 ingest 读放大 | 单笔入账更快 → 锁占用更短 → 投影少排队 | 预检无感 |
| **I6** pending 按公司并行 | 突发流量下 ledger 堆积更短 | 预检无感 |
| **P1** 投影单批减负 | 单批耗时↓ → `gateway_soft_*` 刷新更频 → 锁更短 | 预检仍 1× PG；ingest 等锁更少 |
| **I5** 拆 advisory 锁（评审后） | ingest 与投影可并行 → 高峰 lag 明显缩短 | 预检不变；ingest 吞吐↑ |
| **P2** 合并 River 碎 job | worker 线程少被 `wallet_sync` / `rebalance` 占满 → 投影先跑 | 背景侧；热路径无 RTT 变化 |

**建议 lag 视角下的顺序**：在 §9 第一批之后，优先 **I1 → P1 → I6 → I5**，再视 `projection_lag_seconds` 调 §10.3 的背景参数。

### 10.3 背景侧调参（不动热路径，略增 River / 投影 CPU）

| 编号 | 做法 | 现状 | 减 lag 原理 | 性能代价 |
| --- | --- | --- | --- | --- |
| **L1** | 缩短 `budget_projection` Unique `ByPeriod` | `1s`（`args.go`） | 同公司连续 ingest 时，重复入队不被 dedupe 太久，投影更密 | 仅 River 调度与投影批次数↑；**Gateway / ingest 无额外 RTT** |
| **L2** | 提高 `RIVER_MAX_WORKERS` 或给投影独立 queue 份额 | default 与 `platform_sync` 等共用 default 池 | 多租户时投影 job 少排队 | 进程 CPU / PG 连接略增；热路径不变 |
| **L3** | 保持 / 强化 self-chain | 满批（500）后 `InsertBudgetProjection` | backlog 大时连续批处理，摘要连续刷新 | 已存在；确保 P1 不拖慢单批即可 |
| **L4** | 突发后手动 / 脚本 `Insert(budget_projection)` | Runbook（架构终态 §15） | 运维兜底，非常态 | 无 |

**L1 建议**：压测下将 `ByPeriod` 从 `1s` 试到 `200ms～500ms`，看 `projection_lag_seconds` P99 与 River queue depth；再往下收益递减。

**L2 建议**：lag degraded（≥1s）时先加 worker，再考虑把 `budget_projection` 提到 `critical` 子池（需评估是否饿死 `newapi_sync`）。

### 10.4 有条件权衡（只在你能接受背景开销时）

| 做法 | 对 lag | 对热路径性能 | 说明 |
| --- | --- | --- | --- |
| **缩小 Projector `batchSize`**（如 500→200） | 单批结束更早写 soft，窗更碎 | ingest **等锁次数**可能↑ | 仅当 P1 后单批仍 >500ms 且 lock_wait 占比低时试 |
| **增大 `batchSize`** | 大批次内 soft 不刷新 | 单批占锁更久，lag **变差** | 为 ingest 吞吐服务，不是为 lag |

### 10.5 不能当「减 lag」做的（会变慢或更脆）

| 做法 | 为何不做 |
| --- | --- |
| ingest 事务内同步 `ApplyIncrement` + 写 `gateway_soft_*` | ingest 事务时间暴增（§7 已列） |
| 预检 JOIN `budget_consumed` 现场算 remain | 预检 SQL 变重，违背 1× PG 设计 |
| 为 freshness 开 Redis 并让预检**只**读 Redis | 预检多 1 RTT，且 Redis 不是摘要权威源 |
| 靠 `budget_reconcile`（30min）追实时 lag | 漂移修复用，不能当实时投影替代 |
| 缩短 `budget_reconcile` 周期追 lag | 全量重算重，抢 PG，拖 ingest/投影 |

**Redis `budgetcheck`**（`precheck.go` 第二次挡）：只在 `version >= PG` 时加强 block，**不让摘要更新更快**；开启后预检多一次可选 GET——减 lag 别指望它，多实例挡一致性另论。

### 10.6 Lag 验收（与 §8 压测分开记）

| 指标 | 怎么量 | 目标（压测环境，对齐架构终态 §14） |
| --- | --- | --- |
| `projection_lag_seconds` | `now - max(ledger.occurred_at)` where id > cursor | P99 **< 1s**（degraded 线） |
| `projection_batch_duration_seconds` | `RunBatch` 耗时 | P95 < 200ms（P1 后校准） |
| `projection_queue_depth` | River `budget_projection` pending | 稳态 ≈ 0；突发后 30s 内回落 |
| ingest → soft 刷新 | 同一 Key 连续请求直到 403 | 与 lag 公式一致，I1+P1 后明显缩短 |

**不变量**：优化 lag 后仍应保持 Gateway **1× PG**、`gateway_sql_calls = 1`；若 lag 优化导致预检 RTT 上升，则方案不合格。

### 10.7 Lag 实施顺序（在 §9 性能批之后）

1. **I1 + P1** — 账本更早、投影更快（零预检代价）  
2. **L1** — 调 `ByPeriod`（纯背景）  
3. **I6 + L2** — 突发多租户追平  
4. **I5** — 锁拆分（评审通过再做）  
5. **L4** — 仅故障 / 大 tenant 手工 chain  

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
| `app/usage_enqueuer.go` | ingest 同事务入队 `budget_projection` |
| `store/postgres/log_repo.go` | log 行读取 |
| `domain/budget/budget_projector.go` | 投影批处理、占锁、self-chain |
| `domain/budget/gateway_summary.go` | `ComputeGatewaySummaryUpdates` |
| `pkg/budget/chain.go` | `GatewayChainRemain`（摘要计算公式） |
| `infra/jobs/args.go` | `budget_projection` Unique `ByPeriod` |
| `infra/river/workers/budget_projection.go` | River 投影 worker |
| `store/postgres/projection_progress_repo.go` | `budget_projection_progress` cursor |
| `config/config.go` | `WORKER_POLL_INTERVAL_SEC`、`INGEST_JOB_BATCH_SIZE`、`RIVER_MAX_WORKERS` |

**相关文档**：[Backend-Ingest架构.md](./Backend-Ingest架构.md) · [Backend-离线任务.md](./Backend-离线任务.md) · [架构终态设计.md](./架构终态设计.md)（§12–§14 执法层级与 `projection_lag_seconds`）
