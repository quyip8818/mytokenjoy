# SMS → TokenJoy 同步架构

> SMS 是模型、渠道、定价的唯一管控入口（配置中心）。
> TokenJoy 通过 River PeriodicJob 按分区 version 增量拉取配置快照。

---

## 数据流

```
┌───────────┐  成本价    ┌───────────────┐  按分区拉取       ┌────────────┐
│ SMS-NewAPI│──────────►│   SMS Admin   │◄────────────────│  TokenJoy  │
│ (供应商   │           │  (配置中心)   │────────────────►│  (客户侧) │
│  路由层)  │◄──────────│               │  返回变更分区数据  │            │
└───────────┘  售价回写  │  运维在此管理: │                   └────────────┘
                        │  • 模型目录    │
                        │  • 渠道配置    │
                        │  • 售价策略    │
                        └───────────────┘
```

---

## 同步协议

### 版本查询

```
GET /api/sync/versions → { channels: int, models: int, currencies: int }
```

TokenJoy 先拉 versions，与 `system_settings` 表中本地 version 比对，仅对差异分区拉取数据。

### 分区数据

```
GET /api/sync/catalog/channels  → { version: int, data: CatalogChannel[] }
GET /api/sync/catalog/models    → { version: int, data: CatalogModel[] }
```

响应体含 version + 全量数据。本地 version 仅在写入成功后更新为响应体的 version。

| 分区 | 变更频率 | 当前状态 |
|------|---------|---------|
| channels | 低 | 已实现（SMS 侧当前返回空） |
| models | 中 | 已实现 |
| currencies | 低 | 待实现 |

---

## TokenJoy 侧架构

### 调度方式

- **River PeriodicJob**：每 15 分钟（`SMS_SYNC_INTERVAL_SEC=900`）
- **手动触发**：`POST /api/sync/sms/trigger`（enqueue UniqueOpts 防重复）
- **去重**：`UniqueOpts{ByArgs: true}` — 同 kind 最多一个 pending job
- **重试**：MaxAttempts: 3

### Execute() 流程

```
1. FetchVersions() → 远端各分区 version
2. 逐分区比对本地 system_settings (初始 0)
3. 对 version 不同的分区:
   a. FetchChannels / FetchModels
   b. SyncTarget.Replace*()
   c. 写入成功 → 更新本地 version
4. 分区独立：一个失败不影响其他分区 version 推进
```

### SyncTarget 写入策略

| 分区 | 方法 | 行为 |
|------|------|------|
| channels | `ReplaceChannels` | upsert (前缀 `sms:`) + diff-delete 同前缀旧条目 + RebuildAbilities |
| models | `ReplaceModels` | per-company `DELETE WHERE source='sms'` + INSERT |
| models (pricing) | `ReplaceModelRatios` | 全局 UpsertModelRatio（NewAPI 不分 company） |

**安全约束**：
- 空 catalog 时跳过（防误删）
- 仅 `sms:` 前缀 channel 参与 diff-delete（不影响 provider key sync 创建的渠道）
- 手动模型（source='manual'）不受影响

### 多租户

`listActiveCompanyIDs()` 查全部激活 company → 逐个 ReplaceModels。私有化返回 1 个，SaaS 返回 N 个，代码统一。

---

## 认证

```
TokenJoy ── OAuth2 Client Credentials ──► SMS /api/oauth/token
         ── Bearer JWT (scope: sync:read) ──► SMS /api/sync/*
```

401 时 invalidate cached token → 重新获取 → 重试一次。

---

## 容错

| 场景 | 处理 |
|------|------|
| SMS 不可达 | Job error → 重试 3 次 → discard |
| 全部 version 相同 | Skip，不发请求 |
| 单分区写入失败 | 该分区 version 不更新，下次自动重试 |
| OAuth 401 | invalidate + retry once |
| 多副本部署 | River leader 调度，单实例执行 |

---

## 代码结构

```
apps/backend/internal/
├── infra/jobs/kinds_smssync.go            # SMSSyncArgs (River job 定义)
├── infra/river/workers/sms_sync.go        # River adapter → SMSSyncExecutor
├── infra/river/periodic/jobs.go           # PeriodicJob 注册（含 SMS sync）
├── integration/sms/
│   ├── client.go                          # HTTP Client + OAuth2 + 401 retry
│   └── types.go                           # PartitionVersions, CatalogModel, etc.
├── store/system_settings.go               # SystemSettings KV 接口
├── worker/smssync/
│   ├── execute.go                         # SMSSyncExecutor.Execute()
│   └── target.go                          # SyncTarget 实现
├── http/handler/sync/smssync.go           # POST /api/sync/sms/trigger
└── domain/adminport/port.go               # ListChannels + DeleteChannel

sms/backend/
├── internal/domain/sync/service.go        # GetVersions / GetModels / GetChannels
├── internal/http/handler/sync/handler.go  # /api/sync/* 路由
├── internal/store/postgres/sync_store.go  # sync_versions 表 + models 查询
└── schema.sql                             # sync_versions 表 + 自动 bump trigger
```

---

## 环境变量

| 变量 | 说明 | 默认 |
|------|------|------|
| `SMS_SYNC_ENABLED` | 总开关 | false |
| `SMS_API_BASE_URL` | SMS 地址 | — |
| `SMS_CLIENT_ID` | OAuth2 client_id | — |
| `SMS_CLIENT_SECRET` | OAuth2 client_secret | — |
| `SMS_SYNC_INTERVAL_SEC` | 轮询间隔（秒） | 900 |

---

## 后续

| 文档 | 内容 |
|------|------|
| `sms-currency-sync.md` | currencies 分区同步 |
| `global-models-catalog.md` | 全局 catalog + per-company allowlist（消除 per-company 复制） |
