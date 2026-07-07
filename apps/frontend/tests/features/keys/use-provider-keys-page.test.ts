import { describe, expect, it, vi } from 'vitest'
import { waitFor } from '@testing-library/react'
import { useProviderKeysPage } from '@/features/keys/hooks/use-provider-keys-page'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'

describe('useProviderKeysPage', () => {
  it('loads provider keys on mount', async () => {
    const items = [{ id: 'vk-1', name: 'OpenAI Key', provider: 'openai', status: 'active' }]
    const apis = createMockApis({
      providerKeyApi: {
        list: vi.fn().mockResolvedValue(items),
      },
    })

    const { result } = renderHookWithProviders(() => useProviderKeysPage(apis), { apis })

    await waitForLoaded(result, 'loading')
    await waitFor(() => {
      expect(result.current.keys).toEqual(items)
    })

    expect(apis.providerKeyApi.list).toHaveBeenCalled()
  })
})
