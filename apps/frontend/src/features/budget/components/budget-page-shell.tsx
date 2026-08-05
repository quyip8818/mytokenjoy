import { PageShell } from '@/components/layout/page-shell'
import { SplitPanel } from '@/components/layout/split-panel'
import { ContextHeader } from '@/components/layout/context-header'
import { DataSection } from '@/components/layout/data-section'
import { EmptyState } from '@/components/ui/empty-state'
import { Wallet } from 'lucide-react'
import type { useBudgetPage } from '@/features/budget'
import { BudgetTreePanel } from './budget-tree-panel'
import { BudgetDetailTeam } from './budget-detail-team'
import { ProjectDetail } from './project-detail'
import { BudgetPeriodBar } from './budget-period-bar'

type BudgetPageShellProps = ReturnType<typeof useBudgetPage>

export function BudgetPageShell({
  tree,
  projects,
  period,
  editable,
  canMutateEntities,
  shiftPeriod,
  selectedTeamId,
  selectedNode,
  activeProject,
  loading,
  error,
  refresh,
  handleSelectTeam,
  setActiveProjectId,
  updateDepartment,
  projectsForNode,
  departmentMembers,
  departmentMembersLoading,
  projectMembers,
  createProject,
  updateProject,
  deleteProject,
  openCreateProjectKey,
  getMemberBudgets,
  updateMemberBudget,
  applyAverageBudget,
  getDepartmentTree,
  getMembers,
  getAllDeptMembers,
  searchMembers,
}: BudgetPageShellProps) {
  const breadcrumb =
    activeProject && selectedNode
      ? [selectedNode.name, activeProject.name]
      : selectedNode
        ? [selectedNode.name]
        : undefined

  return (
    <PageShell className="flex min-h-0 flex-1 flex-col">
      <DataSection
        loading={loading}
        error={error}
        onRetry={() => void refresh()}
        loadingVariant="spinner"
      >
        <SplitPanel
          master={
            <BudgetTreePanel tree={tree} selectedId={selectedTeamId} onSelect={handleSelectTeam} />
          }
          detail={
            <>
              <BudgetPeriodBar
                period={period}
                onShiftPeriod={shiftPeriod}
                onPreConfigNext={() => shiftPeriod(1)}
              />
              {breadcrumb && (
                <ContextHeader
                  breadcrumb={breadcrumb}
                  actions={
                    activeProject ? (
                      <button
                        type="button"
                        className="text-xs text-muted-foreground hover:text-foreground"
                        onClick={() => setActiveProjectId(null)}
                      >
                        返回
                      </button>
                    ) : undefined
                  }
                />
              )}
              <div className="min-h-0 flex-1 overflow-y-auto">
                {activeProject ? (
                  <ProjectDetail
                    project={activeProject}
                    editable={editable}
                    canMutateEntities={canMutateEntities}
                    members={projectMembers}
                    departmentMembers={departmentMembers}
                    membersLoading={departmentMembersLoading}
                    onUpdateProject={updateProject}
                    onDeleteProject={deleteProject}
                    onCreateProjectKey={openCreateProjectKey}
                    onUpdated={() => void refresh()}
                    onDeleted={() => {
                      setActiveProjectId(null)
                    }}
                  />
                ) : selectedNode ? (
                  <BudgetDetailTeam
                    node={selectedNode}
                    editable={editable}
                    canMutateEntities={canMutateEntities}
                    projects={projects.filter((project) =>
                      projectsForNode(selectedNode.id).some((group) => group.id === project.id),
                    )}
                    onUpdated={() => void refresh()}
                    onNavigateToProject={setActiveProjectId}
                    onUpdateDepartment={updateDepartment}
                    onCreateProject={createProject}
                    getMemberBudgets={getMemberBudgets}
                    updateMemberBudget={updateMemberBudget}
                    applyAverageBudget={applyAverageBudget}
                    getDepartmentTree={getDepartmentTree}
                    getMembers={getMembers}
                    getAllDeptMembers={getAllDeptMembers}
                    searchMembers={searchMembers}
                  />
                ) : (
                  <EmptyState variant="minimal" icon={Wallet} title="选择左侧节点查看预算详情" />
                )}
              </div>
            </>
          }
        />
      </DataSection>
    </PageShell>
  )
}
