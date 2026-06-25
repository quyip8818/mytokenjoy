import { MoreHorizontal } from 'lucide-react'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { StatusBadge } from '@/components/ui/status-badge'
import { BudgetProgressCell } from '@/components/budget/budget-progress-cell'
import type { PlatformKey } from '@/api/types'
import { KeyPrefixBadge, KeyStatusBadge } from '@/components/keys/status-badges'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'

interface PlatformKeyTableProps {
  keys: PlatformKey[]
  rowClass: (id: string) => string
  onRevoke: (id: string) => void
}

export function PlatformKeyTable({ keys, rowClass, onRevoke }: PlatformKeyTableProps) {
  return (
    <Table>
      <TableHeader>
        <TableRow className="hover:bg-transparent">
          <TableHead>名称</TableHead>
          <TableHead>绑定</TableHead>
          <TableHead>预算组</TableHead>
          <TableHead>Key 前缀</TableHead>
          <TableHead>状态</TableHead>
          <TableHead className="w-36">额度使用</TableHead>
          <TableHead>模型白名单</TableHead>
          <TableHead>到期时间</TableHead>
          <TableHead className="w-[120px]">操作</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {keys.map((key) => (
          <TableRow key={key.id} className={rowClass(key.id)}>
            <TableCell className="font-medium">{key.name}</TableCell>
            <TableCell className="text-sm">{key.memberName ?? key.appName ?? '-'}</TableCell>
            <TableCell className="text-sm text-muted-foreground">
              {key.budgetGroupName ?? '-'}
            </TableCell>
            <TableCell>
              <KeyPrefixBadge prefix={key.keyPrefix} />
            </TableCell>
            <TableCell>
              <KeyStatusBadge status={key.status} />
            </TableCell>
            <TableCell>
              <BudgetProgressCell value={key.used} total={key.quota} />
            </TableCell>
            <TableCell>
              <div className="flex flex-wrap gap-1">
                {key.modelWhitelist.slice(0, 2).map((m) => (
                  <StatusBadge key={m} variant="info" className="text-xs">
                    {m}
                  </StatusBadge>
                ))}
                {key.modelWhitelist.length > 2 && (
                  <StatusBadge variant="info" className="text-xs">
                    +{key.modelWhitelist.length - 2}
                  </StatusBadge>
                )}
              </div>
            </TableCell>
            <TableCell className="text-sm text-muted-foreground">
              {key.expiresAt ?? '永不'}
            </TableCell>
            <TableCell>
              {key.status === 'active' && (
                <DropdownMenu>
                  <DropdownMenuTrigger
                    render={
                      <Button variant="ghost" size="icon" className="h-8 w-8">
                        <MoreHorizontal className="h-4 w-4" />
                      </Button>
                    }
                  />
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem className="text-red-600" onClick={() => onRevoke(key.id)}>
                      吊销
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              )}
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}
