import { DataSection } from '@/components/layout/data-section'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import type { TopConsumer } from '@/api/types'
import { formatTokenCount } from '@/features/dashboard'

interface CostTopConsumersTableProps {
  topConsumers: TopConsumer[]
  loading: boolean
}

export function CostTopConsumersTable({ topConsumers, loading }: CostTopConsumersTableProps) {
  return (
    <DataSection
      title="消费排行 Top 5"
      loading={loading}
      skeletonColumns={6}
      className="border-border shadow-xs"
    >
      <Table>
        <TableHeader>
          <TableRow className="border-border/50 hover:bg-transparent">
            <TableHead className="text-xs font-semibold text-muted-foreground">排名</TableHead>
            <TableHead className="text-xs font-semibold text-muted-foreground">成员</TableHead>
            <TableHead className="text-xs font-semibold text-muted-foreground">部门</TableHead>
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
          {topConsumers.map((c, i) => (
            <TableRow
              key={c.memberId}
              className="border-border-subtle even:bg-muted/40 hover:bg-muted/50"
            >
              <TableCell>
                <div
                  className={`flex h-6 w-6 items-center justify-center rounded-full text-[11px] font-bold text-white ${i < 3 ? 'bg-primary' : 'bg-slate-300'}`}
                >
                  {i + 1}
                </div>
              </TableCell>
              <TableCell className="font-medium">{c.memberName}</TableCell>
              <TableCell className="text-sm text-muted-foreground">{c.department}</TableCell>
              <TableCell className="text-right font-semibold">{c.cost.toLocaleString()}</TableCell>
              <TableCell className="text-right text-muted-foreground">
                {formatTokenCount(c.tokens)}
              </TableCell>
              <TableCell className="text-right text-muted-foreground">
                {c.requests.toLocaleString()}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </DataSection>
  )
}
