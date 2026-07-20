import { useState } from 'react'
import { toast } from 'sonner'
import { ApiError } from '@/api/client'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { displayToQuota, formatDisplayCurrency } from '@/lib/quota-display'
import { Wallet } from 'lucide-react'

interface BudgetInitPromptProps {
  departmentId: string
  departmentName: string
  onUpdateDepartment: (id: string, data: { budget: number }) => Promise<void>
}

export function BudgetInitPrompt({
  departmentId,
  departmentName,
  onUpdateDepartment,
}: BudgetInitPromptProps) {
  const [dialogOpen, setDialogOpen] = useState(false)
  const [draft, setDraft] = useState('')
  const [saving, setSaving] = useState(false)

  async function handleSave() {
    const value = parseFloat(draft)
    if (Number.isNaN(value) || value <= 0) {
      toast.error('请输入有效的额度')
      return
    }
    setSaving(true)
    try {
      await onUpdateDepartment(departmentId, { budget: displayToQuota(value) })
      setDialogOpen(false)
      toast.success(
        `已设置${departmentName}总额度为 ${formatDisplayCurrency(displayToQuota(value))}`,
      )
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : '设置失败，请重试')
    } finally {
      setSaving(false)
    }
  }

  return (
    <>
      <div className="flex flex-col items-center gap-4 rounded-lg border border-dashed border-border p-8 text-center">
        <Wallet className="size-8 text-muted-foreground" />
        <div>
          <p className="text-sm font-medium text-foreground">尚未设置预算额度</p>
          <p className="mt-1 text-xs text-muted-foreground">
            请先设置总额度，然后再分配部门和成员额度
          </p>
        </div>
        <Button onClick={() => setDialogOpen(true)}>设置总额度</Button>
      </div>

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="sm:max-w-sm">
          <DialogHeader>
            <DialogTitle>设置{departmentName}总额度</DialogTitle>
          </DialogHeader>
          <div className="py-2">
            <p className="mb-3 text-xs text-muted-foreground">
              预算周期为每月，额度将在每月 1 号自动刷新。
            </p>
            <Input
              type="number"
              min={0}
              value={draft}
              onChange={(e) => setDraft(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') void handleSave()
              }}
              placeholder="输入总额度（元）"
              className="tabular-nums"
              autoFocus
            />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)} disabled={saving}>
              取消
            </Button>
            <Button onClick={handleSave} disabled={saving || !draft.trim()}>
              {saving ? '设置中…' : '确定'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
