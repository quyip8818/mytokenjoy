import { useState } from 'react'
import type { BudgetNode, ProjectView } from '@/api/types'
import { Button } from '@/components/ui/button'
import { Pencil } from 'lucide-react'
import { BudgetAllocationTable } from './budget-allocation-table'
import { BudgetAllocationDialog } from './budget-allocation-dialog'

interface BudgetEditAllocationProps {
  node: BudgetNode
  projects: ProjectView[]
  onUpdated: () => void
  onUpdateDepartment: (
    departmentId: string,
    data: { budget: number; reservedPool?: number },
  ) => Promise<void>
}

export function BudgetEditAllocation({
  node,
  projects,
  onUpdated,
  onUpdateDepartment,
}: BudgetEditAllocationProps) {
  const [dialogOpen, setDialogOpen] = useState(false)

  const children = node.children ?? []
  const nodeProjects = projects.filter((p) => p.departmentId === node.id)

  if (children.length === 0 && nodeProjects.length === 0) return null

  return (
    <div>
      <div className="mb-3 flex items-center justify-between">
        <h4 className="text-sm font-semibold text-foreground">子节点分配</h4>
        <Button
          variant="ghost"
          size="sm"
          className="h-7 gap-1.5 text-xs text-muted-foreground"
          onClick={() => setDialogOpen(true)}
        >
          <Pencil className="size-3.5" />
          编辑
        </Button>
      </div>

      <BudgetAllocationTable
        node={node}
        children={children}
        nodeProjects={nodeProjects}
        editing={false}
        drafts={{}}
        onUpdateDraft={() => {}}
      />

      <BudgetAllocationDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        node={node}
        projects={projects}
        onUpdated={onUpdated}
        onUpdateDepartment={onUpdateDepartment}
      />
    </div>
  )
}
