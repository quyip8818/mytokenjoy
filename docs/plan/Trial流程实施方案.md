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

原地升级：`type` 改为 `standard`，Banner 消失，Gateway 切换到正式 NewAPI，充值解锁，数据全部保留。

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
4. 灌入模拟资金（`TRIAL_CREDIT_AMOUNT`）
5. 签发 JWT → Set-Cookie

错误：`409` 手机号已注册；`403` 注册未开放。

---

## 3. Gateway

- precheck 识别 `type='trial'` → 标记 `isTrialCompany`
- Director **强制** rewrite 到 `TRIAL_MOCK_LLM_URL`，忽略请求中指定的模型
- 后端拒绝 Trial 租户调用非 Mock LLM 模型（返回 `403 { message: "试用环境仅支持模拟模型" }`）
- Mock LLM 返回模拟 response → 正常走 ingest 写 ledger
- 模拟资金正常扣费，看板可见消费数据
- 升级为 standard 后切换到正式 NewAPI，解锁全部模型

---

## 4. 通知

Trial 租户：全 channel 通知正常（Trial 注册时已绑定手机号，具备 SMS/email 投递能力）。

---

## 5. 功能限制

| 功能 | Trial 行为 |
| --- | --- |
| Gateway | **仅 Mock LLM**，后端拒绝真实模型调用（403） |
| 模型管理 | 可查看全部模型，但调用时只允许 Mock 模型 |
| 预算 / Key / 组织 / 审计 / 看板 | 全功能 |
| 充值 / 钱包 | 禁用真实充值，提示"试用环境使用模拟资金" |
| 邀请成员 | 正常投递 |
| 数据源（飞书/钉钉/企微） | 全功能 |
| 通知 | 全 channel |
| 成员上限 | 50 人（`TRIAL_MEMBER_LIMIT`） |

**前端提示**：Trial 用户尝试调用非 Mock 模型时，弹出提示"试用模式仅支持模拟模型，升级后可接入真实模型"。

---

## 6. 数据库

```sql
-- companies 表变更
type TEXT NOT NULL DEFAULT 'selfhosted'
     CHECK (type IN ('standard', 'trial', 'demo', 'selfhosted', 'testing'))

onboarding_status TEXT NOT NULL DEFAULT 'pending'
     CHECK (onboarding_status IN ('pending', 'completed', 'skipped'))
```

不需要 `trial_expires_at` 列（Trial 无到期机制）。

---

## 7. Session

```typescript
interface AppSession {
  companyType: 'standard' | 'trial' | 'demo' | 'selfhosted' | 'testing'
  onboardingStatus: 'pending' | 'completed' | 'skipped'
}
```

---

## 8. 配置

| 环境变量 | 默认值 | 说明 |
| --- | --- | --- |
| `REGISTRATION_ENABLED` | `true` | 允许公开注册 |
| `TRIAL_MEMBER_LIMIT` | `50` | 成员上限 |
| `TRIAL_CREDIT_AMOUNT` | `10000` | 模拟资金（points） |
| `TRIAL_MOCK_LLM_URL` | `http://127.0.0.1:8765` | Mock LLM 地址 |
| `VITE_REGISTRATION_ENABLED` | — | 前端控制试用按钮显示 |

---

## 9. 前端文件结构

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

## 10. type 流转

| 流转 | 允许 |
| --- | --- |
| `trial` → `standard` | ✅ 付费升级，原地保留 |
| 其他 → 任何 | ❌ |

---

## 11. 实施清单

### 后端

- [ ] schema: `onboarding_status` 列
- [ ] `POST /auth/register`（创建 trial company + 灌模拟资金）
- [ ] `GET /auth/trial/csv-template`
- [ ] CSV 解析（部门层级 + 成员 + 角色映射）
- [ ] Gateway Trial guard（Mock LLM rewrite）
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

- [ ] `dev-mock-llm` 生产部署确认
- [ ] 升级流程（Trial → Standard）后续文档
