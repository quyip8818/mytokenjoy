import { useState } from 'react'
import type { BudgetNode, BudgetProjectView, Role } from '@/api/types'
import type { AlertRuleView } from '@/features/budget/lib/alerts'
import { groupProjectsByTeam } from '@/features/budget/lib/alerts'
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
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Checkbox } from '@/components/ui/checkbox'
import { Badge } from '@/components/ui/badge'
import { DepartmentTreeSelect } from './department-tree-select'
import { cn } from '@/lib/utils'
import { Plus, X } from 'lucide-react'

interface AlertRuleDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  rule: AlertRuleView | null
  tree: BudgetNode[]
  projects: BudgetProjectView[]
  roles: Role[]
  onSave: (view: AlertRuleView, existingId?: string) => Promise<void>
}

const PRESET_THRESHOLDS = [
  { label: '80%, 90%, 100%', value: [80, 90, 100] },
  { label: '90%, 100%', value: [90, 100] },
  { label: '仅 100%', value: [100] },
]

export function AlertRuleDialog({
  open,
  onOpenChange,
  rule,
  tree,
  projects,
  roles,
  onSave,
}: AlertRuleDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      {open && (
        <AlertRuleDialogContent
          rule={rule}
          tree={tree}
          projects={projects}
          roles={roles}
          onOpenChange={onOpenChange}
          onSave={onSave}
        />
      )}
    </Dialog>
  )
}

