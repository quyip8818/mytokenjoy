import type { AppApis } from '@/api/app-apis'
import { useBudgetQueries } from './use-budget-queries'
import { useBudgetSelection } from './use-budget-selection'
import { useBudgetActions } from './use-budget-actions'

/**
 * Page-level orchestrator for the Budget page.
 * Composes query, selection/UI-state, and action hooks into a single return value
 * consumed by BudgetPageShell.
 */
export function useBudgetPage(injectedApis?: AppApis) {
  const queries = useBudgetQueries(injectedApis)

  const selection = useBudgetSelection({
    injectedApis,
    tree: queries.tree,
    projects: queries.projects,
    projectsData: queries.projectsData,
  })

  const actions = useBudgetActions({
    injectedApis,
    refresh: queries.refresh,
  })

  return { ...queries, ...selection, ...actions }
}
