import { describe, expect, it, vi } from 'vitest'
import { screen, fireEvent } from '@testing-library/react'
import { BillingPageShell } from '@/features/billing/components/billing-page-shell'
import { createMockApis, renderWithProviders } from '@tests/utils'

// ponytail: mock IS_SAAS at module level — toggled per test via mockReturnValue
const isSaasMock = vi.fn(() => true)
vi.mock('@/config/app', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/config/app')>()
  return {
    ...actual,
    get IS_SAAS() {
      return isSaasMock()
    },
  }
})

const baseProps = {
  wallet: {
    companyId: '00000000-0000-7000-8000-000000000002',
    billingCurrency: 'CNY',
    balances: [{ currency: 'CNY', balance: 50, totalTopup: 100, totalConsumed: 50 }],
    walletRemainQuota: 50000,
    giftQuota: 0,
    overdraftQuota: 0,
    totalRequests: 5,
  },
  loading: false,
  error: null,
  refresh: vi.fn(),
  topUpRecords: [],
}

describe('BillingPageShell recharge button', () => {
  it('shows recharge button and opens workflow on SaaS standard account', () => {
    isSaasMock.mockReturnValue(true)
    const apis = createMockApis()

    renderWithProviders(<BillingPageShell {...baseProps} />, {
      apis,
      companyType: 'saas',
    })

    const btn = screen.getByRole('button', { name: '充值' })
    expect(btn).toBeInTheDocument()

    // Clicking should NOT show the local hint dialog
    fireEvent.click(btn)
    expect(screen.queryByText('在线充值')).not.toBeInTheDocument()
  })

  it('shows recharge button with local hint dialog on non-SaaS deployment', () => {
    isSaasMock.mockReturnValue(false)
    const apis = createMockApis()

    renderWithProviders(<BillingPageShell {...baseProps} />, {
      apis,
      companyType: 'selfhosted',
    })

    const btn = screen.getByRole('button', { name: '充值' })
    expect(btn).toBeInTheDocument()

    fireEvent.click(btn)
    expect(screen.getByText('在线充值')).toBeInTheDocument()
    expect(screen.getByText(/www\.tokenjoy\.me\/billing/)).toBeInTheDocument()
  })
})
