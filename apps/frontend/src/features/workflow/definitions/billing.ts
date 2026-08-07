import { defineWorkflow } from '../types'
import { LotAuditWorkflow } from '../workflows/lot-audit'

export const billingWorkflowDefinitions = {
  'lot-audit': defineWorkflow(LotAuditWorkflow, { title: 'Lot 审计' }),
}
