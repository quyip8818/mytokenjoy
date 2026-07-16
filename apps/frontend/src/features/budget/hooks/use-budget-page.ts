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
    refreshApprovals: queries.refreshApprovals,
  })

  return {
    // --- Queries ---
    tree: queries.tree,
    projects: queries.projects,
    period: queries.period,
    periodLabel: queries.periodLabel,
    overrunPolicy: queries.overrunPolicy,
    approvals: queries.approvals,
    pendingCount: queries.pendingCount,
    loading: queries.loading,
    error: queries.error,
    refresh: queries.refresh,
    refreshApprovals: queries.refreshApprovals,
    shiftPeriod: queries.shiftPeriod,

    // --- Selection / UI State ---
    selectedTeamId: selection.selectedTeamId,
    selectedNode: selection.selectedNode,
    activeProjectId: selection.activeProjectId,
    activeProject: selection.activeProject,
    approvalOpen: selection.approvalOpen,
    setApprovalOpen: selection.setApprovalOpen,
    handleSelectTeam: selection.handleSelectTeam,
    setActiveProjectId: selection.setActiveProjectId,
    projectsForNode: selection.projectsForNode,
    departmentMembers: selection.departmentMembers,
    departmentMembersLoading: selection.departmentMembersLoading,
    projectMembers: selection.projectMembers,

    // --- Actions ---
    updateDepartment: actions.updateDepartment,
    resolveApproval: actions.resolveApproval,
    createProject: actions.createProject,
    updateProject: actions.updateProject,
    deleteProject: actions.deleteProject,
    openCreateProjectKey: actions.openCreateProjectKey,
    getMemberBudgets: actions.getMemberBudgets,
    updateMemberBudget: actions.updateMemberBudget,
    applyAverageBudget: actions.applyAverageBudget,
    getDepartmentTree: actions.getDepartmentTree,
    getMembers: actions.getMembers,
    getAllDeptMembers: actions.getAllDeptMembers,
    searchMembers: actions.searchMembers,
  }
}
