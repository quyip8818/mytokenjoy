import { defineWorkflow } from '../types'
import { ApprovalSubmitWorkflow } from '../workflows/approval-submit'
import { ApprovalReviewWorkflow } from '../workflows/approval-review'

export const approvalWorkflowDefinitions = {
  'approval-submit': defineWorkflow(ApprovalSubmitWorkflow, { defaultLayer: 1, title: '发起申请' }),
  'approval-review': defineWorkflow(ApprovalReviewWorkflow, { defaultLayer: 1, title: '审批处理' }),
}
