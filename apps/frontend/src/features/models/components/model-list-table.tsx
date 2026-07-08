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
import { StatusBadge } from '@/components/ui/status-badge'
import { PROVIDER_LABELS } from '@/lib/provider-labels'

interface ModelListTableProps {
  models: ModelInfo[]
  canManage: boolean
  showActions?: boolean
  rowClass: (id: string | number) => string | undefined
  onToggle: (model: ModelInfo) => void
  onEdit: (model: ModelInfo) => void
  onDelete: (model: ModelInfo) => void
}

export function ModelListTable({
  models,
  canManage,
  showActions = true,
  rowClass,
  onToggle,
  onEdit,
  onDelete,
}: ModelListTableProps) {
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
          <TableHead className="text-xs font-medium uppercase text-muted-foreground">
            来源
          </TableHead>
          <TableHead className="text-xs font-medium uppercase text-muted-foreground">
            描述
          </TableHead>
          <TableHead className="text-xs font-medium uppercase text-muted-foreground">
            部署地址
          </TableHead>
          {showActions && canManage && (
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
            <TableCell>
              <Badge variant="outline" className="border-0 bg-muted text-xs">
                {isCustomModel(model) ? (PROVIDER_LABELS[model.provider] ?? '自定义') : '内置'}
              </Badge>
            </TableCell>
            <TableCell className="max-w-xs truncate text-sm text-muted-foreground">
              {model.description || '—'}
            </TableCell>
            <TableCell className="max-w-xs truncate font-mono text-xs text-muted-foreground">
              {isCustomModel(model) ? (model.endpoint ?? '—') : '—'}
            </TableCell>
            {showActions && canManage && (
              <TableCell>
                <div className="flex items-center gap-2">
                  {isCustomModel(model) && (
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
                      <Button
                        variant="ghost"
                        size="sm"
                        className="h-8"
                        onClick={() => onToggle(model)}
                      >
                        <StatusBadge variant={model.enabled ? 'success' : 'neutral'}>
                          {model.enabled ? '启用' : '禁用'}
                        </StatusBadge>
                      </Button>
                    </>
                  )}
                  {!isCustomModel(model) && (
                    <StatusBadge variant={model.enabled ? 'success' : 'neutral'}>
                      {model.enabled ? '启用' : '禁用'}
                    </StatusBadge>
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
