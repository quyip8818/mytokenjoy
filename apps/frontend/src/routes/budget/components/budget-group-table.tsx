import { MoreHorizontal } from 'lucide-react'
import type { BudgetGroup } from '@/api/types'
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
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'

interface BudgetGroupTableProps {
  groups: BudgetGroup[]
  canWrite: boolean
  rowClass: (id: string) => string | undefined
  onEdit: (group: BudgetGroup) => void
  onManageKeys: (group: BudgetGroup) => void
  onDelete: (id: string) => void
}

export function BudgetGroupTable({
  groups,
  canWrite,
  rowClass,
  onEdit,
  onManageKeys,
  onDelete,
}: BudgetGroupTableProps) {
  return (
    <Table>
      <TableHeader>
        <TableRow className="hover:bg-transparent">
          <TableHead>名称</TableHead>
          <TableHead className="text-right">预算 (¥)</TableHead>
          <TableHead className="text-right">已消耗 (¥)</TableHead>
          <TableHead className="w-40">进度</TableHead>
          <TableHead>关联</TableHead>
          <TableHead className="w-[120px]">操作</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {groups.map((g) => (
          <TableRow key={g.id} className={rowClass(g.id)}>
            <TableCell className="font-medium">{g.name}</TableCell>
            <TableCell className="text-right">{g.budget.toLocaleString()}</TableCell>
            <TableCell className="text-right">{g.consumed.toLocaleString()}</TableCell>
            <TableCell className="w-40">
              <BudgetProgressCell value={g.consumed} total={g.budget} />
            </TableCell>
            <TableCell>
              <StatusBadge variant="info">{g.memberIds.length} 人</StatusBadge>
              <StatusBadge variant="info" className="ml-1">
                {g.departmentIds.length} 部门
              </StatusBadge>
            </TableCell>
            <TableCell>
              {canWrite ? (
                <DropdownMenu>
                  <DropdownMenuTrigger
                    render={
                      <Button variant="ghost" size="icon" className="h-8 w-8">
                        <MoreHorizontal className="h-4 w-4" />
                      </Button>
                    }
                  />
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem onClick={() => onEdit(g)}>管理</DropdownMenuItem>
                    <DropdownMenuItem onClick={() => onManageKeys(g)}>管理 Key</DropdownMenuItem>
                    <DropdownMenuItem className="text-red-600" onClick={() => onDelete(g.id)}>
                      删除
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              ) : null}
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}
