# 禁用按钮审计 — Disabled 状态需展示 Tooltip 说明原因

## 现状

当前项目中，所有 disabled 按钮均仅变灰（`opacity-50 + pointer-events-none`），**没有任何 tooltip/hover 提示告知用户为什么无法操作**。唯一例外是 `step-field-mapping.tsx` 中以纯文本形式在按钮旁边显示原因。

## 需要添加 disabled tooltip 的场景清单

### 1. Key 管理

| 按钮 | 文件 | 禁用条件 | 建议 Tooltip |
|------|------|----------|-------------|
| 新建 Key / 创建 Key | `features/keys/components/my-keys-admin-page-shell.tsx` | `budgetSummary.remaining <= 0` | "可用额度不足，请先充值或申请额度" |
| 新建 Key | `features/keys/components/member-keys-page-shell.tsx` | `budgetSummary.remaining <= 0` | "可用额度为 0，无法创建新 Key" |
| 创建 Key（表单提交） | `features/workflow/workflows/key-form/index.tsx` | `!name.trim()` | "请输入 Key 名称" |
| 同上 | 同上 | `requiresMemberPick && !targetMemberId` | "请先选择绑定成员" |
| 同上 | 同上 | `budgetInsufficient` | "账户额度为 0" |
| 同上 | 同上 | `budgetExceedsRemaining` | "申请额度超过账户剩余" |
| 同上 | 同上 | `projectBudgetExceeds` | "申请额度超过项目剩余" |
| 同上 | 同上 | `subBudgetExceeds` | "申请额度超过成员子额度剩余" |
| 保存 Key（编辑表单） | 同上 | `!name.trim()` | "请输入 Key 名称" |

### 2. 充值 / 计费

| 按钮 | 文件 | 禁用条件 | 建议 Tooltip |
|------|------|----------|-------------|
| 确认充值 | `features/billing/components/recharge-panel.tsx` | `selectedAmount <= 0` | "请选择充值金额" |
| 同上 | 同上 | `rechargePending` | "充值处理中，请稍候" |
| 兑换额度 | 同上 | 永久 disabled | "兑换码能力即将上线" |
| 批量开票 | `features/billing/components/recharge-records-table.tsx` | 永久 disabled | "批量开票即将上线" |
| 全部开票 | 同上 | 永久 disabled | "全部开票即将上线" |

### 3. 预算管理

| 按钮 | 文件 | 禁用条件 | 建议 Tooltip |
|------|------|----------|-------------|
| 应用到全部 | `features/budget/components/budget-edit-member-budget.tsx` | `!averageDraft.trim()` | "请先输入平均额度" |
| 同上 | 同上 | `savingAverage` | "正在保存…" |
| 成员额度保存/取消 | 同上 | `saving` | "正在保存…" |
| 总额度确定 | `features/budget/components/budget-init-prompt.tsx` | `!draft.trim()` | "请输入总额度" |
| 同上 | 同上 | `saving` | "正在保存…" |
| 超额策略保存 | `features/budget/components/budget-overrun-policy-section.tsx` | `saving` | "正在保存…" |
| 删除项目（确认） | `features/budget/components/project-delete-action.tsx` | `deleting` | "正在删除…" |
| 成员选择器 | `features/budget/components/budget-member-picker.tsx` | `disabled`（来自 parent） | "请先选择团队" |

### 4. 组织 / 数据源

| 按钮 | 文件 | 禁用条件 | 建议 Tooltip |
|------|------|----------|-------------|
| 测试按钮 | `features/org/components/data-source/step-field-mapping.tsx` | `!testKeyword.trim()` | "请输入测试关键字" |
| 下一步 | 同上 | `!testPassed \|\| missingRequired.length > 0` | 动态："请先完成映射测试" / "请为必填字段选择目标" |
| 平台选择-下一步 | `features/org/components/data-source/platform-select.tsx` | `!selected` | "请选择一个平台" |
| 测试连接 | `features/org/components/data-source/step-credentials.tsx` | `testing` | "正在测试…" |
| 保存并完成 | `features/org/components/data-source/step-sync-schedule.tsx` | `saving` | "正在保存…" |
| 保存凭证 | `features/org/components/credential-form.tsx` | `!testSuccess` | "请先通过连接测试" |
| 同上 | 同上 | `saving` | "正在保存…" |
| 分步器步骤 | `features/org/components/data-source/stepper.tsx` | `!isClickable` | "请先完成当前步骤" |

