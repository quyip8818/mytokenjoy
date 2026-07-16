import { useState } from 'react'
import type { BudgetNode, ProjectView } from '@/api/types'
import { toast } from 'sonner'
import { ApiError } from '@/api/client'
import { FormDialog } from '@/components/ui/form-dialog'
import { Input } from '@/components/ui/input'
import { displayToPoints, formatDisplayCurrency, pointsToDisplay } from '@/lib/points'

interface BudgetAllocationDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  node: BudgetNode
  projects: ProjectView[]
  onUpdated: () => void
  onUpdateDepartment: (
    departmentId: string,
    data: { budget: number; reservedPool?: number },
  ) => Promise<void>
}

export function BudgetAllocationDialog({
  open,
  onOpenChange,
  node,
  projects,
  onUpdated,
  onUpdateDepartment,
}: BudgetAllocationDialogProps) {
  const children = node.children ?? []
  const nodeProjects = projects.filter((p) => p.departmentId === node.id)
  const [drafts, setDrafts] = useState<Record<string, string>>({})
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  function handleOpenChange(value: boolean) {
    if (!value) {
      setDrafts({})
      setError(null)
    } else {
      const initial: Record<string, string> = {}
      for (const child of children) {
        initial[child.id] = String(pointsToDisplay(child.budget))
      }
      setDrafts(initial)
      setError(null)
    }
    onOpenChange(value)
  }

  async function handleSave() {
    setError(null)
    const updates: { id: string; budget: number }[] = []
    for (const child of children) {
      const raw = drafts[child.id] ?? String(pointsToDisplay(child.budget))
      const value = parseFloat(raw)
      if (Number.isNaN(value) || value < 0) {
        setError(`"${child.name}" 额度无效`)
        return
      }
      const budget = displayToPoints(value)
      if (budget !== child.budget) {
        updates.push({ id: child.id, budget })
      }
    }

    const newChildSum = children.reduce((sum, child) => {
      const raw = drafts[child.id] ?? String(pointsToDisplay(child.budget))
      return sum + displayToPoints(parseFloat(raw))
    }, 0)
    const projectSum = nodeProjects.reduce((sum, p) => sum + p.budget, 0)
    if (newChildSum + projectSum > node.budget) {
      setError(
        `分配总额 ${formatDisplayCurrency(newChildSum + projectSum)} 超出节点额度 ${formatDisplayCurrency(node.budget)}`,
      )
      return
    }

    setSaving(true)
    try {
      for (const { id, budget } of updates) {
        await onUpdateDepartment(id, { budget })
      }
      onUpdated()
      onOpenChange(false)
      toast.success('子节点分配已更新')
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : '保存失败，请重试')
    } finally {
      setSaving(false)
    }
  }

  return (
    <FormDialog
      open={open}
      onOpenChange={handleOpenChange}
      title="编辑子节点分配"
      error={error}
      busy={saving}
      submitLabel="保存"
      onSubmit={handleSave}
    >
      <div className="grid gap-3">
        {children.map((child) => (
          <div key={child.id} className="flex items-center gap-3">
            <span className="w-20 truncate text-sm font-medium">{child.name}</span>
            <Input
              type="number"
              min={0}
              value={drafts[child.id] ?? String(pointsToDisplay(child.budget))}
              onChange={(e) => {
                setDrafts((prev) => ({ ...prev, [child.id]: e.target.value }))
                setError(null)
              }}
              className="h-8 flex-1 tabular-nums"
              placeholder="元"
            />
            <span className="w-20 text-right text-xs tabular-nums text-muted-foreground">
              已消耗 {formatDisplayCurrency(child.consumed)}
            </span>
          </div>
        ))}
      </div>
    </FormDialog>
  )
}
