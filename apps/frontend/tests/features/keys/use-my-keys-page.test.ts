import { describe, expect, it, vi } from 'vitest'
import { act, waitFor } from '@testing-library/react'
import { useMyKeysPage } from '@/features/keys/hooks/use-my-keys-page'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'
import { mockPlatformKeys, mockBudgetSummary } from '@tests/fixtures/keys'

describe('useMyKeysPage', () => {
  it('loads member keys and budget summary on mount', async () => {
    const apis = createMockApis({
      platformKeyApi: {
        list: vi
          .fn()
          .mockResolvedValue({ items: mockPlatformKeys, total: mockPlatformKeys.length }),
      },
      budgetApi: {
        getMemberSummary: vi.fn().mockResolvedValue(mockBudgetSummary),
      },
    })

    const { result } = renderHookWithProviders(() => useMyKeysPage(apis), { apis })

    await waitForLoaded(result, 'loading')
    await waitFor(() => {
      expect(result.current.keys).toEqual([mockPlatformKeys[0], mockPlatformKeys[2]])
    })

    expect(apis.platformKeyApi.list).toHaveBeenCalled()
    expect(apis.budgetApi.getMemberSummary).toHaveBeenCalled()
    expect(result.current.budgetSummary).toEqual(mockBudgetSummary)
  })

  it('opens create workflow with member scope', async () => {
    const apis = createMockApis({
      platformKeyApi: {
        list: vi.fn().mockResolvedValue({ items: [], total: 0 }),
      },
      budgetApi: {
        getMemberSummary: vi.fn().mockResolvedValue(mockBudgetSummary),
      },
    })

    const { result } = renderHookWithProviders(() => useMyKeysPage(apis), { apis })
    await waitForLoaded(result, 'loading')

    await act(async () => {
      result.current.openCreateKey()
    })

    expect(result.current.budgetSummary?.remaining).toBeGreaterThan(0)
  })
})
