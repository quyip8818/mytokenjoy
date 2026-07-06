import { useState } from 'react'
import { budgetApi } from '@/api/budget'
import type { BudgetNode, BudgetProject } from '@/api/types'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { cn } from '@/lib/utils'
import { Pencil, Users, Wallet, X, Check } from 'lucide-react'

interface BudgetEditMemberQuotaProps {
  node: BudgetNode
  projects: BudgetProject[]
  onUpdated: () => void
}

export function BudgetEditMemberQuota({ node, projects, onUpdated }: BudgetEditMemberQuotaProps) {
  const [editing, setEditing] = useState(false)
  const [quotaDraft, setQuotaDraft] = useState('')
  const [reservedDraft, setReservedDraft] = useState('')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  function startEdit() {
    setQuotaDraft(String(node.memberQuota))
    setReservedDraft(String(node.reserved))
    setError(null)
    setEditing(true)
  }

  function cancelEdit() {
    setError(null)
    setEditing(false)
  }

  function computeAllocated(newReserved: number) {
    const childSum = node.children?.reduce((s, c) => s + c.budget, 0) ?? 0
    const projectSum = projects.filter((p) => p.departmentId === node.id).reduce((s, p) => s + p.budget, 0)
    return childSum + projectSum + newReserved
  }

  async function handleSave() {
    const quota = parseFloat(quotaDraft)
    const reserved = parseFloat(reservedDraft)
    if (isNaN(quota) || quota < 0) {
      setError('平均额度无效')
      return
    }
    if (isNaN(reserved) || reserved < 0) {
      setError('预留池余额无效')
      return
    }
    const allocated = computeAllocated(reserved)
    if (allocated > node.budget) {
      setError(`分配总额 ¥${allocated.toLocaleString()} 超出节点额度 ¥${node.budget.toLocaleString()}`)
      return
    }
    setSaving(true)
    try {
      await budgetApi.updateNode(node.id, { memberQuota: quota, reserved })
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
        <div className={cn('flex items-center gap-3 rounded-md bg-muted/50 px-3 py-2.5', editing && 'items-start pt-3')}>
          <Wallet className="mt-0.5 size-4 shrink-0 text-muted-foreground" />
          <div className="min-w-0 flex-1">
            <p className="text-xs text-muted-foreground">平均额度/人</p>
            {editing ? (
              <Input
                type="number"
                min={0}
                value={quotaDraft}
                onChange={(e) => {
                  setQuotaDraft(e.target.value)
                  setError(null)
                }}
                className="mt-1 h-7 tabular-nums"
                placeholder="元"
              />
            ) : (
              <p className="text-sm font-medium tabular-nums">¥{node.memberQuota.toLocaleString()}</p>
            )}
          </div>
        </div>

        <div className={cn('flex items-center gap-3 rounded-md bg-muted/50 px-3 py-2.5', editing && 'items-start pt-3')}>
          <Users className="mt-0.5 size-4 shrink-0 text-muted-foreground" />
          <div className="min-w-0 flex-1">
            <p className="text-xs text-muted-foreground">预留池余额</p>
            {editing ? (
              <Input
                type="number"
                min={0}
                value={reservedDraft}
                onChange={(e) => {
                  setReservedDraft(e.target.value)
                  setError(null)
                }}
                className="mt-1 h-7 tabular-nums"
                placeholder="元"
              />
            ) : (
              <p className="text-sm font-medium tabular-nums">¥{node.reserved.toLocaleString()}</p>
            )}
          </div>
        </div>
      </div>

      {error && <p className="mt-2 text-xs text-red-600">{error}</p>}

      <p className="mt-2 text-xs text-muted-foreground">
        成员个人额度用尽后可申请从预留池追加，需 TL 审批通过
      </p>
    </div>
  )
}
