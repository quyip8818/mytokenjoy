import { describe, expect, it, vi } from 'vitest'
import { waitFor } from '@testing-library/react'
import { useProviderKeysPage } from '@/features/keys/hooks/use-provider-keys-page'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'
import { mockProviderKeys } from '@tests/fixtures/keys'

describe('useProviderKeysPage', () => {
  it('loads provider keys on mount', async () => {
    const apis = createMockApis({
      providerKeyApi: {
        list: vi.fn().mockResolvedValue(mockProviderKeys),
      },
    })

    const { result } = renderHookWithProviders(() => useProviderKeysPage(apis), { apis })

    await waitForLoaded(result, 'loading')
    await waitFor(() => {
      expect(result.current.keys).toEqual(mockProviderKeys)
    })

    expect(apis.providerKeyApi.list).toHaveBeenCalled()
  })
})
