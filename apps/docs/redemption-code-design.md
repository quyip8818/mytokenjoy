# 兑换码充值功能设计文档

> **状态**：Draft  
> **最后更新**：2026-08-07

---

## 1. 背景与动机

TokenJoy 当前支持两种充值路径：

1. **自助充值**（Self Recharge）：客户在线支付 → 创建 pending order → 确认支付 → 生成 paid lot
2. **平台运营充值**（Platform Recharge/Gift/Adjust）：平台管理员手动为企业充值或赠送额度

缺少一种 **离线分发** 路径——即平台生成兑换码，通过线下渠道（销售邮件、活动赠送、合作伙伴分发等）交付给客户，客户自行输入兑换码完成充值。

### 业务场景

| 场景 | 说明 |
|------|------|
| 销售赠送 | BD 谈客户时提供体验额度，无需客户绑定支付方式 |
| 活动促销 | 线上/线下活动批量发放额度码 |
| 渠道合作 | 代理商批量采购兑换码分发给终端客户 |
| 内部测试 | 给测试账号快速注入额度 |
| 客服补偿 | 服务质量问题时发放补偿码 |

### 行业参考

| 产品 | 方案要点 |
|------|---------|
| **NewAPI / One-API** | 管理员批量生成码（批次名+面值+数量），用户输入码充值到余额。曾暴露并发竞态漏洞（未正确加锁导致同一码被多人同时兑换） |
| **OpenAI Credit Grants** | 多来源 credits（注册赠金、promo codes、researcher program），有效期 12 个月，按来源优先级消耗 |
| **通用 SaaS** | redemption_codes 表 + batch 管理，码有状态机（unused → used / expired），需要行级锁或乐观锁保证并发安全 |

---

## 2. 产品定义

### 2.1 核心概念

| 概念 | 定义 |
|------|------|
| **兑换码（Redemption Code）** | 一个可被兑换一次的字符串凭证，持有面值额度 |
| **批次（Batch）** | 一次性批量生成的一组兑换码，共享相同面值和有效期 |
| **面值（Face Value）** | 兑换码对应的金额（billing currency），兑换后按当时 quota_per_unit 折算为 quota |
| **兑换（Redeem）** | 持有 billing:manage 权限的成员输入兑换码，将其面值充入公司钱包 |

### 2.2 用户角色

| 角色 | 能力 |
|------|------|
| **平台管理员** | 批量生成码、查看/搜索码、禁用码 |
| **企业管理员**（billing:manage） | 兑换码（输入码充值到本公司钱包） |
| **企业普通成员**（billing:read） | 查看兑换记录（作为充值记录的一部分） |

### 2.3 兑换码规格

| 属性 | 规格 |
|------|------|
| 格式 | `TJ-XXXX-XXXX-XXXX`（16 位字母数字，分 4 段，前缀 TJ，大写） |
| 字符集 | `0-9A-Z` 去掉易混淆字符 `0O1IL`，实际集 = `23456789ABCDEFGHJKMNPQRSTUVWXYZ`（30 chars） |
| 碰撞空间 | 30^12 ≈ 5.3×10^17，批量生成时 DB unique 约束兜底 |
| 使用次数 | 单次（一码一兑） |
| 有效期 | 批次级设置，默认 365 天，最短 1 天，最长 3 年 |
| 面值范围 | 0.01 ~ 100,000（billing currency） |

---

## 3. 数据模型

### 3.1 新增表

单表设计，批次概念退化为 `batch_name` 标签字段。后续如需批次级统计/策略再拆表，当前无旧数据无需过度设计。

```sql
CREATE TABLE redemption_codes (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code                TEXT NOT NULL UNIQUE,             -- 兑换码明文
    batch_name          TEXT NOT NULL DEFAULT '',         -- 批次标签（运营备注，批量生成时统一值）
    face_value          NUMERIC(12,2) NOT NULL,          -- 面值（billing currency）
    currency            TEXT NOT NULL DEFAULT 'CNY',
    status              TEXT NOT NULL DEFAULT 'unused',   -- unused | used | disabled
    -- 兑换信息（status=used 时填充）
    redeemed_by_company UUID,                            -- 兑换的公司
    redeemed_by_member  UUID,                            -- 操作成员
    redeemed_at         TIMESTAMPTZ,
    recharge_order_id   UUID,                            -- 关联的充值订单
    -- 元数据
    expires_at          TIMESTAMPTZ NOT NULL,
    created_by          UUID NOT NULL,                   -- 创建人（platform admin）
    note                TEXT NOT NULL DEFAULT '',         -- 备注
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_redemption_codes_code ON redemption_codes(code);
CREATE INDEX idx_redemption_codes_batch ON redemption_codes(batch_name) WHERE batch_name != '';
CREATE INDEX idx_redemption_codes_status ON redemption_codes(status) WHERE status = 'unused';
```

