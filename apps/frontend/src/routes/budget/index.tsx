import { useState, useEffect, useMemo } from 'react'
import { budgetApi } from '@/api/budget'
import type { BudgetNode, BudgetProject } from '@/api/types'
import { BudgetTreePanel } from '@/components/budget/budget-tree-panel'
import { BudgetDetailTeam } from '@/components/budget/budget-detail-team'
import { BudgetDetailProject } from '@/components/budget/budget-detail-project'
import { BudgetApprovalDrawer } from '@/components/budget/budget-approval-drawer'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { ChevronLeft, ChevronRight, ClipboardCheck } from 'lucide-react'

function formatPeriod(period: string) {
  const [year, month] = period.split('-')
  return `${year} 年 ${parseInt(month)} 月`
}

function shiftPeriod(period: string, delta: number) {
  const [y, m] = period.split('-').map(Number)
  const d = new Date(y, m - 1 + delta, 1)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`
}

function findNode(nodes: BudgetNode[], id: string): BudgetNode | undefined {
  for (const n of nodes) {
    if (n.id === id) return n
    if (n.children) {
      const found = findNode(n.children, id)
      if (found) return found
    }
  }
  return undefined
}

export default function BudgetPage() {
  const [tree, setTree] = useState<BudgetNode[]>([])
  const [projects, setProjects] = useState<BudgetProject[]>([])
  const [period, setPeriod] = useState('2026-06')
  const [selectedTeamId, setSelectedTeamId] = useState<string | undefined>()
  const [activeProjectId, setActiveProjectId] = useState<string | null>(null)
  const [approvalOpen, setApprovalOpen] = useState(false)
  const [pendingCount, setPendingCount] = useState(0)

  const reload = () => {
    budgetApi.getTree(period).then((data) => {
      setTree(data)
      if (!selectedTeamId && data.length > 0) setSelectedTeamId(data[0].id)
    })
    budgetApi.getProjects(period).then(setProjects)
    budgetApi.getApprovals().then((list) => {
      setPendingCount(list.filter((a) => a.status === 'pending').length)
    })
  }

  useEffect(() => {
    reload()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [period])

  const selectedNode = useMemo(
    () => (selectedTeamId ? findNode(tree, selectedTeamId) : undefined),
    [tree, selectedTeamId]
  )

  const activeProject = useMemo(
    () => (activeProjectId ? projects.find((p) => p.id === activeProjectId) : undefined),
    [projects, activeProjectId]
  )

  function handleSelectTeam(nodeId: string) {
    setSelectedTeamId(nodeId)
    setActiveProjectId(null)
  }

  return (
    <div className="flex h-[calc(100dvh-7.5rem)] flex-col">
      <div className="flex flex-1 flex-col overflow-hidden rounded-lg border border-border bg-card shadow-xs">
        {/* Toolbar */}
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
              onClick={() => setPeriod(shiftPeriod(period, -1))}
            >
              <ChevronLeft className="size-4" />
            </Button>
            <Badge variant="outline" className="border-border text-xs tabular-nums">
              {formatPeriod(period)}
            </Badge>
            <Button
              variant="ghost"
              size="icon"
              className="size-7"
              aria-label="下一月"
              onClick={() => setPeriod(shiftPeriod(period, 1))}
            >
              <ChevronRight className="size-4" />
            </Button>
          </div>
        </div>

        {/* Content */}
        <div className="flex flex-1 overflow-hidden">
          <BudgetTreePanel
            tree={tree}
            selectedId={selectedTeamId}
            onSelect={handleSelectTeam}
          />
          <div className="flex flex-1 flex-col overflow-hidden">
            {/* Breadcrumb when drilled into a project */}
            {activeProject && selectedNode && (
              <div className="flex items-center gap-1.5 border-b border-border px-5 py-2.5">
                <button
                  className="text-xs text-muted-foreground hover:text-foreground"
                  onClick={() => setActiveProjectId(null)}
                >
                  {selectedNode.name}
                </button>
                <ChevronRight className="size-3 text-muted-foreground" />
                <span className="text-xs font-medium text-foreground">{activeProject.name}</span>
              </div>
            )}

            <div className="flex-1 overflow-y-auto">
              {activeProject ? (
                <BudgetDetailProject
                  project={activeProject}
                  onUpdated={reload}
                  onDeleted={() => {
                    setActiveProjectId(null)
                    reload()
                  }}
                />
              ) : selectedNode ? (
                <BudgetDetailTeam
                  node={selectedNode}
                  projects={projects}
                  onUpdated={reload}
                  onNavigateToProject={setActiveProjectId}
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

      <BudgetApprovalDrawer
        open={approvalOpen}
        onOpenChange={setApprovalOpen}
        onResolved={reload}
      />
    </div>
  )
}
