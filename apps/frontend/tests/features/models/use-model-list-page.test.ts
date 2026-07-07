import { describe, expect, it, vi } from 'vitest'
import { useModelListPage } from '@/features/models/hooks/use-model-list-page'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'

describe('useModelListPage', () => {
  it('loads models on mount', async () => {
    const models = [
      {
        id: 'm1',
        name: 'gpt-4o',
        provider: 'openai',
        enabled: true,
        type: 'preset' as const,
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
