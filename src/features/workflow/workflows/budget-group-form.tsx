import { useEffect, useMemo, useState } from 'react'
import { toast } from 'sonner'
import { budgetApi } from '@/api/budget'
import type { BudgetGroup, BudgetNode, Member } from '@/api/types'
import type { WorkflowComponentProps } from '../types'
import { WorkflowPanelChrome, WorkflowPanelFooter } from '../components/workflow-panel-chrome'
import { WorkflowFormLayout } from '../components/workflow-form-layout'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { useWorkflow } from '../use-workflow'
import { findBudgetNode } from '@/lib/budget'
import { flattenDepartments } from '@/lib/org'
import { departmentApi, memberApi } from '@/api/org'

export function BudgetGroupFormWorkflow({
  entry,
  onClose,
  onSetDirty,
  onPush,
}: WorkflowComponentProps<'budget-group-form'>) {
  const { closeAll } = useWorkflow()
  const group = entry.payload.group as BudgetGroup | undefined
  const tree = useMemo(() => (entry.payload.tree as BudgetNode[]) ?? [], [entry.payload.tree])
  const onSuccess = entry.payload.onSuccess as ((id?: string) => void) | undefined
  const [name, setName] = useState(group?.name ?? '')
  const [budget, setBudget] = useState(String(group?.budget ?? ''))
  const [departmentId, setDepartmentId] = useState(group?.departmentIds[0] ?? '')
  const [departmentName, setDepartmentName] = useState('')
  const [memberIds, setMemberIds] = useState<string[]>(group?.memberIds ?? [])
  const [memberNames, setMemberNames] = useState<string[]>([])
  const [submitting, setSubmitting] = useState(false)

  const budgetError = useMemo(() => {
    if (!departmentId || !budget) return ''
    const node = findBudgetNode(tree, departmentId)
    if (!node) return ''
    const remaining = node.budget - node.consumed - (node.reservedPool ?? 0)
    const alloc = Number(budget)
    if (alloc > remaining) {
      return `超出部门可分配额度，剩余约 ¥${Math.max(0, remaining).toLocaleString()}`
    }
    return ''
  }, [departmentId, budget, tree])

  useEffect(() => {
    if (!departmentId) return
    departmentApi.getTree().then((depts) => {
      const flat = flattenDepartments(depts)
      setDepartmentName(flat.find((d) => d.id === departmentId)?.name ?? '')
    })
    memberApi.list({ departmentId, page: 1, pageSize: 100 }).then((res) => {
      const names = memberIds
        .map((id) => res.items.find((m) => m.id === id)?.name)
        .filter((n): n is string => !!n)
      setMemberNames(names)
    })
  }, [departmentId, memberIds])

  const openPickDept = () => {
    onPush('pick-dept', {
      selectedId: departmentId,
      onConfirm: (deptId: string) => {
        setDepartmentId(deptId)
        setMemberIds([])
        setMemberNames([])
        onSetDirty(true)
      },
    })
  }

  const openPickMembers = () => {
    if (!departmentId) return
    onPush('pick-members', {
      departmentId,
      selectedIds: memberIds,
      onConfirm: (ids: string[], members: Member[]) => {
        setMemberIds(ids)
        setMemberNames(members.map((m) => m.name))
        onSetDirty(true)
      },
    })
  }

  const handleSubmit = async () => {
    if (!name.trim() || !budget || !departmentId || budgetError) return
    setSubmitting(true)
    try {
      if (group) {
        await budgetApi.updateGroup(group.id, {
          name,
          budget: Number(budget),
          departmentIds: [departmentId],
          memberIds,
        })
        toast.success('预算组已更新')
        onSuccess?.(group.id)
      } else {
        const created = await budgetApi.createGroup({
          name,
          budget: Number(budget),
          memberIds,
          departmentIds: [departmentId],
        })
        toast.success('预算组已创建')
        onSuccess?.(created.id)
      }
      closeAll()
    } catch {
      toast.error('保存失败')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <WorkflowPanelChrome
      title={group ? '编辑预算组' : '新建预算组'}
      onClose={onClose}
      footer={
        <WorkflowPanelFooter
          onCancel={onClose}
          primaryLabel={submitting ? '保存中...' : '保存'}
          onPrimary={handleSubmit}
          primaryDisabled={!name.trim() || !budget || !departmentId || !!budgetError || submitting}
        />
      }
    >
      <WorkflowFormLayout>
        <div className="space-y-1.5">
          <Label>名称</Label>
          <Input
            value={name}
            onChange={(e) => {
              setName(e.target.value)
              onSetDirty(true)
            }}
            placeholder="AI 创新项目组"
          />
        </div>
        <div className="space-y-1.5">
          <Label>预算额度 (¥)</Label>
          <Input
            type="number"
            value={budget}
            onChange={(e) => {
              setBudget(e.target.value)
              onSetDirty(true)
            }}
          />
        </div>
        <div className="space-y-1.5">
          <Label>来源部门</Label>
          <Button variant="outline" className="w-full justify-start" onClick={openPickDept}>
            {departmentName || '选择部门'}
          </Button>
        </div>
        <div className="space-y-1.5">
          <Label>关联成员</Label>
          <Button
            variant="outline"
            className="w-full justify-start"
            onClick={openPickMembers}
            disabled={!departmentId}
          >
            {memberNames.length > 0 ? `已选 ${memberNames.length} 人` : '选择成员'}
          </Button>
          {memberNames.length > 0 && (
            <p className="text-xs text-muted-foreground">{memberNames.join('、')}</p>
          )}
        </div>
        {budgetError && (
          <div className="rounded-md bg-amber-50 border border-amber-200 px-3 py-2 text-sm text-amber-800">
            {budgetError}
          </div>
        )}
      </WorkflowFormLayout>
    </WorkflowPanelChrome>
  )
}
