import { useCallback, useEffect, useMemo, useState } from 'react'
import type { Department, ModelInfo, RoutingRule } from '@/api/types'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
import { Checkbox } from '@/components/ui/checkbox'
import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { cn } from '@/lib/utils'
import { PROVIDER_LABELS, PROVIDER_CHIP_STYLES } from '@/lib/provider-labels'
import { Save, Shield, ArrowDownUp } from 'lucide-react'

interface RoutingDetailPanelProps {
  department: Department
  rule: RoutingRule
  parentRule: RoutingRule | undefined
  models: ModelInfo[]
  saving: boolean
  onSave: (input: {
    inherited: boolean
    allowedModelIds: number[]
    defaultModelId: number | null
    fallbackModelId: number | null
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
  const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set(rule.allowedModelIds))
  const [defaultModelId, setDefaultModelId] = useState<number | null>(
    rule.defaultModelId ?? null,
  )
  const [fallbackModelId, setFallbackModelId] = useState<number | null>(
    rule.fallbackModelId ?? null,
  )

  // Reset state when rule changes (switching nodes)
  useEffect(() => {
    setInherited(rule.inherited)
    setSelectedIds(new Set(rule.allowedModelIds))
    setDefaultModelId(rule.defaultModelId ?? null)
    setFallbackModelId(rule.fallbackModelId ?? null)
  }, [rule])

  // Group models by provider
  const groupedModels = useMemo(() => {
    const groups = new Map<string, ModelInfo[]>()
    for (const model of models) {
      const list = groups.get(model.provider) ?? []
      list.push(model)
      groups.set(model.provider, list)
    }
    return Array.from(groups.entries()).map(([provider, items]) => ({
      provider,
      label: PROVIDER_LABELS[provider] ?? provider,
      models: items,
    }))
  }, [models])

  const parentModelIds = useMemo(
    () => new Set(parentRule?.allowedModelIds ?? rule.allowedModelIds),
    [parentRule, rule],
  )

  const effectiveIds = inherited ? Array.from(parentModelIds) : Array.from(selectedIds)

  const toggleModel = useCallback((modelId: number) => {
    setSelectedIds((prev) => {
      const next = new Set(prev)
      if (next.has(modelId)) next.delete(modelId)
      else next.add(modelId)
      return next
    })
  }, [])

  const selectAll = useCallback(() => {
    setSelectedIds(new Set(models.map((m) => m.modelId)))
  }, [models])

  const deselectAll = useCallback(() => {
    setSelectedIds(new Set())
  }, [])

  const handleSave = () => {
    void onSave({
      inherited,
      allowedModelIds: effectiveIds,
      defaultModelId,
      fallbackModelId,
    })
  }

  const selectedModels = models.filter((m) => effectiveIds.includes(m.modelId))

  return (
    <div className="flex flex-1 flex-col gap-5 overflow-y-auto p-5">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-semibold text-foreground">{department.name}</h3>
        <Button
          size="sm"
          className="gap-1.5"
          onClick={handleSave}
          disabled={saving}
        >
          <Save className="size-3.5" />
          {saving ? '保存中...' : '保存'}
        </Button>
      </div>

      {/* Inherit toggle */}
      <div className="flex items-center justify-between rounded-xl border border-border p-4">
        <div>
          <Label className="text-sm font-medium">继承父级配置</Label>
          <p className="mt-0.5 text-xs text-muted-foreground">
            开启后使用父级的模型配置（{parentModelIds.size} 个模型），关闭可自定义
          </p>
        </div>
        <Switch checked={inherited} onCheckedChange={setInherited} />
      </div>

      {/* Model selection */}
      {!inherited && (
        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <Label className="text-sm font-medium">
              可用模型 <span className="text-muted-foreground">({selectedIds.size})</span>
            </Label>
            <div className="flex gap-2">
              <Button variant="ghost" size="sm" className="h-7 text-xs" onClick={selectAll}>
                全选
              </Button>
              <Button variant="ghost" size="sm" className="h-7 text-xs" onClick={deselectAll}>
                清空
              </Button>
            </div>
          </div>

          <div className="space-y-4 rounded-xl border border-border p-4">
            {groupedModels.map((group) => (
              <div key={group.provider}>
                <div className="mb-2 flex items-center gap-2">
                  <Badge
                    variant="outline"
                    className={cn('text-xs', PROVIDER_CHIP_STYLES[group.provider])}
                  >
                    {group.label}
                  </Badge>
                  <span className="text-xs text-muted-foreground">
                    {group.models.filter((m) => selectedIds.has(m.modelId)).length}/{group.models.length}
                  </span>
                </div>
                <div className="grid grid-cols-2 gap-2">
                  {group.models.map((model) => (
                    <label
                      key={model.modelId}
                      className={cn(
                        'flex cursor-pointer items-center gap-2.5 rounded-lg border px-3 py-2 text-sm transition-colors',
                        selectedIds.has(model.modelId)
                          ? 'border-primary/30 bg-primary/5'
                          : 'border-border hover:bg-muted/50',
                      )}
                    >
                      <Checkbox
                        checked={selectedIds.has(model.modelId)}
                        onCheckedChange={() => toggleModel(model.modelId)}
                      />
                      <span className="truncate">{model.name || model.type}</span>
                    </label>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {inherited && (
        <div className="rounded-xl border border-dashed border-border p-6 text-center">
          <p className="text-sm text-muted-foreground">
            当前继承父级配置，共 {parentModelIds.size} 个可用模型
          </p>
        </div>
      )}

      {/* Default & fallback model */}
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <div className="flex items-center gap-1.5">
            <Shield className="size-3.5 text-muted-foreground" />
            <Label className="text-sm font-medium">默认模型</Label>
          </div>
          <Select
            value={defaultModelId ? String(defaultModelId) : 'none'}
            onValueChange={(v) => setDefaultModelId(v === 'none' ? null : Number(v))}
          >
            <SelectTrigger className="h-9">
              <SelectValue placeholder="未设置" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="none">未设置</SelectItem>
              {selectedModels.map((m) => (
                <SelectItem key={m.modelId} value={String(m.modelId)}>
                  {m.name || m.type}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <div className="space-y-2">
          <div className="flex items-center gap-1.5">
            <ArrowDownUp className="size-3.5 text-muted-foreground" />
            <Label className="text-sm font-medium">降级模型</Label>
          </div>
          <Select
            value={fallbackModelId ? String(fallbackModelId) : 'none'}
            onValueChange={(v) => setFallbackModelId(v === 'none' ? null : Number(v))}
          >
            <SelectTrigger className="h-9">
              <SelectValue placeholder="未设置" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="none">未设置</SelectItem>
              {selectedModels.map((m) => (
                <SelectItem key={m.modelId} value={String(m.modelId)}>
                  {m.name || m.type}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>
    </div>
  )
}
