import { useCallback, useEffect, useState } from 'react'
import type { MemberBudget, UpdateMemberBudgetInput } from '@/api/types'
import { ApiError } from '@/api/client'
import { toast } from 'sonner'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { displayToPoints, formatDisplayCurrency, pointsToDisplay } from '@/lib/points'
import { cn } from '@/lib/utils'
import { Pencil, Check, X, Loader2 } from 'lucide-react'

interface BudgetMemberBudgetDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  departmentId: string
  getMemberBudgets: (departmentId: string) => Promise<MemberBudget[]>
  updateMemberBudget: (memberId: string, data: UpdateMemberBudgetInput) => Promise<MemberBudget>
}

export function BudgetMemberBudgetDialog({
  open,
  onOpenChange,
  departmentId,
  getMemberBudgets,
  updateMemberBudget,
}: BudgetMemberBudgetDialogProps) {
  function handleOpenChange(value: boolean) {
    onOpenChange(value)
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>成员额度配置</DialogTitle>
        </DialogHeader>

        {open && departmentId ? (
          <BudgetMemberBudgetDialogBody
            key={departmentId}
            departmentId={departmentId}
            getMemberBudgets={getMemberBudgets}
            updateMemberBudget={updateMemberBudget}
          />
        ) : null}
      </DialogContent>
    </Dialog>
  )
}

interface BudgetMemberBudgetDialogBodyProps {
  departmentId: string
  getMemberBudgets: (departmentId: string) => Promise<MemberBudget[]>
  updateMemberBudget: (memberId: string, data: UpdateMemberBudgetInput) => Promise<MemberBudget>
}

function BudgetMemberBudgetDialogBody({
  departmentId,
  getMemberBudgets,
  updateMemberBudget,
}: BudgetMemberBudgetDialogBodyProps) {
  const [memberBudgets, setMemberBudgets] = useState<MemberBudget[]>([])
  const [loading, setLoading] = useState(true)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [draft, setDraft] = useState('')
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    let cancelled = false
    void getMemberBudgets(departmentId)
      .then((data) => {
        if (!cancelled) {
          setMemberBudgets(data ?? [])
        }
      })
      .catch((err) => {
        if (!cancelled) {
          toast.error(err instanceof ApiError ? err.message : '加载成员额度失败')
        }
      })
      .finally(() => {
        if (!cancelled) {
          setLoading(false)
        }
      })
    return () => {
      cancelled = true
    }
  }, [departmentId, getMemberBudgets])

  const startEdit = useCallback((member: MemberBudget) => {
    setEditingId(member.memberId)
    setDraft(String(pointsToDisplay(member.personalBudget)))
  }, [])

  const cancelEdit = useCallback(() => {
    setEditingId(null)
    setDraft('')
  }, [])

  const handleSave = useCallback(
    async (memberId: string) => {
      const value = parseFloat(draft)
      if (Number.isNaN(value) || value < 0) {
        toast.error('请输入有效的额度数值')
        return
      }

      setSaving(true)
      try {
        const updated = await updateMemberBudget(memberId, {
          personalBudget: displayToPoints(value),
        })
        setMemberBudgets((prev) =>
          prev.map((item) => (item.memberId === memberId ? updated : item)),
        )
        setEditingId(null)
        setDraft('')
        toast.success('成员额度已更新')
      } catch (err) {
        toast.error(err instanceof ApiError ? err.message : '修改失败，请重试')
      } finally {
        setSaving(false)
      }
    },
    [draft, updateMemberBudget],
  )

  if (loading) {
    return (
      <div className="flex items-center justify-center py-8">
        <Loader2 className="size-5 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (memberBudgets.length === 0) {
    return <p className="py-8 text-center text-sm text-muted-foreground">暂无成员</p>
  }

  return (
    <div className="max-h-80 overflow-y-auto">
      <table className="w-full text-sm">
        <thead className="sticky top-0 bg-background">
          <tr className="border-b border-border text-left text-xs text-muted-foreground">
            <th className="pb-2 font-medium">成员</th>
            <th className="pb-2 font-medium">个人额度</th>
            <th className="pb-2 font-medium">已消耗</th>
            <th className="pb-2 text-right font-medium">操作</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-border">
          {memberBudgets.map((member) => (
            <MemberRow
              key={member.memberId}
              member={member}
              editing={editingId === member.memberId}
              draft={draft}
              saving={saving}
              onDraftChange={setDraft}
              onStartEdit={() => startEdit(member)}
              onCancel={cancelEdit}
              onSave={() => void handleSave(member.memberId)}
            />
          ))}
        </tbody>
      </table>
    </div>
  )
}

interface MemberRowProps {
  member: MemberBudget
  editing: boolean
  draft: string
  saving: boolean
  onDraftChange: (value: string) => void
  onStartEdit: () => void
  onCancel: () => void
  onSave: () => void
}

function MemberRow({
  member,
  editing,
  draft,
  saving,
  onDraftChange,
  onStartEdit,
  onCancel,
  onSave,
}: MemberRowProps) {
  return (
    <tr className={cn('h-10', editing && 'bg-muted/30')}>
      <td className="py-2 font-medium text-foreground">{member.memberName}</td>
      <td className="py-2 tabular-nums">
        {editing ? (
          <Input
            type="number"
            min={0}
            value={draft}
            onChange={(e) => onDraftChange(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') onSave()
              if (e.key === 'Escape') onCancel()
            }}
            className="h-7 w-28 tabular-nums"
            placeholder="元"
            autoFocus
          />
        ) : (
          <span className="text-muted-foreground">
            {formatDisplayCurrency(member.personalBudget)}
          </span>
        )}
      </td>
      <td className="py-2 tabular-nums text-muted-foreground">
        {formatDisplayCurrency(member.consumed)}
      </td>
      <td className="py-2 text-right">
        {editing ? (
          <div className="flex items-center justify-end gap-1">
            <Button
              variant="ghost"
              size="icon"
              className="size-7"
              onClick={onCancel}
              disabled={saving}
              aria-label="取消"
            >
              <X className="size-3.5" />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="size-7"
              onClick={onSave}
              disabled={saving}
              aria-label="保存"
            >
              {saving ? (
                <Loader2 className="size-3.5 animate-spin" />
              ) : (
                <Check className="size-3.5" />
              )}
            </Button>
          </div>
        ) : (
          <Button
            variant="ghost"
            size="icon"
            className="size-7 text-muted-foreground"
            onClick={onStartEdit}
            aria-label={`编辑${member.memberName}的额度`}
          >
            <Pencil className="size-3.5" />
          </Button>
        )}
      </td>
    </tr>
  )
}
