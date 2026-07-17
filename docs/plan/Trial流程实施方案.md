# Trial 免费试用实施方案

> **范围**：SaaS 免费试用（Trial）流程。正式注册升级、私有化 Setup、成员邀请等见 [登录注册方案设计.md](./登录注册方案设计.md)。

---

## 1. 用户流程

```
/login → [免费试用] → 手机号验证码 → 创建企业 → 进入平台 → Onboarding 导入组织架构
```

### 1.1 登录页

```
┌────────────────────────────────────────┐
│            登录 TokenJoy               │
│                                        │
│  手机号: [+86 ___________]             │
│  验证码: [______]  [获取验证码 60s]    │
│                                        │
│  [登录 / 注册]                         │
│  ──────────── 或 ────────────          │
│  [✨ 免费试用]                          │
│  其他方式：邮箱密码登录                │
└────────────────────────────────────────┘
```

### 1.2 注册流程

1. 输入手机号 + 验证码
2. 填写公司名
3. 创建 Trial 企业 → 自动登录
4. 触发 Onboarding（右侧栏：CSV / 手动 / 飞书 / 跳过）

### 1.3 CSV 模板

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

### 1.4 Trial Banner

```
┌──────────────────────────────────────────────────────────────────┐
│ 🎯 试用环境 · 使用模拟资金体验，升级后接入真实模型   [联系升级]  │
└──────────────────────────────────────────────────────────────────┘
```

纯信息提示 + 升级 CTA，不涉及到期/倒计时。

### 1.5 升级

原地升级：`type` 改为 `standard`，Banner 消失，Gateway 切 allowlist 解锁全部模型，充值解锁，消费记录保留。

升级时模拟资金处理：Trial lot 全部 expire，`wallet_remain` 按剩余 active lot 重算（无真实充值则归零）。

---

## 2. 后端 API

| 方法 | 路径 | Body | 响应 |
| --- | --- | --- | --- |
| POST | `/auth/sms/send` | `{ phone }` | `{ sent: true }` |
| POST | `/auth/sms/verify` | `{ phone, code }` | `SmsVerifyResult` |
| POST | `/auth/register` | `{ companyName, phone, token }` | `{ memberId, needsOnboarding }` |
| GET | `/auth/trial/csv-template` | — | CSV file |

`POST /auth/register` 逻辑：
1. 验证临时 token
2. 创建 Company（`type='trial'`）
3. 创建超管 Member + 预设角色 + Root Dept
4. 灌入模拟资金 — 创建 `lot_kind='trial'` 的 Lot（见 §3.1）
5. 签发 JWT → Set-Cookie

错误：`409` 手机号已注册；`403` 注册未开放。

---

## 3. 模拟资金

### 3.1 设计思路

| 约束 | 实现方式 |
| --- | --- |
| 模拟资金只能用于 mock 模型 | Gateway allowlist 仅含 mock model → 非 mock 请求 403，根本不进 ingest |
| 升级后模拟资金不可用 | 升级时 expire 所有 `lot_kind='trial'` 的 lot，`wallet_remain` 按实际 active lot 重算 |
| Trial 期间看板可见消费 | Trial lot 正常走 FIFO 消费 → ledger 有条目 → 看板展示 |

### 3.2 数据模型

新增 lot kind：

```go
LotKindTrial = "trial"  // 模拟资金，仅 trial 期间可用
```

注册时灌入：

```go
// 在 POST /auth/register 事务内
order := store.RechargeOrder{
    ID: fmt.Sprintf("trial-%d-%d", companyID, now.UnixNano()),
    CompanyID: companyID, Amount: 0, Currency: currency,
    PointsPerUnit: ppu, PointsGranted: cfg.TrialCreditAmount,
    Source: "system", LotKind: store.LotKindTrial,
    Status: "confirmed", CreatedBy: "system",
}
lot := BuildTrialLot(order, currency) // 类似 BuildGiftLot，UnitPriceDisplay=0
billinglot.CreditFromLot(ctx, st, order, lot, cfg.TrialCreditAmount)
```

### 3.3 升级清零

升级接口（`trial` → `standard`）执行：

```sql
-- 1. 冻结所有 trial lot
UPDATE recharge_lots
   SET status = 'expired', updated_at = now()
 WHERE company_id = $1 AND lot_kind = 'trial' AND status = 'active';

-- 2. wallet_remain 按剩余 active lot 重算（非 trial lot 余额之和）
UPDATE companies
   SET wallet_remain = (
       SELECT COALESCE(SUM(points_remaining), 0)
         FROM recharge_lots
        WHERE company_id = $1 AND status = 'active'
   ),
   type = 'standard'
 WHERE id = $1;
```

