import type { ModelInfo } from '@/api/types'
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

const VISIBILITY_LABELS: Record<string, string> = {
  all: '全员可见',
  department: '部门可见',
  custom: '自定义',
}

interface ModelListTableProps {
  models: ModelInfo[]
  canManage: boolean
  showActions?: boolean
  rowClass: (id: string) => string | undefined
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
            模型 ID
          </TableHead>
          <TableHead className="text-xs font-medium uppercase text-muted-foreground">
            模型类型
          </TableHead>
          <TableHead className="text-xs font-medium uppercase text-muted-foreground">
            描述
          </TableHead>
          <TableHead className="text-xs font-medium uppercase text-muted-foreground">
            可见范围
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
            key={model.id}
            className={`even:bg-muted/40 ${rowClass(model.id)} ${!model.enabled ? 'opacity-50' : ''}`}
          >
            <TableCell className="font-medium">{model.displayName}</TableCell>
            <TableCell className="font-mono text-xs text-muted-foreground">{model.name}</TableCell>
            <TableCell>
              <Badge variant="outline" className="border-0 bg-muted text-xs">
                {model.type === 'builtin' ? '内置' : (PROVIDER_LABELS[model.provider] ?? '自定义')}
              </Badge>
            </TableCell>
            <TableCell className="max-w-xs truncate text-sm text-muted-foreground">
              {model.description || '—'}
            </TableCell>
            <TableCell className="text-sm text-muted-foreground">
              {VISIBILITY_LABELS[model.visibility] ?? model.visibility}
            </TableCell>
            <TableCell className="max-w-xs truncate font-mono text-xs text-muted-foreground">
              {model.type === 'custom' ? (model.endpoint ?? '—') : '—'}
            </TableCell>
            {showActions && canManage && (
              <TableCell>
                <div className="flex items-center gap-2">
                  {model.type === 'custom' && (
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
                  <Button variant="ghost" size="sm" className="h-8" onClick={() => onToggle(model)}>
                    <StatusBadge variant={model.enabled ? 'success' : 'neutral'}>
                      {model.enabled ? '启用' : '禁用'}
                    </StatusBadge>
                  </Button>
                </div>
              </TableCell>
            )}
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}