> **为什么不拆 batches 表？** 当前码生成量小，批次只是运营分组标签，不需要独立生命周期管理。`WHERE batch_name = ?` 即可满足按批次查询/导出需求。

### 3.2 与现有模型集成

兑换码充值复用现有 `RechargeOrder` + `RechargeLot` 路径：

```
兑换码 → 新建 RechargeOrder(source="redemption", lot_kind="gift")
       → BuildLot → CreditFromLot → syncWallet
```

| 字段 | 值 |
|------|---|
| `RechargeOrder.Source` | `"redemption"`（新增常量） |
| `RechargeOrder.LotKind` | `"gift"`（兑换码本质是赠送额度） |
| `RechargeOrder.Amount` | `0`（无实际支付） |
| `RechargeLot.PaidAmount` | `0` |
| `RechargeOrder.CreatedBy` | 兑换操作人 memberID |

新增 store 常量：

```go
RechargeSourceRedemption = "redemption"
```

---

## 4. 接口设计

### MVP 接口清单（3 个）

| # | 方法 | 路径 | 权限 | 说明 |
|---|------|------|------|------|
| 1 | POST | `/api/platform/redemption-codes/generate` | `platform:manage` | 批量生成码 |
| 2 | GET | `/api/platform/redemption-codes` | `platform:manage` | 列出码（含筛选分页） |
| 3 | POST | `/api/billing/redeem` | `billing:manage` | 兑换码充值 |

> Phase 2 再加：`POST .../disable`（禁用）。导出 CSV 由前端从列表数据生成，不需要后端接口。

### 4.1 平台管理接口（SaaS only, platform:manage）

#### POST /api/platform/redemption-codes/generate

批量生成兑换码。

```json
// Request
{
  "batchName": "2026春节活动",
  "faceValue": 100.00,
  "quantity": 50,
  "expiresInDays": 365,
  "note": "春节促销活动赠送"
}

// Response 201
{
  "batchName": "2026春节活动",
  "faceValue": 100.00,
  "quantity": 50,
  "expiresAt": "2027-08-07T00:00:00Z"
}
```

#### GET /api/platform/redemption-codes

列出兑换码（支持按 batch_name / status 筛选，分页）。

```json
// Query: ?batchName=2026春节活动&status=unused&page=1&pageSize=20

// Response 200
{
  "items": [
    {
      "id": "uuid",
      "code": "TJ-A3B4-C5D6-E7F8",
      "batchName": "2026春节活动",
      "faceValue": 100.00,
      "status": "unused",
      "redeemedByCompany": null,
      "redeemedAt": null,
      "expiresAt": "..."
    }
  ],
  "total": 50
}
```

### 4.2 企业兑换接口（billing:manage, SaaS only）

#### POST /api/billing/redeem

兑换码充值。

```json
// Request
{
  "code": "TJ-A3B4-C5D6-E7F8"
}

// Response 200
{
  "faceValue": 100.00,
  "currency": "CNY",
  "quotaGranted": 100000
}
```

前端收到成功响应后 `invalidateQueries(['billing', 'wallet'])` 刷新钱包余额，不依赖 response 返回 balance。

```json
// Error 400 — 码无效/已使用/已过期/已禁用
{
  "code": "INVALID_REDEMPTION_CODE",
  "message": "兑换码无效或已被使用"
}
```

### 4.3 路由注册方式

`/billing/redeem` 不能混进现有 billing handler 的 `RegisterRoutes`（因为 billing handler 在 SaaS/Local 都注册），需要**条件注册**：

```go
// billing handler 入口
func Mount(r chi.Router, d httpdeps.Deps) {
    h := NewHandler(d.Protected(), d.BillingSvc, d.ModelDiscount())
    h.RegisterRoutes(r) // 现有路由，SaaS/Local 都注册

    // 兑换码路由仅 SaaS 注册
    if d.Config().SupportSaas {
        write := httpmiddleware.ReadRoutes(r, d.Protected(), grants.BillingManage)
        write.Post("/billing/redeem", h.Redeem)
    }
}
```

平台侧路由无需额外处理——已有 `RequirePlatformAdmin` 三层隔离（见 permission-hierarchy.md §7）。

---

## 5. 核心流程

### 5.1 生成流程（平台管理员）

