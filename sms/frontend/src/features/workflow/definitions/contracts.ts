import { defineWorkflow } from '../types'
import { ContractDetailWorkflow } from '../workflows/contract-detail'

export const contractsWorkflowDefinitions = {
  'contract-detail': defineWorkflow(ContractDetailWorkflow, { title: '合同详情' }),
}
