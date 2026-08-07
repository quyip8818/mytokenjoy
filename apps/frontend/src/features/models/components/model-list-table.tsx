import { useCallback, useRef, useState } from 'react'
import type { DiscountEntry, ModelInfo } from '@/api/types'
import { isCustomModel } from '../lib/model-kind'
import { ModelTable } from './model-table'
import { Badge } from '@/components/ui/badge'
import { Pencil, Power } from 'lucide-react'

function formatDiscount(d: number): string {
  if (d === 1) return '—'
  if (d < 1) return `${(d * 10).toFixed(1).replace(/\.0$/, '')}折`
  return `加价${Math.round((d - 1) * 100)}%`
}

interface ModelListTableProps {
  models: ModelInfo[]
  canManage: boolean
  discountMap: Map<string, DiscountEntry>
  rowClass: (id: string) => string | undefined
  onToggle: (model: ModelInfo) => void | Promise<void>
  onEdit: (model: ModelInfo) => void
}

export function ModelListTable({
  models,
  canManage,
  discountMap,
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

  // ponytail: hasAnyDiscount gates column render — no discount data = no column
  const hasAnyDiscount = discountMap.size > 0

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
        ...(hasAnyDiscount
          ? [
              {
                header: '折扣',
                render: (model: ModelInfo) => {
                  const entry = discountMap.get(model.type) ?? discountMap.get('*')
                  if (!entry || entry.discount === 1)
                    return <span className="text-muted-foreground">—</span>
                  return (
                    <span className="text-xs tabular-nums">{formatDiscount(entry.discount)}</span>
                  )
                },
              },
            ]
          : []),
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
