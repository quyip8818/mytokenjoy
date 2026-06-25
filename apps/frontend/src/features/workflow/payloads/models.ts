import type { RoutingRule } from '@/api/types'

export interface ModelsWorkflowPayloads {
  'model-create': {
    onSuccess?: (id?: string) => void
  }
  'whitelist-config': {
    rule: RoutingRule
    onSuccess?: () => void
  }
  'model-picker': {
    selectedModels?: string[]
    parentWhitelist?: string[]
    onConfirm?: (models: string[]) => void
  }
}
