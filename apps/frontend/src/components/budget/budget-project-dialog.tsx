import { useState, useMemo } from 'react'
import { budgetApi } from '@/api/budget'
import type { BudgetNode } from '@/api/types'
import { BudgetMemberPicker } from './budget-member-picker'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

interface BudgetProjectDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  department: BudgetNode
  onCreated: () => void
}

export function BudgetProjectDialog({
  open,
  onOpenChange,
  department,
  onCreated,
}: BudgetProjectDialogProps) {
  const [name, setName] = useState('')
  const [budget, setBudget] = useState('')
  const [memberIds, setMemberIds] = useState<string[]>([])
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const available = useMemo(() => {
    const childrenSum = department.children?.reduce((s: number, c: BudgetNode) => s + c.budget, 0) ?? 0
    return department.budget - childrenSum - department.reserved
  }, [department])

  function resetForm() {
    setName('')
    setBudget('')
    setMemberIds([])
    setError(null)
  }

  function handleOpenChange(v: boolean) {
    if (!v) resetForm()
    onOpenChange(v)
  }

  async function handleCreate() {
    setError(null)

    const trimmedName = name.trim()
    if (!trimmedName) {
      setError('请输入项目名称')
      return
    }
    const budgetNum = parseFloat(budget)
    if (!budget || isNaN(budgetNum) || budgetNum < 0) {
      setError('请输入有效的项目额度')
      return
    }
    if (budgetNum > available) {
      setError(`团队可用额度为 ¥${available.toLocaleString()}，请调低项目额度`)
      return
    }

    setSaving(true)
    try {
      await budgetApi.createProject({
        name: trimmedName,
        departmentId: department.id,
        departmentName: department.name,
        budget: budgetNum,
        memberIds,
        overrunPolicy: 'hard_reject',
        period: department.period,
      })
      resetForm()
      onCreated()
      onOpenChange(false)
    } catch {
      setError('创建失败，请重试')
    } finally {
      setSaving(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>创建项目</DialogTitle>
        </DialogHeader>

        <div className="grid gap-4 py-2">
          {/* Project name */}
          <div className="grid gap-1.5">
            <Label htmlFor="proj-name" className="text-xs font-medium">
              项目名称
            </Label>
            <Input
              id="proj-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="输入项目名称"
              className="h-8 text-sm"
            />
          </div>

          {/* Budget */}
          <div className="grid gap-1.5">
            <Label htmlFor="proj-budget" className="text-xs font-medium">
              项目额度（元）
            </Label>
            <Input
              id="proj-budget"
              type="number"
              min={0}
              value={budget}
              onChange={(e) => setBudget(e.target.value)}
              placeholder="输入额度"
              className="h-8 text-sm tabular-nums"
            />
            <p className="text-xs text-muted-foreground">
              可用额度：¥{available.toLocaleString()}
            </p>
          </div>

          {/* Member picker */}
          <div className="grid gap-1.5">
            <Label className="text-xs font-medium">关联成员</Label>
            <BudgetMemberPicker
              departmentId={department.id}
              selectedIds={memberIds}
              onChange={setMemberIds}
            />
          </div>
        </div>

        {error && <p className="text-xs text-red-600">{error}</p>}

        <DialogFooter>
          <Button
            variant="outline"
            size="sm"
            onClick={() => handleOpenChange(false)}
            disabled={saving}
          >
            取消
          </Button>
          <Button size="sm" onClick={handleCreate} disabled={saving}>
            {saving ? '创建中…' : '创建'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
