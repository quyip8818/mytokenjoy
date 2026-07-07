import { useState } from 'react'
import type { BudgetNode, BudgetProjectView } from '@/api/types'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { nodeReservedPool } from '@/features/budget/lib/mappers'
import { cn } from '@/lib/utils'
import { Pencil, Users, Wallet, X, Check } from 'lucide-react'

interface BudgetEditMemberQuotaProps {
  node: BudgetNode
  projects: BudgetProjectView[]
  onUpdated: () => void
  onUpdateDepartment: (
    departmentId: string,
    data: { budget: number; reservedPool?: number },
  ) => Promise<void>
}

export function BudgetEditMemberQuota({
  node,
  projects,
  onUpdated,
  onUpdateDepartment,
}: BudgetEditMemberQuotaProps) {
  const [editing, setEditing] = useState(false)
  const [reservedDraft, setReservedDraft] = useState('')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  function startEdit() {
    setReservedDraft(String(nodeReservedPool(node)))
    setError(null)
    setEditing(true)
  }

  function cancelEdit() {
    setError(null)
    setEditing(false)
  }

  function computeAllocated(newReserved: number) {
    const childSum = node.children?.reduce((sum, child) => sum + child.budget, 0) ?? 0
    const projectSum = projects
      .filter((project) => project.departmentId === node.id)
      .reduce((sum, project) => sum + project.budget, 0)
    return childSum + projectSum + newReserved
  }

  async function handleSave() {
    const reserved = parseFloat(reservedDraft)
    if (Number.isNaN(reserved) || reserved < 0) {
      setError('预留池余额无效')
      return
    }
    const allocated = computeAllocated(reserved)
    if (allocated > node.budget) {
      setError(
        `分配总额 ¥${allocated.toLocaleString()} 超出节点额度 ¥${node.budget.toLocaleString()}`,
      )
      return
    }
    setSaving(true)
    try {
      await onUpdateDepartment(node.id, { budget: node.budget, reservedPool: reserved })
      setEditing(false)
      onUpdated()
    } catch {
      setError('保存失败，请重试')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="rounded-lg border border-border p-4">
      <div className="mb-3 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Users className="size-4 text-muted-foreground" />
          <h4 className="text-sm font-semibold text-foreground">成员额度</h4>
        </div>
        {!editing ? (
          <Button
            variant="ghost"
            size="sm"
            className="h-7 gap-1.5 text-xs text-muted-foreground"
            onClick={startEdit}
          >
            <Pencil className="size-3.5" />
            编辑配置
          </Button>
        ) : (
          <div className="flex items-center gap-1">
            <Button
              variant="ghost"
              size="sm"
              className="h-7 gap-1.5 text-xs text-muted-foreground"
              onClick={cancelEdit}
              disabled={saving}
              aria-label="取消编辑"
            >
              <X className="size-3.5" />
              取消
            </Button>
            <Button
              size="sm"
              className="h-7 gap-1.5 text-xs"
              onClick={handleSave}
              disabled={saving}
              aria-label="保存配置"
            >
              <Check className="size-3.5" />
              保存
            </Button>
          </div>
        )}
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="flex items-center gap-3 rounded-md bg-muted/50 px-3 py-2.5">
          <Wallet className="size-4 shrink-0 text-muted-foreground" />
          <div className="min-w-0 flex-1">
            <p className="text-xs text-muted-foreground">平均额度/人</p>
            <p className="text-sm font-medium tabular-nums text-muted-foreground">—</p>
          </div>
        </div>

        <div
          className={cn(
            'flex items-center gap-3 rounded-md bg-muted/50 px-3 py-2.5',
            editing && 'items-start pt-3',
          )}
        >
          <Users className="mt-0.5 size-4 shrink-0 text-muted-foreground" />
          <div className="min-w-0 flex-1">
            <p className="text-xs text-muted-foreground">预留池余额</p>
            {editing ? (
              <Input
                type="number"
                min={0}
                value={reservedDraft}
                onChange={(event) => {
                  setReservedDraft(event.target.value)
                  setError(null)
                }}
                className="mt-1 h-7 tabular-nums"
                placeholder="元"
              />
            ) : (
              <p className="text-sm font-medium tabular-nums">
                ¥{nodeReservedPool(node).toLocaleString()}
              </p>
            )}
          </div>
        </div>
      </div>

      {error && <p className="mt-2 text-xs text-red-600">{error}</p>}

      <p className="mt-2 text-xs text-muted-foreground">
        成员个人额度请在成员配额配置中管理；预留池余额可在上方编辑。
      </p>
    </div>
  )
}
