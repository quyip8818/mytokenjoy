import type { ModelInfo, ModelRef } from '@/api/types'

export function modelRefLabel(ref: Pick<ModelRef, 'name' | 'type'>): string {
  return ref.name || ref.type
}

export function buildModelIndex(models: ModelInfo[]): Map<string, ModelInfo> {
  return new Map(models.map((model) => [model.modelId, model]))
}

export function modelIdLabel(modelId: string, index: Map<string, ModelInfo>): string {
  const model = index.get(modelId)
  if (!model) return `#${modelId}`
  return model.name || model.type
}

export function modelIdsToLabels(modelIds: string[], index: Map<string, ModelInfo>): string[] {
  return modelIds.map((id) => modelIdLabel(id, index))
}
