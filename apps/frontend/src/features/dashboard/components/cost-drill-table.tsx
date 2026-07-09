import { ChevronRight, Users } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { DataSection } from '@/components/layout/data-section'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import type { DepartmentCost, DepartmentCostMember } from '@/api/types'
import { formatTokenCount, type DrillState } from '@/features/dashboard'

interface CostDrillTableProps {
  drill: DrillState
  drillTitle: string
  deptCosts: DepartmentCost[]
  memberCosts: DepartmentCostMember[]
  loading: boolean
  canDrillBack: boolean
  onDrillBack: () => void
  onDrillDept: (dept: DepartmentCost) => void
}

export function CostDrillTable({
  drill,
  drillTitle,
  deptCosts,
  memberCosts,
  loading,
  canDrillBack,
  onDrillBack,
  onDrillDept,
}: CostDrillTableProps) {
  return (
    <DataSection
      title={drillTitle}
      loading={loading}
      skeletonColumns={5}
      className="border-border shadow-xs"
      headerAction={
        canDrillBack ? (
          <Button variant="outline" size="sm" onClick={onDrillBack}>
            返回上级
          </Button>
        ) : undefined
      }
    >
      {drill.level === 'members' ? (
        <Table>
          <TableHeader>
            <TableRow className="border-border/50 hover:bg-transparent">
              <TableHead className="text-xs font-semibold text-muted-foreground">成员</TableHead>
              <TableHead className="text-right text-xs font-semibold text-muted-foreground">
                花费 (¥)
              </TableHead>
              <TableHead className="text-right text-xs font-semibold text-muted-foreground">
                Token 数
              </TableHead>
              <TableHead className="text-right text-xs font-semibold text-muted-foreground">
                请求数
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {memberCosts.map((m) => (
              <TableRow
                key={m.memberId}
                className="border-border-subtle hover:bg-muted/50 transition-colors"
              >
                <TableCell className="font-medium pl-6">
                  <Users className="mr-2 inline h-4 w-4 text-muted-foreground" />
                  {m.memberName}
                </TableCell>
                <TableCell className="text-right font-semibold tabular-nums">
                  {m.cost.toFixed(2)}
                </TableCell>
                <TableCell className="text-right text-muted-foreground tabular-nums">
                  {formatTokenCount(m.tokens)}
                </TableCell>
                <TableCell className="text-right text-muted-foreground tabular-nums">
                  {m.requests.toLocaleString()}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      ) : (
        <Table>
          <TableHeader>
            <TableRow className="border-border/50 hover:bg-transparent">
              <TableHead className="w-6" />
              <TableHead className="text-xs font-semibold text-muted-foreground">部门</TableHead>
              <TableHead className="text-right text-xs font-semibold text-muted-foreground">
                花费 (¥)
              </TableHead>
              <TableHead className="text-right text-xs font-semibold text-muted-foreground">
                占比
              </TableHead>
              <TableHead className="w-24" />
            </TableRow>
          </TableHeader>
          <TableBody>
            {deptCosts.map((dept) => (
              <TableRow
                key={dept.departmentId}
                className="border-border-subtle hover:bg-muted/50 transition-colors"
              >
                <TableCell />
                <TableCell className="font-medium">{dept.departmentName}</TableCell>
                <TableCell className="text-right font-semibold tabular-nums">
                  {dept.cost.toFixed(2)}
                </TableCell>
                <TableCell className="text-right text-muted-foreground tabular-nums">
                  {dept.percentage}%
                </TableCell>
                <TableCell className="text-right">
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-8 text-blue-600"
                    onClick={() => onDrillDept(dept)}
                  >
                    下钻
                    <ChevronRight className="ml-1 h-4 w-4" />
                  </Button>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}
    </DataSection>
  )
}
