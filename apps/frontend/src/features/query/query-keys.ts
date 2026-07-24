import { approvalKeys } from '@/features/approval'
import { auditKeys } from '@/features/audit/query-keys'
import { budgetKeys } from '@/features/budget/query-keys'
import { dashboardKeys } from '@/features/dashboard/query-keys'
import { keysKeys } from '@/features/keys/query-keys'
import { mydashboardKeys } from '@/features/mydashboard/query-keys'
import { modelsKeys } from '@/features/models/query-keys'
import { orgKeys } from '@/features/org/query-keys'
import { billingKeys } from '@/features/billing/query-keys'

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
  billing: billingKeys,
  mydashboard: mydashboardKeys,
  dashboard: dashboardKeys,
}
