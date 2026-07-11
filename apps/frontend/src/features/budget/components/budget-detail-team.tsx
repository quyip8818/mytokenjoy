import { useState } from 'react'
import type { BudgetNode, BudgetProjectView, Member, MemberBudgetQuota, UpdateMemberBudgetInput } from '@/api/types'
import { Progress } from '@/components/ui/progress'
import { Button } from '@/components/ui/button'
import { BudgetEditAllocation } from './budget-edit-allocation'
import { BudgetEditMemberBudget } from './budget-edit-member-budget'
import { BudgetProjectDialog } from './budget-project-dialog'
import { nodeReservedPool } from '@/features/budget'
import { formatDisplayCurrency } from '@/lib/points'
import { cn } from '@/lib/utils'
import { Plus, ChevronRight } from 'lucide-react'

interface BudgetDetailTeamProps {
  node: BudgetNode
  projects: BudgetProjectView[]
  overrunPolicyLabel: string
  departmentMembers: Member[]
  membersLoading?: boolean
  onUpdated: () => void
  onNavigateToProject: (projectId: string) => void
  onUpdateDepartment: (
    departmentId: string,
    data: { budget: number; reservedPool?: number },
  ) => Promise<void>
  onCreateGroup: (data: {
    name: string
    budget: number
    memberIds: string[]
    departmentIds: string[]
  }) => Promise<void>
  getMemberBudgets: (departmentId: string) => Promise<MemberBudgetQuota[]>
  updateMemberBudget: (memberId: string, data: UpdateMemberBudgetInput) => Promise<MemberBudgetQuota>
}

function SummaryCard({
  label,
  value,
  muted,
  highlight,
}: {
  label: string
  value: number
  muted?: boolean
  highlight?: boolean
}) {
  return (
    <div className="rounded-lg border border-border p-3">
      <p className="text-xs text-muted-foreground">{label}</p>
      <p
        className={
          highlight
            ? 'mt-1 text-lg font-semibold tabular-nums text-red-600'
            : muted
              ? 'mt-1 text-lg font-semibold tabular-nums text-muted-foreground'
              : 'mt-1 text-lg font-semibold tabular-nums text-foreground'
        }
      >
        {formatDisplayCurrency(value)}
      </p>
    </div>
  )
}

export function BudgetDetailTeam({
  node,
  projects,
  overrunPolicyLabel,
  departmentMembers,
  membersLoading = false,
  onUpdated,
  onNavigateToProject,
  onUpdateDepartment,
  onCreateGroup,
  getMemberBudgets,
  updateMemberBudget,
}: BudgetDetailTeamProps) {
  const [createDialogOpen, setCreateDialogOpen] = useState(false)

  const nodeProjects = projects.filter((project) => project.departmentId === node.id)
  const childrenBudgetSum = node.children?.reduce((sum, child) => sum + child.budget, 0) ?? 0
  const projectBudgetSum = nodeProjects.reduce((sum, project) => sum + project.budget, 0)
  const reservedPool = nodeReservedPool(node)
  const allocated = childrenBudgetSum + projectBudgetSum + reservedPool
  const unallocated = node.budget - allocated
  const pct = node.budget > 0 ? Math.round((node.consumed / node.budget) * 100) : 0

  return (
    <div className="flex flex-1 flex-col gap-6 overflow-y-auto p-5">
      <div className="flex items-center gap-3">
        <h3 className="text-sm font-semibold text-foreground">{node.name}</h3>
      </div>

      <div className="grid grid-cols-4 gap-4">
        <SummaryCard label="总额度" value={node.budget} />
        <SummaryCard label="已分配" value={allocated} muted />
        <SummaryCard label="预留池" value={reservedPool} />
        <SummaryCard label="未分配" value={unallocated} highlight={unallocated < 0} />
      </div>

      <div className="rounded-lg border border-border p-4">
        <div className="mb-2 flex items-center justify-between text-xs">
          <span className="text-muted-foreground">本月消耗</span>
          <span className="font-medium tabular-nums">
            {formatDisplayCurrency(node.consumed)} / {formatDisplayCurrency(node.budget)}
          </span>
        </div>
        <Progress value={pct} className="h-2" />
        <p className="mt-1.5 text-xs text-muted-foreground">
          已使用 {pct}%，剩余 {formatDisplayCurrency(node.budget - node.consumed)}
        </p>
      </div>

      <BudgetEditAllocation
        node={node}
        projects={projects}
        overrunPolicyLabel={overrunPolicyLabel}
        onUpdated={onUpdated}
        onUpdateDepartment={onUpdateDepartment}
      />

      <div>
        <div className="mb-3 flex items-center justify-between">
          <h4 className="text-sm font-semibold text-foreground">项目预算</h4>
          <Button
            variant="ghost"
            size="sm"
            aria-label="创建项目"
            onClick={() => setCreateDialogOpen(true)}
          >
            <Plus className="h-4 w-4" />
            创建项目
          </Button>
        </div>

        <div className="divide-y divide-border rounded-lg border border-border">
          {nodeProjects.length === 0 ? (
            <div className="flex flex-col items-center gap-3 px-4 py-8 text-center">
              <p className="text-sm text-muted-foreground">暂无项目</p>
              <Button variant="outline" size="sm" onClick={() => setCreateDialogOpen(true)}>
                <Plus className="h-4 w-4" />
                创建第一个项目
              </Button>
            </div>
          ) : (
            nodeProjects.map((project) => {
              const projectPct =
                project.budget > 0 ? Math.round((project.consumed / project.budget) * 100) : 0
              return (
                <div
                  key={project.id}
                  role="button"
                  tabIndex={0}
                  className={cn(
                    'flex cursor-pointer items-center gap-3 px-4 py-3 hover:bg-muted/50',
                    'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring',
                  )}
                  onClick={() => onNavigateToProject(project.id)}
                  onKeyDown={(event) => {
                    if (event.key === 'Enter' || event.key === ' ') {
                      event.preventDefault()
                      onNavigateToProject(project.id)
                    }
                  }}
                >
                  <span className="flex-1 text-sm font-medium text-foreground">{project.name}</span>
                  <span className="text-xs tabular-nums text-muted-foreground">
                    {formatDisplayCurrency(project.budget)} /{' '}
                    {formatDisplayCurrency(project.consumed)}
                  </span>
                  <div className="w-24">
                    <Progress value={projectPct} className="h-1.5" />
                  </div>
                  <span className="w-8 text-right text-xs tabular-nums text-muted-foreground">
                    {projectPct}%
                  </span>
                  <ChevronRight className="size-4 text-muted-foreground" />
                </div>
              )
            })
          )}
        </div>
      </div>

      <BudgetEditMemberBudget
        node={node}
        projects={projects}
        onUpdated={onUpdated}
        onUpdateDepartment={onUpdateDepartment}
        getMemberBudgets={getMemberBudgets}
        updateMemberBudget={updateMemberBudget}
      />

      <BudgetProjectDialog
        open={createDialogOpen}
        onOpenChange={setCreateDialogOpen}
        department={node}
        departmentMembers={departmentMembers}
        membersLoading={membersLoading}
        onCreateGroup={onCreateGroup}
      />
    </div>
  )
}