```
平台管理员 → POST /api/platform/redemption-codes/generate
           → 校验参数（面值>0, 数量≤1000, 有效期合规）
           → 批量生成 N 个随机码 → batch INSERT redemption_codes (同一 batch_name)
           → 返回批次信息
```

生成码的随机性：使用 `crypto/rand` 生成，DB unique 约束防碰撞，碰撞时重试（概率极低）。

### 5.2 兑换流程（企业管理员）

```
企业管理员 → POST /api/billing/redeem { code }
           → 标准化输入（去空格、统一大写、去掉分隔符后重新格式化）
           → BEGIN TX
             → SELECT * FROM redemption_codes WHERE code = $1 FOR UPDATE
             → 校验：status='unused' ∧ expires_at > now()
             → 创建 RechargeOrder(source=redemption, lot_kind=gift)
             → BuildLot → CreditFromLot（原子写入 lot + wallet delta）
             → UPDATE redemption_codes SET status='used', redeemed_by_*, recharge_order_id
           → COMMIT
           → recordCreditTransaction（审计）
           → syncWalletBestEffort（同步 NewAPI）
           → 返回成功 + 面值 + quotaGranted
```

**并发安全**：`SELECT ... FOR UPDATE` 行级锁确保同一码不会被并发兑换（吸取 One-API 教训）。

### 5.3 状态机

```
unused ──兑换──→ used      （终态）
unused ──过期──→ expired   （逻辑状态，查询时判断 expires_at < now()）
unused ──禁用──→ disabled  （终态）
```

expired 不写入数据库，查询时动态判断：`status = 'unused' AND expires_at < now()` → 展示为 expired。

---

## 6. 前端设计

### 6.1 现状

充值页面（/billing）已包含兑换码充值区域：
- 上半部分：在线充值（自定义金额 + 预设金额卡片 + 支付方式选择）
- 下半部分：兑换码充值卡片（输入框 + 兑换按钮）

### 6.2 企业侧 — 兑换码 UX 优化

当前实现的改进点：

| 优先级 | 改动 | 理由 |
|--------|------|------|
| **P0** | 兑换按钮加 loading + disabled 防重复提交 | 避免双击导致并发请求 |
| **P0** | 输入框自动大写 + 去空格标准化 | 减少因大小写/空格导致的兑换失败 |
| **P1** | 成功反馈明确显示面值金额（"成功充值 ¥100"） | 用户需要知道充了多少钱 |
| **P1** | 失败原因细分展示（过期/已使用/无效/已禁用） | 比统一"失败"更友好 |
| **P2** | 输入框 placeholder 带格式提示（`TJ-XXXX-XXXX-XXXX`） | 引导正确输入 |
| **P3** | 输入框做格式化 mask（粘贴时自动加连字符） | 锦上添花 |

#### 错误码映射

| 后端错误码 | 前端展示 |
|-----------|---------|
| `INVALID_REDEMPTION_CODE` | 兑换码无效，请检查输入 |
| `CODE_ALREADY_USED` | 该兑换码已被使用 |
| `CODE_EXPIRED` | 该兑换码已过期 |
| `CODE_DISABLED` | 该兑换码已被禁用 |
| `TRIAL_NO_TOPUP` | 试用环境不支持兑换，升级后可使用 |

### 6.3 平台管理侧 — 兑换码管理

在 **平台管理** nav group 下新增页面 `/platform/redemption`：

- 码列表（支持按 batch_name / status 筛选）
- 生成 Dialog（批次名、面值、数量、有效天数、备注）
- 码值部分脱敏显示（`TJ-A3B4-****-****`）
- 单码禁用

权限：`platform:manage`

### 6.4 前端文件结构

```
apps/frontend/src/
├── features/billing/
│   ├── components/redeem-section.tsx      ← 兑换码输入区（已存在，需优化）
│   └── hooks/use-redeem.ts               ← 兑换 mutation
├── features/platform-redemption/          ← 平台管理功能模块
│   ├── index.ts
│   ├── components/
│   │   ├── code-list.tsx
│   │   └── generate-codes-dialog.tsx
│   └── hooks/
│       └── use-redemption-codes.ts
├── routes/platform/redemption.tsx         ← 页面
└── api/platform-redemption.ts            ← API 层
```

---

## 7. 后端模块结构

