import { describe, expect, it, vi } from 'vitest'
import { waitFor } from '@testing-library/react'
import { useMyKeysPage } from '@/features/keys/hooks/use-my-keys-page'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'

describe('useMyKeysPage', () => {
  it('loads member keys and quota on mount', async () => {
    const items = [{ id: 'pk-1', name: 'My Key', status: 'active' }]
    const quota = { remaining: 1000, used: 200, limit: 1200 }
    const apis = createMockApis({
      platformKeyApi: {
        list: vi.fn().mockResolvedValue({ items, total: 1 }),
        getBudgetSummary: vi.fn().mockResolvedValue(quota),
      },
    })

    const { result } = renderHookWithProviders(() => useMyKeysPage(apis), { apis })

    await waitForLoaded(result, 'loading')
    await waitFor(() => {
      expect(result.current.keys).toEqual(items)
    })

    expect(apis.platformKeyApi.list).toHaveBeenCalled()
    expect(apis.platformKeyApi.getBudgetSummary).toHaveBeenCalled()
    expect(result.current.quota).toEqual(quota)
  })
})
