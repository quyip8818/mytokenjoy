import { useState, useMemo } from 'react'
import type { BudgetNode, Department, Member } from '@/api/types'
import { displayToPoints, formatDisplayCurrency } from '@/lib/points'
import { BudgetOrgMemberPicker } from './budget-org-member-picker'
import { FormDialog } from '@/components/ui/form-dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

interface ProjectDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  department: BudgetNode
  existingProjectsBudget?: number
  memberBudgetSum?: number
  onCreateProject: (data: {
    name: string
    budget: number
    memberIds: string[]
    ownerDepartmentId: string
  }) => Promise<void>
  getDepartmentTree: () => Promise<Department[]>
  getMembers: (departmentId: string) => Promise<Member[]>
  getAllDeptMembers: (departmentId: string) => Promise<Member[]>
  searchMembers: (keyword: string) => Promise<Member[]>
}

export function ProjectDialog({
  open,
  onOpenChange,
  department,
  existingProjectsBudget = 0,
  memberBudgetSum = 0,
  onCreateProject,
  getDepartmentTree,
  getMembers,
  getAllDeptMembers,
  searchMembers,
}: ProjectDialogProps) {
  const [name, setName] = useState('')
  const [budget, setBudget] = useState('')
  const [memberIds, setMemberIds] = useState<string[]>([])
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const available = useMemo(() => {
    const childrenSum =
      department.children?.reduce((sum: number, child: BudgetNode) => sum + child.budget, 0) ?? 0
    return department.budget - childrenSum - existingProjectsBudget - memberBudgetSum
  }, [department, existingProjectsBudget, memberBudgetSum])

  function resetForm() {
    setName('')
    setBudget('')
    setMemberIds([])
    setError(null)
  }

  function handleOpenChange(value: boolean) {
    if (!value) resetForm()
    onOpenChange(value)
  }

  async function handleCreate() {
    setError(null)

    const trimmedName = name.trim()
    if (!trimmedName) {
      setError('请输入项目名称')
      return
    }
    const budgetNum = parseFloat(budget)
    if (!budget || Number.isNaN(budgetNum) || budgetNum < 0) {
      setError('请输入有效的项目额度')
      return
    }
    if (displayToPoints(budgetNum) > available) {
      setError(`团队可用额度为 ${formatDisplayCurrency(available)}，请调低项目额度`)
      return
    }

    setSaving(true)
    try {
      await onCreateProject({
        name: trimmedName,
        budget: displayToPoints(budgetNum),
        memberIds,
        ownerDepartmentId: department.id,
      })
      resetForm()
      onOpenChange(false)
    } catch {
      setError('创建失败，请重试')
    } finally {
      setSaving(false)
    }
  }

  return (
    <FormDialog
      open={open}
      onOpenChange={handleOpenChange}
      title="创建项目"
      error={error}
      busy={saving}
      submitLabel="创建"
      onSubmit={handleCreate}
    >
      <div className="grid gap-1.5">
        <Label htmlFor="proj-name" className="text-xs font-medium">
          项目名称
        </Label>
        <Input
          id="proj-name"
          value={name}
          onChange={(event) => setName(event.target.value)}
          placeholder="输入项目名称"
          className="h-8 text-sm"
        />
      </div>

      <div className="grid gap-1.5">
        <Label htmlFor="proj-budget" className="text-xs font-medium">
          项目额度（元）
        </Label>
        <Input
          id="proj-budget"
          type="number"
          min={0}
          value={budget}
          onChange={(event) => setBudget(event.target.value)}
          placeholder="输入额度"
          className="h-8 text-sm tabular-nums"
        />
        <p className="text-xs text-muted-foreground">
          可用额度：{formatDisplayCurrency(available)}
        </p>
      </div>

      <div className="grid gap-1.5">
        <Label className="text-xs font-medium">关联成员</Label>
        <BudgetOrgMemberPicker
          selectedIds={memberIds}
          onChange={setMemberIds}
          defaultExpandDepartmentId={department.id}
          getDepartmentTree={getDepartmentTree}
          getMembers={getMembers}
          getAllDeptMembers={getAllDeptMembers}
          searchMembers={searchMembers}
        />
      </div>
    </FormDialog>
  )
}