```
apps/backend/internal/
├── domain/billing/
│   ├── redeem.go                          ← 兑换逻辑（核心，SaaS only）
│   └── (现有 lot_confirm.go 等保持不变)
├── domain/redemption/                     ← 兑换码管理 domain（SaaS only）
│   ├── service.go                         ← 码生成、查询、禁用
│   └── codegen.go                         ← 码生成算法
├── store/
│   ├── redemption_repo.go                 ← repository interface
│   └── pg/redemption.go                   ← PG 实现（SaaS schema 含此表）
└── http/handler/
    ├── billing/handler.go                 ← 新增 POST /billing/redeem（SaaS only 注册）
    └── platform/redemption_handler.go     ← 平台管理路由（已有 platform 隔离）
```

---

## 8. 权限设计

### 企业侧

| 操作 | 权限 |
|------|------|
| 兑换码充值 | `billing:manage` |
| 查看兑换记录（通过充值记录） | `billing:read` |

无需新增权限，复用现有 billing 权限体系。

### 平台侧

| 操作 | 权限 |
|------|------|
| 全部管理操作 | `platform:manage`（RequirePlatformAdmin） |

无需新增权限，兑换码管理是平台运营功能。

---

## 9. 安全考虑

| 风险 | 缓解措施 |
|------|---------|
| 并发兑换（竞态） | `SELECT ... FOR UPDATE` 行级锁，事务内完成状态变更 |
| 暴力枚举 | 30^12 码空间 + 接口限流（每用户 5次/分钟） + 连续失败锁定 |
| 码泄露 | 平台列表中码值部分脱敏显示；导出文件建议加密分发 |
| 试用公司兑换 | 复用现有 `isTrialOrDemoCompany` 检查，试用/Demo 公司禁止兑换 |
| 批量滥用 | 单批次上限 1000 码；单次兑换不做批量（一码一请求） |
| SQL 注入 | 参数化查询（标准做法） |
| Local 误调用 | `/billing/redeem` 仅 SaaS 模式注册，Local 路由不存在（404）；前端条件渲染不显示 |

---

## 10. 可观测性

| 事件 | 记录方式 |
|------|---------|
| 批次创建 | platform operation log |
| 码兑换成功 | LotTransaction(action="credit", source="redemption") + 充值记录 |
| 码兑换失败 | 结构化日志（code hash, 原因, 请求者 IP） |
| 码禁用 | platform operation log |

---

## 11. SaaS / Local 架构边界

### 核心原则

**兑换码全生命周期仅在 SaaS 侧。** Local 部署不持有 `redemption_codes` 表，不暴露兑换相关接口。

### 数据流

```
┌─────────────────────────────────────────────────┐
│                    SaaS                          │
│                                                 │
│  platform admin → 生成码 → redemption_codes     │
│  企业用户 → /billing/redeem → 校验码 →          │
│    → RechargeOrder + RechargeLot (写入 SaaS DB) │
│    → syncWalletBestEffort → bump sync_version   │
│                                                 │
└──────────────────────┬──────────────────────────┘
                       │ catalog sync (wallet_lots)
                       ▼
┌─────────────────────────────────────────────────┐
│                   Local                          │
│                                                 │
│  sync worker → UpsertOrderFromSync              │
│             → UpsertLotFromSync                 │
│  企业钱包余额自动更新                            │
│                                                 │
│  ❌ 无 redemption_codes 表                      │
│  ❌ 无 /billing/redeem 路由                     │
│  ❌ 前端不渲染兑换码输入区域                     │
└─────────────────────────────────────────────────┘
```

### 实现要点

| 层 | SaaS | Local |
|----|------|-------|
| DB schema | `redemption_codes` 表存在 | 不建表 |
| 后端路由 | `POST /billing/redeem` 注册 | 不注册（`if cfg.SupportSaas` 或直接放 billing handler 不 gate，因为 Local 无表即 500 兜底，但推荐显式不注册） |
| 后端平台路由 | `/api/platform/redemption-codes/*` | 不注册（已有 platform 路由隔离） |
| 前端充值页 | 兑换码区域正常显示 | `SUPPORT_SAAS` flag 控制隐藏兑换码区域 |
| Lot 同步 | 兑换产生的 order/lot 写入后 bump `wallet_lots` version | sync worker 拉取并 upsert，钱包余额自动生效 |

### 用户体验（Local 企业管理员视角）

1. 从 SaaS 平台方或 BD 拿到兑换码
2. 登录 SaaS 控制台（不是 Local）
3. 在 SaaS 的 /billing 页面兑换
4. 等待 sync（通常秒级），Local 钱包余额更新

> 这与现有的「平台充值」流程一致——平台管理员在 SaaS 给 Local 企业充值，lot 通过 sync 下发。兑换码只是把"平台手动操作"变成"用户自助输入码"。

