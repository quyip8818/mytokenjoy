import { useCallback, useEffect, useState } from 'react'
import type { MemberBudgetQuota } from '@/api/types'
import { ApiError } from '@/api/client'
import { toast } from 'sonner'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { displayToPoints, formatDisplayCurrency, pointsToDisplay } from '@/lib/points'
import { cn } from '@/lib/utils'
import { Pencil, Check, X, Loader2 } from 'lucide-react'

interface BudgetMemberQuotaDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  departmentId: string
  getMemberQuotas: (departmentId: string) => Promise<MemberBudgetQuota[]>
  updateMemberQuota: (memberId: string, data: { personalQuota: number }) => Promise<MemberBudgetQuota>
}

export function BudgetMemberQuotaDialog({
  open,
  onOpenChange,
  departmentId,
  getMemberQuotas,
  updateMemberQuota,
}: BudgetMemberQuotaDialogProps) {
  const [quotas, setQuotas] = useState<MemberBudgetQuota[]>([])
  const [loading, setLoading] = useState(false)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [draft, setDraft] = useState('')
  const [saving, setSaving] = useState(false)

  const fetchQuotas = useCallback(async () => {
    setLoading(true)
    try {
      const data = await getMemberQuotas(departmentId)
      setQuotas(data ?? [])
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : '加载成员额度失败')
    } finally {
      setLoading(false)
    }
  }, [departmentId, getMemberQuotas])

  useEffect(() => {
    if (open && departmentId) {
      fetchQuotas()
    }
  }, [open, departmentId, fetchQuotas])

  function startEdit(member: MemberBudgetQuota) {
    setEditingId(member.memberId)
    setDraft(String(pointsToDisplay(member.personalQuota)))
  }

  function cancelEdit() {
    setEditingId(null)
    setDraft('')
  }

  async function handleSave(memberId: string) {
    const value = parseFloat(draft)
    if (Number.isNaN(value) || value < 0) {
      toast.error('请输入有效的额度数值')
      return
    }

    setSaving(true)
    try {
      const updated = await updateMemberQuota(memberId, {
        personalQuota: displayToPoints(value),
      })
      setQuotas((prev) =>
        prev.map((q) => (q.memberId === memberId ? updated : q)),
      )
      setEditingId(null)
      setDraft('')
      toast.success('成员额度已更新')
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : '修改失败，请重试')
    } finally {
      setSaving(false)
    }
  }

  function handleOpenChange(value: boolean) {
    if (!value) {
      setEditingId(null)
      setDraft('')
    }
    onOpenChange(value)
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>成员额度配置</DialogTitle>
        </DialogHeader>

        {loading ? (
          <div className="flex items-center justify-center py-8">
            <Loader2 className="size-5 animate-spin text-muted-foreground" />
          </div>
        ) : quotas.length === 0 ? (
          <p className="py-8 text-center text-sm text-muted-foreground">暂无成员</p>
        ) : (
          <div className="max-h-80 overflow-y-auto">
            <table className="w-full text-sm">
              <thead className="sticky top-0 bg-background">
                <tr className="border-b border-border text-left text-xs text-muted-foreground">
                  <th className="pb-2 font-medium">成员</th>
                  <th className="pb-2 font-medium">个人额度</th>
                  <th className="pb-2 font-medium">已用</th>
                  <th className="pb-2 text-right font-medium">操作</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {quotas.map((member) => (
                  <MemberRow
                    key={member.memberId}
                    member={member}
                    editing={editingId === member.memberId}
                    draft={draft}
                    saving={saving}
                    onDraftChange={setDraft}
                    onStartEdit={() => startEdit(member)}
                    onCancel={cancelEdit}
                    onSave={() => handleSave(member.memberId)}
                  />
                ))}
              </tbody>
            </table>
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}

interface MemberRowProps {
  member: MemberBudgetQuota
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
            {formatDisplayCurrency(member.personalQuota)}
          </span>
        )}
      </td>
      <td className="py-2 tabular-nums text-muted-foreground">
        {formatDisplayCurrency(member.used)}
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
