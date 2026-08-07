import { defineWorkflow } from '../types'
import { LotAuditWorkflow } from '../workflows/lot-audit'
import { RechargeWorkflow } from '../workflows/recharge'

export const billingWorkflowDefinitions = {
  'lot-audit': defineWorkflow(LotAuditWorkflow, { title: 'Lot 审计' }),
  recharge: defineWorkflow(RechargeWorkflow, { title: '账户充值' }),
}
