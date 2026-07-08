import { describe, expect, it, vi } from 'vitest'
import { useModelListPage } from '@/features/models/hooks/use-model-list-page'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'

describe('useModelListPage', () => {
  it('loads models on mount', async () => {
    const models = [
      {
        modelId: 1,
        type: 'gpt-4o',
        name: 'GPT-4o',
        provider: 'openai',
        description: '',
        inputPrice: 0,
        outputPrice: 0,
        maxContext: 0,
        enabled: true,
        capabilities: [],
      },
    ]
    const apis = createMockApis({
      modelApi: {
        list: vi.fn().mockResolvedValue(models),
      },
    })

    const { result } = renderHookWithProviders(() => useModelListPage(apis), { apis })

    await waitForLoaded(result, 'loading')

    expect(apis.modelApi.list).toHaveBeenCalled()
    expect(result.current.models).toEqual(models)
  })
})
