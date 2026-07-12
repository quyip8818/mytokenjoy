import { useState } from 'react'
import type { ProjectView } from '@/api/types'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { cn } from '@/lib/utils'
import { Check, Pencil, X } from 'lucide-react'
import { POLICY_LABELS } from '@/features/budget'
import { displayToPoints, formatDisplayCurrency, pointsToDisplay } from '@/lib/points'

type ProjectSettingsFormProps = {
  project: ProjectView
  onUpdateProject: (projectId: string, data: { budget: number }) => Promise<void>
  onUpdated: () => void
}

export function ProjectSettingsForm({
  project,
  onUpdateProject,
  onUpdated,
}: ProjectSettingsFormProps) {
  const policy = POLICY_LABELS[project.overrunPolicy]
  const pct = project.budget > 0 ? Math.round((project.consumed / project.budget) * 100) : 0

  const [editingSettings, setEditingSettings] = useState(false)
  const [draftBudget, setDraftBudget] = useState('')
  const [settingsError, setSettingsError] = useState<string | null>(null)
  const [savingSettings, setSavingSettings] = useState(false)

  function startEditSettings() {
    setDraftBudget(String(pointsToDisplay(project.budget)))
    setSettingsError(null)
    setEditingSettings(true)
  }

  function cancelEditSettings() {
    setEditingSettings(false)
    setSettingsError(null)
  }

  async function saveSettings() {
    setSettingsError(null)
    const budgetNum = parseFloat(draftBudget)
    if (!draftBudget || isNaN(budgetNum) || budgetNum < 0) {
      setSettingsError('请输入有效的额度')
      return
    }
    if (budgetNum < pointsToDisplay(project.consumed)) {
      setSettingsError(`额度不能低于已消耗 ${formatDisplayCurrency(project.consumed)}`)
      return
    }
    setSavingSettings(true)
    try {
      await onUpdateProject(project.id, { budget: displayToPoints(budgetNum) })
      setEditingSettings(false)
      onUpdated()
    } catch {
      setSettingsError('保存失败，请重试')
    } finally {
      setSavingSettings(false)
    }
  }

  return (
    <div>
      <div className="mb-3 flex items-center justify-between">
        <h4 className="text-sm font-semibold text-foreground">项目设置</h4>
        {!editingSettings ? (
          <Button
            variant="ghost"
            size="sm"
            className="h-7 gap-1.5 text-xs text-muted-foreground"
            onClick={startEditSettings}
            aria-label="编辑项目设置"
          >
            <Pencil className="size-3.5" />
            编辑
          </Button>
        ) : (
          <div className="flex items-center gap-1">
            <Button
              variant="ghost"
              size="sm"
              className="h-7 gap-1.5 text-xs text-muted-foreground"
              onClick={cancelEditSettings}
              disabled={savingSettings}
              aria-label="取消编辑设置"
            >
              <X className="size-3.5" />
              取消
            </Button>
            <Button
              size="sm"
              className="h-7 gap-1.5 text-xs"
              onClick={saveSettings}
              disabled={savingSettings}
              aria-label="保存设置"
            >
              <Check className="size-3.5" />
              保存
            </Button>
          </div>
        )}
      </div>

      <div className="rounded-lg border border-border p-4">
        <div className="grid grid-cols-2 gap-4">
          <div className="grid gap-1.5">
            <Label htmlFor="proj-settings-budget" className="text-xs text-muted-foreground">
              项目额度（元）
            </Label>
            {editingSettings ? (
              <Input
                id="proj-settings-budget"
                type="number"
                min={0}
                value={draftBudget}
                onChange={(e) => setDraftBudget(e.target.value)}
                className="h-8 text-sm tabular-nums"
              />
            ) : (
              <p className="text-sm font-medium tabular-nums">
                {formatDisplayCurrency(project.budget)}
              </p>
            )}
          </div>

          <div className="grid gap-1.5">
            <Label htmlFor="proj-settings-policy" className="text-xs text-muted-foreground">
              超限策略
            </Label>
            <Badge variant="outline" className={cn(policy.className, 'w-fit text-xs font-normal')}>
              {policy.label}
            </Badge>
          </div>
        </div>

        {editingSettings && settingsError && (
          <p className="mt-2 text-xs text-red-600">{settingsError}</p>
        )}

        {!editingSettings && (
          <div className="mt-3 border-t border-border pt-3">
            <div className="flex justify-between text-xs text-muted-foreground">
              <span>本月消耗进度</span>
              <span className="tabular-nums">
                {formatDisplayCurrency(project.consumed)} / {formatDisplayCurrency(project.budget)}{' '}
                ({pct}%)
              </span>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
