import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import type { useBudgetPage } from '@/features/budget'
import { BudgetTreePanel } from './budget-tree-panel'
import { BudgetDetailTeam } from './budget-detail-team'
import { BudgetDetailProject } from './budget-detail-project'
import { ChevronRight } from 'lucide-react'

type BudgetPageShellProps = ReturnType<typeof useBudgetPage>

export function BudgetPageShell({
  tree,
  projects,
  selectedTeamId,
  selectedNode,
  activeProject,
  loading,
  error,
  refresh,
  handleSelectTeam,
  setActiveProjectId,
  updateDepartment,
  groupsForNode,
  overrunPolicyLabel,
  departmentMembers,
  departmentMembersLoading,
  projectMembers,
  createBudgetGroup,
  updateBudgetGroup,
  deleteBudgetGroup,
  getMemberBudgets,
  updateMemberBudget,
  applyAverageBudget,
  getDepartmentTree,
  getMembers,
  getAllDeptMembers,
  searchMembers,
}: BudgetPageShellProps) {
  return (
    <PageShell layout="fill" className="min-h-0 flex-1">
      <DataSection
        loading={loading}
        error={error}
        onRetry={() => void refresh()}
        className="flex min-h-0 flex-1 flex-col"
        contentClassName="flex min-h-0 flex-1 flex-col"
      >
        <div className="flex min-h-0 flex-1 flex-col overflow-hidden rounded-lg border border-border bg-card shadow-xs">
          <div className="flex min-h-0 flex-1 overflow-hidden">
            <BudgetTreePanel tree={tree} selectedId={selectedTeamId} onSelect={handleSelectTeam} />
            <div className="flex min-h-0 flex-1 flex-col overflow-hidden">
              {activeProject && selectedNode && (
                <nav
                  aria-label="breadcrumb"
                  className="flex items-center gap-1.5 border-b border-border px-5 py-2.5"
                >
                  <button
                    type="button"
                    className="text-xs text-muted-foreground hover:text-foreground"
                    onClick={() => setActiveProjectId(null)}
                  >
                    {selectedNode.name}
                  </button>
                  <ChevronRight className="size-3 text-muted-foreground" aria-hidden="true" />
                  <span className="text-xs font-medium text-foreground" aria-current="page">
                    {activeProject.name}
                  </span>
                </nav>
              )}

              <div className="min-h-0 flex-1 overflow-y-auto">
                {activeProject ? (
                  <BudgetDetailProject
                    project={activeProject}
                    members={projectMembers}
                    departmentMembers={departmentMembers}
                    membersLoading={departmentMembersLoading}
                    onUpdateGroup={updateBudgetGroup}
                    onDeleteGroup={deleteBudgetGroup}
                    onUpdated={() => void refresh()}
                    onDeleted={() => {
                      setActiveProjectId(null)
                    }}
                  />
                ) : selectedNode ? (
                  <BudgetDetailTeam
                    node={selectedNode}
                    projects={projects.filter((project) =>
                      groupsForNode(selectedNode.id).some((group) => group.id === project.id),
                    )}
                    overrunPolicyLabel={overrunPolicyLabel}
                    onUpdated={() => void refresh()}
                    onNavigateToProject={setActiveProjectId}
                    onUpdateDepartment={updateDepartment}
                    onCreateGroup={createBudgetGroup}
                    getMemberBudgets={getMemberBudgets}
                    updateMemberBudget={updateMemberBudget}
                    applyAverageBudget={applyAverageBudget}
                    getDepartmentTree={getDepartmentTree}
                    getMembers={getMembers}
                    getAllDeptMembers={getAllDeptMembers}
                    searchMembers={searchMembers}
                  />
                ) : (
                  <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
                    选择左侧节点查看预算详情
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>
      </DataSection>
    </PageShell>
  )
}
