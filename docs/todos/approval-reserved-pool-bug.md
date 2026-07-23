# 审批类型整合：产品分析与修正方案

> 2026-07-22 发现 · 2026-07-23 产品分析完成

---

## 结论

**不是产品设计需要三种审批类型，是迭代遗留导致的实现冗余。**

产品实际只有两种审批需求：
1. **申请 Key**（附带额度）→ `key` 类型
2. **申请额度追加**（不创建 Key）→ 应由 `member_budget` 类型承载

`budget` 类型是早期有缺陷的实现，`member_budget` 是后来的修正版但前端未切换。

---

## 三种类型的来龙去脉

### 演进时间线

1. **第一阶段**：只有 `key` 类型 — 成员申请创建 Platform Key，附带额度分配
2. **第二阶段**：加了 `budget` 类型 — 成员只想追加额度不要 Key。放在 keys domain 复用了 keys service 的预算计算逻辑，但**漏了扣减预留池**
3. **第三阶段**：发现 `budget` 缺陷后，在 budget domain 正确实现了 `member_budget` 类型（扣预留池 + 触发 rebalance），但**前端从未切换过来**

### 代码证据

| 类型 | 所在 domain | 前端可选？ | 扣预留池？ | 触发 rebalance？ | 状态 |
|------|------------|-----------|-----------|----------------|------|
| `key` | keys | ✅ | ❌ | ✅（通过 PostApprove 同步 NewAPI） | **有缺陷**：补差额度不扣预留池 |
| `budget` | keys | ✅ | ❌ | ❌ | **有缺陷**：校验了预留池但不扣减 |
| `member_budget` | budget | ❌（前端未暴露） | ✅ | ✅ | **正确实现**但未接入 |

### 为什么 `budget` 放在 keys domain？

历史原因。`budget` 类型的 handler (`BudgetApprovalHandler`) 定义在 `internal/domain/keys/approval_handler.go`，和 `KeyApprovalHandler` 同文件。因为早期只有 keys service 有完整的预算计算逻辑（`LoadBudgetContext`、`GetReservedPoolForMember`），所以"额度追加"就顺手放这里了。这是个 domain 归属错误 —— 纯额度操作应该归 budget domain。

### `budget` 的具体缺陷

`BudgetApprovalHandler.OnApprovedTx` 做了两件事：
1. 获取锁 → 读取预留池 → **校验**预留池 >= 申请额度（有 validation 错误返回）
2. 增加 personalBudget

但关键的第三步 —— **扣减预留池** —— 缺失了。这导致：
- 预留池数字永远不变
- 管理员设 5000 预留池，可以无限次批出 5000（每次校验都通过因为从不扣减）
- 预留池对管理员没有实际参考价值

### `key` 类型的缺陷

`KeyApprovalHandler.OnApprovedTx` 在需要补差额度时（`personalBudgetAdded > 0`），只增加了 personalBudget，没有从预留池扣减补差部分。PreApprove 校验了预留池足够，但批准时不扣。

---

## 修正方案

### 目标状态：只保留两种审批类型

| 类型 | 用途 | 实现 |
|------|------|------|
| `key` | 申请创建 Key + 分配额度 | 修补：补差时扣预留池 |
| `member_budget` | 申请额度追加 | 已正确实现，前端切过来即可 |

`budget` 类型废弃删除。

### 实施步骤

#### 前端

1. `approval-submit.tsx` — "额度追加" 选项的 `value` 从 `"budget"` 改为 `"member_budget"`
2. 提交时 metadata 格式从 `{ reason, requestedBudget, requestedModels }` 改为 `{ amount: displayToQuota(requestedBudget), reason }`
3. `api/types/approval.ts` — 可选：从 `ApprovalType` 联合类型中移除 `'budget'`

#### 后端

1. `internal/domain/keys/approval_handler.go` — `KeyApprovalHandler.OnApprovedTx`：当 `personalBudgetAdded > 0` 时，扣减申请人所在部门的预留池
2. `internal/domain/keys/approval_handler.go` — 删除整个 `BudgetApprovalHandler`
3. `internal/app/compose_domain_wire.go` — 移除 `NewBudgetApprovalHandler(keysSvc)` 注册
4. `internal/domain/types/approval.go` — 移除 `ApprovalTypeBudget` 常量和 `BudgetApprovalMeta` 结构体

### 代码位置

| 文件 | 改动 |
|------|------|
| `apps/frontend/src/features/workflow/workflows/approval-submit.tsx` | type 值 + metadata 格式 |
| `apps/frontend/src/api/types/approval.ts` | 移除 `'budget'` |
| `internal/domain/keys/approval_handler.go` | Key handler 加预留池扣减；删 BudgetApprovalHandler |
| `internal/domain/budget/approval_handler.go` | 无改动（已正确） |
| `internal/app/compose_domain_wire.go` | 移除 budget handler 注册 |
| `internal/domain/types/approval.go` | 移除 budget 常量和 meta struct |

---

## 验证清单

1. 额度追加审批通过 → 预留池数值减少、personalBudget 增加、触发 rebalance
2. Key 审批通过（需补差时）→ 预留池扣减补差额、personalBudget 增加、Key 创建、同步 NewAPI
3. Key 审批通过（不需补差时）→ 预留池不变、直接从已有额度切分
4. 预留池不够 → PreApprove 拦截，审批无法通过
5. 预留池扣到 0 → 新申请被拦截
6. 管理员调高预留池后 → 新审批可正常通过
7. 旧的 `budget` 类型的 pending 请求 → 需一次性处理（手动 reject 或迁移为 member_budget）