---

## 12. 实现优先级

### Phase 1（MVP — 3 个 API + UX 优化）

- [x] 数据模型（redemption_codes 单表，SaaS schema only）
- [x] 平台：批量生成码（POST generate）
- [x] 平台：列出码（GET list，含筛选分页）
- [x] 企业：兑换码充值后端接口 + 并发安全（行级锁）+ 条件注册
- [x] 企业：前端 UX 优化（P0: loading/防重复 + 输入标准化；P1: 面值反馈 + 错误细分）
- [ ] 接口限流（5次/分钟/用户）
- [x] 前端：Local 模式隐藏兑换码区域

### Phase 2（增强）

- [ ] 平台：单码/批量禁用（POST disable）
- [ ] 平台：按 batch_name 统计（兑换率、使用趋势）
- [ ] 企业：兑换历史独立列表
- [ ] 兑换码过期自动清理 job
- [ ] 兑换成功通知（站内信/邮件）

### Phase 3（扩展）

- [ ] 兑换码限定企业（指定企业才能兑换）
- [ ] 兑换码限定模型/用途
- [ ] 渠道追踪（每个码关联渠道标签）
- [ ] Local 前端充值页引导跳转 SaaS 兑换（deeplink）

---

## 13. 开发工作量估算

| 模块 | 预估 |
|------|------|
| 数据模型（单表 + migration） | 0.5d |
| 后端 domain + store（生成+兑换，3 个 API） | 1.5d |
| 后端 HTTP handler + 条件注册 + 限流 | 0.5d |
| 前端 平台管理页面（列表 + 生成 Dialog） | 1.5d |
| 前端 兑换区 UX 优化 + Local 条件隐藏 | 0.5d |
| 测试 | 1d |
| **合计** | **~5.5d** |

---

## 14. 附录：码生成伪代码

```go
const charset = "23456789ABCDEFGHJKMNPQRSTUVWXYZ" // 30 chars, no 0O1IL

func GenerateCode() string {
    code := make([]byte, 12)
    for i := range code {
        // Rejection sampling 消除 modulo bias (256 % 30 = 16 → 前16字符概率偏高)
        // 取 [0, 240) 范围内的随机字节，240 是 30 的最大整数倍 ≤ 256
        for {
            b := make([]byte, 1)
            _, _ = crypto_rand.Read(b)
            if b[0] < 240 {
                code[i] = charset[b[0]%30]
                break
            }
        }
    }
    // Format: TJ-XXXX-XXXX-XXXX
    return fmt.Sprintf("TJ-%s-%s-%s", string(code[0:4]), string(code[4:8]), string(code[8:12]))
}
```

> 为什么用 rejection sampling：直接 `byte % 30` 有 ~2.3% 概率偏差（256 不是 30 的倍数）。取 `[0, 240)` 范围后 `% 30` 分布均匀。实际 reject 率仅 6.25%，性能影响可忽略。

---

## 15. 附录：兑换核心逻辑伪代码

```go
func (s *service) Redeem(ctx context.Context, code string, memberID uuid.UUID) (RedeemResult, error) {
    companyID := company.CompanyID(ctx)
    
    // 标准化
    code = normalizeCode(code) // 去空格、大写、去连字符后重新格式化
    
    return s.store.WithTx(ctx, func(tx store.Store) error {
        // 1. 锁定兑换码行
        rc, err := tx.Redemption().GetCodeForUpdate(ctx, code)
        if err != nil || rc == nil {
            return domain.BadRequest("INVALID_REDEMPTION_CODE", "兑换码无效")
        }
        
        // 2. 校验状态
        if rc.Status == "disabled" {
            return domain.BadRequest("CODE_DISABLED", "该兑换码已被禁用")
        }
        if rc.Status != "unused" {
            return domain.BadRequest("CODE_ALREADY_USED", "该兑换码已被使用")
        }
        if rc.ExpiresAt.Before(time.Now()) {
            return domain.BadRequest("CODE_EXPIRED", "该兑换码已过期")
        }
        
        // 3. 创建 RechargeOrder + Lot（复用现有路径）
        order := buildRedemptionOrder(companyID, rc, memberID)
        lot := billing.BuildLot(order, rc.Currency, store.LotKindGift, 0)
        _, err = billinglot.CreditFromLot(ctx, tx, order, lot, lot.QuotaGranted)
        
        // 4. 标记码为已使用
        tx.Redemption().MarkUsed(ctx, rc.ID, companyID, memberID, order.ID)
        
        return nil
    })
}
```
