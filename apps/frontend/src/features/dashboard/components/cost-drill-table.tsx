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
import { formatTokenCount, type DrillState } from '@/features/dashboard/lib/dashboard'

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
            <TableRow className="hover:bg-transparent">
              <TableHead>成员</TableHead>
              <TableHead className="text-right">花费 (¥)</TableHead>
              <TableHead className="text-right">Token 数</TableHead>
              <TableHead className="text-right">请求数</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {memberCosts.map((m) => (
              <TableRow key={m.memberId}>
                <TableCell className="font-medium">
                  <Users className="mr-2 inline h-4 w-4 text-muted-foreground" />
                  {m.memberName}
                </TableCell>
                <TableCell className="text-right font-semibold">
                  {m.cost.toLocaleString()}
                </TableCell>
                <TableCell className="text-right text-muted-foreground">
                  {formatTokenCount(m.tokens)}
                </TableCell>
                <TableCell className="text-right text-muted-foreground">
                  {m.requests.toLocaleString()}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      ) : (
        <Table>
          <TableHeader>
            <TableRow className="hover:bg-transparent">
              <TableHead className="w-6" />
              <TableHead>部门</TableHead>
              <TableHead className="text-right">花费 (¥)</TableHead>
              <TableHead className="text-right">占比</TableHead>
              <TableHead className="w-24" />
            </TableRow>
          </TableHeader>
          <TableBody>
            {deptCosts.map((dept) => (
              <TableRow key={dept.departmentId}>
                <TableCell />
                <TableCell className="font-medium">{dept.departmentName}</TableCell>
                <TableCell className="text-right font-semibold">
                  {dept.cost.toLocaleString()}
                </TableCell>
                <TableCell className="text-right text-muted-foreground">
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
