import { describe, expect, it, vi } from 'vitest'
import { useWalletPage } from '@/features/wallet/hooks/use-wallet-page'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'

describe('useWalletPage', () => {
  it('loads wallet balance and recharge records on mount', async () => {
    const apis = createMockApis({
      billingApi: {
        getWallet: vi.fn().mockResolvedValue({
          balance: 100,
          currency: 'CNY',
          totalConsumed: 50,
          totalRequests: 10,
        }),
        listRechargeRecords: vi.fn().mockResolvedValue([]),
      },
    })

    const { result } = renderHookWithProviders(() => useWalletPage(apis), { apis })

    await waitForLoaded(result, 'loading')

    expect(apis.billingApi.getWallet).toHaveBeenCalled()
    expect(result.current.balance).toBe(100)
    expect(result.current.currency).toBe('CNY')
    expect(result.current.topUpRecords).toEqual([])
  })
})
