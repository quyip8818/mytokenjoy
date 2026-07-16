import { MoreHorizontal } from 'lucide-react'
import type { PlatformKey } from '@/api/types'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { KeyStatusBadge } from './status-badges'
import { Progress } from '@/components/ui/progress'
import { formatDisplayCurrency } from '@/lib/points'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { PermissionGate } from '@/features/session'
import { PERMISSION } from '@/lib/permissions'

interface MyKeysTableProps {
  keys: PlatformKey[]
  rowClass: (id: string) => string | undefined
  onEdit: (key: PlatformKey) => void
  onRotate: (key: PlatformKey) => void
  onToggle: (key: PlatformKey) => void
  onDelete: (key: PlatformKey) => void
}

function keyBudgetPercent(key: PlatformKey) {
  if (key.budget <= 0) return 0
  return Math.min(100, Math.round((key.consumed / key.budget) * 100))
}

function KeyRowActions({
  keyItem,
  onEdit,
  onRotate,
  onToggle,
  onDelete,
}: {
  keyItem: PlatformKey
  onEdit: (key: PlatformKey) => void
  onRotate: (key: PlatformKey) => void
  onToggle: (key: PlatformKey) => void
  onDelete: (key: PlatformKey) => void
}) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon" className="h-8 w-8">
          <MoreHorizontal className="h-4 w-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem onClick={() => onEdit(keyItem)}>编辑</DropdownMenuItem>
        <DropdownMenuItem onClick={() => onRotate(keyItem)}>重新生成</DropdownMenuItem>
        <DropdownMenuItem onClick={() => void onToggle(keyItem)}>
          {keyItem.status === 'active' ? '禁用' : '启用'}
        </DropdownMenuItem>
        <DropdownMenuItem className="text-red-600" onClick={() => onDelete(keyItem)}>
          删除
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}

export function MyKeysTable({
  keys,
  rowClass,
  onEdit,
  onRotate,
  onToggle,
  onDelete,
}: MyKeysTableProps) {
  return (
    <Table>
      <TableHeader>
        <TableRow className="hover:bg-transparent">
          <TableHead>名称</TableHead>
          <TableHead>Key 前缀</TableHead>
          <TableHead>额度</TableHead>
          <TableHead>模型</TableHead>
          <TableHead>状态</TableHead>
          <TableHead className="w-[120px]">操作</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {keys.map((key) => (
          <TableRow key={key.id} className={rowClass(key.id)}>
            <TableCell className="font-medium">{key.name}</TableCell>
            <TableCell className="font-mono text-sm text-muted-foreground">
              {key.keyPrefix}
            </TableCell>
            <TableCell>
              <div className="min-w-28 space-y-1">
                <div className="text-xs text-muted-foreground">
                  {formatDisplayCurrency(key.consumed)} / {formatDisplayCurrency(key.budget)}
                </div>
                <Progress value={keyBudgetPercent(key)} className="h-1.5" />
              </div>
            </TableCell>
            <TableCell className="text-sm text-muted-foreground">
              {key.modelWhitelist.length} 个
            </TableCell>
            <TableCell>
              <KeyStatusBadge status={key.status} />
            </TableCell>
            <TableCell>
              <PermissionGate write permission={PERMISSION.SELF_KEYS}>
                <KeyRowActions
                  keyItem={key}
                  onEdit={onEdit}
                  onRotate={onRotate}
                  onToggle={onToggle}
                  onDelete={onDelete}
                />
              </PermissionGate>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}
