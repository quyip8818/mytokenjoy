import type { BudgetNode, BudgetProjectView, Role } from '@/api/types'
import { groupProjectsByTeam } from '@/features/budget/lib/alerts'
import { ALERT_PRESET_THRESHOLDS } from '@/features/budget/lib/constants'
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
import type { useAlertRuleForm } from '../hooks/use-alert-rule-form'

type AlertRuleFormFieldsProps = {
  tree: BudgetNode[]
  projects: BudgetProjectView[]
  roles: Role[]
  form: ReturnType<typeof useAlertRuleForm>
}

export function AlertRuleFormFields({ tree, projects, roles, form }: AlertRuleFormFieldsProps) {
  const {
    state,
    customThreshold,
    setCustomThreshold,
    setTargetTypeAndReset,
    setTarget,
    addCustomThreshold,
    removeThreshold,
    toggleRole,
    selectPreset,
  } = form

  return (
    <div className="grid gap-5 py-2">
      <div className="grid gap-1.5">
        <Label className="text-xs font-medium">监控类型</Label>
        <RadioGroup
          value={state.targetType}
          onValueChange={(value) => setTargetTypeAndReset(value as 'team' | 'project')}
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
        {state.targetType === 'team' ? (
          <DepartmentTreeSelect
            tree={tree}
            value={state.targetId}
            onChange={setTarget}
            placeholder="选择团队…"
          />
        ) : (
          <Select
            value={state.targetId}
            onValueChange={(value) => {
              const project = projects.find((item) => item.id === value)
              setTarget(value, project?.name ?? '')
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
          {ALERT_PRESET_THRESHOLDS.map((preset) => {
            const isActive = JSON.stringify(state.thresholds) === JSON.stringify(preset.value)
            return (
              <button
                key={preset.label}
                type="button"
                onClick={() => selectPreset([...preset.value])}
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
          {state.thresholds.map((threshold) => (
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
                checked={state.notifyRoleIds.includes(role.id)}
                onCheckedChange={() => toggleRole(role.id)}
              />
              <span className="truncate">{role.name}</span>
            </label>
          ))}
        </div>
      </div>
    </div>
  )
}
