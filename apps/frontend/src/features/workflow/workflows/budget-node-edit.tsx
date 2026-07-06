import { useMemo, useState } from 'react'
import { computeUnallocated, sumChildrenBudget } from '@/lib/budget'
import type { BudgetNode } from '@/api/types'
import type { WorkflowComponentProps } from '../types'
import { WorkflowPanelChrome, WorkflowPanelFooter } from '../components/workflow-panel-chrome'
import { WorkflowInfoBox } from '../components/workflow-info-box'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

export function BudgetNodeEditWorkflow({
  entry,
  onClose,
  onPush,
  onSetDirty,
}: WorkflowComponentProps<'budget-node-edit'>) {
  const node = entry.payload.node as BudgetNode
  const parent = entry.payload.parent as BudgetNode | null | undefined
  const onSuccess = entry.payload.onSuccess as (() => void) | undefined
  const [budget, setBudget] = useState(String(node.budget))
  const [reservedPool, setReservedPool] = useState(String(node.reservedPool ?? 0))

  const error = useMemo(() => {
    const b = Number(budget)
    const r = Number(reservedPool)
    const childrenSum = sumChildrenBudget(node)
    if (parent) {
      const parentUnalloc =
        parent.budget - (parent.reservedPool ?? 0) - sumChildrenBudget(parent) + node.budget
      if (b + r > parentUnalloc + node.budget) {
        return `超出可分配额度，父级剩余约 ¥${Math.max(0, parentUnalloc).toLocaleString()}`
      }
    }
    if (b < childrenSum + r) {
      return `子节点预算之和 ¥${childrenSum.toLocaleString()} + 预留池不能超过节点预算`
    }
    return ''
  }, [budget, reservedPool, node, parent])

  const handlePreview = () => {
    if (error) return
    onPush('budget-impact-preview', {
      nodeId: node.id,
      nodeName: node.name,
      before: { budget: node.budget, reservedPool: node.reservedPool ?? 0 },
      after: { budget: Number(budget), reservedPool: Number(reservedPool) },
      onSuccess,
    })
    onSetDirty(false)
  }

  return (
    <WorkflowPanelChrome
      title="编辑节点预算"
      onClose={onClose}
      contextBar={node.name}
      footer={
        <WorkflowPanelFooter
          onCancel={onClose}
          primaryLabel="预览并保存"
          onPrimary={handlePreview}
          primaryDisabled={!!error || !budget.trim()}
        />
      }
    >
      <div className="grid grid-cols-5 gap-8">
        <div className="col-span-3 space-y-5">
          <div className="space-y-1.5">
            <Label>月度预算 (¥)</Label>
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
            <Label>预留池 (¥)</Label>
            <Input
              type="number"
              value={reservedPool}
              onChange={(e) => {
                setReservedPool(e.target.value)
                onSetDirty(true)
              }}
            />
          </div>
          {error && (
            <div className="rounded-md bg-red-50 border border-red-200 px-3 py-2 text-sm text-red-700">
              {error}
            </div>
          )}
        </div>
        <WorkflowInfoBox fullWidth className="space-y-2 text-muted-foreground">
          <h4 className="font-semibold text-foreground/80">摘要</h4>
          <p>已消耗：¥{node.consumed.toLocaleString()}</p>
          <p>子节点合计：¥{sumChildrenBudget(node).toLocaleString()}</p>
          <p>未分配：¥{computeUnallocated(node).toLocaleString()}</p>
          {parent && <p>父节点：{parent.name}</p>}
        </WorkflowInfoBox>
      </div>
    </WorkflowPanelChrome>
  )
}
