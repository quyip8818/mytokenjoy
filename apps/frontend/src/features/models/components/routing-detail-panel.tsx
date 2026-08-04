import { useCallback, useMemo, useState } from 'react'
import type { Department, ModelInfo, RoutingRule } from '@/api/types'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Save, Shield, ArrowDownUp } from 'lucide-react'

interface RoutingDetailPanelProps {
  department: Department
  rule: RoutingRule
  parentRule: RoutingRule | undefined
  models: ModelInfo[]
  saving: boolean
  onSave: (input: {
    inherited: boolean
    allowedModelIds: string[]
    defaultModelId: string | null
    fallbackModelId: string | null
  }) => Promise<void>
}

export function RoutingDetailPanel({
  department,
  rule,
  parentRule,
  models,
  saving,
  onSave,
}: RoutingDetailPanelProps) {
  const [inherited, setInherited] = useState(rule.inherited)
  const [enabledIds, setEnabledIds] = useState<Set<string>>(new Set(rule.allowedModelIds ?? []))
  const [defaultModelId, setDefaultModelId] = useState<string | null>(rule.defaultModelId ?? null)
  const [fallbackModelId, setFallbackModelId] = useState<string | null>(
    rule.fallbackModelId ?? null,
  )

  const [prevRuleId, setPrevRuleId] = useState(rule.id)
  if (rule.id !== prevRuleId) {
    setPrevRuleId(rule.id)
    setInherited(rule.inherited)
    setEnabledIds(new Set(rule.allowedModelIds ?? []))
    setDefaultModelId(rule.defaultModelId ?? null)
    setFallbackModelId(rule.fallbackModelId ?? null)
  }

  const parentModelIds = useMemo(
    () => new Set(parentRule?.allowedModelIds ?? rule.allowedModelIds ?? []),
    [parentRule, rule],
  )

  const effectiveIds = inherited ? Array.from(parentModelIds) : Array.from(enabledIds)

  const toggleModel = useCallback((modelId: string) => {
    setEnabledIds((prev) => {
      const next = new Set(prev)
      if (next.has(modelId)) next.delete(modelId)
      else next.add(modelId)
      return next
    })
  }, [])

  const handleSave = () => {
    void onSave({
      inherited,
      allowedModelIds: effectiveIds,
      defaultModelId,
      fallbackModelId,
    })
  }

  const enabledModels = models.filter((m) => effectiveIds.includes(m.modelId))

  return (
    <div className="flex flex-1 flex-col overflow-y-auto">
      {/* Header */}
      <div className="sticky top-0 z-10 flex items-center justify-between border-b border-border/60 bg-card px-5 py-3">
        <h3 className="text-sm font-semibold text-foreground">{department.name}</h3>
        <Button
          className="gap-1.5"
          onClick={handleSave}
          disabled={saving}
          disabledReason={saving ? '保存中…' : undefined}
        >
          <Save className="size-3.5" />
          {saving ? '保存中...' : '保存'}
        </Button>
      </div>

      <div className="flex flex-col gap-4 p-5">
        {/* Inherit toggle */}
        <div className="flex items-center justify-between rounded-lg border border-border/60 px-4 py-3">
          <Label className="text-sm font-medium">
            继承父级配置
            <span className="ml-2 text-xs font-normal text-muted-foreground">
              ({parentModelIds.size} 个模型)
            </span>
          </Label>
          <Switch checked={inherited} onCheckedChange={setInherited} />
        </div>

        {/* Model list with switches */}
        {!inherited && (
          <div className="divide-y divide-border/60 rounded-lg border border-border/60">
            {models.map((model) => {
              const isEnabled = enabledIds.has(model.modelId)
              const isDefault = defaultModelId === model.modelId
              const isFallback = fallbackModelId === model.modelId
              return (
                <div key={model.modelId} className="flex items-center justify-between px-4 py-2.5">
                  <div className="flex items-center gap-2">
                    <span className="text-sm text-foreground">{model.name || model.type}</span>
                    {isDefault && (
                      <Badge
                        variant="outline"
                        className="border-indigo-200 bg-indigo-50 text-[10px] text-indigo-700"
                      >
                        默认
                      </Badge>
                    )}
                    {isFallback && (
                      <Badge
                        variant="outline"
                        className="border-amber-200 bg-amber-50 text-[10px] text-amber-700"
                      >
                        降级
                      </Badge>
                    )}
                  </div>
                  <Switch
                    checked={isEnabled}
                    onCheckedChange={() => toggleModel(model.modelId)}
                    aria-label={`${model.name || model.type} 开关`}
                  />
                </div>
              )
            })}
          </div>
        )}

        {inherited && (
          <div className="rounded-lg border border-dashed border-border/60 px-4 py-6 text-center">
            <p className="text-sm text-muted-foreground">
              继承父级配置，共 {parentModelIds.size} 个可用模型
            </p>
          </div>
        )}

        {/* Default & fallback */}
        {!inherited && (
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <div className="flex items-center gap-1.5">
                <Shield className="size-3.5 text-indigo-500" />
                <Label className="text-sm font-medium">默认模型</Label>
              </div>
              <Select
                value={defaultModelId ?? 'none'}
                onValueChange={(v) => setDefaultModelId(v === 'none' ? null : v)}
              >
                <SelectTrigger className="h-9">
                  <SelectValue placeholder="未设置" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="none">未设置</SelectItem>
                  {enabledModels.map((m) => (
                    <SelectItem key={m.modelId} value={m.modelId}>
                      {m.name || m.type}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-1.5">
              <div className="flex items-center gap-1.5">
                <ArrowDownUp className="size-3.5 text-amber-500" />
                <Label className="text-sm font-medium">降级模型</Label>
              </div>
              <Select
                value={fallbackModelId ?? 'none'}
                onValueChange={(v) => setFallbackModelId(v === 'none' ? null : v)}
              >
                <SelectTrigger className="h-9">
                  <SelectValue placeholder="未设置" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="none">未设置</SelectItem>
                  {enabledModels.map((m) => (
                    <SelectItem key={m.modelId} value={m.modelId}>
                      {m.name || m.type}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
