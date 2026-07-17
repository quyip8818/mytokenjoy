import { useMemo, type ReactNode } from 'react'
import type { PermissionKey } from '@/lib/permission-keys'
import { BillingExchangeProvider } from '@/features/session/billing-exchange-provider'
import { SessionReactContext } from '@/features/session/context'
import type { AppSession } from '@/features/session/types'
import { createBillingExchange } from '@/lib/points'
import { mockMember } from '@tests/fixtures/members'

export function TestSessionProvider({
  children,
  permissions,
  readOnly,
}: {
  children: ReactNode
  permissions: PermissionKey[]
  readOnly: boolean
}) {
  const session = useMemo<AppSession>(
    () => ({
      companyId: 1,
      companyType: 'selfhosted',
      authzRevision: 0,
      memberId: mockMember.id,
      member: mockMember,
      permissions,
      readOnly,
      billingCurrency: 'CNY',
      pointsPerUnit: 1000,
      loading: false,
      sessionError: null,
      refreshSession: async () => {},
    }),
    [permissions, readOnly],
  )

  const exchange = useMemo(() => createBillingExchange(1000, 'CNY'), [])

  return (
    <SessionReactContext.Provider value={session}>
      <BillingExchangeProvider exchange={exchange}>{children}</BillingExchangeProvider>
    </SessionReactContext.Provider>
  )
}
