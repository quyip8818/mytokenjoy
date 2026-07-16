import type { ModelInfo, ModelRef } from '@/api/types'

export const mockModelRefs: ModelRef[] = [
  { modelId: 1, type: 'gpt-4', name: 'GPT-4', provider: 'openai', enabled: true },
  { modelId: 2, type: 'gpt-4o', name: 'GPT-4o', provider: 'openai', enabled: true },
  { modelId: 3, type: 'claude-3', name: 'Claude 3', provider: 'anthropic', enabled: true },
]

export const mockModels: ModelInfo[] = [
  {
    modelId: 1,
    type: 'gpt-4',
    name: 'GPT-4',
    provider: 'openai',
    description: '',
    inputPrice: 0,
    outputPrice: 0,
    maxContext: 128000,
    enabled: true,
    capabilities: [],
  },
  {
    modelId: 2,
    type: 'gpt-4o',
    name: 'GPT-4o',
    provider: 'openai',
    description: '',
    inputPrice: 0,
    outputPrice: 0,
    maxContext: 128000,
    enabled: true,
    capabilities: [],
  },
]
