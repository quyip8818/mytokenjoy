import type { ModelInfo } from '@/api/types'
import type { RoutingRule } from '@/api/types'

export interface ModelsWorkflowPayloads {
  'model-create': {
    onSuccess?: (id?: string | number) => void
  }
  'model-edit': {
    model: ModelInfo
    onSuccess?: (id?: string | number) => void
  }
  'whitelist-config': {
    rule: RoutingRule
    onSuccess?: () => void
  }
  'model-picker': {
    selectedModelIds?: number[]
    parentAllowedModelIds?: number[]
    onConfirm?: (modelIds: number[]) => void
  }
}
