# 预算管理模块审计修复执行方案

> 生成日期: 2026-07-11
> 问题总计: 58 个 (Critical 1 / High 8 / Medium 25 / Low 24)

---

## 执行原则

1. **按影响分阶段**：P0 安全/数据一致性 → P1 正确性/可靠性 → P2 性能/可维护性 → P3 质量提升
2. **每个修复独立可验证**：改完跑测试，不依赖后续步骤
3. **遵循现有模式**：后端用 `WithTx` + `AcquireBudgetLock`；前端用 `try/catch + setError` 或 `toast + ApiError`
4. **补测试与修复同步**：修一个问题，补对应的单元/集成测试

---

## P0 — 安全与数据一致性（立即修复）

### P0-1. ResolveApproval TOCTOU 竞态修复
- **问题**: 预留池余额检查在事务外，扣减在事务内，并发审批可双重扣减
- **文件**: `apps/backend/internal/domain/budget/approvals.go`
- **修复方案**:
  1. 将余额读取和验证逻辑移入 `WithTx` 回调内部
  2. 在 `AcquireBudgetLock` 之后重新读取 tree 和 approvals
  3. 验证审批仍为 pending 状态且预留池余额充足
  4. 事务内完成扣减和状态更新
- **测试**: 新增并发测试 `TestResolveApprovalConcurrentRace`，用 goroutine 并发解决同一笔审批

### P0-2. UpdateMemberBudget 事务保护
- **问题**: 读取-验证-写入全程无事务，丢失更新风险
- **文件**: `apps/backend/internal/domain/budget/tree.go:87-111`
- **修复方案**:
  1. 用 `s.store.WithTx(ctx, func(txStore store.Store) error { ... })` 包裹整个函数体
  2. 在事务内调用 `txStore.Budget().AcquireBudgetLock(ctx, companyID)`
  3. 事务内重新读取 tree、members、platform keys
  4. 验证 + 写入在同一事务内完成
- **测试**: 新增 `TestUpdateMemberBudgetConcurrentSafety`

### P0-3. 负值输入验证（全端点）
- **问题**: 无任何端点校验 budget >= 0，可提交负预算绕过限额
- **文件**:
  - `apps/backend/internal/http/handler/budget/handler.go`
  - `apps/backend/internal/domain/budget/groups.go`
  - `apps/backend/internal/domain/budget/tree.go`
- **修复方案**:
  1. 在 `handler.go` 的 `UpdateNode` 和 `UpdateMemberBudget` 解码后加验证:
     ```go
     if req.Budget < 0 { return domain.Validation("budget must be non-negative") }
     ```
  2. 在 `CreateGroup` 和 `UpdateGroup` domain 层加 `group.Budget < 0` 校验
  3. 对 `personalBudget` 同样校验非负
- **测试**: 每个端点补负值测试用例

### P0-4. UpdateGroup 预算值验证
- **问题**: 直接赋值无任何约束，可设为负值或低于已消耗
- **文件**: `apps/backend/internal/domain/budget/groups.go:47-92`
- **修复方案**:
  1. 校验 `patch.Budget >= 0`
  2. 校验 `patch.Budget` 不低于该组已分配 keys 的总预算
  3. 校验组名非空且长度 <= 100
- **测试**: `TestUpdateGroupRejectsNegativeBudget`, `TestUpdateGroupRejectsBelowAllocated`

### P0-5. evaluateOverrun 事务保护
- **问题**: 多次读取后禁用 keys，无事务，读取与操作间状态可能已变
- **文件**: `apps/backend/internal/domain/budget/overrun.go:51-113`
- **修复方案**:
  1. 用 `WithTx` 包裹整个 `evaluateOverrun` 方法
  2. 在事务内获取 advisory lock
  3. 读取所有 consumed/budget 值
  4. 在同一事务内禁用 keys
  5. 如果通知发送失败，记录 slog.Error（不阻塞事务）
