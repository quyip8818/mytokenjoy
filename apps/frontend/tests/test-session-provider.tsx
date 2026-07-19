import { useMemo, type ReactNode } from 'react'
import type { PermissionKey } from '@/lib/permission-keys'
import type { CompanyType } from '@/api/types/common'
import { BillingExchangeProvider } from '@/features/session/billing-exchange-provider'
import { SessionReactContext } from '@/features/session/context'
import type { AppSession } from '@/features/session/types'
import { createBillingExchange } from '@/lib/points'
import { mockMember } from '@tests/fixtures/members'

export function TestSessionProvider({
  children,
  permissions,
  readOnly,
  companyType = 'selfhosted',
}: {
  children: ReactNode
  permissions: PermissionKey[]
  readOnly: boolean
  companyType?: CompanyType
}) {
  const session = useMemo<AppSession>(
    () => ({
      companyId: '00000000-0000-7000-8000-000000000002',
      companyType,
      authzRevision: 0,
      memberId: mockMember.id,
      member: mockMember,
      permissions,
      readOnly,
      billingCurrency: 'CNY',
      quotaPerUnit: 500000,
      loading: false,
      sessionError: null,
      refreshSession: async () => {},
    }),
    [permissions, readOnly, companyType],
  )

  const exchange = useMemo(() => createBillingExchange(1000, 'CNY'), [])

  return (
    <SessionReactContext.Provider value={session}>
      <BillingExchangeProvider exchange={exchange}>{children}</BillingExchangeProvider>
    </SessionReactContext.Provider>
  )
}
