import { useCallback, useRef, useState } from 'react'
import type { ModelInfo } from '@/api/types'
import { isCustomModel } from '../lib/model-kind'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import { PROVIDER_LABELS } from '@/lib/provider-labels'
import { cn } from '@/lib/utils'
import { Pencil, Trash2 } from 'lucide-react'

interface ModelListTableProps {
  models: ModelInfo[]
  canManage: boolean
  showActions?: boolean
  showProviderColumn?: boolean
  rowClass: (id: string) => string | undefined
  onToggle: (model: ModelInfo) => void | Promise<void>
  onEdit: (model: ModelInfo) => void
  onDelete: (model: ModelInfo) => void
}

export function ModelListTable({
  models,
  canManage,
  showActions = true,
  showProviderColumn = true,
  rowClass,
  onToggle,
  onEdit,
  onDelete,
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
    <Table>
      <TableHeader>
        <TableRow className="hover:bg-transparent border-border/60">
          <TableHead className="text-xs font-medium text-muted-foreground">
            模型
          </TableHead>
          {showProviderColumn && (
            <TableHead className="text-xs font-medium text-muted-foreground">
              来源
            </TableHead>
          )}
          <TableHead className="text-right text-xs font-medium text-muted-foreground">
            输入价格
          </TableHead>
          <TableHead className="text-right text-xs font-medium text-muted-foreground">
            输出价格
          </TableHead>
          {showProviderColumn && (
            <TableHead className="text-xs font-medium text-muted-foreground">
              部署地址
            </TableHead>
          )}
          {canManage && (
            <TableHead className="w-[140px] text-right text-xs font-medium text-muted-foreground">
              操作
            </TableHead>
          )}
        </TableRow>
      </TableHeader>
      <TableBody>
        {models.map((model) => (
          <TableRow
            key={model.modelId}
            className={cn(
              'group transition-colors',
              !model.enabled && 'opacity-45',
              rowClass(model.modelId),
            )}
          >
            {/* Model name + type combined cell */}
            <TableCell>
              <div className="flex flex-col gap-0.5">
                <span className="text-sm font-medium leading-tight text-foreground">
                  {model.name}
                </span>
                <span className="font-mono text-xs text-muted-foreground/80">
                  {model.type}
                </span>
              </div>
            </TableCell>
            {showProviderColumn && (
              <TableCell>
                <Badge
                  variant="outline"
                  className={cn(
                    'text-[11px] font-medium',
                    isCustomModel(model)
                      ? 'border-violet-200 bg-violet-50 text-violet-700'
                      : 'border-indigo-200 bg-indigo-50 text-indigo-700',
                  )}
                >
                  {isCustomModel(model) ? (PROVIDER_LABELS[model.provider] ?? '自定义') : '内置'}
                </Badge>
              </TableCell>
            )}
            <TableCell className="text-right">
              <span className="font-mono text-xs tabular-nums text-foreground/80">
                {model.inputPrice > 0 ? `¥${model.inputPrice}` : '—'}
              </span>
              {model.inputPrice > 0 && (
                <span className="ml-0.5 text-[10px] text-muted-foreground">/M</span>
              )}
            </TableCell>
            <TableCell className="text-right">
              <span className="font-mono text-xs tabular-nums text-foreground/80">
                {model.outputPrice > 0 ? `¥${model.outputPrice}` : '—'}
              </span>
              {model.outputPrice > 0 && (
                <span className="ml-0.5 text-[10px] text-muted-foreground">/M</span>
              )}
            </TableCell>
            {showProviderColumn && (
              <TableCell className="max-w-[200px]">
                <span className="block truncate font-mono text-xs text-muted-foreground">
                  {isCustomModel(model) ? (model.endpoint ?? '—') : '—'}
                </span>
              </TableCell>
            )}
            {canManage && (
              <TableCell className="text-right">
                <div className="flex items-center justify-end gap-1.5 opacity-70 transition-opacity group-hover:opacity-100">
                  <Switch
                    checked={model.enabled}
                    disabled={togglingIds.has(model.modelId)}
                    onCheckedChange={() => void handleToggle(model)}
                    aria-label={model.enabled ? '禁用模型' : '启用模型'}
                  />
                  {showActions && isCustomModel(model) && (
                    <>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="size-7"
                        onClick={() => onEdit(model)}
                        aria-label="编辑"
                      >
                        <Pencil className="size-3.5" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="size-7 text-red-600 hover:bg-red-50 hover:text-red-700"
                        onClick={() => onDelete(model)}
                        aria-label="删除"
                      >
                        <Trash2 className="size-3.5" />
                      </Button>
                    </>
                  )}
                </div>
              </TableCell>
            )}
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}
