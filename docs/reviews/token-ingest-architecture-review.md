# Token 消费/入账架构评审与优化建议

> 评审范围：Gateway → Ingest → Fanout → Projection 全链路
> 日期：2026-07-16
> 最近变更：middleware 层 + ingest queue 合并到 River

---

## 1. 现有架构概览

### 1.1 端到端数据流

```
用户请求 → /v1/chat/completions
         │
         ├─ RealIP → RequestID → AccessLog → Recover → SecurityHeaders → CORS
         │
         ▼
┌──────────────────────────────────────────────────────┐
│  Gateway Layer (/v1)                                 │
│  1. parseBearerSecret (API Key)                      │
│  2. Rate Limit (per-key, Redis token bucket)         │
│  3. Precheck:                                        │
│     - LoadPrecheckContext (PG: key状态/model白名单)   │
│     - budgetRemainCheck (Redis: remain + version)    │
│  4. ReverseProxy → NewAPI                            │
└──────────────────────────┬───────────────────────────┘
                           │ (请求完成后 NewAPI 写 consume log)
                           ▼
                  consume log 写入 (logs 表)
                           │
                  ┌────────▼────────┐
                  │  River IngestJob │  ← webhook 触发 InsertIngest
                  │  (critical queue)│
                  └────────┬────────┘
                           │
          ┌────────────────▼────────────────┐
          │      IngestService.IngestRaw     │
          │  (单公司事务，company row lock)   │
          │                                  │
          │  1. 幂等检查 (idempotency_key)   │
          │  2. FIFO lot 消耗 → ledger 分段  │
          │  3. budget_consumed 增量 UPSERT  │
          │  4. combined_key_remain 递减     │
          │     → RefreshRedisCache          │
          │  5. 事务内 enqueue River jobs    │
          └────────────────┬────────────────┘
                           │
       ┌───────────────────┼───────────────────┐
       ▼                   ▼                   ▼
┌─────────────┐   ┌──────────────┐   ┌────────────────┐
│ WalletSync  │   │ Dashboard    │   │ Overrun (条件) │
│ 5s 去重     │   │ Project      │   │ by args 去重   │
└─────────────┘   │ 1h 去重      │   └────────────────┘
                  └──────────────┘
```

### 1.2 双层额度控制

```
┌─────────────────────────────────────────────────────────────┐
│  Layer 1: Gateway (快，best-effort)                          │
│                                                             │
│  Redis: { remain: 95.2, version: 42 }                       │
│  判断：remain ≤ 0 → 拦截；否则放行                           │
│  特点：亚毫秒，但有延迟窗口                                   │
└─────────────────────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────────────────┐
│  Layer 2: Ingest (慢，精确)                                  │
│                                                             │
│  PG 事务内精确扣减 → commit 后 SET Redis (精确覆盖)          │
│  延迟：请求发出 → 执行完 → webhook → River处理 = 5~60秒      │
└─────────────────────────────────────────────────────────────┘
```

### 1.3 Ingest 事务实际开销

单笔 ~10 条 SQL，常量级，毫秒完成。不是瓶颈。

---

## 2. 核心问题

**Gateway 放行时看到的余额是过期快照**。

```
t0               t1            t2           t3
│                │             │            │
Gateway放行     请求执行中     webhook      Ingest完成
remain=10       (2~30秒)                   remain=5, Redis更新
│
│  窗口内所有并发请求都看到 remain=10
│  余额充裕时无所谓；余额紧张时导致超额
```

---

## 3. 优化方案

### P0：Gateway Precheck 零 PG 查询

#### 现状

每个 /v1 请求执行一次 PG 查询 `LoadPrecheckContext`，返回：
- key 状态、过期时间、model 白名单 → **几乎不变**
- CombinedKeyRemainVersion → **每次 ingest 递增**

两类数据变化频率完全不同，却放在同一条 SQL 里每请求查一次。

#### 目标状态

```
┌─────────────────────────────────────────────────────────┐
│  Gateway 热路径（每请求执行）                              │
│                                                         │
│  ① Key 合法性检查 → 进程内缓存（invalidate-on-write）    │
│     - CompanyStatus                                     │
│     - KeyStatus / KeyExpiresAt                          │
│     - AllowlistTypes                                    │
│     缓存永远命中，除非管理端做了写操作                     │
│     写操作时主动 invalidate → 下一个请求穿透 PG           │
│                                                         │
│  ② Budget 拦截 → 直接查 Redis（不查 PG）                 │
│     - Redis GET remain                                  │
│     - remain ≤ 0 → 拦截                                 │
│     - remain > 0 或 miss → 放行                         │
│                                                         │
│  PG 查询次数：0 次/请求                                   │
│  Redis 查询：1 次/请求（亚毫秒，本来就有）                 │
└─────────────────────────────────────────────────────────┘
```

#### 具体做法

**1. 拆分 `LoadPrecheckContext` 为两个职责**

```
KeyValidityCache（进程内）:
  数据：CompanyStatus, KeyStatus, KeyExpiresAt, HasAllowlist, AllowlistTypes
  策略：invalidate-on-write（管理端修改时清除）
  命中率：接近 100%（这些值每天改几次）

budgetRemainCheck（Redis）:
  数据：remain
  策略：每请求实时查 Redis
  不再需要 version 比对
```

**2. Invalidate 触发点（同进程，无跨服务通信）**

