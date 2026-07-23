import { useState } from 'react'
import type { ProjectView, Member } from '@/api/types'
import { ApiError } from '@/api/client'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Pencil } from 'lucide-react'
import { displayToQuota, formatDisplayCurrency, quotaToDisplay } from '@/lib/quota-display'

type ProjectSettingsFormProps = {
  project: ProjectView
  members?: Member[]
  onUpdateProject: (
    projectId: string,
    data: { budget?: number; ownerId?: string },
  ) => Promise<void>
  onUpdated: () => void
}

export function ProjectSettingsForm({
  project,
  members = [],
  onUpdateProject,
  onUpdated,
}: ProjectSettingsFormProps) {
  const pct = project.budget > 0 ? Math.round((project.consumed / project.budget) * 100) : 0
  const [dialogOpen, setDialogOpen] = useState(false)
  const [draftBudget, setDraftBudget] = useState('')
  const [draftOwnerId, setDraftOwnerId] = useState('')
  const [saving, setSaving] = useState(false)

  const ownerName = members.find((m) => m.id === project.ownerId)?.alias ?? '未设置'

  function openDialog() {
    setDraftBudget(String(quotaToDisplay(project.budget)))
    setDraftOwnerId(project.ownerId ?? '')
    setDialogOpen(true)
  }

  async function handleSave() {
    const budgetNum = parseFloat(draftBudget)
    if (!draftBudget || isNaN(budgetNum) || budgetNum < 0) {
      toast.error('请输入有效的额度')
      return
    }
    if (budgetNum < quotaToDisplay(project.consumed)) {
      toast.error(`额度不能低于已消耗 ${formatDisplayCurrency(project.consumed)}`)
      return
    }
    setSaving(true)
    try {
      const data: { budget?: number; ownerId?: string } = { budget: displayToQuota(budgetNum) }
      if (draftOwnerId && draftOwnerId !== project.ownerId) {
        data.ownerId = draftOwnerId
      }
      await onUpdateProject(project.id, data)
      setDialogOpen(false)
      onUpdated()
      toast.success('项目设置已更新')
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : '保存失败，请重试')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div>
      <div className="mb-3 flex items-center justify-between">
        <h4 className="text-sm font-semibold text-foreground">项目设置</h4>
        <Button
          variant="ghost"
          size="sm"
          className="h-7 gap-1.5 text-xs text-muted-foreground"
          onClick={openDialog}
          aria-label="编辑项目设置"
        >
          <Pencil className="size-3.5" />
          编辑
        </Button>
      </div>

      <div className="rounded-lg border border-border p-4">
        <div className="grid grid-cols-2 gap-4">
          <div className="grid gap-1.5">
            <Label className="text-xs text-muted-foreground">项目额度（元）</Label>
            <p className="text-sm font-medium tabular-nums">
              {formatDisplayCurrency(project.budget)}
            </p>
          </div>
          <div className="grid gap-1.5">
            <Label className="text-xs text-muted-foreground">负责人</Label>
            <p className="text-sm font-medium">{ownerName}</p>
          </div>
        </div>
        <div className="mt-3 border-t border-border pt-3">
          <div className="flex justify-between text-xs text-muted-foreground">
            <span>本月消耗进度</span>
            <span className="tabular-nums">
              {formatDisplayCurrency(project.consumed)} / {formatDisplayCurrency(project.budget)} (
              {pct}%)
            </span>
          </div>
        </div>
      </div>

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="sm:max-w-sm">
          <DialogHeader>
            <DialogTitle>编辑项目设置</DialogTitle>
          </DialogHeader>
          <div className="grid gap-3 py-2">
            <div className="grid gap-1.5">
              <Label htmlFor="proj-budget-edit" className="text-xs font-medium">
                项目额度（元）
              </Label>
              <Input
                id="proj-budget-edit"
                type="number"
                min={0}
                value={draftBudget}
                onChange={(e) => setDraftBudget(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter') void handleSave()
                }}
                className="h-8 tabular-nums"
                autoFocus
              />
              <p className="text-xs text-muted-foreground">
                已消耗：{formatDisplayCurrency(project.consumed)}
              </p>
            </div>
            {members.length > 0 && (
              <div className="grid gap-1.5">
                <Label className="text-xs font-medium">负责人</Label>
                <Select value={draftOwnerId} onValueChange={setDraftOwnerId}>
                  <SelectTrigger className="h-8">
                    <SelectValue placeholder="选择负责人" />
                  </SelectTrigger>
                  <SelectContent>
                    {members.map((m) => (
                      <SelectItem key={m.id} value={m.id}>
                        {m.alias}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            )}
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setDialogOpen(false)}
              disabled={saving}
            >
              取消
            </Button>
            <Button size="sm" onClick={handleSave} disabled={saving}>
              {saving ? '保存中…' : '保存'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
