import { defineWorkflow } from '../types'
import { BudgetNodeEditWorkflow } from '../workflows/budget-node-edit'
import { BudgetImpactPreviewWorkflow } from '../workflows/budget-impact-preview'
import { BudgetGroupFormWorkflow } from '../workflows/budget-group-form'
import { OverrunPolicyWorkflow } from '../workflows/overrun-policy'
import { MemberQuotaConfigWorkflow } from '../workflows/member-quota-config'

export const budgetWorkflowDefinitions = {
  'budget-node-edit': defineWorkflow(BudgetNodeEditWorkflow, {
    defaultLayer: 1,
    title: '编辑节点预算',
  }),
  'budget-group-form': defineWorkflow(BudgetGroupFormWorkflow, {
    defaultLayer: 1,
    title: '新建预算组',
  }),
  'overrun-policy': defineWorkflow(OverrunPolicyWorkflow, {
    defaultLayer: 1,
    title: '全局超限策略',
  }),
  'budget-impact-preview': defineWorkflow(BudgetImpactPreviewWorkflow, {
    defaultLayer: 2,
    title: '影响范围预览',
  }),
  'member-quota-config': defineWorkflow(MemberQuotaConfigWorkflow, {
    defaultLayer: 1,
    title: '成员额度配置',
  }),
}