- **测试**: `TestEvaluateOverrunAtomicity`

---

## P1 — 正确性与可靠性（本迭代内）

### P1-1. 前端 overrunPolicy 硬编码修复
- **问题**: `groupToProjectView` 始终设 `overrunPolicy: 'hard_reject'`
- **文件**: `apps/frontend/src/features/budget/lib/mappers.ts:92-108`
- **修复方案**:
  1. 确认后端 `BudgetGroup` API 响应是否已包含 `overrunPolicy` 字段
  2. 若无：后端 `Groups` 查询增加 `overrun_policy` join
  3. 前端 `BudgetGroup` 类型增加 `overrunPolicy?: string`
  4. mapper 使用 `group.overrunPolicy ?? 'hard_reject'` 作为 fallback
- **测试**: 前端 mapper 单元测试覆盖有/无 policy 的情况

### P1-2. 前端变更操作统一错误处理
- **问题**: `use-budget-page.ts` 中 mutation callbacks 无 try/catch
- **文件**: `apps/frontend/src/features/budget/hooks/use-budget-page.ts:114-157`
- **修复方案**:
  1. 为所有 async callback 添加 try/catch
  2. catch 块使用 `toast.error(err instanceof ApiError ? err.message : '操作失败，请重试')`
  3. 统一导入 `toast` 和 `ApiError`
  4. 确保调用方（组件层）的 `void xxx()` 不再产生 unhandled rejection
- **测试**: hook 测试中 mock API 抛错，验证 toast 被调用

### P1-3. 前端分配编辑并行保存改串行
- **问题**: `Promise.all` 并行更新导致部分失败和不一致
- **文件**: `apps/frontend/src/features/budget/hooks/use-budget-allocation-edit.ts:101-119`
- **修复方案**:
  1. 将 `Promise.all` 改为 `for...of` 顺序执行
  2. 如果任一更新失败，立即中断并 setError
  3. 失败后调用 refresh 回滚 UI 至服务端状态
  4. 考虑添加后端批量更新端点（后续优化）
- **测试**: mock 第二个 API 调用失败，验证第一个成功的不被回滚且 error 正确显示

### P1-4. 前端 useEffect 竞态与内存泄漏修复
- **问题**: `budget-project-members-section.tsx` 和 `budget-org-member-picker.tsx` 中 effect 无清理
- **文件**:
  - `apps/frontend/src/features/budget/components/budget-project-members-section.tsx:40-45`
  - `apps/frontend/src/features/budget/components/budget-org-member-picker.tsx:42-57, 132-148`
- **修复方案**:
  1. 添加 `let cancelled = false` 模式（与项目现有风格一致，不引入 AbortController）:
     ```typescript
     useEffect(() => {
       let cancelled = false
       budgetApi.getGroupMemberConsumed(project.id).then(data => {
         if (!cancelled) setConsumedMap(data)
       }).catch(() => {})
       return () => { cancelled = true }
     }, [project.id])
     ```
  2. org-member-picker 的搜索同理：在 setTimeout 回调内检查 cancelled
  3. org-member-picker 的树加载同理
- **测试**: 组件测试验证快速切换 project.id 不产生 React warnings

### P1-5. 前端 departmentName 数据错误修复
- **问题**: 所有项目被赋予当前选中节点名作为 departmentName
- **文件**: `apps/frontend/src/features/budget/hooks/use-budget-page.ts:85-88`
- **修复方案**:
  1. `mapGroupsToProjectViews` 需要按 group 的 `departmentIds` 查找对应 node 名称
  2. 修改 mapper 接收整棵 tree 或 node map，按 group.departmentIds[0] 查找正确的部门名
  3. 如果 group 关联多个部门，取第一个部门名（与现有 UI 一致）
- **测试**: mapper 测试验证不同部门的 group 各自显示正确部门名

