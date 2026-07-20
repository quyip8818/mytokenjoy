import type { ModelInfo, ModelRef } from '@/api/types'

export const mockModelRefs: ModelRef[] = [
  {
    modelId: '00000000-0000-7000-8000-0000000000b1',
    type: 'deepseek-v4-pro',
    name: 'DeepSeek V4 Pro',
    provider: 'deepseek',
    enabled: true,
  },
  {
    modelId: '00000000-0000-7000-8000-0000000000bb',
    type: 'deepseek-v4-flash',
    name: 'DeepSeek V4 Flash',
    provider: 'deepseek',
    enabled: true,
  },
]

export const mockModels: ModelInfo[] = [
  {
    modelId: '00000000-0000-7000-8000-0000000000b1',
    type: 'deepseek-v4-pro',
    name: 'DeepSeek V4 Pro',
    provider: 'deepseek',
    description: 'DeepSeek旗舰推理模型',
    inputPrice: 1.0,
    outputPrice: 2.5,
    maxContext: 128000,
    maxTokens: 4096,
    enabled: true,
    capabilities: ['chat'],
  },
  {
    modelId: '00000000-0000-7000-8000-0000000000bb',
    type: 'deepseek-v4-flash',
    name: 'DeepSeek V4 Flash',
    provider: 'deepseek',
    description: 'DeepSeek高速经济模型',
    inputPrice: 0.15,
    outputPrice: 0.6,
    maxContext: 128000,
    maxTokens: 4096,
    enabled: true,
    capabilities: ['chat'],
  },
]
