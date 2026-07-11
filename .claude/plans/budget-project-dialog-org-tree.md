# Plan: 创建项目 Dialog — 组织树选人 + 额度校验

## 需求

1. 「关联成员」改为组织树形式选人，默认展开当前创建项目所属的部门节点
2. 额度校验不能超过当前团队剩余预算（未分配额度）

## 改动范围

### 1. 新建 `budget-org-member-picker.tsx`

替代现有的 `BudgetMemberPicker`（flat checkbox list），实现为：
- Popover 触发器（保持现有外观）
- 内部渲染组织树（调用 `departmentApi.getTree()` 获取）
- 每个部门节点可展开，展开后显示该部门成员列表（调用 `memberApi.list`）
- 成员行带 Checkbox，支持多选
- Props: `selectedIds`, `onChange`, `defaultExpandDepartmentId`（当前部门 ID，默认展开）
- 树节点复用 budget-tree-panel 的视觉风格（缩进、图标、展开动画一致）

### 2. 修改 `budget-project-dialog.tsx`

- 将 `BudgetMemberPicker` 替换为 `BudgetOrgMemberPicker`
- 传入 `defaultExpandDepartmentId={department.id}`
- 额度校验逻辑：已有 `available` 计算（`department.budget - childrenSum - reservedPool`），但当前仅在提交时校验。需要在输入框旁显示可用额度提示（已有），并在校验时使用正确的"团队剩余未分配额度"

### 3. 修改 `budget-detail-team.tsx` + `budget-page-shell.tsx` + `use-budget-page.ts`

- 传递 `getDepartmentTree` 和 `getMembers` 方法到 Dialog，供组织树选人组件使用

## 额度校验细节

当前 `available` 的计算：
```ts
const available = department.budget - childrenSum - nodeReservedPool(department)
```
这个计算已经是正确的"当前团队未分配额度"。现有校验逻辑 `displayToPoints(budgetNum) > available` 已经正确。无需改动校验逻辑本身。

## 组织树选人组件设计

```
┌─ Popover ─────────────────────────┐
│ 🔍 搜索成员...                     │
│ ─────────────────────────────────  │
│ ▼ 📂 总公司                        │
│   ▼ 📂 技术部 (当前默认展开)        │
│     ☐ 张三  技术部                  │
│     ☑ 李四  技术部                  │
│   ▶ 📁 产品部                      │
│   ▶ 📁 运营部                      │
│                                    │
│ 已选 2 人                          │
└────────────────────────────────────┘
```

## 技术要点

- `departmentApi.getTree()` 返回 `Department[]`（与 BudgetNode 结构不同但类似）
- 成员列表用 `memberApi.list({ departmentId, page: 1, pageSize: 200 })`
- 组织树节点只在展开时按需加载成员（避免一次性加载所有部门的成员）
- 搜索走 `memberApi.list({ keyword })` 全局搜索
