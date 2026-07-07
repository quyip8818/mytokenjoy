import { describe, expect, it, vi } from 'vitest'
import { waitFor } from '@testing-library/react'
import { usePlatformKeysPage } from '@/features/keys/hooks/use-platform-keys-page'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'

describe('usePlatformKeysPage', () => {
  it('loads platform keys on mount', async () => {
    const items = [{ id: 'pk-1', name: 'Admin Key', status: 'active' }]
    const apis = createMockApis({
      platformKeyApi: {
        list: vi.fn().mockResolvedValue({ items, total: 1 }),
      },
    })

    const { result } = renderHookWithProviders(() => usePlatformKeysPage(apis), { apis })

    await waitForLoaded(result, 'loading')
    await waitFor(() => {
      expect(result.current.keys).toEqual(items)
    })

    expect(apis.platformKeyApi.list).toHaveBeenCalled()
  })
})
