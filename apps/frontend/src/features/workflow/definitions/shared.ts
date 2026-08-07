import { defineWorkflow } from '../types'
import { RejectReasonWorkflow } from '../workflows/reject-reason'
import { ModelPickerWorkflow } from '../workflows/model-picker'
import { BudgetCheckWorkflow } from '../workflows/budget-check'

export const sharedWorkflowDefinitions = {
  'model-picker': defineWorkflow(ModelPickerWorkflow, { title: '选择模型' }),
  'budget-check': defineWorkflow(BudgetCheckWorkflow, { title: '额度不足' }),
  'reject-reason': defineWorkflow(RejectReasonWorkflow, { title: '拒绝理由' }),
}
