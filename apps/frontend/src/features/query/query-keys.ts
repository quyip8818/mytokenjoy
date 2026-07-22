import { approvalKeys } from '@/features/approval'
import { auditKeys } from '@/features/audit/query-keys'
import { budgetKeys } from '@/features/budget/query-keys'
import { dashboardKeys } from '@/features/dashboard/query-keys'
import { keysKeys } from '@/features/keys/query-keys'
import { memberKeys } from '@/features/member/query-keys'
import { modelsKeys } from '@/features/models/query-keys'
import { orgKeys } from '@/features/org/query-keys'
import { walletKeys } from '@/features/wallet/query-keys'

export const queryKeys = {
  session: {
    all: ['session'] as const,
    current: () => [...queryKeys.session.all, 'current'] as const,
  },
  approval: approvalKeys,
  budget: budgetKeys,
  org: orgKeys,
  keys: keysKeys,
  models: modelsKeys,
  audit: auditKeys,
  wallet: walletKeys,
  member: memberKeys,
  dashboard: dashboardKeys,
}