### P1-6. 前端项目创建可用预算计算修复
- **问题**: 可用预算不扣除已有项目（BudgetGroup）的预算总和
- **文件**: `apps/frontend/src/features/budget/components/budget-project-dialog.tsx:47-51`
- **修复方案**:
  1. `available` 计算需减去当前部门下所有 groups 的 budget 总和
  2. `available = department.budget - childrenSum - projectBudgetSum - nodeReservedPool(department)`
  3. 需从 props 传入或从 parent hook 获取当前部门的 groups 列表
- **测试**: 组件/hook 测试验证有已有项目时可用预算正确减少

### P1-7. 后端 N+1 查询批量化
- **问题**: `mergeBudgetTreeConsumed` 和 `GetGroupMemberConsumed` 逐个查询
- **文件**:
  - `apps/backend/internal/pkg/budget/snapshotload.go:110-140`
  - `apps/backend/internal/domain/budget/tree.go:113-148`
- **修复方案**:
  1. `mergeBudgetTreeConsumed`: 先遍历 tree 收集所有 (nodeID, period) 对，调用 `ListConsumedByPeriods` 批量获取，再赋值回 nodes
  2. `GetGroupMemberConsumed`: 收集所有 memberIDs，调用批量接口获取 consumed map，再按 member 分配
  3. 参考已有的 `LoadPlatformKeysWithUsed` 和 `LoadBudgetGroupsWithConsumed` 的批量模式
- **测试**: 验证结果正确性不变 + benchmark 对比查询数

### P1-8. 后端审计日志
- **问题**: 预算变更操作无任何日志记录
- **文件**: 所有 domain/budget/ 文件
- **修复方案**:
  1. `Service` struct 增加 `logger *slog.Logger` 字段（参考 OverrunService 模式）
  2. 构造函数 `NewService` 接受 `*slog.Logger` 参数
  3. 每个变更操作在成功后记录:
     ```go
     s.logger.Info("budget.group.created",
       "company_id", companyID,
       "group_id", group.ID,
       "group_name", group.Name,
       "budget", group.Budget,
     )
     ```
  4. 审批解决额外记录 operator_id（从 ctx 获取）
  5. 更新所有调用点传入 logger
- **测试**: 使用 `slogtest` 或检查日志输出包含预期字段

### P1-9. 后端 Rebalance/Notification 错误不再静默丢弃
- **问题**: `enqueueRebalanceAxis` 和 `notifier.Send` 错误被 `_ =` 丢弃
- **文件**:
  - `apps/backend/internal/domain/budget/approvals.go:120`
  - `apps/backend/internal/domain/budget/overrun.go:120`
- **修复方案**:
  1. 不阻塞主流程（不返回 error），但用 `slog.Error` 记录失败:
     ```go
     if err := s.enqueueRebalanceAxis(ctx, ...); err != nil {
       s.logger.Error("enqueue rebalance failed", "error", err, "member_id", memberID)
     }
     ```
  2. 同理处理 `notifier.Send` 失败
- **测试**: 使用 `FailingNotifier` 测试通知失败不影响主操作结果

### P1-10. "预留池"语义统一
- **问题**: mappers.ts、budget-detail-team.tsx、nodeReservedPool() 三处计算语义不同
- **文件**:
  - `apps/frontend/src/features/budget/lib/mappers.ts` (`computeUnallocated`)
  - `apps/frontend/src/features/budget/components/budget-detail-team.tsx:88`
- **修复方案**:
  1. 明确定义：`reservedPool` = 服务端返回的 `node.reservedPool` 值（唯一来源）
  2. `computeUnallocated` 重命名为 `computeAvailable`，语义改为"部门可再分配额度"
  3. `budget-detail-team.tsx` 中 SummaryCard 统一使用 `nodeReservedPool(node)` 显示预留池
  4. 删除冗余的本地计算
- **测试**: 更新 mappers.test.ts

---

## P2 — 性能与健壮性（下迭代）

