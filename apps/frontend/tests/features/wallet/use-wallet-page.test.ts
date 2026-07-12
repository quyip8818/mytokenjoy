import { describe, expect, it, vi } from 'vitest'
import { useWalletPage } from '@/features/wallet/hooks/use-wallet-page'
import { primaryWalletBalance } from '@/features/wallet/lib/selectors'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'

describe('useWalletPage', () => {
  it('loads wallet balance and recharge records on mount', async () => {
    const wallet = {
      companyId: 1,
      billingCurrency: 'CNY',
      balances: [{ currency: 'CNY', balance: 100, totalTopup: 150, totalConsumed: 50 }],
      walletRemainPoint: 100000,
      giftPoints: 0,
      overdraftPoints: 0,
      totalRequests: 10,
    }
    const apis = createMockApis({
      billingApi: {
        getWallet: vi.fn().mockResolvedValue(wallet),
        listRechargeRecords: vi.fn().mockResolvedValue([]),
      },
    })

    const { result } = renderHookWithProviders(() => useWalletPage(apis), { apis })

    await waitForLoaded(result, 'loading')

    expect(apis.billingApi.getWallet).toHaveBeenCalled()
    expect(primaryWalletBalance(result.current.wallet)?.balance).toBe(100)
    expect(result.current.topUpRecords).toEqual([])
  })
})
