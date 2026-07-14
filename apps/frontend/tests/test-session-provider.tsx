import { useMemo, type ReactNode } from 'react'
import type { PermissionKey } from '@/lib/permission-keys'
import { BillingExchangeProvider } from '@/features/session/billing-exchange-provider'
import { SessionReactContext } from '@/features/session/context'
import type { AppSession } from '@/features/session/types'
import type { Member } from '@/api/types'
import { createBillingExchange } from '@/lib/points'

const mockMember: Member = {
  id: 'm-admin',
  companyId: 1,
  name: '管理员',
  phone: '13800000000',
  email: 'admin@test.com',
  departmentId: 'd1',
  departmentName: '总部',
  status: 'active',
  roles: ['超级管理员'],
  source: 'manual',
}

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
