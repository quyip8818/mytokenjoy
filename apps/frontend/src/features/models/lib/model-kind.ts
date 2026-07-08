import type { ModelInfo } from '@/api/types'
import { CUSTOM_MODEL_PROVIDER, isCustomModel } from '@/api/types/models'

export { isCustomModel }

export function isBuiltinModel(model: Pick<ModelInfo, 'provider'>): boolean {
  return !isCustomModel(model)
}

export function matchesModelListTab(
  model: Pick<ModelInfo, 'provider'>,
  tab: 'all' | 'custom' | 'builtin',
): boolean {
  if (tab === 'all') return true
  if (tab === 'custom') return isCustomModel(model)
  return model.provider !== CUSTOM_MODEL_PROVIDER
}
