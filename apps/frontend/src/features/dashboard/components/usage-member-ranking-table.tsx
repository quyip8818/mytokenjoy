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

interface UsageMemberRankingTableProps {
  topConsumers: TopConsumer[]
  loading: boolean
}

export function UsageMemberRankingTable({ topConsumers, loading }: UsageMemberRankingTableProps) {
  const title = topConsumers.length > 0 ? `成员消耗排行 Top ${topConsumers.length}` : '成员消耗排行'

  return (
    <DataSection
      title={title}
      loading={loading}
      skeletonColumns={4}
      className="border-border shadow-xs"
    >
      <Table>
        <TableHeader>
          <TableRow className="border-border/50 hover:bg-transparent">
            <TableHead className="text-xs font-semibold text-muted-foreground">排名</TableHead>
            <TableHead className="text-xs font-semibold text-muted-foreground">成员</TableHead>
            <TableHead className="text-right text-xs font-semibold text-muted-foreground">
              花费 (¥)
            </TableHead>
            <TableHead className="text-right text-xs font-semibold text-muted-foreground">
              请求数
            </TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {topConsumers.length === 0 && !loading && (
            <TableRow>
              <TableCell colSpan={4} className="text-center text-muted-foreground">
                暂无成员消耗数据
              </TableCell>
            </TableRow>
          )}
          {topConsumers.map((c, i) => (
            <TableRow
              key={c.memberId}
              className="border-border-subtle hover:bg-muted/50 transition-colors"
            >
              <TableCell>
                <div
                  className={`flex h-6 w-6 items-center justify-center rounded-full text-[11px] font-bold text-white ${i < 3 ? 'bg-primary' : 'bg-slate-300'}`}
                >
                  {i + 1}
                </div>
              </TableCell>
              <TableCell className="font-medium">{c.memberName}</TableCell>
              <TableCell className="text-right font-semibold tabular-nums">
                {c.cost.toLocaleString(undefined, { maximumFractionDigits: 2 })}
              </TableCell>
              <TableCell className="text-right text-muted-foreground tabular-nums">
                {c.requests.toLocaleString()}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </DataSection>
  )
}
