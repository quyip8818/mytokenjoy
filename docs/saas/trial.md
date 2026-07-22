# Trial 免费试用

> 本文定义 SaaS 免费试用（Trial）流程：注册、模拟资金、Gateway mock、功能限制、升级。  
> 数据模型（Company type/status）见 [identity-model.md](./identity-model.md)。  
> 认证流程（注册 API）见 [auth-flow.md](./auth-flow.md)。  
> 开户 Domain Service 见 [invite-and-onboarding.md](./invite-and-onboarding.md)。

---

## 1. 用户流程

```
/login → [进入] → 手机号验证码 → 创建企业 → 进入平台 → Onboarding 导入组织架构
```

> **核心语义**：用户创建的是一家**真实企业**，组织架构、Key、预算等数据永久保留。Trial 阶段使用模拟资金 + mock 模型体验全部功能。升级为正式版后数据原地保留，仅切换资金来源和模型 allowlist。不存在独立的"免费试用"入口——创建公司本身就是试用的开始。

### 1.1 注册流程

1. 输入手机号 + 验证码（`/auth/verify-code/verify` → action: onboard）
2. 填写公司名（`/auth/register/company`）
3. 创建 Trial 企业 → 自动登录
4. 触发 Onboarding（右侧栏：CSV / 手动 / 飞书 / 跳过）

### 1.2 CSV 模板

文件名：`tokenjoy-组织架构模板.csv`

```csv
姓名,邮箱,手机号,部门,角色
张三,zhang@example.com,13800001111,技术部,超级管理员
李四,li@example.com,13800002222,技术部/后端组,普通成员
王五,wang@example.com,13800003333,产品部,普通成员
```

- 部门用 `/` 表示层级（自动创建）
- 角色：超级管理员 / 组织管理员 / 普通成员（默认）
- 邮箱和手机号选填

### 1.3 Trial Banner

```
┌──────────────────────────────────────────────────────────────────┐
│ 🎯 试用环境 · 使用模拟资金体验，升级后接入真实模型   [联系升级]  │
└──────────────────────────────────────────────────────────────────┘
```

纯信息提示 + 升级 CTA，不涉及到期/倒计时。

---

## 2. 模拟资金

### 2.1 设计思路

| 约束 | 实现方式 |
| --- | --- |
| 模拟资金只能用于 mock 模型 | Gateway allowlist 仅含 mock model → 非 mock 请求 403，根本不进 ingest |
| 升级后模拟资金不可用 | 升级时 expire 所有 `lot_kind='mock'` 的 lot，`wallet_remain` 按实际 active lot 重算 |
| Trial 期间看板可见消费 | Mock lot 正常走 FIFO 消费 → ledger 有条目 → 看板展示 |

核心逻辑：**Gateway allowlist 做模型隔离（消费侧守卫），lot_kind='mock' 做资金标记（升级清零依据）**，不给 ingest/lot-consume 加任何 if-else。

### 2.2 数据模型

新增 lot kind：

```go
LotKindMock = "mock"  // 模拟资金，Trial 期间使用
```

注册时灌入：

```go
// 在 POST /auth/register/company 事务内
order := store.RechargeOrder{
    ID: fmt.Sprintf("trial-%d-%d", companyID, now.UnixNano()),
    CompanyID: companyID, Amount: 0, Currency: currency,
    QuotaPerUnit: ppu, QuotaGranted: trialQuota,
    Source: "system", LotKind: store.LotKindMock,
    Status: "confirmed", CreatedBy: "system",
}
lot := BuildMockLot(order, currency)
billinglot.CreditFromLot(ctx, st, order, lot, trialQuota)
```

### 2.3 升级清零

升级接口（`trial` → `standard`）执行：

```sql
-- 1. 冻结所有 mock lot
UPDATE recharge_lots
   SET status = 'expired', updated_at = now()
 WHERE company_id = $1 AND lot_kind = 'mock' AND status = 'active';

-- 2. wallet_remain 按剩余 active lot 重算（非 mock lot 余额之和）
UPDATE companies
   SET wallet_remain = (
       SELECT COALESCE(SUM(quota_remaining), 0)
         FROM recharge_lots
        WHERE company_id = $1 AND status = 'active'
   ),
   type = 'standard'
 WHERE id = $1;
```

无真实充值时 `wallet_remain` 归零；若升级前已有 paid/gift lot，保留其余额。

### 2.4 为什么不用双账本

| 方案 | 优点 | 缺点 |
| --- | --- | --- |
| Sandbox 双账本 | 语义最隔离 | 双路径查询，改动大，Trial 数据升级后要迁移或丢弃 |
| **标记型 Mock Lot**（采用） | 零侵入计费核心、升级只需 expire + 重算 | 无 |

---

## 3. Gateway