### P2-1. 后端 List 端点分页
- **文件**: `budget_repo_approvals.go`, `budget_repo_groups.go`, `budget_repo_alerts.go`
- **方案**: 增加 `ListOptions{Limit, Offset}` 参数，SQL 加 `LIMIT $n OFFSET $m`，默认 limit=100

### P2-2. 后端 Advisory Lock 命名空间
- **文件**: `apps/backend/internal/store/postgres/budget_repo.go:15`
- **方案**: 改用双参数 `pg_advisory_xact_lock(namespace, companyID)`，budget 模块用固定 namespace 常量（如 `100`）

### P2-3. 后端请求体大小限制
- **文件**: `apps/backend/internal/http/httputil/decode.go`
- **方案**: `r.Body = http.MaxBytesReader(w, r.Body, 1<<20)` (1MB limit)

### P2-4. 后端 Alert Rule Thresholds 验证
- **文件**: `apps/backend/internal/domain/budget/alerts.go:17-39`
- **方案**: 校验每个 threshold 在 [1, 100] 范围内，去重，排序

### P2-5. 后端 Group 名称验证
- **文件**: `apps/backend/internal/domain/budget/groups.go`
- **方案**: `strings.TrimSpace(name)` 非空，长度 <= 100

### P2-6. 后端 Float64 精度改进
- **文件**: `internal/pkg/budget/remain.go`, `internal/domain/types/budget.go`
- **方案**:
  1. 短期：所有比较使用 epsilon（`math.Abs(a-b) < 0.001`），负值结果 clamp 到 0
  2. 长期（后续版本）：考虑迁移至整数分（points * 1000）

### P2-7. 后端 DeleteGroup 触发 Rebalance
- **文件**: `apps/backend/internal/domain/budget/groups.go:94-114`
- **方案**: 删除成功后 enqueue rebalance for affected platform keys

### P2-8. 后端 UpdateOverrunPolicy 加事务
- **文件**: `apps/backend/internal/domain/budget/policy.go:15-23`
- **方案**: 包裹 `WithTx` + `AcquireBudgetLock`

### P2-9. 前端搜索防抖竞态修复
- **文件**: `apps/frontend/src/features/budget/components/budget-org-member-picker.tsx:132-148`
- **方案**: 在 setTimeout 回调内使用 `cancelled` flag 或 request ID 防止旧响应覆盖新响应

### P2-10. 前端 budget-project-members-section DI 修复
- **文件**: `apps/frontend/src/features/budget/components/budget-project-members-section.tsx:3`
- **方案**: 将 `getGroupMemberConsumed` 通过 props 从 parent hook 传入，移除直接 import

### P2-11. 前端审批 resolving 状态细化
- **文件**: `apps/frontend/src/features/budget/components/budget-approval-drawer.tsx`
- **方案**: 将 `resolving: boolean` 改为 `resolvingId: string | null`，仅禁用正在处理的那一条

### P2-12. 前端 parseFloat 精度处理
- **文件**: `apps/frontend/src/features/budget/hooks/use-budget-allocation-edit.ts`
- **方案**: 使用 `Math.round(value * 100) / 100` 或引入 `toFixed(2)` 在显示层处理

---

## P3 — 代码质量与无障碍（持续改进）

### P3-1. 前端无障碍修复
| 组件 | 问题 | 修复 |
|------|------|------|
| `budget-tree-panel.tsx` | 缺 Arrow 键导航 | 实现 WAI-ARIA TreeView 键盘交互 |
| `budget-allocation-table.tsx` | 输入框无 label | 添加 `aria-label={node.name + ' 预算'}` |
| `department-tree-select.tsx` | 无键盘导航 | 添加 onKeyDown 处理 ArrowUp/Down/Enter |
| `budget-detail-team.tsx` | 进度条无 aria-label | 添加 `aria-label="预算使用率"` |
| 面包屑区域 | 无 nav landmark | 包裹 `<nav aria-label="breadcrumb">` |

