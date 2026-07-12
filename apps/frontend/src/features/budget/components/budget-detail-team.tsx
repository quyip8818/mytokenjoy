import { useEffect, useState } from 'react'
import type {
  BudgetNode,
  ProjectView,
  Department,
  Member,
  MemberBudget,
  UpdateMemberBudgetInput,
} from '@/api/types'
import { Progress } from '@/components/ui/progress'
import { Button } from '@/components/ui/button'
import { BudgetEditAllocation } from './budget-edit-allocation'
import { BudgetEditMemberBudget } from './budget-edit-member-budget'
import { ProjectDialog } from './project-dialog'
import { BudgetInitPrompt } from './budget-init-prompt'
import { formatDisplayCurrency } from '@/lib/points'
import { cn } from '@/lib/utils'
import { Plus, ChevronRight } from 'lucide-react'

interface BudgetDetailTeamProps {
  node: BudgetNode
  projects: ProjectView[]
  overrunPolicyLabel: string
  onUpdated: () => void
  onNavigateToProject: (projectId: string) => void
  onUpdateDepartment: (
    departmentId: string,
    data: { budget: number; reservedPool?: number },
  ) => Promise<void>
  onCreateProject: (data: {
    name: string
    budget: number
    memberIds: string[]
    ownerDepartmentId: string
  }) => Promise<void>
  getMemberBudgets: (departmentId: string) => Promise<MemberBudget[]>
  updateMemberBudget: (memberId: string, data: UpdateMemberBudgetInput) => Promise<MemberBudget>
  applyAverageBudget: (
    departmentId: string,
    data: { personalBudget: number; recursive: boolean },
  ) => Promise<void>
  getDepartmentTree: () => Promise<Department[]>
  getMembers: (departmentId: string) => Promise<Member[]>
  getAllDeptMembers: (departmentId: string) => Promise<Member[]>
  searchMembers: (keyword: string) => Promise<Member[]>
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
  onUpdated,
  onNavigateToProject,
  onUpdateDepartment,
  onCreateProject,
  getMemberBudgets,
  updateMemberBudget,
  applyAverageBudget,
  getDepartmentTree,
  getMembers,
  getAllDeptMembers,
  searchMembers,
}: BudgetDetailTeamProps) {
  const [createDialogOpen, setCreateDialogOpen] = useState(false)
  const [memberRefreshKey, setMemberRefreshKey] = useState(0)

  // Load member budget sum for this department (must be before any early return)
  const [memberBudgetSum, setMemberBudgetSum] = useState(0)
  useEffect(() => {
    if (node.budget === 0) return
    getMemberBudgets(node.id).then((members) => {
      setMemberBudgetSum(members.reduce((sum, m) => sum + m.personalBudget, 0))
    })
  }, [node.id, node.budget, memberRefreshKey, getMemberBudgets])

  // Show initialization prompt if budget is not set
  if (node.budget === 0) {
    const isRoot = node.parentId === null || node.parentId === undefined
    return (
      <div className="flex flex-1 flex-col gap-6 overflow-y-auto p-5">
        <h3 className="text-sm font-semibold text-foreground">{node.name}</h3>
        {isRoot ? (
          <BudgetInitPrompt
            departmentId={node.id}
            departmentName={node.name}
            onUpdateDepartment={onUpdateDepartment}
          />
        ) : (
          <div className="flex flex-col items-center gap-3 rounded-lg border border-dashed border-border p-8 text-center">
            <p className="text-sm font-medium text-foreground">当前部门尚未分配额度</p>
            <p className="text-xs text-muted-foreground">请在上级部门中为该部门分配预算额度</p>
          </div>
        )}
      </div>
    )
  }

  const nodeProjects = projects.filter((project) => project.departmentId === node.id)
  const childrenBudgetSum = node.children?.reduce((sum, child) => sum + child.budget, 0) ?? 0
  const projectBudgetSum = nodeProjects.reduce((sum, project) => sum + project.budget, 0)

  const allocated = childrenBudgetSum + projectBudgetSum + memberBudgetSum
  const reservedPool = node.budget - allocated
  const pct = node.budget > 0 ? Math.round((node.consumed / node.budget) * 100) : 0

  return (
    <div className="flex flex-1 flex-col gap-6 overflow-y-auto p-5">
      <div className="flex items-center gap-3">
        <h3 className="text-sm font-semibold text-foreground">{node.name}</h3>
      </div>

      <div className="grid grid-cols-3 gap-4">
        <SummaryCard label="总额度" value={node.budget} />
        <SummaryCard label="已分配" value={allocated} muted />
        <SummaryCard label="预留池" value={reservedPool} />
      </div>

      <div className="rounded-lg border border-border p-4">
        <div className="mb-2 flex items-center justify-between text-xs">
          <span className="text-muted-foreground">本月消耗</span>
          <span className="font-medium tabular-nums">
            {formatDisplayCurrency(node.consumed)} / {formatDisplayCurrency(node.budget)}
          </span>
        </div>
        <Progress value={pct} className="h-2" aria-label="预算使用进度" />
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
        onUpdated={() => {
          onUpdated()
          setMemberRefreshKey((k) => k + 1)
        }}
        getMemberBudgets={getMemberBudgets}
        updateMemberBudget={updateMemberBudget}
        applyAverageBudget={applyAverageBudget}
      />

      <ProjectDialog
        open={createDialogOpen}
        onOpenChange={setCreateDialogOpen}
        department={node}
        existingProjectsBudget={projectBudgetSum}
        memberBudgetSum={memberBudgetSum}
        onCreateProject={onCreateProject}
        getDepartmentTree={getDepartmentTree}
        getMembers={getMembers}
        getAllDeptMembers={getAllDeptMembers}
        searchMembers={searchMembers}
      />
    </div>
  )
}
