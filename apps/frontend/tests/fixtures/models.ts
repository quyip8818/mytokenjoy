import type { ModelInfo, ModelRef } from '@/api/types'

export const mockModelRefs: ModelRef[] = [
  { modelId: 1, type: 'deepseek-v4', name: 'DeepSeek V4', provider: 'deepseek', enabled: true },
  { modelId: 2, type: 'qwen-3.5-plus', name: 'Qwen 3.5 Plus', provider: 'qwen', enabled: true },
  {
    modelId: 3,
    type: 'claude-sonnet-5',
    name: 'Claude Sonnet 5',
    provider: 'anthropic',
    enabled: true,
  },
]

export const mockModels: ModelInfo[] = [
  {
    modelId: 1,
    type: 'deepseek-v4',
    name: 'DeepSeek V4',
    provider: 'deepseek',
    description: 'DeepSeek旗舰模型',
    inputPrice: 0.3,
    outputPrice: 0.5,
    maxContext: 128000,
    maxTokens: 4096,
    enabled: true,
    capabilities: ['chat'],
  },
  {
    modelId: 2,
    type: 'qwen-3.5-plus',
    name: 'Qwen 3.5 Plus',
    provider: 'qwen',
    description: '通义千问',
    inputPrice: 0.8,
    outputPrice: 2.0,
    maxContext: 1000000,
    maxTokens: 8192,
    enabled: true,
    capabilities: ['chat', 'vision'],
  },
  {
    modelId: 3,
    type: 'custom-model',
    name: 'My Custom',
    provider: 'custom',
    description: '自定义模型',
    inputPrice: 1.0,
    outputPrice: 2.0,
    maxContext: 128000,
    maxTokens: 4096,
    enabled: true,
    capabilities: ['chat'],
  },
]
