import type { ContractDetail } from '@/api/contracts'

export interface ContractsWorkflowPayloads {
  'contract-detail': {
    contract: ContractDetail
    canEdit: boolean
    onRefresh?: () => void
  }
}
