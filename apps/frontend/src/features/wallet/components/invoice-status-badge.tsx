import { Badge } from '@/components/ui/badge'
import type { TopUpRecordView } from '@/features/wallet'

interface InvoiceStatusBadgeProps {
  status: TopUpRecordView['invoiceStatus']
}

export function InvoiceStatusBadge({ status }: InvoiceStatusBadgeProps) {
  if (status === 'none') return <span className="text-xs text-muted-foreground">未申请</span>
  if (status === 'applied')
    return (
      <Badge variant="outline" className="border-amber-200 bg-amber-50 text-xs text-amber-700">
        申请中
      </Badge>
    )
  return (
    <Badge variant="outline" className="border-emerald-200 bg-emerald-50 text-xs text-emerald-700">
      已开票
    </Badge>
  )
}
