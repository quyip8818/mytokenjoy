import type { DepartmentUsage } from '@/api/types'
import { departmentUsagePercent } from '@/features/dashboard'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Progress } from '@/components/ui/progress'
import { Badge } from '@/components/ui/badge'

interface DepartmentUsageTableProps {
  departmentUsage: DepartmentUsage[]
  onSelectDept?: (deptId: string) => void
}

export function DepartmentUsageTable({ departmentUsage, onSelectDept }: DepartmentUsageTableProps) {
  return (
    <Table>
      <TableHeader>
        <TableRow className="border-border/50 hover:bg-transparent">
          <TableHead className="text-xs font-semibold text-muted-foreground">部门</TableHead>
          <TableHead className="text-xs font-semibold text-muted-foreground">额度</TableHead>
          <TableHead className="text-xs font-semibold text-muted-foreground">已消耗</TableHead>
          <TableHead className="text-xs font-semibold text-muted-foreground w-48">
            消耗进度
          </TableHead>
          <TableHead className="text-xs font-semibold text-muted-foreground text-right">
            成员数
          </TableHead>
          <TableHead className="text-xs font-semibold text-muted-foreground">主力模型</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {departmentUsage.map((row) => {
          const pct = departmentUsagePercent(row.consumed, row.budget)
          return (
            <TableRow
              key={row.departmentId}
              className="border-border-subtle hover:bg-muted/50 transition-colors cursor-pointer"
              onClick={() => onSelectDept?.(row.departmentId)}
            >
              <TableCell className="font-medium">{row.departmentName}</TableCell>
              <TableCell className="text-muted-foreground tabular-nums">
                {row.budget.toLocaleString(undefined, { maximumFractionDigits: 2 })}
              </TableCell>
              <TableCell className="font-medium tabular-nums">
                {row.consumed.toLocaleString(undefined, { maximumFractionDigits: 2 })}
              </TableCell>
              <TableCell>
                <div className="flex items-center gap-2.5">
                  <Progress value={pct} className="flex-1 h-2" />
                  <span
                    aria-hidden="true"
                    className={`text-xs font-semibold ${pct >= 90 ? 'text-red-500' : pct >= 70 ? 'text-amber-500' : 'text-primary'}`}
                  >
                    {pct}%
                  </span>
                </div>
              </TableCell>
              <TableCell className="text-right text-muted-foreground">{row.memberCount}</TableCell>
              <TableCell>
                <Badge variant="secondary" className="text-xs font-medium">
                  {row.topModel}
                </Badge>
              </TableCell>
            </TableRow>
          )
        })}
      </TableBody>
    </Table>
  )
}