### 5. 组织结构

| 按钮 | 文件 | 禁用条件 | 建议 Tooltip |
|------|------|----------|-------------|
| 发送邀请（FormDialog） | `features/org/components/structure/invite-dialog.tsx` | `!value.trim()` | "请输入手机号/邮箱" |
| 批量转移（FormDialog） | `features/org/components/structure/transfer-members-dialog.tsx` | `!transferDeptId` | "请选择目标部门" |
| 分页上一页/下一页 | `features/org/components/role-member-table.tsx`、`structure/member-table.tsx` | 边界 | "已是第一页" / "已是最后一页" |
| 搜索按钮 | `features/org/components/role-member-table.tsx` | `loading` | "搜索中…" |

### 6. 模型管理

| 按钮 | 文件 | 禁用条件 | 建议 Tooltip |
|------|------|----------|-------------|
| 模型启/停 Switch | `features/models/components/model-list-table.tsx` | `togglingIds.has(model.modelId)` | "操作中，请稍候" |
| 路由保存 | `features/models/components/routing-detail-panel.tsx` | `saving` | "正在保存…" |

### 7. 认证 / 登录

| 按钮 | 文件 | 禁用条件 | 建议 Tooltip |
|------|------|----------|-------------|
| 登录（手机密码） | `features/auth/components/auth-card.tsx` | `!phone.trim() \|\| !password` | "请填写手机号和密码" |
| 登录（验证码） | 同上 | `!code.trim()` | "请输入验证码" |
| 登录（邮箱密码） | 同上 | `!email.trim() \|\| !password` | "请填写邮箱和密码" |
| 重置密码 | 同上 | `newPassword.length < 8 \|\| newPassword !== confirmNewPassword` | "密码至少8位且需一致" |
| 创建企业 | 同上 | `!companyName.trim()` | "请输入企业名称" |
| 以上所有 | 同上 | `submitting` | "请求中…" |

### 8. 通用 / 共享组件

| 组件 | 文件 | 禁用条件 | 建议 Tooltip |
|------|------|----------|-------------|
| FormDialog 提交 | `components/ui/form-dialog.tsx` | `busy \|\| submitDisabled` | 由调用方传入 reason |
| WorkflowPanelFooter 主/危险按钮 | `features/workflow/components/workflow-panel-chrome.tsx` | `primaryDisabled` / `destructiveDisabled` | 由调用方传入 reason |
| 分页上/下 | `features/mydashboard/components/call-logs-list.tsx` | 边界 | "已是第一页" / "已是最后一页" |

### 9. Dashboard

| 按钮 | 文件 | 禁用条件 | 建议 Tooltip |
|------|------|----------|-------------|
| 调用日志翻页 | `features/mydashboard/components/call-logs-list.tsx` | `page <= 1` / `page >= totalPages` | "已到边界" |

---

## 分类总结

| 禁用原因类别 | 处理方式 |
|-------------|---------|
| **表单校验未通过**（必填为空、长度不足等） | Tooltip 提示缺什么 |
| **额度/预算不足** | Tooltip 显示余额数字 + 操作建议 |
| **功能未上线** | Tooltip "即将上线" |
| **正在请求中** | 可不加 tooltip（loading 状态已有 spinner），或简短 "处理中" |
| **前置步骤未完成** | Tooltip 告知要先做什么 |
| **分页边界** | Tooltip "已是第一/最后一页" |

## 实现建议

建议封装一个 `DisabledTooltip` wrapper（或扩展 Button 组件的 `disabledReason` prop），统一处理：

```tsx
// ponytail: 最小实现，升级路径 → Radix Tooltip 封装
interface ButtonProps {
  disabled?: boolean
  disabledReason?: string  // 有值时自动 wrap Tooltip
}
```

当 `disabled && disabledReason` 时，用 `<Tooltip>` 包裹按钮（注意 disabled button 无法触发 hover，需要在外层包一个 `<span>`）。