- Trial 公司的 Platform Key 白名单仅包含 mock 模型（`trial-mock-model`）
- Gateway precheck 正常执行 allowlist 检查 → Trial Key 调用真实模型自然被 "model not allowed" 拦截
- mock 模型请求正常通过 → proxy 到 NewAPI → NewAPI 路由到 mock channel
- Mock LLM 返回模拟 response → 正常走 ingest 写 ledger
- Mock lot 正常 FIFO 扣费，看板可见消费数据
- 升级为 standard 后：扩展 Key allowlist 解锁全部模型，mock lot 已 expired 不参与消费

> **无需特殊 Gateway 路由或第二个 proxy**——复用现有 allowlist 机制和 NewAPI 内的 mock channel。
>
> **Mock 模型使用产线级 model_id**（≥100），避免触碰 `IsLocalOnlyCallType` 的 dev-only 拦截逻辑。

---

## 4. 功能限制

| 功能 | Trial 行为 |
| --- | --- |
| Gateway | 仅 mock 模型（allowlist 控制），真实模型被 "model not allowed" 拦截 |
| 模型管理 | 可查看全部模型，但 Key allowlist 只含 `trial-mock-model` |
| 预算 / Key / 组织 / 审计 / 看板 | 全功能 |
| 充值 / 钱包 | 禁用真实充值，提示"试用环境使用模拟资金" |
| 邀请成员 | 正常投递 |
| 数据源（飞书/钉钉/企微） | 全功能 |
| 通知 | 全 channel |
| 成员上限 | 50 人（`TRIAL_MEMBER_LIMIT`） |

**前端提示**：Trial 用户尝试调用非 mock 模型时，Gateway 返回 "model not allowed"，前端展示"试用模式仅支持模拟模型，升级后可接入真实模型"。

---

## 5. 升级（Trial → Standard）

### 5.1 触发

原地升级：`type` 改为 `standard`，Banner 消失，Gateway 切 allowlist 解锁全部模型，充值解锁，消费记录保留。

### 5.2 充值即升级

Trial 公司的充值流程变为**升级确认**：

```
用户点击充值（或系统自动弹出充值引导）
    │
    ▼
┌────────────────────────────────────────┐
│  升级为正式版                           │
│                                        │
│  充值后您的企业将升级为正式版：         │
│  · 接入真实 AI 模型                    │
│  · 模拟资金清零，使用真实余额           │
│  · 所有数据保留                         │
│                                        │
│  充值金额: [________] 元               │
│                                        │
│  [取消]         [确认充值并升级]        │
└────────────────────────────────────────┘
```

确认后：
1. 创建充值订单（paid lot）
2. 执行 mock lot expire + wallet_remain 重算
3. `companies.type` 改为 `standard`
4. Banner 消失，Key allowlist 解锁

---

## 6. 通知

Trial 租户：全 channel 通知正常（Trial 注册时已绑定手机号，具备 SMS/email 投递能力）。

---

## 7. 配置

| 环境变量 | 默认值 | 说明 |
| --- | --- | --- |
| `REGISTRATION_ENABLED` | `true` | 允许公开注册 |
| `TRIAL_MEMBER_LIMIT` | `50` | 成员上限 |
| `TRIAL_CREDIT_AMOUNT` | `10000` | 模拟资金（quota） |
| `VITE_REGISTRATION_ENABLED` | — | 前端控制试用按钮显示 |

---

## 8. 前端文件结构

```
features/trial/
├── index.ts
├── hooks/use-trial-status.ts
├── components/
│   └── trial-banner.tsx
└── lib/trial-utils.ts

features/onboarding/
├── index.ts
├── hooks/use-onboarding-page.ts
├── components/
│   ├── onboarding-panel.tsx
│   ├── csv-upload.tsx
│   └── import-method-picker.tsx
└── lib/csv-parser.ts

api/trial.ts
routes/register.tsx
```

---

## 9. 实施清单

### 后端

- [ ] store: `LotKindMock = "mock"` 已有
- [ ] `POST /auth/register/company` 内创建 trial company + 灌入 mock lot
- [ ] `BuildMockLot` 构建函数（类似 `BuildGiftLot`，`UnitPriceDisplay=0`）
- [ ] `GET /auth/trial/csv-template`
- [ ] CSV 解析（部门层级 + 成员 + 角色映射）
- [ ] Gateway: 产线级 `trial-mock-model` 定义（model_id ≥ 100）
- [ ] NewAPI: mock channel 配置（路由到 echo LLM 服务）
- [ ] 升级接口: expire mock lots + wallet_remain 重算
- [ ] `REGISTRATION_ENABLED` 守卫
- [ ] 成员上限守卫
- [ ] 充值接口 Trial 拦截（禁止真实充值）

### 前端

- [ ] `features/trial/`（Banner）
- [ ] `features/onboarding/`（滑出栏 + CSV）
- [ ] `/login` 改手机号验证码 + 试用按钮
- [ ] Trial 功能限制 UI（充值禁用提示）
- [ ] 升级确认弹窗

### 其他

- [ ] Echo LLM 服务（OpenAI-compatible mock server）生产部署
- [ ] 升级流程（Trial → Standard）后续文档
