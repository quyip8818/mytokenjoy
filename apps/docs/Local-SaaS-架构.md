# Local 私有化部署与 SaaS 平台关系架构

> TokenJoy 的 Local（客户私有化）部署如何与 SaaS 平台协作。
>
> 核心原则：**SaaS 是 Local 的默认 LLM 供应商。Local Ingest 与 SaaS 代码路径完全一致（含 lot 扣减 + wallet 维护）。Catalog Sync 同时同步模型定价和 SaaS 余额，定期对账修正偏差。组织预算树管控所有 channel 的内部使用。**

---

## 1. 一句话总结

客户在 SaaS 注册 Company，获得总 key。Local 默认 channel 用总 key 调 SaaS 模型。Local Ingest 与 SaaS 代码路径完全一致（含 lot 扣减 + wallet 维护）。Catalog Sync 定期从 SaaS 拉取模型定价 + 公司余额，余额覆盖写入 Local wallet 以修正偏差。组织预算树管控所有 channel 的员工使用。

---

## 2. 两层分离：结算 vs 管控

```
┌─ 结算层（各供应商各自收费，Local 不介入）───────────────────┐
│                                                             │
│  TokenJoy channel（默认）:                                  │
│    请求经过 SaaS Gateway → SaaS Ingest 扣公司 wallet（lot） │
│    充值：管理员登录 SaaS 付款                               │
│    SOT：SaaS wallet_remain_quota（lot 体系）                │
│                                                             │
│  自管 channel（阿里云 / Azure / 自建等）:                   │
│    请求直达供应商 API → 供应商直接向客户出账单               │
│    充值：客户在供应商平台操作                               │
│    SOT：供应商那边                                          │
│                                                             │
└─────────────────────────────────────────────────────────────┘

┌─ 管控层（管理员在 Local 操作）──────────────────────────────┐
│                                                             │
│  Local Wallet（SaaS 余额镜像 + Local Ingest 维护）:          │
│    初始值 = Catalog Sync 从 SaaS 拉取                       │
│    运行时 = Local Ingest 扣减（与 SaaS Ingest 同步扣）      │
│    定期对账 = Catalog Sync 覆盖写入修正偏差                  │
│    用途 = Gateway 预检拦截 TokenJoy channel 请求            │
│                                                             │
│  组织预算树（所有 channel 的消耗统一计入）:                  │
│    根部门总预算（可选，默认不限）                            │
│      ├─ 技术部: 30000                                      │
│      │   ├─ 员工 A / Key A: 5000                           │
│      │   └─ 项目 X / Key X: 10000                          │
│      └─ 产品部: 20000                                      │
│          └─ 员工 B / Key B: 3000                           │
│                                                             │
│  Gateway 预检:                                              │
│    ✓ 组织预算检查（所有 channel）                           │
│    ✓ Local wallet 检查（仅 TokenJoy channel）              │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**管理员心智模型**：
- 充钱（结算）→ 去各供应商平台（TokenJoy SaaS、阿里云等）
- 管人（管控）→ 在 Local 设预算分配；Local wallet 自动同步 SaaS 余额

---

## 3. 调用链路

### 3.1 走 TokenJoy channel（默认）

```
员工 SDK → Bearer sk-employee-xxx
    │
    ▼
