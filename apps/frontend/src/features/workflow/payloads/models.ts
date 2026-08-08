import type { ModelInfo, PlatformModel } from '@/api/types'
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
  'platform-model-create': {
    onSuccess?: () => void
  }
  'platform-model-edit': {
    model: PlatformModel
    onSuccess?: () => void
  }
  'platform-model-channels': {
    model: PlatformModel
  }
  'discount-config': {
    companyId: string
    companyName: string
    onSuccess?: () => void
  }
}
