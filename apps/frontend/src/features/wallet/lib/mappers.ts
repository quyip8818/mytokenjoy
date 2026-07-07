import type { TopUpRecord } from '@/api/billing'
import type { PaymentMethod, TopUpRecordView } from '@/features/wallet'

export function toTopUpRecordView(record: TopUpRecord): TopUpRecordView {
  return {
    id: record.id,
    orderId: record.orderId,
    method: record.method as PaymentMethod,
    amount: record.amount,
    paidAmount: record.paidAmount,
    invoiceStatus: record.invoiceStatus,
    createdAt: record.createdAt,
  }
}

export function filterTopUpRecords(records: TopUpRecordView[], search: string) {
  if (!search) return records
  return records.filter((record) => record.orderId.toLowerCase().includes(search.toLowerCase()))
}
