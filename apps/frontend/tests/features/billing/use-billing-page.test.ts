import { describe, expect, it, vi } from 'vitest'
import { useBillingPage } from '@/features/billing/hooks/use-billing-page'
import { primaryWalletBalance } from '@/features/billing/lib/selectors'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'

describe('useBillingPage', () => {
  it('loads wallet balance and recharge records on mount', async () => {
    const wallet = {
      companyId: 1,
      billingCurrency: 'CNY',
      balances: [{ currency: 'CNY', balance: 100, totalTopup: 150, totalConsumed: 50 }],
      walletRemainQuota: 100000,
      giftQuota: 0,
      overdraftQuota: 0,
      totalRequests: 10,
    }
    const apis = createMockApis({
      billingApi: {
        getWallet: vi.fn().mockResolvedValue(wallet),
        listRechargeRecords: vi.fn().mockResolvedValue([]),
      },
    })

    const { result } = renderHookWithProviders(() => useBillingPage(apis), { apis })

    await waitForLoaded(result, 'loading')

    expect(apis.billingApi.getWallet).toHaveBeenCalled()
    expect(primaryWalletBalance(result.current.wallet)?.balance).toBe(100)
    expect(result.current.topUpRecords).toEqual([])
  })
})
