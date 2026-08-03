import { useCallback, useRef, useState } from 'react'
import type { ModelInfo } from '@/api/types'
import { isCustomModel } from '../lib/model-kind'
import { ModelTable } from './model-table'
import { Badge } from '@/components/ui/badge'
import { Pencil, Power } from 'lucide-react'

interface ModelListTableProps {
  models: ModelInfo[]
  canManage: boolean
  rowClass: (id: string) => string | undefined
  onToggle: (model: ModelInfo) => void | Promise<void>
  onEdit: (model: ModelInfo) => void
}

export function ModelListTable({
  models,
  canManage,
  rowClass,
  onToggle,
  onEdit,
}: ModelListTableProps) {
  const [togglingIds, setTogglingIds] = useState<Set<string>>(new Set())
  const inflightRef = useRef<Set<string>>(new Set())

  const handleToggle = useCallback(
    async (model: ModelInfo) => {
      if (inflightRef.current.has(model.modelId)) return
      inflightRef.current.add(model.modelId)
      setTogglingIds(new Set(inflightRef.current))
      try {
        await onToggle(model)
      } finally {
        inflightRef.current.delete(model.modelId)
        setTogglingIds(new Set(inflightRef.current))
      }
    },
    [onToggle],
  )

  // ponytail: only custom models can be toggled — global models are read-only on SaaS
  const canToggle = (model: ModelInfo) => isCustomModel(model)

  return (
    <ModelTable
      models={models}
      extraColumns={[
        {
          header: '状态',
          render: (model) => (
            <Badge variant={model.deprecated ? 'outline' : 'default'}>
              {model.deprecated ? '已下线' : '启用'}
            </Badge>
          ),
        },
      ]}
      rowClass={(model) => rowClass(model.modelId)}
      renderActions={
        canManage && models.some(isCustomModel)
          ? (model) => (
              <div className="inline-flex items-center gap-1">
                {isCustomModel(model) && (
                  <button
                    onClick={() => onEdit(model)}
                    className="rounded p-1.5 text-muted-foreground hover:bg-muted hover:text-foreground"
                    title="编辑"
                  >
                    <Pencil className="h-3.5 w-3.5" />
                  </button>
                )}
                {canToggle(model) && (
                  <button
                    onClick={() => void handleToggle(model)}
                    disabled={togglingIds.has(model.modelId)}
                    className={`rounded p-1.5 hover:bg-muted ${model.deprecated ? 'text-green-500' : 'text-amber-500'}`}
                    title={model.deprecated ? '恢复' : '下线'}
                  >
                    <Power className="h-3.5 w-3.5" />
                  </button>
                )}
              </div>
            )
          : undefined
      }
    />
  )
}
