import { defineWorkflow } from '../types'
import { ApprovalSubmitWorkflow } from '../workflows/approval-submit'
import { ApprovalReviewWorkflow } from '../workflows/approval-review'

export const approvalWorkflowDefinitions = {
  'approval-submit': defineWorkflow(ApprovalSubmitWorkflow, { title: '发起申请' }),
  'approval-review': defineWorkflow(ApprovalReviewWorkflow, { title: '审批处理' }),
}
