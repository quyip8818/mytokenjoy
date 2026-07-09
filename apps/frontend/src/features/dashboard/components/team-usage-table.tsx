import type { TeamUsage } from '@/api/types'
import { teamUsagePercent } from '@/features/dashboard'
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

interface TeamUsageTableProps {
  teamUsage: TeamUsage[]
}

export function TeamUsageTable({ teamUsage }: TeamUsageTableProps) {
  return (
    <Table>
      <TableHeader>
        <TableRow className="border-border/50 hover:bg-transparent">
          <TableHead className="text-xs font-semibold text-muted-foreground">部门</TableHead>
          <TableHead className="text-xs font-semibold text-muted-foreground">额度 (¥)</TableHead>
          <TableHead className="text-xs font-semibold text-muted-foreground">已消耗 (¥)</TableHead>
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
        {teamUsage.map((t) => {
          const pct = teamUsagePercent(t.consumed, t.quota)
          return (
            <TableRow key={t.departmentId} className="border-border-subtle hover:bg-muted/50 transition-colors">
              <TableCell className="font-medium">{t.departmentName}</TableCell>
              <TableCell className="text-muted-foreground tabular-nums">{t.quota.toLocaleString()}</TableCell>
              <TableCell className="font-medium tabular-nums">{t.consumed.toLocaleString()}</TableCell>
              <TableCell>
                <div className="flex items-center gap-2.5">
                  <Progress value={pct} className="flex-1 h-2" />
                  <span
                    className={`text-xs font-semibold ${pct >= 90 ? 'text-red-500' : pct >= 70 ? 'text-amber-500' : 'text-primary'}`}
                  >
                    {pct}%
                  </span>
                </div>
              </TableCell>
              <TableCell className="text-right text-muted-foreground">{t.memberCount}</TableCell>
              <TableCell>
                <Badge variant="secondary" className="text-xs font-medium">
                  {t.topModel}
                </Badge>
              </TableCell>
            </TableRow>
          )
        })}
      </TableBody>
    </Table>
  )
}