| 管理操作 | Invalidate 方法 |
|---------|----------------|
| 禁用/删除 key | `cache.Invalidate(keyHash)` |
| 修改 model 白名单 | `cache.Invalidate(keyHash)` |
| 冻结/解冻公司 | `cache.InvalidateCompany(companyID)` |

这些操作在 keys service / company service 里执行，注入 cache 引用即可。

**3. 去掉 version 比对**

当前 version 比对的作用：防止 Redis 中的陈旧数据误拦截。

去掉后的风险分析：
- Redis refresh 失败（Ingest 写 PG 成功但写 Redis 失败）→ Redis 里 remain 偏高（旧值）→ 结果是"该拦截的没拦截"
- 但这等价于当前的 fail-open 设计（Redis 不可用时全部放行）
- 下一次 Ingest 完成后 Redis 会被覆盖（自动修复，窗口 < 1 分钟）

结论：**version 比对可以去掉**。它保护的是一个极端边缘情况，而 fail-open 设计已经接受了这个风险。

#### 效果

```
当前：每请求 1 次 PG + 1 次 Redis
优化后：0 次 PG + 1 次 Redis

100 QPS × 10 个 key = 1000 请求/秒 → PG 负载从 1000 次/秒降至 ~0
（仅管理端操作后偶尔穿透 1 次）
```

#### 副作用

| 场景 | 影响 |
|------|------|
| Key 被禁用 | **立即生效**（invalidate 是同步的） |
| Redis refresh 失败 | 最多 1 分钟放行本应拦截的请求（和当前 fail-open 一致） |
| 进程重启 | 缓存丢失，第一个请求穿透 PG 重建（毫秒级） |

**工作量**：约 60 行代码 + 3 个注入点。

---

### P1：Gateway 乐观预扣（解决余额紧张时超额）

#### 现状

`budgetRemainCheck` 只读 Redis，不修改值。并发请求都看到同一个 remain。

#### 目标状态

```
Gateway precheck:
  DECRBY remain, estimated_cost    ← 原子递减
  if 结果 ≥ 0 → 放行
  if 结果 < 0 → 拒绝 + INCRBY 回滚

Ingest 完成后:
  SET remain = 精确值              ← 覆盖累积偏差
```

```
示例：remain=10, estimated_cost=5

  请求A: DECRBY 5 → remain=5  → 放行
  请求B: DECRBY 5 → remain=0  → 放行
  请求C: DECRBY 5 → remain=-5 → 拒绝 + INCRBY 回滚

  结果：只放行 2 个请求 ✓
```

#### estimated_cost 取值

用全局 P90 cost 或 key 近期平均值。不需要精确：
- 偏大 → 提前 1~2 个请求开始拒绝（保守，少超额）
- 偏小 → 可能多放行 1 个请求
- Ingest 的 SET 覆盖会持续修正偏差

#### 副作用

| 场景 | 影响 |
|------|------|
| 请求被 NewAPI 拒绝（非余额原因） | 预扣不会回滚，下次 Ingest SET 时修正（< 1 分钟） |
| Redis 不可用 | 退回 fail-open（不预扣，当前行为） |
| estimated_cost 不准 | 偏差在下次 Ingest SET 时消除 |

**工作量**：约 50 行。

---

### P2：WalletSync 去重窗口扩大

**做法**：`ByPeriod: 5s` → `ByPeriod: 60s`

**收益**：外部 API 调用降低 12 倍。

**副作用**：NewAPI 后台 quota 显示延迟从 5s 变为 60s（用户无感知）。

**工作量**：改 1 个常量。

---

### ✅ 已完成：合并 ingest_jobs 到 River

ingest_jobs 表 + 自建 claim/retry/dead-letter 已完全迁移到 River。
- 减少 ~400 行自维护代码
- 统一监控（River UI）
- 统一重试策略（River 内置指数退避 + MaxAttempts 20）
- Reconcile 作为 River 周期任务自动执行

---

## 4. 不应该做的事情

| 想法 | 为什么不做 |
|------|-----------|
| Lot 消耗 CTE 批量化 | 99% 只更新 1 个 lot，不是瓶颈 |
| budget_consumed 移到 post-commit | 同公司串行 job 看不到前一笔消耗 |
| 引入 Kafka/RabbitMQ | PG + River 够用 |
| 按 platform_key 拆锁 | Lot 是公司级共享池 |
| Precheck 整体 TTL 缓存 | version 变太快导致命中率低；拆开后根本不需要 TTL |

---

## 5. 总结

```
优先级    做什么                     解决什么问题              状态
────────────────────────────────────────────────────────────────────
P0       Precheck 零 PG 查询       Gateway 热路径 PG 压力    待实施
         (invalidate cache +       （key 合法性缓存 +
          去掉 version 比对)         budget 直接查 Redis）

P1       Gateway 乐观预扣           余额紧张时并发超额        待实施
         (DECRBY + Ingest SET)

P2       WalletSync 窗口 5s→60s    外部 API 调用过频         待实施

✅       合并 ingest_jobs 到 River  维护两套 queue 的负担     已完成
```

核心判断：
- Ingest 事务本身没有问题——10 条 SQL，毫秒完成
- **P0 消除 Gateway 唯一的 PG 依赖**——hotpath 上只剩 Redis
- **P1 解决真正的业务问题**——余额紧张时的并发超额
- P2 是常量级改动，随时可做