function AlertRuleDialogContent({
  rule,
  tree,
  projects,
  roles,
  onOpenChange,
  onSave,
}: {
  rule: AlertRuleView | null
  tree: BudgetNode[]
  projects: BudgetProjectView[]
  roles: Role[]
  onOpenChange: (open: boolean) => void
  onSave: (view: AlertRuleView, existingId?: string) => Promise<void>
}) {
  const [targetType, setTargetType] = useState<'team' | 'project'>(rule?.targetType ?? 'team')
  const [targetId, setTargetId] = useState(rule?.targetId ?? '')
  const [targetName, setTargetName] = useState(rule?.targetName ?? '')
  const [thresholds, setThresholds] = useState<number[]>(
    rule?.thresholds ? [...rule.thresholds] : [80, 90, 100],
  )
  const [customThreshold, setCustomThreshold] = useState('')
  const [notifyRoleIds, setNotifyRoleIds] = useState<string[]>(
    rule?.notifyRoleIds ? [...rule.notifyRoleIds] : [],
  )
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  function addCustomThreshold() {
    const num = parseInt(customThreshold, 10)
    if (Number.isNaN(num) || num <= 0 || num > 100) return
    if (thresholds.includes(num)) return
    setThresholds([...thresholds, num].sort((a, b) => a - b))
    setCustomThreshold('')
  }

  function removeThreshold(value: number) {
    setThresholds(thresholds.filter((threshold) => threshold !== value))
  }

  function toggleRole(roleId: string) {
    setNotifyRoleIds((prev) =>
      prev.includes(roleId) ? prev.filter((id) => id !== roleId) : [...prev, roleId],
    )
  }

  function selectPreset(preset: number[]) {
    setThresholds([...preset])
  }

  async function handleSave() {
    setError(null)
    if (!targetId) {
      setError('请选择监控对象')
      return
    }
    if (thresholds.length === 0) {
      setError('请至少设置一个阈值')
      return
    }
    if (notifyRoleIds.length === 0) {
      setError('请选择通知角色')
      return
    }

    setSaving(true)
    try {
      await onSave(
        {
          id: rule?.id ?? '',
          nodeId: targetId,
          nodeName: targetName,
          targetType,
          targetId,
          targetName,
          thresholds,
          notifyRoleIds,
          enabled: rule?.enabled ?? true,
        },
        rule?.id,
      )
      onOpenChange(false)
    } catch {
      setError('保存失败，请重试')
    } finally {
      setSaving(false)
    }
  }

  return (
    <DialogContent className="sm:max-w-lg">
      <DialogHeader>
        <DialogTitle>{rule ? '编辑预警规则' : '创建预警规则'}</DialogTitle>
      </DialogHeader>

      <div className="grid gap-5 py-2">
        <div className="grid gap-1.5">
          <Label className="text-xs font-medium">监控类型</Label>
          <RadioGroup
            value={targetType}
            onValueChange={(value) => {
              setTargetType(value as 'team' | 'project')
              setTargetId('')
              setTargetName('')
            }}
            className="flex gap-4"
          >
            <div className="flex items-center gap-2">
              <RadioGroupItem value="team" id="type-team" />
              <Label htmlFor="type-team" className="text-sm">
                团队
              </Label>
            </div>
            <div className="flex items-center gap-2">
              <RadioGroupItem value="project" id="type-project" />
              <Label htmlFor="type-project" className="text-sm">
                项目
              </Label>
            </div>
          </RadioGroup>
        </div>

        <div className="grid gap-1.5">
          <Label className="text-xs font-medium">监控对象</Label>
          {targetType === 'team' ? (
            <DepartmentTreeSelect
              tree={tree}
              value={targetId}
              onChange={(id, name) => {
                setTargetId(id)
                setTargetName(name)
              }}
              placeholder="选择团队…"
            />
          ) : (
            <Select
              value={targetId}
              onValueChange={(value) => {
                setTargetId(value)
                const project = projects.find((item) => item.id === value)
                setTargetName(project?.name ?? '')
              }}
            >
              <SelectTrigger className="h-8 text-sm">
                <SelectValue placeholder="选择项目…" />
              </SelectTrigger>
              <SelectContent>
                {groupProjectsByTeam(projects, tree).map((group) => (
                  <div key={group.teamId}>
                    <div className="px-2 py-1.5 text-xs font-medium text-muted-foreground">
                      {group.teamName}
                    </div>
                    {group.projects.map((project) => (
                      <SelectItem key={project.id} value={project.id} className="pl-6">
                        {project.name}
                      </SelectItem>
                    ))}
                  </div>
                ))}
              </SelectContent>
            </Select>
          )}
        </div>

        <div className="grid gap-2">
          <Label className="text-xs font-medium">阈值设置</Label>
          <div className="flex flex-wrap gap-2">
            {PRESET_THRESHOLDS.map((preset) => {
              const isActive = JSON.stringify(thresholds) === JSON.stringify(preset.value)
              return (
                <button
                  key={preset.label}
                  type="button"
                  onClick={() => selectPreset(preset.value)}
                  className={cn(
                    'rounded-md border px-2.5 py-1 text-xs',
                    isActive
                      ? 'border-primary bg-primary/8 text-primary'
                      : 'border-border text-muted-foreground hover:bg-muted',
                  )}
                >
                  {preset.label}
                </button>
              )
            })}
          </div>
          <div className="flex flex-wrap items-center gap-1.5">
            {thresholds.map((threshold) => (
              <Badge key={threshold} variant="outline" className="gap-1 tabular-nums">
                {threshold}%
                <button
                  type="button"
                  onClick={() => removeThreshold(threshold)}
                  className="ml-0.5"
                  aria-label={`移除 ${threshold}%`}
                >
                  <X className="size-3" />
                </button>
              </Badge>
            ))}
            <div className="flex items-center gap-1">
              <Input
                type="number"
                min={1}
                max={100}
                value={customThreshold}
                onChange={(event) => setCustomThreshold(event.target.value)}
                onKeyDown={(event) => {
                  if (event.key === 'Enter') {
                    event.preventDefault()
                    addCustomThreshold()
                  }
                }}
                placeholder="%"
                className="h-6 w-14 text-xs tabular-nums"
              />
              <Button
                type="button"
                variant="ghost"
                size="icon"
                className="size-6"
                aria-label="添加阈值"
                onClick={addCustomThreshold}
              >
                <Plus className="size-3" />
              </Button>
            </div>
          </div>
        </div>

        <div className="grid gap-2">
          <Label className="text-xs font-medium">通知角色</Label>
          <div className="grid grid-cols-2 gap-2">
            {roles.map((role) => (
              <label
                key={role.id}
                className="flex cursor-pointer items-center gap-2 rounded-md border border-border px-3 py-2 text-sm hover:bg-muted/50"
              >
                <Checkbox
                  checked={notifyRoleIds.includes(role.id)}
                  onCheckedChange={() => toggleRole(role.id)}
                />
                <span className="truncate">{role.name}</span>
              </label>
            ))}
          </div>
        </div>
      </div>

      {error && <p className="text-xs text-red-600">{error}</p>}

      <DialogFooter>
        <Button variant="outline" size="sm" onClick={() => onOpenChange(false)} disabled={saving}>
          取消
        </Button>
        <Button size="sm" onClick={handleSave} disabled={saving}>
          {saving ? '保存中…' : rule ? '保存' : '创建'}
        </Button>
      </DialogFooter>
    </DialogContent>
  )
}
