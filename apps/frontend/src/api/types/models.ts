import type { ProviderType } from './keys'

export const CUSTOM_MODEL_PROVIDER: ProviderType = 'custom'

export interface ModelInfo {
  modelId: number
  provider: ProviderType
  type: string
  name: string
  description: string
  endpoint?: string
  apiKey?: string
  endpointModelName?: string
  inputPrice: number
  outputPrice: number
  maxContext: number
  maxTokens: number
  enabled: boolean
  capabilities: string[]
}

export function isCustomModel(model: Pick<ModelInfo, 'provider'>): boolean {
  return model.provider === CUSTOM_MODEL_PROVIDER
}

export interface ModelRef {
  modelId: number
  type: string
  name: string
  provider: ProviderType
  enabled: boolean
}

export interface CreateModelInput {
  type: string
  name: string
  baseUrl: string
  apiKey?: string
  endpointModelName?: string
  inputPrice: number
  outputPrice: number
  maxContext: number
  maxTokens?: number
  capabilities?: string[]
}

export interface UpdateModelInput {
  name?: string
  type?: string
  description?: string
  endpoint?: string
  apiKey?: string
  endpointModelName?: string
  inputPrice?: number
  outputPrice?: number
  maxContext?: number
  maxTokens?: number
  capabilities?: string[]
}

export interface RoutingRule {
  id: string
  nodeId: string
  nodeName: string
  allowedModelIds: number[]
  defaultModelId: number | null
  fallbackModelId: number | null
  inherited: boolean
  allowedModels?: ModelRef[]
  defaultModel?: ModelRef | null
  fallbackModel?: ModelRef | null
}

export interface ResolvedWhitelist {
  inherited: boolean
  allowedModelIds: number[]
  parentCount: number
  allowedModels?: ModelRef[]
}
