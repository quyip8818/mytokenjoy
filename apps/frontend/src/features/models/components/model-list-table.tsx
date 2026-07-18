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
        <TableRow className="hover:bg-transparent">
          <TableHead className="text-xs font-medium uppercase text-muted-foreground">
            模型名称
          </TableHead>
          <TableHead className="text-xs font-medium uppercase text-muted-foreground">
            模型类型
          </TableHead>
          {showProviderColumn && (
            <TableHead className="text-xs font-medium uppercase text-muted-foreground">
              来源
            </TableHead>
          )}
          <TableHead className="text-xs font-medium uppercase text-muted-foreground">
            描述
          </TableHead>
          {showProviderColumn && (
            <TableHead className="text-xs font-medium uppercase text-muted-foreground">
              部署地址
            </TableHead>
          )}
          {canManage && (
            <TableHead className="text-xs font-medium uppercase text-muted-foreground">
              操作
            </TableHead>
          )}
        </TableRow>
      </TableHeader>
      <TableBody>
        {models.map((model) => (
          <TableRow
            key={model.modelId}
            className={`even:bg-muted/40 ${rowClass(model.modelId)} ${!model.enabled ? 'opacity-50' : ''}`}
          >
            <TableCell className="font-medium">{model.name}</TableCell>
            <TableCell className="font-mono text-xs text-muted-foreground">{model.type}</TableCell>
            {showProviderColumn && (
              <TableCell>
                <Badge variant="outline" className="border-0 bg-muted text-xs">
                  {isCustomModel(model) ? (PROVIDER_LABELS[model.provider] ?? '自定义') : '内置'}
                </Badge>
              </TableCell>
            )}
            <TableCell className="max-w-xs truncate text-sm text-muted-foreground">
              {model.description || '—'}
            </TableCell>
            {showProviderColumn && (
              <TableCell className="max-w-xs truncate font-mono text-xs text-muted-foreground">
                {isCustomModel(model) ? (model.endpoint ?? '—') : '—'}
              </TableCell>
            )}
            {canManage && (
              <TableCell>
                <div className="flex items-center gap-2">
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
                        size="sm"
                        className="h-8"
                        onClick={() => onEdit(model)}
                      >
                        编辑
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="h-8 text-red-600 hover:text-red-700"
                        onClick={() => onDelete(model)}
                      >
                        删除
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
