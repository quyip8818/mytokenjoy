import { useCallback, useRef, useState } from 'react'
import type { ModelInfo } from '@/api/types'
import { isCustomModel } from '../lib/model-kind'
import { ModelTable } from '@/components/ui/model-table'
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

  return (
    <ModelTable
      models={models}
      extraColumns={[
        {
          header: '状态',
          render: (model) => (
            <Badge variant={model.active ? 'default' : 'outline'}>
              {model.active ? '启用' : '禁用'}
            </Badge>
          ),
        },
      ]}
      rowClass={(model) => rowClass(model.modelId)}
      renderActions={
        canManage
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
                <button
                  onClick={() => void handleToggle(model)}
                  disabled={togglingIds.has(model.modelId)}
                  className={`rounded p-1.5 hover:bg-muted ${model.active ? 'text-amber-500' : 'text-green-500'}`}
                  title={model.active ? '禁用' : '启用'}
                >
                  <Power className="h-3.5 w-3.5" />
                </button>
              </div>
            )
          : undefined
      }
    />
  )
}
