import type { ProviderType } from './keys'

export interface ModelInfo {
  id: string
  provider: ProviderType
  name: string
  displayName: string
  inputPrice: number
  outputPrice: number
  maxContext: number
  enabled: boolean
  capabilities: string[]
}

export interface RoutingRule {
  id: string
  nodeId: string
  nodeName: string
  allowedModels: string[]
  defaultModel: string | null
  fallbackModel: string | null
  inherited: boolean
}
