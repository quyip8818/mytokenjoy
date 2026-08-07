import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import type { DiscountEntry } from '@/api/types'

interface Props {
  discounts: DiscountEntry[]
}

function formatDiscount(d: number): string {
  if (d === 1) return '无优惠'
  if (d < 1) return `${(d * 10).toFixed(1).replace(/\.0$/, '')}折`
  return `加价${Math.round((d - 1) * 100)}%`
}

/**
 * 企业侧只读优惠列表。条件渲染：无 discount 时不展示。
 */
export function DiscountSection({ discounts }: Props) {
  if (discounts.length === 0) return null

  return (
    <div
      data-testid="billing-discount-section"
      className="rounded-lg border border-border bg-card p-5"
    >
      <h3 className="text-sm font-medium mb-3">当前优惠</h3>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>模型</TableHead>
            <TableHead>折扣</TableHead>
            <TableHead>说明</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {discounts.map((d) => (
            <TableRow key={d.modelType}>
              <TableCell className="font-mono text-xs">
                {d.modelType === '*' ? '其他所有模型' : d.modelType}
              </TableCell>
              <TableCell className="tabular-nums">{formatDiscount(d.discount)}</TableCell>
              <TableCell className="text-muted-foreground text-xs">{d.note || '-'}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
      <p className="text-xs text-muted-foreground mt-2">
        * 折扣由平台管理员配置，直接影响调用扣费金额。
      </p>
    </div>
  )
}
