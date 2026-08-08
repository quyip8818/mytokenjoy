# CompanyType 收口设计

## 最终方案：`standard` → `saas` + `IsTestingAccount` helper

### 类型表

| 类型 | IsTestingAccount | 语义 |
|------|:---:|------|
| `saas` | false | SaaS 正式付费客户（原 standard，trial 升级后） |
| `selfhosted` | false | 私有化部署客户 |
| `trial` | true | SaaS 试用 |
| `demo` | true | SaaS 演示 |
| `testing` | true | 纯测试 |
| `platform` | false | TokenJoy 内部 |

### 新增 helper（后端 + 前端各一个）

```go
// domain/company/testing.go
package company

import "github.com/tokenjoy/backend/internal/store"

// IsTestingAccount returns true for non-production account types that allow
// test-model access, simulated consumption, and mock quota grants.
func IsTestingAccount(companyType string) bool {
    switch companyType {
    case store.CompanyTypeTrial, store.CompanyTypeDemo, store.CompanyTypeTesting:
        return true
    default:
        return false
    }
}
```

```typescript
// frontend: lib/company.ts
import type { CompanyType } from '@/api/types/common'

export function isTestingAccount(type: CompanyType): boolean {
  return type === 'trial' || type === 'demo' || type === 'testing'
}
```

---

## 逐处改动 + 风险评估

### 1. `gateway/gateway_service.go` — `isTestModelAllowed`

**现在**：`switch demo|trial|testing → true`
**改为**：`return company.IsTestingAccount(companyType)`
**风险**：无。逻辑完全等价。

### 2. `handler/keys/handler.go` — `isSimulateAllowed`

**现在**：`switch demo|trial|testing → true`
**改为**：`return company.IsTestingAccount(companyType)`
**风险**：无。逻辑完全等价。

### 3. `handler/billing/handler.go` — `isTrialOrDemoCompany`

**现在**：`trial|demo` → 禁止充值/兑换
**改为**：`company.IsTestingAccount(type)` → testing 类型也被禁止充值
**风险**：⚠️ 行为变化——`testing` 类型原来可以充值，改后不能。
**评估**：testing 类型是纯测试环境，不需要真实充值，禁止是合理的。如果 testing 环境需要额度，走 mock 赠送即可。**可接受。**

### 4. `domain/company/service_create.go` — 创建时赠送 mock 额度

**现在**：`trial|demo` → 赠送
**改为**：`company.IsTestingAccount(type)` → testing 类型创建时也赠送
**风险**：⚠️ 行为变化——`testing` 类型创建时会额外赠送 mock 额度。
**评估**：testing 本来就是测试环境，多给点 mock 额度无害。**可接受。**

### 5. `domain/company/service.go` — Upgrade（升级为 saas）

**现在**：`trial|demo` → 允许升级
**改为**：`company.IsTestingAccount(type)` → testing 也可以升级
**风险**：⚠️ testing 类型原来不能升级，改后可以。
**评估**：testing 是开发环境，让它能升级便于测试升级流程。**可接受。** 但如果不想改行为，这里可以保持原样不用 helper。
**建议**：保持原逻辑 `trial|demo` 不改，升级是业务决策而非 testing 属性。

### 6. `domain/company/iterate.go` — 跳过批处理

**现在**：`== testing` → 跳过
**改为**：**不改**。这个语义是"跳过纯测试公司的定时任务"，和 IsTestingAccount 语义不同（trial/demo 需要参与批处理）。
**风险**：无改动。

### 7. `domain/org/member_helpers.go` — 试用成员上限

**现在**：`== trial` → 限制
**改为**：**不改**。成员限制只针对 trial，demo 和 testing 不需要限制。
**风险**：无改动。

### 8. `gateway/precheck.go` — SkipWallet

**现在**：`== selfhosted` → 跳过钱包检查
**改为**：PrecheckService 构造注入 `!cfg.SupportSaas`
**风险**：⚠️ SaaS 模式下的 selfhosted 公司（如果存在）会被检查钱包。
**评估**：SaaS 模式下不存在 selfhosted 公司。selfhosted 只在私有化部署创建。**无风险。**

### 9. 前端 `model-list-page` — 显示自定义模型 tab

**现在**：`companyType === 'selfhosted'`
**改为**：`!IS_SAAS`
**风险**：无。私有化部署 = !IS_SAAS，语义一致。

### 10. 前端 `header-dev-backend-chrome` — 模拟消费工具栏

**现在**：`['demo', 'trial', 'testing'].includes(companyType)`
**改为**：`isTestingAccount(companyType)`
**风险**：无。逻辑等价。

### 11. `store.CompanyTypeStandard` → `store.CompanyTypeSaas`

**改动**：常量重命名 + DB 值从 `"standard"` 改为 `"saas"`
**风险**：⚠️ 需要确认没有外部系统读取 companies.type 字段依赖 "standard" 值。
**评估**：项目未上线，无历史数据，schema 可重建。**无风险。**

---

## 不适合用 IsTestingAccount 的判断

| 位置 | 原因 |
|------|------|
| Upgrade（`trial\|demo`） | 升级是业务决策，testing 不需要升级 |
| iterate 跳过（`== testing`） | trial/demo 需要参与批处理 |
| 成员上限（`== trial`） | 只有 trial 有限制，demo/testing 不限 |
| authz 平台权限（`== platform`） | 独立语义 |
| precheck SkipWallet | 部署模式问题，不是 testing 属性 |

---

## 总结

| 可直接替换为 `IsTestingAccount` | 保持不变 |
|---|---|
| isTestModelAllowed | Upgrade（trial\|demo） |
| isSimulateAllowed | iterate 跳过（testing only） |
| isTrialOrDemoCompany（充值禁止） | 成员上限（trial only） |
| service_create 赠送额度 | authz 平台权限（platform only） |
| 前端 simulate toolbar | precheck SkipWallet（部署模式） |

## 执行步骤

1. 后端：`store.CompanyTypeStandard` 重命名为 `store.CompanyTypeSaas`，值改为 `"saas"`
2. 后端：新增 `domain/company/testing.go`，实现 `IsTestingAccount`
3. 后端：替换 isTestModelAllowed / isSimulateAllowed / isTrialOrDemoCompany / 赠送额度判断
4. 后端：precheck SkipWallet 改为构造注入 `!cfg.SupportSaas`
5. 前端：CompanyType union 里 `standard` → `saas`
6. 前端：新增 `lib/company.ts`，实现 `isTestingAccount`
7. 前端：model-list-page 改 `!IS_SAAS`，simulate toolbar 改 `isTestingAccount`
8. 更新 seed/bootstrap 中 Upgrade 目标类型为 `saas`
9. schema 重建（项目未上线）
