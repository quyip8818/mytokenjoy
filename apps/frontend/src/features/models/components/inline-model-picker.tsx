import { useMemo, useState } from 'react'
import { Search, X } from 'lucide-react'
import type { ModelInfo } from '@/api/types'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Label } from '@/components/ui/label'

const PROVIDER_LABELS: Record<string, string> = {
  openai: 'OpenAI',
  anthropic: 'Anthropic',
  deepseek: 'DeepSeek',
  qwen: '通义千问',
  custom: '自定义',
}

export interface InlineModelPickerProps {
  /** Currently selected model IDs */
  value: string[]
  /** Called when selection changes */
  onChange: (ids: string[]) => void
  /** Available models to display (caller fetches these) */
  models: ModelInfo[]
  /** Label shown above the picker */
  label?: string
  /** Hint text shown next to the label */
  hint?: string
}

export function InlineModelPicker({
  value,
  onChange,
  models,
  label = '模型白名单',
  hint,
}: InlineModelPickerProps) {
  const [search, setSearch] = useState('')

  const allIds = useMemo(() => models.map((m) => m.modelId), [models])

  const groupedModels = useMemo(() => {
    const groups = new Map<string, ModelInfo[]>()
    for (const model of models) {
      const provider = model.provider ?? 'custom'
      const matchesSearch =
        !search ||
        model.name.toLowerCase().includes(search.toLowerCase()) ||
        model.type.toLowerCase().includes(search.toLowerCase())
      if (!matchesSearch) continue
      if (!groups.has(provider)) groups.set(provider, [])
      groups.get(provider)!.push(model)
    }
    return groups
  }, [models, search])

  const toggle = (modelId: string) => {
    onChange(
      value.includes(modelId) ? value.filter((id) => id !== modelId) : [...value, modelId],
    )
  }

  const toggleAll = () => {
    onChange(value.length === allIds.length ? [] : [...allIds])
  }

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <Label>
          {label}
          {hint && (
            <span className="ml-2 text-xs font-normal text-muted-foreground">{hint}</span>
          )}
          {value.length > 0 && (
            <span className="ml-2 text-xs font-normal text-muted-foreground">
              已选 {value.length}/{allIds.length}
            </span>
          )}
        </Label>
        {allIds.length > 0 && (
          <Button variant="ghost" size="sm" className="h-6 px-2 text-xs" onClick={toggleAll}>
            {value.length === allIds.length ? '取消全选' : '全选'}
          </Button>
        )}
      </div>

      {/* Search */}
      <div className="relative">
        <Search className="absolute left-2.5 top-2.5 h-3.5 w-3.5 text-muted-foreground" />
        <Input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="搜索模型..."
          className="h-8 pl-8 pr-8 text-sm"
        />
        {search && (
          <button
            className="absolute right-2 top-2 text-muted-foreground hover:text-foreground"
            onClick={() => setSearch('')}
          >
            <X className="h-3.5 w-3.5" />
          </button>
        )}
      </div>

      {/* Grouped model list */}
      <div className="max-h-56 overflow-y-auto rounded-md border">
        {models.length === 0 ? (
          <p className="p-3 text-sm text-muted-foreground">加载中...</p>
        ) : groupedModels.size === 0 ? (
          <p className="p-3 text-sm text-muted-foreground">无匹配模型</p>
        ) : (
          Array.from(groupedModels.entries()).map(([provider, items]) => (
            <div key={provider}>
              <div className="sticky top-0 z-10 bg-muted/80 backdrop-blur-sm px-3 py-1.5 text-xs font-medium text-muted-foreground border-b">
                {PROVIDER_LABELS[provider] ?? provider}
              </div>
              <div className="divide-y">
                {items.map((model) => (
                  <label
                    key={model.modelId}
                    className="flex items-center gap-3 px-3 py-2 cursor-pointer hover:bg-accent/50 transition-colors"
                  >
                    <Checkbox
                      checked={value.includes(model.modelId)}
                      onCheckedChange={() => toggle(model.modelId)}
                    />
                    <div className="flex-1 min-w-0">
                      <span className="text-sm truncate block">{model.name}</span>
                      {model.type !== model.name && (
                        <span className="text-xs text-muted-foreground truncate block">
                          {model.type}
                        </span>
                      )}
                    </div>
                  </label>
                ))}
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  )
}
