import { defineWorkflow } from '../types'
import { RejectReasonWorkflow } from '../workflows/reject-reason'
import { QuotaCheckWorkflow } from '../workflows/quota-check'
import { ModelPickerWorkflow } from '../workflows/model-picker'
import { PermissionPickerWorkflow } from '../workflows/permission-picker'

export const sharedWorkflowDefinitions = {
  'model-picker': defineWorkflow(ModelPickerWorkflow, { defaultLayer: 2, title: '选择模型' }),
  'quota-check': defineWorkflow(QuotaCheckWorkflow, { defaultLayer: 3, title: '额度不足' }),
  'reject-reason': defineWorkflow(RejectReasonWorkflow, { defaultLayer: 2, title: '拒绝理由' }),
  'permission-picker': defineWorkflow(PermissionPickerWorkflow, {
    defaultLayer: 2,
    title: '选择权限',
  }),
}
