import { useEffect, useState } from 'react'
import { modelApi } from '@/api/models'
import type { ModelInfo } from '@/api/types'
import type { WorkflowComponentProps } from '../types'
import { WorkflowPanelChrome, WorkflowPanelFooter } from '../components/workflow-panel-chrome'
import { Input } from '@/components/ui/input'
import { Checkbox } from '@/components/ui/checkbox'
import { cn } from '@/lib/utils'

export function ModelPickerWorkflow({ entry, onPop, onClose, onSetDirty }: WorkflowComponentProps) {
  const selected = (entry.payload.selectedModels as string[]) ?? []
  const parentWhitelist = (entry.payload.parentWhitelist as string[] | undefined) ?? undefined
  const onConfirm = entry.payload.onConfirm as ((models: string[]) => void) | undefined
  const [models, setModels] = useState<ModelInfo[]>([])
  const [search, setSearch] = useState('')
  const [picked, setPicked] = useState<string[]>(selected)

  useEffect(() => {
    modelApi.list().then((list) => {
      let enabled = list.filter((m) => m.enabled)
      if (parentWhitelist?.length) {
        enabled = enabled.filter((m) => parentWhitelist.includes(m.name))
      }
      setModels(enabled)
    })
  }, [parentWhitelist])

  const filtered = models.filter(
    (m) =>
      m.displayName.toLowerCase().includes(search.toLowerCase()) ||
      m.name.toLowerCase().includes(search.toLowerCase()),
  )

  const toggle = (name: string) => {
    setPicked((prev) => (prev.includes(name) ? prev.filter((n) => n !== name) : [...prev, name]))
    onSetDirty(true)
  }

  const handleConfirm = () => {
    onConfirm?.(picked)
    onPop()
  }

  return (
    <WorkflowPanelChrome
      title="选择模型"
      showBack
      onBack={onPop}
      onClose={onClose}
      footer={
        <WorkflowPanelFooter
          onCancel={onPop}
          primaryLabel={`确认 (${picked.length})`}
          onPrimary={handleConfirm}
          primaryDisabled={picked.length === 0}
        />
      }
    >
      <div className="space-y-4">
        <Input
          placeholder="搜索模型..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
        <div className="space-y-1 max-h-[60vh] overflow-y-auto">
          {filtered.map((m) => (
            <label
              key={m.id}
              className={cn(
                'flex items-center gap-3 rounded-lg border border-border/50 px-4 py-3 cursor-pointer hover:bg-indigo-50/30',
                picked.includes(m.name) && 'border-indigo-200 bg-indigo-50/40',
              )}
            >
              <Checkbox checked={picked.includes(m.name)} onCheckedChange={() => toggle(m.name)} />
              <div className="flex-1">
                <div className="font-medium text-sm">{m.displayName}</div>
                <div className="text-xs text-muted-foreground">{m.name}</div>
              </div>
            </label>
          ))}
        </div>
      </div>
    </WorkflowPanelChrome>
  )
}
