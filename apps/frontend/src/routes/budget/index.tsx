import {
  BudgetTreePanel,
  BudgetDetailTeam,
  BudgetDetailProject,
  BudgetApprovalDrawer,
  useBudgetPage,
} from '@/features/budget'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { ChevronLeft, ChevronRight, ClipboardCheck } from 'lucide-react'

export default function BudgetPage() {
  const {
    tree,
    projects,
    periodLabel,
    selectedTeamId,
    selectedNode,
    activeProject,
    approvalOpen,
    setApprovalOpen,
    pendingCount,
    approvals,
    resolveApproval,
    loading,
    error,
    refresh,
    shiftPeriod,
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
  } = useBudgetPage()

  return (
    <PageShell layout="fill">
      <DataSection
        loading={loading}
        error={error}
        onRetry={() => void refresh()}
        className="flex min-h-0 flex-1 flex-col"
        contentClassName="flex min-h-0 flex-1 flex-col"
      >
        <div className="flex min-h-0 flex-1 flex-col overflow-hidden rounded-lg border border-border bg-card shadow-xs">
          <div className="flex items-center justify-between border-b border-border px-4 py-2">
            <Button
              variant="ghost"
              size="sm"
              className="gap-1.5 text-xs"
              onClick={() => setApprovalOpen(true)}
            >
              <ClipboardCheck className="size-4" />
              审批
              {pendingCount > 0 && (
                <Badge className="ml-1 size-5 items-center justify-center rounded-full bg-red-500 p-0 text-[10px] text-white">
                  {pendingCount}
                </Badge>
              )}
            </Button>
            <div className="flex items-center gap-2">
              <Button
                variant="ghost"
                size="icon"
                className="size-7"
                aria-label="上一月"
                onClick={() => shiftPeriod(-1)}
              >
                <ChevronLeft className="size-4" />
              </Button>
              <Badge variant="outline" className="border-border text-xs tabular-nums">
                {periodLabel}
              </Badge>
              <Button
                variant="ghost"
                size="icon"
                className="size-7"
                aria-label="下一月"
                onClick={() => shiftPeriod(1)}
              >
                <ChevronRight className="size-4" />
              </Button>
            </div>
          </div>

          <div className="flex min-h-0 flex-1 overflow-hidden">
            <BudgetTreePanel tree={tree} selectedId={selectedTeamId} onSelect={handleSelectTeam} />
            <div className="flex min-h-0 flex-1 flex-col overflow-hidden">
              {activeProject && selectedNode && (
                <div className="flex items-center gap-1.5 border-b border-border px-5 py-2.5">
                  <button
                    type="button"
                    className="text-xs text-muted-foreground hover:text-foreground"
                    onClick={() => setActiveProjectId(null)}
                  >
                    {selectedNode.name}
                  </button>
                  <ChevronRight className="size-3 text-muted-foreground" />
                  <span className="text-xs font-medium text-foreground">{activeProject.name}</span>
                </div>
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
                    departmentMembers={departmentMembers}
                    membersLoading={departmentMembersLoading}
                    onUpdated={() => void refresh()}
                    onNavigateToProject={setActiveProjectId}
                    onUpdateDepartment={updateDepartment}
                    onCreateGroup={createBudgetGroup}
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

      <BudgetApprovalDrawer
        open={approvalOpen}
        onOpenChange={setApprovalOpen}
        approvals={approvals}
        onResolve={resolveApproval}
        onResolved={() => void refresh()}
      />
    </PageShell>
  )
}
