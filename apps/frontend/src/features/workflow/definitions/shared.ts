import { defineWorkflow } from '../types'
import { defineAlertWorkflow } from '../define-alert-workflow'
import { RejectReasonWorkflow } from '../workflows/reject-reason'
import { ModelPickerWorkflow } from '../workflows/model-picker'
import { PermissionPickerWorkflow } from '../workflows/permission-picker'

export const sharedWorkflowDefinitions = {
  'model-picker': defineWorkflow(ModelPickerWorkflow, { defaultLayer: 2, title: '选择模型' }),
  'quota-check': defineAlertWorkflow({
    id: 'quota-check',
    title: '额度不足',
    defaultLayer: 3,
    alert: (payload) => {
      const reservedPool = payload.reservedPool ?? 0
      const requested = payload.requested ?? 0
      return {
        title: '预留池额度不足，无法通过审批',
        description: `申请额度 ¥${requested.toLocaleString()}，当前预留池剩余 ¥${reservedPool.toLocaleString()}。请先调整预算分配或拒绝此申请。`,
      }
    },
  }),
  'reject-reason': defineWorkflow(RejectReasonWorkflow, { defaultLayer: 2, title: '拒绝理由' }),
  'permission-picker': defineWorkflow(PermissionPickerWorkflow, {
    defaultLayer: 2,
    title: '选择权限',
  }),
}
