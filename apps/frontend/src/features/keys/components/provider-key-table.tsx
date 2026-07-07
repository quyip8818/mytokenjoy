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
import type { ProviderKey } from '@/api/types'
import { KeyPrefixBadge, KeyStatusBadge, ProviderBadge } from './status-badges'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'

interface ProviderKeyTableProps {
  keys: ProviderKey[]
  rowClass: (id: string) => string
  onToggle: (key: ProviderKey) => void
  onDelete: (id: string) => void
}

function getBalanceClass(balance: number | null) {
  if (balance === null) return 'text-muted-foreground'
  if (balance > 1000) return 'font-medium text-emerald-600'
  if (balance < 500) return 'font-medium text-amber-600'
  return ''
}

export function ProviderKeyTable({ keys, rowClass, onToggle, onDelete }: ProviderKeyTableProps) {
  return (
    <Table>
      <TableHeader>
        <TableRow className="hover:bg-transparent">
          <TableHead className="text-xs font-semibold uppercase text-muted-foreground">
            名称
          </TableHead>
          <TableHead className="text-xs font-semibold uppercase text-muted-foreground">
            供应商
          </TableHead>
          <TableHead className="text-xs font-semibold uppercase text-muted-foreground">
            Key 前缀
          </TableHead>
          <TableHead className="text-xs font-semibold uppercase text-muted-foreground">
            状态
          </TableHead>
          <TableHead className="text-right text-xs font-semibold uppercase text-muted-foreground">
            余额
          </TableHead>
          <TableHead className="text-xs font-semibold uppercase text-muted-foreground">
            最后使用
          </TableHead>
          <TableHead className="w-[120px] text-xs font-semibold uppercase text-muted-foreground">
            操作
          </TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {keys.map((key) => (
          <TableRow key={key.id} className={`even:bg-muted/40 ${rowClass(key.id)}`}>
            <TableCell className="font-medium">{key.name}</TableCell>
            <TableCell>
              <ProviderBadge provider={key.provider} />
            </TableCell>
            <TableCell>
              <KeyPrefixBadge prefix={key.keyPrefix} />
            </TableCell>
            <TableCell>
              <KeyStatusBadge status={key.status} />
            </TableCell>
            <TableCell className={`text-right ${getBalanceClass(key.balance)}`}>
              {key.balance !== null ? `$${key.balance.toFixed(2)}` : '-'}
            </TableCell>
            <TableCell className="text-sm text-muted-foreground">{key.lastUsed ?? '-'}</TableCell>
            <TableCell>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="ghost" size="icon" className="h-8 w-8">
                    <MoreHorizontal className="h-4 w-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuItem onClick={() => onToggle(key)}>
                    {key.status === 'active' ? '禁用' : '启用'}
                  </DropdownMenuItem>
                  <DropdownMenuItem className="text-red-600" onClick={() => onDelete(key.id)}>
                    删除
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}
