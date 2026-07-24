import { useState } from 'react'
import { Receipt } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { InvoiceStatusBadge, type TopUpRecordView } from '@/features/billing'
import { filterTopUpRecords } from '../lib/mappers'

interface RechargeRecordsTableProps {
  records: TopUpRecordView[]
}

export function RechargeRecordsTable({ records }: RechargeRecordsTableProps) {
  const [searchOrderId, setSearchOrderId] = useState('')
  const [activeTab, setActiveTab] = useState<'topup' | 'invoice'>('topup')
  const filteredRecords = filterTopUpRecords(records, searchOrderId)

  return (
    <div className="rounded-lg border border-border bg-card shadow-xs">
      <div className="border-b border-border px-5 py-3">
        <div className="flex items-center gap-2">
          <Receipt className="size-4 text-muted-foreground" strokeWidth={1.5} />
          <h2 className="text-sm font-semibold">充值开票</h2>
          <span className="text-xs text-muted-foreground">管理充值记录与发票申请</span>
        </div>
      </div>
      <div className="space-y-4 p-5">
        <div className="flex items-center justify-between rounded-md border border-amber-200 bg-amber-50 px-4 py-2.5">
          <span className="text-xs text-amber-800">请先完成实名认证后再申请开具发票</span>
          <Button
            size="sm"
            variant="outline"
            className="h-7 border-amber-300 text-xs text-amber-800 hover:bg-amber-100"
          >
            去认证
          </Button>
        </div>

        <div className="flex gap-4 border-b border-border">
          <button
            type="button"
            onClick={() => setActiveTab('topup')}
            className={cn(
              'pb-2 text-sm transition-colors duration-100',
              activeTab === 'topup'
                ? 'border-b-2 border-primary font-medium text-foreground'
                : 'text-muted-foreground hover:text-foreground',
            )}
          >
            充值记录
          </button>
          <button
            type="button"
            onClick={() => setActiveTab('invoice')}
            className={cn(
              'pb-2 text-sm transition-colors duration-100',
              activeTab === 'invoice'
                ? 'border-b-2 border-primary font-medium text-foreground'
                : 'text-muted-foreground hover:text-foreground',
            )}
          >
            开票记录
          </button>
        </div>

        <div className="flex items-center gap-2">
          <Input
            placeholder="订单号"
            value={searchOrderId}
            onChange={(event) => setSearchOrderId(event.target.value)}
            className="h-8 w-44 text-sm"
          />
          <Button variant="ghost" size="sm" className="h-8 text-xs">
            查询
          </Button>
          <Button
            variant="ghost"
            size="sm"
            className="h-8 text-xs"
            onClick={() => setSearchOrderId('')}
          >
            重置
          </Button>
          <div className="ml-auto flex gap-2">
            <Button variant="ghost" size="sm" className="h-8 text-xs" disabled>
              批量开票
            </Button>
            <Button variant="outline" size="sm" className="h-8 text-xs" disabled>
              全部开票
            </Button>
          </div>
        </div>

        {activeTab === 'topup' && (
          <div className="overflow-hidden rounded-md border border-border">
            <table className="w-full text-sm">
              <thead>
                <tr className="bg-muted text-xs font-medium text-muted-foreground">
                  <th className="px-4 py-2.5 text-left font-medium">订单号</th>
                  <th className="px-4 py-2.5 text-left font-medium">支付方式</th>
                  <th className="px-4 py-2.5 text-right font-medium">充值额度</th>
                  <th className="px-4 py-2.5 text-right font-medium">支付金额</th>
                  <th className="px-4 py-2.5 text-center font-medium">开票</th>
                  <th className="px-4 py-2.5 text-left font-medium">充值时间</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {filteredRecords.map((record, index) => (
                  <tr
                    key={record.id}
                    className={cn(index % 2 === 1 && 'bg-muted/40', 'hover:bg-muted/50')}
                  >
                    <td className="px-4 py-2.5 font-mono text-xs">{record.orderId}</td>
                    <td className="px-4 py-2.5">
                      {record.method === 'alipay' ? '支付宝' : '微信'}
                    </td>
                    <td className="px-4 py-2.5 text-right tabular-nums">
                      ¥{record.amount.toFixed(2)}
                    </td>
                    <td className="px-4 py-2.5 text-right tabular-nums">
                      ¥{record.paidAmount.toFixed(2)}
                    </td>
                    <td className="px-4 py-2.5 text-center">
                      <InvoiceStatusBadge status={record.invoiceStatus} />
                    </td>
                    <td className="px-4 py-2.5 text-xs text-muted-foreground">
                      {record.createdAt}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            {filteredRecords.length === 0 && (
              <p className="px-5 py-8 text-center text-sm text-muted-foreground">暂无充值记录</p>
            )}
          </div>
        )}

        {activeTab === 'invoice' && (
          <div className="rounded-md border border-border">
            <p className="px-5 py-8 text-center text-sm text-muted-foreground">暂无开票记录</p>
          </div>
        )}
      </div>
    </div>
  )
}
