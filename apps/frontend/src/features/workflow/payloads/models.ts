import type { ModelInfo } from '@/api/types'
import type { RoutingRule } from '@/api/types'

export interface ModelsWorkflowPayloads {
  'model-create': {
    onSuccess?: (id?: string) => void
  }
  'model-edit': {
    model: ModelInfo
    onSuccess?: (id?: string) => void
  }
  'whitelist-config': {
    rule: RoutingRule
    onSuccess?: () => void
  }
  'model-picker': {
    selectedModelIds?: string[]
    parentAllowedModelIds?: string[]
    onConfirm?: (modelIds: string[]) => void
  }
}