### P3-2. 前端死代码清理
- 删除 `flattenBudgetNodes`（mappers.ts:53-65）
- 删除 `updateBudgetNodeInTree`（mappers.ts:67-86）— 仅测试使用则移至测试文件
- 删除 `reservedDraft`/`onUpdateReservedDraft` 未使用 props（budget-allocation-table.tsx）

### P3-3. 前端 alertRuleToView departmentName 修复
- **文件**: `apps/frontend/src/features/budget/lib/alerts.ts:23`
- **方案**: 从 tree 中查找 departmentId 对应的 name，或重命名字段为 `departmentId`

### P3-4. 前端 shiftBudgetPeriod 防御性编程
- **文件**: `apps/frontend/src/features/budget/lib/mappers.ts:41-44`
- **方案**: 加正则校验 `/^\d{4}-\d{2}$/`，不匹配则返回当前期间

### P3-5. 前端 member name 作为 React key 修复
- **文件**: `apps/frontend/src/features/budget/components/budget-org-member-picker.tsx:175`
- **方案**: 改用 `member.id` 作为 key

### P3-6. 后端 UpdateAlert 缺少 Delayer
- **文件**: `apps/backend/internal/domain/budget/alerts.go:42`
- **方案**: 添加 `s.delayer.Wait(ctx, 300*time.Millisecond)`

### P3-7. 后端 OverrunService Logger 实际使用
- **文件**: `apps/backend/internal/domain/budget/overrun.go`
- **方案**: 在 evaluateOverrun 中记录评估决策（哪些维度触发了 overrun）

### P3-8. 后端 CreateGroup 重名检查
- **文件**: `apps/backend/internal/domain/budget/groups.go:18-44`
- **方案**: 读取所有 groups 后检查是否存在同名，存在则返回 `domain.Conflict`

### P3-9. 后端 DefaultPersonalBudget 可配置化
- **文件**: `apps/backend/internal/pkg/common/constants.go`
- **方案**: 移至 `config.Config`，通过环境变量 `DEFAULT_PERSONAL_BUDGET` 配置

### P3-10. 补充测试覆盖
- **后端**:
  - Handler 层: 所有 budget endpoints 的 happy path + error path
  - Budget guard: member/group/platform key 各维度超支
  - 并发测试: approval + member budget 竞态
- **前端**:
  - `useBudgetAllocationEdit`: 验证逻辑、保存成功/失败、边界值
  - `useAlertRuleForm`: threshold 增删、验证
  - 组件渲染测试: approval-drawer、project-dialog
  - 错误场景: API 500、网络超时

---

## 执行时间估算

| 阶段 | 预计工时 | 交付物 |
|------|---------|--------|
| P0 | 3-4 天 | 5 个修复 + 对应测试，`make test-unit` 全绿 |
| P1 | 5-7 天 | 10 个修复 + 测试，`pnpm verify` 全绿 |
| P2 | 4-5 天 | 12 个修复，性能基准对比 |
| P3 | 3-4 天 | 10 个修复，无障碍测试通过 |
| **总计** | **15-20 天** | 全部 58 个问题修复完毕 |

---

## 验证清单

每个阶段完成后执行:

```bash
# 后端
cd apps/backend && make lint && make test-unit

# 前端
pnpm -F @tokenjoy/frontend build && pnpm -F @tokenjoy/frontend test

# 全栈
pnpm verify
```

---

## 风险与注意事项

1. **P0-1 和 P0-2 涉及事务边界变更**，需确认现有测试全部通过后再合并
2. **P1-1 可能需要后端 API 变更**（BudgetGroup 增加 overrunPolicy 字段），前后端需协调
3. **P2-6 Float64 改整数是破坏性变更**，需数据迁移，建议作为独立版本规划
4. **P1-7 N+1 批量化**需要新增 store 接口方法，影响 store.Store interface
5. **测试补充（P3-10）可与其他阶段并行**，不阻塞功能修复
