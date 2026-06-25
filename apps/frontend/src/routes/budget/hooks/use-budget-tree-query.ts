import type { AppApis } from '@/api/app-apis'
import { queryKeys } from '@/features/query/query-keys'
import { useInjectedQuery } from '@/features/query/use-injected-query'

export function useBudgetTreeQuery(injectedApis?: AppApis) {
  return useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.budget.tree(),
    queryFn: (apis) => apis.budgetApi.getTree(),
  })
}
