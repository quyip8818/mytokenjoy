import type { ModelInfo, ModelRef } from '@/api/types'

export function modelRefLabel(ref: Pick<ModelRef, 'name' | 'type'>): string {
  return ref.name || ref.type
}

export function buildModelIndex(models: ModelInfo[]): Map<number, ModelInfo> {
  return new Map(models.map((model) => [model.modelId, model]))
}

export function modelIdLabel(modelId: number, index: Map<number, ModelInfo>): string {
  const model = index.get(modelId)
  if (!model) return `#${modelId}`
  return model.name || model.type
}

export function modelIdsToLabels(modelIds: number[], index: Map<number, ModelInfo>): string[] {
  return modelIds.map((id) => modelIdLabel(id, index))
}