> 无真实充值时 `wallet_remain` 归零；若升级前已有 paid/gift lot，保留其余额。

### 3.4 为什么不用双账本

| 方案 | 优点 | 缺点 |
| --- | --- | --- |
| Sandbox 双账本 | 语义最隔离 | 双路径查询，改动大，Trial 数据升级后要迁移或丢弃 |
| **标记型 Trial Lot**（采用） | 零侵入计费核心、升级只需 expire + 重算 | 无 |

核心逻辑：**Gateway allowlist 做模型隔离（消费侧守卫），lot_kind 做资金标记（升级清零依据）**，不给 ingest/lot-consume 加任何 if-else。

---

## 4. Gateway

- Trial 公司的 Platform Key 白名单仅包含 mock 模型（`trial-mock-model`）
- Gateway precheck 正常执行 allowlist 检查 → Trial Key 调用真实模型自然被 "model not allowed" 拦截
- mock 模型请求正常通过 → proxy 到 NewAPI → NewAPI 路由到 mock channel
- Mock LLM 返回模拟 response → 正常走 ingest 写 ledger
- Trial lot 正常 FIFO 扣费，看板可见消费数据
- 升级为 standard 后：扩展 Key allowlist 解锁全部模型，trial lot 已 expired 不参与消费

> **无需特殊 Gateway 路由或第二个 proxy**——复用现有 allowlist 机制和 NewAPI 内的 mock channel。
> 
> **Mock 模型使用产线级 model_id**（≥100），避免触碰 `IsLocalOnlyCallType` 的 dev-only 拦截逻辑。

---

## 5. 通知

Trial 租户：全 channel 通知正常（Trial 注册时已绑定手机号，具备 SMS/email 投递能力）。

---

## 6. 功能限制

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

**前端提示**：Trial 用户尝试调用非 mock 模型时，Gateway 返回 "model not allowed"，前端展示提示"试用模式仅支持模拟模型，升级后可接入真实模型"。

---

## 7. 数据库

```sql
-- companies 表变更
type TEXT NOT NULL DEFAULT 'selfhosted'
     CHECK (type IN ('standard', 'trial', 'demo', 'selfhosted', 'testing'))

onboarding_status TEXT NOT NULL DEFAULT 'pending'
     CHECK (onboarding_status IN ('pending', 'completed', 'skipped'))
```

不需要 `trial_expires_at` 列（Trial 无到期机制）。

---

## 8. Session

```typescript
interface AppSession {
  companyType: 'standard' | 'trial' | 'demo' | 'selfhosted' | 'testing'
  onboardingStatus: 'pending' | 'completed' | 'skipped'
}
```

---

## 9. 配置

| 环境变量 | 默认值 | 说明 |
| --- | --- | --- |
| `REGISTRATION_ENABLED` | `true` | 允许公开注册 |
| `TRIAL_MEMBER_LIMIT` | `50` | 成员上限 |
| `TRIAL_CREDIT_AMOUNT` | `10000` | 模拟资金（points） |
| `VITE_REGISTRATION_ENABLED` | — | 前端控制试用按钮显示 |

---

## 10. 前端文件结构

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

## 11. type 流转

| 流转 | 允许 |
| --- | --- |
| `trial` → `standard` | ✅ 付费升级，原地保留 |
| 其他 → 任何 | ❌ |

---

## 12. 实施清单

### 后端

- [ ] schema: `onboarding_status` 列
- [ ] store: `LotKindTrial = "trial"` 常量
- [ ] `POST /auth/register`（创建 trial company + 灌入 trial lot）
- [ ] `BuildTrialLot` 构建函数（类似 `BuildGiftLot`，`UnitPriceDisplay=0`）
- [ ] `GET /auth/trial/csv-template`
- [ ] CSV 解析（部门层级 + 成员 + 角色映射）
- [ ] Gateway: 产线级 `trial-mock-model` 定义（model_id ≥ 100）
- [ ] NewAPI: mock channel 配置（路由到 echo LLM 服务）
- [ ] 升级接口: expire trial lots + wallet_remain 重算
- [ ] Session 增加 `companyType`、`onboardingStatus`
- [ ] `REGISTRATION_ENABLED` 守卫
- [ ] 成员上限守卫
- [ ] 充值接口 Trial 拦截（禁止真实充值）

### 前端

- [ ] `/register` 页面
- [ ] `features/trial/`（Banner）
- [ ] `features/onboarding/`（滑出栏 + CSV）
- [ ] `/login` 改手机号验证码 + 试用按钮
- [ ] `AppSession` 扩展
- [ ] Trial 功能限制 UI（充值禁用提示）

### 其他

- [ ] Echo LLM 服务（OpenAI-compatible mock server）生产部署
- [ ] 升级流程（Trial → Standard）后续文档
