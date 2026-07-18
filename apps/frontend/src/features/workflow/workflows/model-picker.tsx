import { useEffect, useMemo, useState } from 'react'
import type { ModelInfo } from '@/api/types'
import { useInjectedApis } from '@/api/use-apis'
import type { WorkflowComponentProps } from '../types'
import { WorkflowPanelChrome, WorkflowPanelFooter } from '../components/workflow-panel-chrome'
import { WorkflowFormLayout } from '../components/workflow-form-layout'
import { WORKFLOW_LIST_ITEM_CLASS, WORKFLOW_LIST_ITEM_SELECTED_CLASS } from '../constants'
import { Input } from '@/components/ui/input'
import { Checkbox } from '@/components/ui/checkbox'
import { cn } from '@/lib/utils'
import { isBuiltinModel } from '@/features/models'

export function ModelPickerWorkflow({
  entry,
  onPop,
  onClose,
  onSetDirty,
}: WorkflowComponentProps<'model-picker'>) {
  const apis = useInjectedApis()
  const selectedModelIds = useMemo(
    () => (entry.payload.selectedModelIds as string[]) ?? [],
    [entry.payload.selectedModelIds],
  )
  const parentAllowedModelIds =
    (entry.payload.parentAllowedModelIds as string[] | undefined) ?? undefined
  const onConfirm = entry.payload.onConfirm as ((modelIds: string[]) => void) | undefined
  const [models, setModels] = useState<ModelInfo[]>([])
  const [search, setSearch] = useState('')
  const [pickedIds, setPickedIds] = useState<string[]>(() => [...selectedModelIds])

  useEffect(() => {
    apis.modelApi.list().then((list) => {
      let enabled = list.filter((m) => m.enabled)
      if (parentAllowedModelIds?.length) {
        const allowed = new Set(parentAllowedModelIds)
        enabled = enabled.filter((m) => allowed.has(m.modelId))
      }
      setModels(enabled)
      setPickedIds((prev) => (prev.length > 0 ? prev : [...selectedModelIds]))
    })
  }, [apis, parentAllowedModelIds, selectedModelIds])

  const filtered = models.filter(
    (m) =>
      m.name.toLowerCase().includes(search.toLowerCase()) ||
      m.type.toLowerCase().includes(search.toLowerCase()),
  )

  const toggle = (modelId: string) => {
    setPickedIds((prev) =>
      prev.includes(modelId) ? prev.filter((id) => id !== modelId) : [...prev, modelId],
    )
    onSetDirty(true)
  }

  const handleConfirm = () => {
    onConfirm?.(pickedIds)
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
          primaryLabel={`确认 (${pickedIds.length})`}
          onPrimary={handleConfirm}
          primaryDisabled={pickedIds.length === 0}
        />
      }
    >
      <WorkflowFormLayout variant="full">
        <Input
          placeholder="搜索模型..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
        <div className="space-y-1 max-h-[60vh] overflow-y-auto">
          {filtered.map((m) => (
            <label
              key={m.modelId}
              className={cn(
                'flex items-center gap-3 px-4 py-3 cursor-pointer',
                WORKFLOW_LIST_ITEM_CLASS,
                pickedIds.includes(m.modelId) && WORKFLOW_LIST_ITEM_SELECTED_CLASS,
                !m.enabled && 'opacity-50',
              )}
            >
              <Checkbox
                checked={pickedIds.includes(m.modelId)}
                disabled={!m.enabled}
                onCheckedChange={() => toggle(m.modelId)}
              />
              <div className="flex-1">
                <div className="font-medium text-sm">{m.name}</div>
                <div className="text-xs text-muted-foreground">
                  {m.type}
                  {isBuiltinModel(m) ? ` · ${m.provider}` : ''}
                </div>
              </div>
            </label>
          ))}
        </div>
      </WorkflowFormLayout>
    </WorkflowPanelChrome>
  )
}
