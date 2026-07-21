# TokenJoy × NewAPI 后端架构 Review

## 1. 架构总览

```
用户请求
  │
  │ Authorization: Bearer <platform_key_secret>
  ▼
┌─────────────────────────────────────────────────────────────────┐
│                      TokenJoy Backend (:8080)                    │
│                                                                 │
│  /api/*  ─── 管理 API (session, auth, org, budget, keys, ...)   │
│                                                                 │
│  /v1/*   ─── Gateway ──┬── Rate Limit (Redis token-bucket)      │
│                        ├── Precheck (PG cache + Redis budget)   │
│                        │   ├─ key 存在 + active                 │
│                        │   ├─ company wallet > 0                │
│                        │   ├─ model 在白名单内                   │
│                        │   └─ combined_key_remain > 0           │
│                        └── Reverse Proxy ──────────────────►    │
│                                                                 │
│  Worker  ─── River Jobs                                         │
│              ├─ create_key (async token creation)                │
│              ├─ upsert_channel (provider key sync)               │
│              └─ rebalance (budget recompute)                     │
│                                                                 │
│  Ingest  ─── consume_log 轮询 → ledger 写入 → budget 扣减       │
└─────────────────────────────────────────────────────────────────┘
                              │
                              │ HTTP (reverse proxy + admin API)
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      NewAPI (:3000)                              │
│                                                                 │
│  接收 /v1/* 请求 → Channel routing (group → upstream LLM)       │
│  产出 consume_log (用量日志)                                     │
│  Token/Channel/User/Group CRUD (admin API)                      │
│  Pricing API (model ratio 定价)                                  │
└─────────────────────────────────────────────────────────────────┘
```

## 2. 数据流

### 请求链路（热路径）

```
用户 → TJ gateway → [rate limit] → [precheck: PG+Redis] → proxy → NewAPI → LLM upstream
                                                                      │
                                                                      ▼
                                                               consume_log 写入
```

### 用量采集（异步）

```
NewAPI consume_log → TJ Ingest 轮询 → token_id 反查 mapping → company context
    → BuildCallSettledEntry → lot 消费 → ledger 写入 → budget_consumed 更新
    → combined_key_remain 扣减 → Redis cache 刷新 → overrun 判断
```

### Key 生命周期同步

```
TJ domain/keys (创建/启禁/轮转/吊销)
    → newapisync (编排层)
        → platformkey/ (实际同步逻辑)
            → adminport.Port (接口)
                → integration/newapi.Client (HTTP 实现)
                    → NewAPI admin API
```

### Channel 同步

```
TJ domain/keys (ProviderKey 变更)
    → newapisync.EnqueueUpsertProviderKey → River job
        → provider.SyncUpsertProviderKey
            → Client.UpsertChannel + RebuildAbilities
```

## 3. 职责边界

| 关注点 | TokenJoy (管控面) | NewAPI (执行面) |
|--------|------------------|----------------|
| API Key 鉴权 | gateway precheck (唯一入口) | 不做验证 (信任 proxy) |
| 模型白名单 | precheck `checkPlatformKey` | 不检查 (model_limits 已移除) |
| 预算限额 | combined_key_remain + overrun | Token 设为 unlimited |
| Rate Limit | per-key Redis bucket | 不做 |
| Key CRUD | PlatformKey 持久化 + 业务规则 | Token 存储 (bearer 持有方) |
| Channel Routing | 决定 group 分配 + RebuildAbilities | 执行 group → upstream 选择 |
| 用量产出 | — | consume_log 记录 |
| 用量计费 | Ingest → ledger → lot 消费 | — |
| 定价 | models 表 (TJ 为准) | Pricing API (同步源) |

## 4. 同步操作清单

### adminport.Port 接口（TJ 调用 NewAPI 的全部 API）

| 方法 | 调用方 | 用途 |
|------|--------|------|
| `CreateToken` | platformkey/create | 创建 bearer token |
| `UpdateToken` | platformkey/update | 推送 status + group |
| `GetToken` | provision/bootstrap | reconcile 检查存活 |
| `GetTokenKey` | platformkey/create | 获取 bearer secret |
| `RegenerateToken` | platformkey/rotate | 轮转 bearer |
| `DeleteToken` | platformkey/revoke | 吊销 token |
| `UpsertChannel` | provider/sync | 同步上游渠道 |
| `EnsureGroup` | platformkey/create | 确保 department group 存在 |
| `RebuildAbilities` | models/service | routing rule 变更后重建 channel-model 映射 |
| `CreateUser` | provision/bootstrap | 创建 wallet owner (仅 dev) |
| `ListModelPricing` | models/service | 定价同步 |

### 同步频率

