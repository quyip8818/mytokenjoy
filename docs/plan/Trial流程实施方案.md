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
剩余 > 7 天：蓝色  "试用期剩余 X 天 · [联系我们升级]"
剩余 ≤ 7 天：橙色警告
已到期：     红色  "试用已到期 · [升级正式版]"
```

### 1.5 到期行为

1. `status` → `suspended`
2. 可登录，只读访问
3. Gateway API 调用拒绝
4. 数据保留 90 天后清理

### 1.6 升级

原地升级：`type` 改为 `standard`，移除 `trial_expires_at`，数据全部保留。

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
2. 创建 Company（`type='trial'`, `trial_expires_at = NOW() + TRIAL_DURATION_DAYS`）
4. 创建超管 Member + 预设角色 + Root Dept
5. 签发 JWT → Set-Cookie

错误：`409` 手机号已注册 / slug 已占用；`403` 注册未开放。

---

## 3. Gateway

- precheck 识别 `type='trial'` → 标记 `isTrialCompany`
- Director rewrite 到 `TRIAL_MOCK_LLM_URL`
- Mock LLM 返回模拟 response → 正常走 ingest 写 ledger
- 初始钱包余额 10000 points
- 升级为 standard 后切换到正式 NewAPI

---

## 4. 通知

Trial 租户：全 channel 通知正常（Trial 注册时已绑定手机号，具备 SMS/email 投递能力）。

---

## 5. 定时任务

**冻结**（每日）：

```go
// trial 到期 + status=active → status=suspended
UPDATE companies SET status = 'suspended'
WHERE type = 'trial' AND status = 'active' AND trial_expires_at < NOW()
```

**清理**（每日）：

```go
// suspended + 到期超 90 天 → CASCADE DELETE，每批 10 个
SELECT id FROM companies
WHERE type = 'trial' AND status = 'suspended'
  AND trial_expires_at < NOW() - INTERVAL '90 days'
LIMIT 10
```

---

## 6. 功能限制

| 功能 | Trial 行为 |
| --- | --- |
| Gateway | Mock LLM，正常扣费记录 |
| 预算 / 模型 / Key / 组织 / 审计 / 看板 | 全功能 |
| 充值 | 禁用，提示"试用期间不支持充值" |
| 邀请成员 | 正常投递 |
| 数据源（飞书/钉钉/企微） | 全功能 |
| 通知 | 全 channel（站内 + 短信 + 邮件） |
| 成员上限 | 50 人 |

---

## 7. 数据库

```sql
-- companies 表变更
type TEXT NOT NULL DEFAULT 'selfhosted'
     CHECK (type IN ('standard', 'trial', 'demo', 'selfhosted', 'testing'))

trial_expires_at TIMESTAMPTZ           -- Trial 到期时间
onboarding_status TEXT NOT NULL DEFAULT 'pending'
     CHECK (onboarding_status IN ('pending', 'completed', 'skipped'))
```

---

## 8. Session

```typescript
interface AppSession {
  companyType: 'standard' | 'trial' | 'demo' | 'selfhosted' | 'testing'
  trialExpiresAt: string | null
  onboardingStatus: 'pending' | 'completed' | 'skipped'
}
```

---

## 9. 配置

| 环境变量 | 默认值 | 说明 |
| --- | --- | --- |
| `REGISTRATION_ENABLED` | `true` | 允许公开注册 |
| `TRIAL_DURATION_DAYS` | `30` | 试用期天数 |
| `TRIAL_MEMBER_LIMIT` | `50` | 成员上限 |
| `TRIAL_CREDIT_AMOUNT` | `10000` | 赠送积分 |
| `TRIAL_MOCK_LLM_URL` | `http://127.0.0.1:8765` | Mock LLM 地址 |
| `TRIAL_DATA_RETENTION_DAYS` | `90` | 到期后数据保留天数 |
| `VITE_REGISTRATION_ENABLED` | — | 前端控制试用按钮显示 |

---

## 10. 前端文件结构

```
features/trial/
├── index.ts
├── hooks/use-trial-status.ts
├── components/
│   ├── trial-banner.tsx
│   └── trial-expired-overlay.tsx
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

- [ ] schema: `trial_expires_at`、`onboarding_status` 列
- [ ] `POST /auth/register`（创建 trial company）
- [ ] `GET /auth/trial/csv-template`
- [ ] CSV 解析（部门层级 + 成员 + 角色映射）
- [ ] Gateway Trial guard（Mock LLM rewrite）
- [ ] Session 增加 `companyType`、`trialExpiresAt`、`onboardingStatus`
- [ ] 冻结 job + 清理 job
- [ ] `REGISTRATION_ENABLED` 守卫
- [ ] 成员上限守卫

### 前端

- [ ] `/register` 页面
- [ ] `features/trial/`（Banner + 到期覆盖层）
- [ ] `features/onboarding/`（滑出栏 + CSV）
- [ ] `/login` 改手机号验证码 + 试用按钮
- [ ] `AppSession` 扩展
- [ ] Trial 功能限制 UI
- [ ] Slug 实时检查

### 其他

- [ ] `dev-mock-llm` 生产部署确认
- [ ] 升级流程（Trial → Standard）后续文档
