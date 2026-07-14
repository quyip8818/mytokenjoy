import { DataSection } from '@/components/layout/data-section'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import type { DepartmentCost } from '@/api/types'
import { COST_CHART_COLORS } from '../lib/dashboard'

interface DeptComparisonTableProps {
  deptCosts: DepartmentCost[]
  loading: boolean
  onSelectDept?: (deptId: string) => void
}

export function DeptComparisonTable({
  deptCosts,
  loading,
  onSelectDept,
}: DeptComparisonTableProps) {
  return (
    <DataSection
      title="子部门费用对比"
      loading={loading}
      skeletonColumns={5}
      className="border-border shadow-xs"
    >
      <Table>
        <TableHeader>
          <TableRow className="border-border/50 hover:bg-transparent">
            <TableHead className="w-12 text-xs font-semibold text-muted-foreground">排名</TableHead>
            <TableHead className="text-xs font-semibold text-muted-foreground">部门</TableHead>
            <TableHead className="text-right text-xs font-semibold text-muted-foreground">
              费用
            </TableHead>
            <TableHead className="text-right text-xs font-semibold text-muted-foreground">
              占比
            </TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {deptCosts.length === 0 ? (
            <TableRow>
              <TableCell colSpan={4} className="h-24 text-center text-sm text-muted-foreground">
                暂无子部门数据
              </TableCell>
            </TableRow>
          ) : (
            deptCosts.map((dept, i) => (
              <TableRow
                key={dept.departmentId}
                className="border-border-subtle hover:bg-muted/50 transition-colors cursor-pointer"
                onClick={() => onSelectDept?.(dept.departmentId)}
              >
                <TableCell>
                  <div
                    className="flex h-6 w-6 items-center justify-center rounded-full text-[11px] font-bold text-white"
                    style={{ backgroundColor: COST_CHART_COLORS[i % COST_CHART_COLORS.length] }}
                  >
                    {i + 1}
                  </div>
                </TableCell>
                <TableCell className="font-medium">{dept.departmentName}</TableCell>
                <TableCell className="text-right font-semibold tabular-nums">
                  {dept.cost.toFixed(2)}
                </TableCell>
                <TableCell className="text-right text-muted-foreground tabular-nums">
                  {dept.percentage.toFixed(1)}%
                </TableCell>
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
    </DataSection>
  )
}