| 操作 | 触发条件 | 频率 |
|------|---------|------|
| CreateToken | Key 创建 | 每个新 Key 一次 |
| UpdateToken | Key 启用/禁用 + whitelist 变更 | 每次 Key 状态变更 |
| UpsertChannel | ProviderKey 创建/更新 | 每次 ProviderKey 变更 |
| RebuildAbilities | Routing rule 变更 | 每次白名单变更 |
| ListModelPricing | 启动 + 手动触发 | 低频 |

## 5. 关键设计决策

### 5.1 Precheck 是唯一鉴权防线

model_limits 同步移除后，NewAPI 侧不再检查模型白名单。所有访问控制由 TJ gateway precheck 完成：
- PG 单次 JOIN 查询加载 key 状态 + company 状态 + 白名单 + budget
- 进程内 LRU cache (TTL 5min, max 10K entries)
- 主动 invalidation: `InvalidateByKeyID` (key 变更) / `InvalidateCompany` (routing rule / company 冻结)

### 5.2 NewAPI PUT 全量替换 → MergeTokenPut

NewAPI 没有 PATCH 语义，PUT 会覆盖所有字段。TJ 的 `UpdateToken` 实际执行 GET + merge + PUT（2 次 HTTP），确保不会意外清空未传字段。

### 5.3 Billing 与 NewAPI 完全解耦

Billing domain 不直接调 NewAPI。充值/钱包/lot 完全在 TJ PG 内完成。Ingest 消费 consume_log 后扣减 lot，与 NewAPI 无同步关系。

### 5.4 Key Secret 生命周期

TJ 用户持有 platform_key_secret（存在 PG `platform_keys.key_hash`），gateway 用 hash 匹配。实际发往 NewAPI 的是 bearer token（存在 NewAPI 侧），TJ 只通过 `GetTokenKey` 获取一次后存入 `full_key`。用户永远不直接接触 NewAPI token。

## 6. 可优化点

### P0: 定价同步缺少定时刷新

`SyncPricingFromUpstream` 只在启动时和手动调用时执行。如果 NewAPI 侧定价变动，TJ 用量计费会用旧价格。

**建议**：scheduler 加 periodic job，每 10 分钟调用 `SyncPricingFromUpstream`。

### P1: CreateToken 未传 UnlimitedQuota

当前 `CreateTokenInput` 不设 `UnlimitedQuota` 和 `RemainQuota`（零值）。依赖 NewAPI 侧默认行为。如果 NewAPI 默认不是 unlimited，新创建的 Token 可能有意外限额。

**建议**：显式传 `UnlimitedQuota: true`。

### P2: UpdateToken 每次 2 次 HTTP (GET + PUT)

受限于 NewAPI PUT 全量替换语义。每次 `SyncUpdatePlatformKey` 实际产生 2 次 HTTP 调用。

**影响**：Key 启禁操作延迟翻倍。如果 NewAPI 支持 PATCH 可消除。

### P3: RebuildAbilities 调用时机可优化

当前 `UpdateRoutingRule` 先调 `RebuildAbilities`（channel routing 重建），再调 `InvalidateCompany`（precheck cache 刷新）。两者是独立的——RebuildAbilities 是为了让 NewAPI 更新 channel-model 映射（哪个 channel 支持哪个模型），与白名单检查无关。

**问题**：如果 `RebuildAbilities` 失败，整个 `UpdateRoutingRule` 会失败回滚，但 routing rules 已经 persist 到了 PG（前面的 `PersistRoutingRules` 不在事务中？需确认）。

### P4: 长期 — Key 双写是最大复杂度源

Key 的 create/update/revoke/rotate 都需要 TJ PG + NewAPI 双写，通过 outbox + River job 保证最终一致。这是正确的分布式设计，但也是系统中最复杂的部分（~400 LOC）。

**终极方向**：如果 NewAPI 可改为 pass-through 模式（不验证 Token，只做 channel routing），整个 Key 同步退化为"初始化时注册一个 bearer → 后续不再同步"。但这需要 NewAPI 配合。

## 7. 架构健康度

| 维度 | 状态 | 说明 |
|------|------|------|
| 职责清晰度 | ✅ | 管控面/执行面分离明确 |
| 同步链路 | ✅ | model_limits 移除后只剩 status+group 同步，复杂度可控 |
| 故障隔离 | ✅ | NewAPI 不可达不影响 routing rule 管理和 precheck |
| 数据一致性 | ✅ | outbox + reconcile + cache invalidation 三重保障 |
| 可观测性 | ⚠️ | 建议给 precheck cache invalidation 加结构化日志 |
| 性能 | ✅ | precheck 热路径纯内存+Redis，不阻塞 proxy |
