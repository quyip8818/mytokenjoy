import type { BudgetNode, BudgetProjectView } from '@/api/types'
import { Button } from '@/components/ui/button'
import { Pencil, X, Check } from 'lucide-react'
import { useBudgetAllocationEdit } from '../hooks/use-budget-allocation-edit'
import { BudgetAllocationTable } from './budget-allocation-table'

interface BudgetEditAllocationProps {
  node: BudgetNode
  projects: BudgetProjectView[]
  overrunPolicyLabel: string
  onUpdated: () => void
  onUpdateDepartment: (
    departmentId: string,
    data: { budget: number; reservedPool?: number },
  ) => Promise<void>
}

export function BudgetEditAllocation({
  node,
  projects,
  overrunPolicyLabel,
  onUpdated,
  onUpdateDepartment,
}: BudgetEditAllocationProps) {
  const allocation = useBudgetAllocationEdit({
    node,
    projects,
    onUpdated,
    onUpdateDepartment,
  })

  if (allocation.children.length === 0 && allocation.nodeProjects.length === 0) return null

  return (
    <div>
      <div className="mb-3 flex items-center justify-between">
        <h4 className="text-sm font-semibold text-foreground">子节点分配</h4>
        {!allocation.editing ? (
          <Button
            variant="ghost"
            size="sm"
            className="h-7 gap-1.5 text-xs text-muted-foreground"
            onClick={allocation.startEdit}
          >
            <Pencil className="size-3.5" />
            编辑
          </Button>
        ) : (
          <div className="flex items-center gap-1">
            <Button
              variant="ghost"
              size="sm"
              className="h-7 gap-1.5 text-xs text-muted-foreground"
              onClick={allocation.cancelEdit}
              disabled={allocation.saving}
              aria-label="取消编辑"
            >
              <X className="size-3.5" />
              取消
            </Button>
            <Button
              size="sm"
              className="h-7 gap-1.5 text-xs"
              onClick={allocation.handleSave}
              disabled={allocation.saving}
              aria-label="保存分配"
            >
              <Check className="size-3.5" />
              保存
            </Button>
          </div>
        )}
      </div>

      <BudgetAllocationTable
        node={node}
        children={allocation.children}
        nodeProjects={allocation.nodeProjects}
        overrunPolicyLabel={overrunPolicyLabel}
        editing={allocation.editing}
        drafts={allocation.drafts}
        reservedDraft={allocation.reservedDraft}
        error={allocation.error}
        onUpdateDraft={allocation.updateDraft}
        onUpdateReservedDraft={allocation.updateReservedDraft}
      />
    </div>
  )
}