Local Gateway (/v1/*)
  ├─ Precheck 1: 组织预算（combined_key_remain > 0）
  ├─ Precheck 2: Local wallet（SaaS 余额缓存 > minEstimate）
  └─ 通过 → Local NewAPI
                │
                ▼
         Channel: SaaS /v1/* (key=sk-company-xxx)
                │ HTTPS
                ▼
SaaS Gateway（验证总 key + SaaS wallet 预检）
  → SaaS NewAPI → LLM 供应商
  → SaaS Ingest: lot FIFO 扣减 → 更新 SaaS wallet_remain_quota
```

### 3.2 走自管 channel（阿里云等）

```
员工 SDK → Bearer sk-employee-xxx
    │
    ▼
Local Gateway (/v1/*)
  ├─ Precheck 1: 组织预算（combined_key_remain > 0）
  ├─ Precheck 2: 无 wallet 检查（自管渠道不经过 TokenJoy）
  └─ 通过 → Local NewAPI
                │
                ▼
         Channel: 阿里云 API (key=客户自己的key)
                │ HTTPS
                ▼
         阿里云 → LLM 供应商
         费用由阿里云直接向客户收取
```

### 3.3 Gateway 预检逻辑

```
请求进来 (员工 sk-xxx, 模型 M)
  │
  ├─ 1. Key 状态 / 过期 / 模型白名单（现有逻辑）
  │
  ├─ 2. 公司类型判断：
  │     ├─ selfhosted → 跳过 wallet 检查（步骤 4）
  │     │   理由：自管渠道可兜底，SaaS Gateway 是最终门槛
  │     └─ standard/trial → 执行步骤 4
  │
  ├─ 3. 组织预算检查（所有 channel，不区分类型）
  │     ├─ combined_key_remain > 0 ?
  │     └─ 失败 → 403 "预算不足"
  │
  └─ 4. Local wallet 检查（仅 standard/trial 公司）
        ├─ wallet_remain_quota > minEstimate ?
        ├─ 失败 → 403 "TokenJoy 余额不足，请充值"
        └─ 通过 → 放行
```

> **实现细节**：`PrecheckService.Run` 检测 `CompanyType == selfhosted` 时自动设置 `SkipWalletCheck = true`。
> 这是公司级别的跳过，不是模型级别。如果 selfhosted 公司的某个模型只有 platform channel，
> 请求仍会放行到 NewAPI → SaaS Gateway → SaaS 自己的 wallet precheck 兜底拒绝。

---

## 4. Local Ingest

Local Ingest 与 SaaS Ingest **代码路径完全一致**（同一份代码、同一条逻辑）。

**分流规则（SaaS 和 Local 都一样）**：
- 平台渠道（TokenJoy channel）消耗 → 完整路径（lot 扣减 + wallet + budget + ledger）
- 自管渠道消耗 → 只记账（budget + ledger），**不扣 lot / wallet**

**如何判断 channel 类型**：NewAPI logs 表有 `channel_id` 字段（patch 添加）。Ingest 比较 `log.channel_id` 与 `system_settings.platform_channel_id`：
- 匹配 → 平台渠道
- 不匹配 → 自管渠道
- `channel_id = 0`（无此字段的旧日志）→ 视为平台渠道（向后兼容）
- `platform_channel_id` 未配置 → 视为平台渠道（SaaS 模式或 Setup 前）

```
NewAPI log 进来:
  │
  ├─ token_id → platform_key_mapping → 归因（员工/项目/部门）
  │
  ├─ 写 usage_ledger（所有 channel）
  │
  ├─ 累加 budget_consumed（所有 channel）
  │     ├─ platform_key 轴
  │     ├─ member / project 轴（按 key scope）
  │     └─ combined_key_remain 递减
  │
  ├─ if 平台渠道:
  │     ├─ FIFO ConsumeLots → 更新 wallet_remain_quota
  │     └─ post-commit: ManageUser set_quota → NewAPI
  │
  └─ if 自管渠道:
        └─ 跳过 lot/wallet（只记账）
```

### 与 SaaS Ingest 对比

| 步骤 | SaaS | Local | 差异 |
|------|------|-------|------|
| 归因 | ✅ | ✅ | 无 |
| 写 usage_ledger | ✅ | ✅ | 无 |
| 累加 budget_consumed | ✅ | ✅ | 无 |
| 递减 combined_key_remain | ✅ | ✅ | 无 |
| 平台渠道：FIFO ConsumeLots | ✅ | ✅ | 无 |
| 平台渠道：SetWalletRemain | ✅ | ✅ | 无 |
| 平台渠道：set_quota → NewAPI | ✅ | ✅ | 无 |
| 自管渠道：跳过 lot/wallet | ✅ | ✅ | 无 |

**零差异。** 同一份代码。分流逻辑在 Ingest 内部根据 channel 类型判断，与部署模式无关。

### Local wallet 的对账

Local Ingest 对平台渠道消耗执行 lot 扣减，所以 Local wallet 实时下降。Catalog Sync 每 10min 拉取 SaaS 真实余额覆盖写入，修正定价时间差导致的微小偏差。

---

## 5. Catalog Sync（含余额对账）

Catalog Sync 是一个统一的定期同步 Worker，通过 version 门控拉取多个独立数据通道。

```
Catalog Sync Worker（每 10min，River PeriodicJob）:

  1. GET SaaS /api/platform/sync/versions （需要 sync token）
     → 返回 { models: N, pricing: N, currencies: N, discounts: N, walletLots: N }
     → 全局版本（models/pricing/currencies）+ per-company 版本（discounts/walletLots）

  2. 比较本地 sync_versions 表中 (GlobalSyncVersion, "models") vs 远端
     → 不同 → GET /api/platform/sync/catalog/models
     → 更新本地 models 表 + NewAPI model_ratio
     → Set 本地 version = resp.Version

  3. 比较本地 sync_versions (GlobalSyncVersion, "pricing") vs 远端
     → 不同 → GET /api/platform/sync/catalog/pricing （需要 sync token）
     → push NewAPI ratio（UpsertModelRatio）
     → Set 本地 version

  4. 比较本地 sync_versions (GlobalSyncVersion, "currencies") vs 远端
     → 不同 → GET /api/platform/sync/catalog/currencies
     → 遍历返回的 rows，逐条 INSERT ON CONFLICT (id) DO NOTHING（id 幂等）
     → Set 本地 version

  5. 比较本地 sync_versions (GlobalSyncVersion, "discounts") vs 远端
     → 不同 → GET /api/platform/sync/catalog/discounts （需要 sync token）
     → 遍历返回的 rows，逐条 INSERT ON CONFLICT (id) DO NOTHING（id 幂等）
     → Set 本地 version

  6. 比较本地 sync_versions (GlobalSyncVersion, "wallet_lots") vs 远端
     → 不同 → GET /api/platform/sync/catalog/wallet_lots （需要 sync token）
     → 返回 { data: [...lots], orders: [...orders], transactions: [...lot_transactions], walletRemainQuota: N }
     → Upsert orders 到本地 company_recharge_orders 表
     → Upsert lots 到本地 company_recharge_lots 表（含 exhausted，保留原始 kind）
     → Upsert lot_transactions 到本地（UUID 幂等，ON CONFLICT DO NOTHING）
     → 覆盖写入 companies.wallet_remain_quota（对账修正）
     → Set 本地 version
```

| 属性 | 说明 |
|------|------|
| 频率 | 默认 10min（`CATALOG_SYNC_INTERVAL_SEC`） |
| 余额对账 | wallet_lots 通道：SaaS 真实余额覆盖 Local wallet |
| lot 同步 | 直接镜像 SaaS 的 lot + order + transaction 列表（含 exhausted），保留原始 kind |
| version 门控 | 各通道独立版本号（`sync_versions` 表），无变化时跳过（避免无谓 IO） |
| version bump | SaaS 侧：充值/退费/Ingest 后 per-company bump `wallet_lots`；折扣变更 per-company bump `discounts`；全局操作 bump `models`/`pricing`/`currencies` |
| 失败处理 | 保留上次数据继续用；SaaS Gateway 兜底 |
| 认证 | /sync/versions + pricing + discounts + wallet_lots 需要 sync token（cst_ 前缀） |

---

## 6. 关键概念

### 6.1 总 key（SaaS 侧分配）

| 属性 | 说明 |
|------|------|
| 格式 | `sk-company-xxx` |
| 归属 | SaaS Company wallet user |
| 预算 | 无上限（消耗直扣公司 SaaS wallet） |
| 用途 | Local NewAPI 默认 channel 的 upstream key |

```
Local NewAPI Channel 配置（自动）:
  name: tokenjoy-upstream
  base_url: https://app.tokenjoy.me/v1
  key: sk-company-xxx
  models: [从 Catalog Sync 获取]
```

### 6.2 Local NewAPI 的角色

| 功能 | 说明 |
|------|------|
| Token 管理 | 每个员工 platform key → NewAPI token |
| 多 Channel 路由 | 按模型匹配 channel（TokenJoy / 阿里云 / ...） |
| 记账 | 写 logs（token_id, model, tokens），供 Local Ingest |
| Quota | 由 Local Ingest post-commit set_quota 维护（与 SaaS 一致） |

### 6.3 管理员账号统一

Setup 时的管理员邮箱+密码同时在 SaaS 创建 User + Member：
- 登录 SaaS：充值、查看公司总消耗
- 登录 Local：管理员工、分配预算、查看调用明细
- 同一个人、同一套凭证、两个入口

---

## 7. 生命周期

### 7.1 首次启动（Setup）

```
1. 管理员部署 Local 实例，首次启动进入 setupServer
2. 填写：公司名 / 行业 / 规模 / 管理员邮箱+密码
3. Local → POST SaaS /api/platform/register-local
   ├─ Header: X-Registration-Secret
   └─ Body: { name, industry, size, adminEmail, adminPassword, adminName, idempotencyKey }
4. SaaS 侧（幂等）：
   ├─ 创建 admin User（或复用已有 email）
   ├─ 创建 type=selfhosted Company（含 NewAPI wallet user + org tree）
   ├─ 创建 admin Member（super_admin 角色）
   ├─ 在 wallet user 上创建 unlimited quota Token（总 key）
   └─ 签发 sync token（cst_ 前缀，用于 Catalog Sync 认证）
5. SaaS 返回 { companyId, walletUserId, platformKey, syncToken }
6. Local 持久化（单事务）：
   ├─ system_settings.setup_company_id = companyId
   ├─ system_settings.setup_company_name = name
   ├─ system_settings.setup_admin_email = adminEmail
   ├─ system_settings.catalog_sync_token = syncToken
   ├─ system_settings.saas_platform_key = platformKey
   ├─ system_settings.saas_wallet_user_id = walletUserId
   └─ 创建本地 admin user（用于 Local 登录）
7. Local 正常启动 → ensurePlatformChannel:
   ├─ 用 platformKey 调 adminport.UpsertChannel 创建 tokenjoy-upstream
   └─ system_settings.platform_channel_id = channelId
8. Local 完成 bootstrap（seed、根部门）
9. 正常运行（根部门 budget 默认不限）
```

### 7.2 日常运行

```
┌─ Local ───────────────────────────────────────────────────┐
│                                                           │
│  员工调 API:                                              │
│    sk-employee → Gateway precheck → NewAPI → channel      │
│                                                           │
│  Local Ingest（与 SaaS 完全一致）:                        │
│    NewAPI logs → 归因 → lot FIFO → wallet → budget        │
│                                                           │
│  Catalog Sync（每 10min，含余额对账）:                     │
│    拉 SaaS 模型+定价+余额 → 更新本地 → 覆盖 wallet 修正  │
│                                                           │
└───────────────────────────────────────────────────────────┘

┌─ SaaS（自动）────────────────────────────────────────────┐
│                                                           │
│  总 key 的请求 → SaaS Gateway → SaaS NewAPI → LLM        │
│  SaaS Ingest → lot FIFO → wallet_remain_quota 更新        │
│                                                           │
└───────────────────────────────────────────────────────────┘
```

### 7.3 充值

| 供应商 | 操作 | Local 影响 |
|--------|------|-----------|
| TokenJoy | 管理员登录 SaaS → 充值 | 下次 Catalog Sync 余额对账自动涨 |
| 阿里云 | 管理员登录阿里云 → 充值 | 无影响（与 Local 无关） |

### 7.4 余额不足

| 场景 | 谁拦截 | 员工看到什么 |
|------|--------|-------------|
| 员工预算用完 | Local Gateway（budget 预检） | "预算不足" |
| TokenJoy 余额不足 | Local Gateway（wallet 预检）或 SaaS Gateway | "余额不足，请充值" |
| 阿里云余额不足 | 阿里云 API 返回错误 | "上游服务返回错误" |

**双重保护（TokenJoy channel）**：Local wallet 预检 + SaaS Gateway 预检。即使 Local 缓存有延迟，SaaS 仍是最终兜底。

---

## 8. Local "钱包"页面

```
┌─ 钱包 ─────────────────────────────────────────────────────┐
│                                                             │
│  ─── TokenJoy 余额（自动同步）───                          │
│  余额: ¥8,320                                              │
│  本月消耗: ¥1,680                                          │
│  [去充值 →]（跳转 SaaS）                                    │
│                                                             │
│  ─── 自管渠道消耗（统计，非结算）───                        │
│  阿里云: 本月 ¥520                                         │
│  Azure: 本月 ¥340                                          │
│                                                             │
│  ─── 合计 ───                                              │
│  本月总消耗: ¥2,540                                        │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

- TokenJoy 余额 = Local wallet（Wallet Sync 写入），有真实结算意义
- 自管渠道消耗 = `usage_ledger` 按 channel 聚合，纯统计展示

---

## 9. 数据 SOT 分布

| 数据 | SOT | Local 怎么获取 |
|------|-----|---------------|
| 模型目录 + 定价 | SaaS（定价 SOT 在 NewAPI ratio store） | Catalog Sync models + pricing 通道（pricing 只 push NewAPI，不写 DB） |
| 公司 TokenJoy 余额 | SaaS (`wallet_remain_quota` + lot 体系) | Catalog Sync wallet_lots 通道（含 lot 列表镜像） |
| 公司 TokenJoy 总消耗 | SaaS | SaaS Ingest 按总 key 记账 |
| 组织预算 (budget/limit) | Local | 管理员在 Local 设置 |
| 员工消耗归因 | Local | Local Ingest → `budget_consumed` + `usage_ledger` |
| 员工/项目/Key 列表 | Local | SaaS 不知道内部结构 |
| 自管渠道消耗明细 | Local | Local NewAPI logs + Local Ingest |
| 币种配置 | SaaS | Catalog Sync currencies 通道 |

---

## 10. SaaS 与 Local 之间的接口

| 方向 | 接口 | 频率 | 认证 | 说明 |
|------|------|------|------|------|
| Local → SaaS | `POST /api/platform/register-local` | 一次性 | X-Registration-Secret | Setup 注册，返回 companyId + platformKey + syncToken + walletUserId |
| Local ← SaaS | `GET /api/platform/sync/versions` | 每 10min | 无 | 各通道版本号（models, pricing, currencies, walletLots） |
| Local ← SaaS | `GET /api/platform/sync/catalog/models` | 按需 | 无 | 模型目录（version 变化时拉取） |
| Local ← SaaS | `GET /api/platform/sync/catalog/pricing` | 按需 | Bearer cst_xxx | 定价（含合约价，version 变化时拉取） |
| Local ← SaaS | `GET /api/platform/sync/catalog/currencies` | 按需 | 无 | 币种列表（version 变化时拉取） |
| Local ← SaaS | `GET /api/platform/sync/catalog/wallet_lots` | 按需 | Bearer cst_xxx | 公司 lot 列表 + order 列表 + walletRemainQuota（version 变化时拉取） |
| Local → SaaS | SaaS `/v1/*` Gateway | 实时 | Bearer sk-company-xxx | LLM 请求（总 key） |

---

## 11. 配置清单

### Local 侧

| 环境变量 | 说明 |
|----------|------|
| `SUPPORT_SAAS=false` | 私有化模式 |
| `SAAS_PLATFORM_URL` | SaaS 地址 |
| `SAAS_REGISTRATION_SECRET` | 注册密钥 |
| `CATALOG_SYNC_INTERVAL_SEC` | Catalog Sync 间隔，默认 600s |

运行时 `system_settings`：

| key | 写入时机 | 说明 |
|-----|----------|------|
| `setup_company_id` | Setup | SaaS 分配的 companyId |
| `setup_company_name` | Setup | 公司名 |
| `setup_admin_email` | Setup | 管理员邮箱 |
| `catalog_sync_token` | Setup | cst_ 前缀 sync token |
| `saas_platform_key` | Setup | 总 key（sk-xxx） |
| `saas_wallet_user_id` | Setup | NewAPI wallet user ID |
| `platform_channel_id` | 启动时（ensurePlatformChannel） | tokenjoy-upstream channel ID |

Local 的 catalog sync 版本号存储在 `sync_versions` 表（统一用 `GlobalSyncVersion` 作为 company_id）：

| type | 说明 |
|------|------|
| `models` | 本地模型版本号 |
| `pricing` | 本地定价版本号 |
| `currencies` | 本地币种版本号 |
| `discounts` | 本地折扣版本号 |
| `wallet_lots` | 本地 wallet lots 版本号 |

### SaaS 侧

| 环境变量 | 说明 |
|----------|------|
| `SUPPORT_SAAS=true` | SaaS 模式 |
| `LOCAL_REGISTRATION_SECRET` | 接受 Local 注册的密钥 |

SaaS `sync_versions` 表（自动维护）：

| company_id | type | bump 时机 | 说明 |
|------------|------|-----------|------|
| GlobalSyncVersion | `models` | 创建/更新/删除模型 + PublishCatalog | 模型目录版本（全局） |
| GlobalSyncVersion | `pricing` | SetGlobalPricing / SetModelPricing / CreateModel（有价时） | 定价版本（全局） |
| GlobalSyncVersion | `currencies` | 创建/更新币种 | 币种版本（全局） |
| `<companyID>` | `discounts` | SetCompanyDiscount | 折扣版本（per-company） |
| `<companyID>` | `wallet_lots` | 充值（CreditFromLot）/ Ingest 扣 lot | wallet lots 版本（per-company） |

---

## 12. 与 SaaS 版代码路径对比

| 模块 | SaaS | Local | 差异 |
|------|------|-------|------|
| 组织预算树 | ✅ | ✅ | 无 |
| budget_consumed | ✅ | ✅ | 无 |
| combined_key_remain | ✅ | ✅ | 无 |
| Gateway precheck | ✅ | ✅ | selfhosted 跳过 wallet check（SaaS 兜底） |
| Ingest - 平台渠道：lot + wallet | ✅ | ✅ | 无 |
| Ingest - 自管渠道：只记账 | ✅ | ✅ | 无 |
| Ingest - set_quota NewAPI | ✅ | ✅ | 无 |
| 充值（lot 创建） | ✅ 真实付款 | ✅ SaaS 充值 → Catalog Sync 同步 lot | 充值入口在 SaaS |
| Catalog Sync | 不需要 | ✅ models + pricing + currencies + wallet_lots | 仅 Local |
| 自管渠道 | ✅ | ✅ | 无 |

**总结**：SaaS 和 Local 的 Ingest/预算代码**完全一致**（含自管渠道分流逻辑）。Gateway precheck 唯一差异：selfhosted 公司跳过 wallet 检查。Local 新增 Catalog Sync Worker（4 个版本门控通道）。

---

## 13. 类比理解

```
TokenJoy SaaS   ≈ 电信运营商（话费余额在运营商那里）
阿里云          ≈ 另一家运营商（另一张卡）
公司总 key      ≈ 企业主卡
Local 实例      ≈ 企业 PBX（电话交换机）
员工 Key        ≈ 分机号
Local wallet    ≈ PBX 上显示的"主卡余额"（每分钟从运营商同步）
组织预算树      ≈ 各部门电话费限额（不管打哪家运营商都算）

PBX（Local）管的是：
  - 谁有分机号
  - 每人限打多少（预算）
  - 通话记录
  - 显示主卡余额（同步自运营商）

运营商（SaaS）管的是：
  - 话费计费（lot 扣减）
  - 余额不够停机
  - 充值
```

---

## 14. 设计决策记录

| 决策 | 选择 | 理由 |
|------|------|------|
| 自管渠道 Ingest 是否扣 wallet | **否，只记账**（SaaS 和 Local 都一样） | 自管渠道费用由供应商收取，不经过 TokenJoy 结算 |
| Local Ingest 是否与 SaaS 有差异 | **无差异，同一份代码** | 分流逻辑是按 channel 类型，不是按部署模式 |
| Local wallet 如何对账 | Catalog Sync wallet_lots 通道覆盖写入 | version 门控，变化时才拉取 |
| Catalog Sync API 结构 | **独立 endpoint + version 门控** | 各通道独立演进，version 避免无谓 IO |
| Lot 同步策略 | **直接镜像 SaaS lot（保留原始 kind）** | 不需要新 lot kind，FIFO 链表直接对齐 SaaS |
| Gateway wallet 检查粒度 | **公司级跳过（selfhosted 跳过，standard 检查）** | 实现简单；SaaS Gateway 兜底平台渠道；未来可升级到模型级 |
| Local wallet 只拦截哪些公司 | standard/trial 公司检查 wallet，selfhosted 跳过 | selfhosted 有自管渠道兜底，SaaS 是最终门槛 |
| 组织预算是否区分 channel | 不区分，统一计入 | 预算管"使用量"，不管钱付给谁 |
| 根部门 budget 是否必填 | 默认不限 | 开箱即用 |
| Local 充值入口 | SaaS 侧 | Local 不处理真实资金流 |
| register-local 做什么 | 创建 User + Company + Member + Token + SyncToken（一次调用） | 减少 Setup 步骤，幂等设计 |

---

## 15. 边界情况

### 15.1 网络中断

| 影响 | TokenJoy channel | 自管 channel |
|------|-----------------|-------------|
| LLM 请求 | 失败 | 取决于该 channel 网络 |
| Local 管理面 | 正常 | 正常 |
| Wallet Sync | 保留上次值 | — |

### 15.2 Catalog Sync 延迟 / 失败

- 两次 sync 之间（最长 10min），Local wallet 由 Ingest 自己扣减维护
- 如果 Ingest 和 SaaS 扣的数值有微小偏差 → 下次 Catalog Sync 自动修正
- Catalog Sync 失败 → 保留上次数据，SaaS Gateway 兜底
- 管理员充值后 → 最多 10min 后 Local 看到新余额

### 15.3 总 key 泄露

- 风险：绕过 Local 直接用总 key 调 SaaS
- 缓解：总 key 只在服务器配置中
- 加固：SaaS IP 白名单

### 15.4 定价时间差

- Catalog Sync 有 5 分钟延迟
- Local `budget_consumed` 的金额可能与 SaaS lot 扣减的金额有微小差异
- 可接受：预算是管控工具，不需要精确到分
