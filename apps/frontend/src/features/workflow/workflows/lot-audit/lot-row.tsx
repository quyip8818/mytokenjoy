import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import type { LotAuditEntry } from '@/api/types'
import { formatDateTime } from '@/lib/date'

const KIND_STYLES: Record<
  string,
  { label: string; variant: 'default' | 'secondary' | 'destructive' | 'outline' }
> = {
  paid: { label: '付费', variant: 'default' },
  gift: { label: '赠送', variant: 'secondary' },
  adjust: { label: '调整', variant: 'outline' },
  overdraft: { label: '透支', variant: 'destructive' },
  mock: { label: '模拟', variant: 'outline' },
}

function fmt(n: number) {
  return n.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

interface LotRowProps {
  lot: LotAuditEntry
  readonly: boolean
  onRefund: (lot: LotAuditEntry) => void
}

export function LotRow({ lot, readonly, onRefund }: LotRowProps) {
  const kind = KIND_STYLES[lot.lotKind] ?? { label: lot.lotKind, variant: 'outline' as const }
  const granted = lot.quotaGranted / lot.quotaPerUnit
  const remaining = lot.quotaRemaining / lot.quotaPerUnit
  const percent = lot.quotaGranted > 0 ? (lot.quotaRemaining / lot.quotaGranted) * 100 : 0
  const canRefund =
    !readonly && lot.status === 'active' && lot.quotaRemaining > 0 && lot.lotKind !== 'overdraft'

  return (
    <div className="rounded-lg border border-border p-4 space-y-3">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Badge variant={kind.variant}>{kind.label}</Badge>
          <Badge variant={lot.status === 'active' ? 'default' : 'secondary'}>
            {lot.status === 'active' ? '活跃' : '已用尽'}
          </Badge>
          <span className="text-xs text-muted-foreground">{formatDateTime(lot.createdAt)}</span>
        </div>
        {canRefund && (
          <Button variant="outline" size="sm" onClick={() => onRefund(lot)}>
            退费
          </Button>
        )}
      </div>

      {/* Stats */}
      <div className="flex gap-4 text-sm">
        <span>配额: ¥{fmt(granted)}</span>
        <span>剩余: ¥{fmt(remaining)}</span>
      </div>

      {/* Progress bar */}
      <div className="h-2 w-full rounded-full bg-muted overflow-hidden">
        <div
          className="h-full rounded-full bg-primary transition-all"
          style={{ width: `${Math.min(percent, 100)}%` }}
        />
      </div>

      {/* Transactions timeline */}
      {lot.transactions.length > 0 && (
        <div className="border-t border-border/50 pt-2 space-y-1">
          <p className="text-xs font-medium text-muted-foreground">变更记录</p>
          {lot.transactions.map((tx) => (
            <div key={tx.id} className="flex items-center gap-2 text-xs text-muted-foreground">
              <span className="w-20 shrink-0">{formatDateTime(tx.createdAt).slice(5, 16)}</span>
              <Badge
                variant={tx.action === 'credit' ? 'default' : 'destructive'}
                className="text-[10px] px-1 py-0"
              >
                {tx.action === 'credit' ? '+' : '-'}¥{fmt(tx.moneyAmount)}
              </Badge>
              <span className="truncate">{tx.note || tx.source}</span>
              <span className="ml-auto shrink-0">
                → ¥{fmt(tx.remainingAfter / lot.quotaPerUnit)}
              </span>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
