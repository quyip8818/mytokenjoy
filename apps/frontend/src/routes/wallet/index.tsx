import { useState } from 'react'
import {
  Wallet,
  TrendingUp,
  BarChart3,
  Gift,
  ShoppingCart,
  Receipt,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import {
  mockWalletSummary,
  mockTopUpRecords,
} from '@/mocks/data'
import type { PaymentMethod, TopUpRecord } from '@/api/types'

const PRESET_AMOUNTS = [10, 20, 50, 100, 200, 500]

function InvoiceStatusBadge({ status }: { status: TopUpRecord['invoiceStatus'] }) {
  if (status === 'none') return <span className="text-xs text-muted-foreground">未申请</span>
  if (status === 'applied')
    return <Badge variant="outline" className="bg-amber-50 text-amber-700 border-amber-200 text-xs">申请中</Badge>
  return <Badge variant="outline" className="bg-emerald-50 text-emerald-700 border-emerald-200 text-xs">已开票</Badge>
}

export default function WalletPage() {
  const [amount, setAmount] = useState('')
  const [paymentMethod, setPaymentMethod] = useState<PaymentMethod>('alipay')
  const [redemptionCode, setRedemptionCode] = useState('')
  const [searchOrderId, setSearchOrderId] = useState('')
  const [activeTab, setActiveTab] = useState<'topup' | 'invoice'>('topup')

  const selectedAmount = amount ? Number(amount) : 0

  return (
    <div className="space-y-6">
      {/* Page header */}
      <div>
        <h1 className="text-xl font-semibold">钱包管理</h1>
        <p className="mt-1 text-sm text-muted-foreground">账户余额与充值管理</p>
      </div>

      {/* Stat cards */}
      <div className="grid grid-cols-3 gap-4">
        <StatCard icon={Wallet} label="当前余额" value={`¥${mockWalletSummary.balance.toFixed(2)}`} />
        <StatCard icon={TrendingUp} label="历史消耗" value={`¥${mockWalletSummary.totalConsumed.toFixed(2)}`} />
        <StatCard icon={BarChart3} label="请求次数" value={String(mockWalletSummary.totalRequests)} />
      </div>

      {/* Top-up section */}
      <div className="rounded-lg border border-border bg-card shadow-xs">
        <div className="flex items-center justify-between border-b border-border px-5 py-3">
          <div className="flex items-center gap-2">
            <ShoppingCart className="size-4 text-muted-foreground" strokeWidth={1.5} />
            <h2 className="text-sm font-semibold">账户充值</h2>
          </div>
          <Button variant="ghost" size="sm" className="gap-1.5 text-xs">
            <Receipt className="size-3.5" />
            账单
          </Button>
        </div>
        <div className="p-5 space-y-5">
          {/* Amount input + payment method */}
          <div className="grid grid-cols-12 gap-4">
            <div className="col-span-5 space-y-1.5">
              <label className="text-xs font-medium text-muted-foreground">充值数量</label>
              <Input
                type="number"
                min="0"
                placeholder="充值数量，最低 ¥0"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                className="h-9"
              />
              <p className="text-xs text-muted-foreground">
                实付金额：<span className="text-destructive font-medium">{selectedAmount} 元</span>
              </p>
            </div>
            <div className="col-span-7 space-y-1.5">
              <label className="text-xs font-medium text-muted-foreground">选择支付方式</label>
              <div className="flex gap-2">
                <Button
                  variant={paymentMethod === 'alipay' ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => setPaymentMethod('alipay')}
                  className="gap-1.5"
                >
                  支付宝
                </Button>
                <Button
                  variant={paymentMethod === 'wechat' ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => setPaymentMethod('wechat')}
                  className="gap-1.5"
                >
                  微信
                </Button>
              </div>
            </div>
          </div>

          {/* Preset amounts */}
          <div className="rounded-md border border-border p-4">
            <div className="flex items-center gap-2 mb-3">
              <span className="text-xs font-medium">选择充值额度</span>
              <span className="text-xs text-muted-foreground">如需开发票，请联系客服</span>
            </div>
            <div className="grid grid-cols-6 gap-2">
              {PRESET_AMOUNTS.map((preset) => (
                <button
                  key={preset}
                  type="button"
                  onClick={() => setAmount(String(preset))}
                  className={cn(
                    'rounded-md border px-3 py-2.5 text-center transition-colors duration-150',
                    String(preset) === amount
                      ? 'border-primary bg-primary/5 text-foreground'
                      : 'border-border bg-card hover:bg-muted text-foreground',
                  )}
                >
                  <span className="text-sm font-semibold tabular-nums">{preset} ¥</span>
                  <p className="mt-0.5 text-xs text-muted-foreground">实付 ¥{preset.toFixed(2)}</p>
                </button>
              ))}
            </div>
          </div>

          {/* Redemption code */}
          <div className="rounded-md border border-border p-4">
            <div className="flex items-center gap-2 mb-3">
              <Gift className="size-4 text-muted-foreground" strokeWidth={1.5} />
              <span className="text-xs font-medium">兑换码充值</span>
            </div>
            <div className="flex gap-2">
              <Input
                placeholder="请输入兑换码"
                value={redemptionCode}
                onChange={(e) => setRedemptionCode(e.target.value)}
                className="h-9 max-w-sm"
              />
              <Button size="sm">兑换额度</Button>
            </div>
          </div>
        </div>
      </div>


      {/* Top-up records & invoices */}
      <div className="rounded-lg border border-border bg-card shadow-xs">
        <div className="border-b border-border px-5 py-3">
          <div className="flex items-center gap-2">
            <Receipt className="size-4 text-muted-foreground" strokeWidth={1.5} />
            <h2 className="text-sm font-semibold">充值开票</h2>
            <span className="text-xs text-muted-foreground">管理充值记录与发票申请</span>
          </div>
        </div>
        <div className="p-5 space-y-4">
          {/* Warning banner */}
          <div className="flex items-center justify-between rounded-md border border-amber-200 bg-amber-50 px-4 py-2.5">
            <span className="text-xs text-amber-800">请先完成实名认证后再申请开具发票</span>
            <Button size="sm" variant="outline" className="h-7 text-xs border-amber-300 text-amber-800 hover:bg-amber-100">
              去认证
            </Button>
          </div>

          {/* Tabs */}
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

          {/* Search bar */}
          <div className="flex items-center gap-2">
            <Input
              placeholder="订单号"
              value={searchOrderId}
              onChange={(e) => setSearchOrderId(e.target.value)}
              className="h-8 w-44 text-sm"
            />
            <Button variant="ghost" size="sm" className="h-8 text-xs">查询</Button>
            <Button variant="ghost" size="sm" className="h-8 text-xs" onClick={() => setSearchOrderId('')}>重置</Button>
            <div className="ml-auto flex gap-2">
              <Button variant="ghost" size="sm" className="h-8 text-xs" disabled>批量开票</Button>
              <Button variant="outline" size="sm" className="h-8 text-xs" disabled>全部开票</Button>
            </div>
          </div>

          {/* Table */}
          {activeTab === 'topup' && (
            <div className="rounded-md border border-border overflow-hidden">
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
                  {filteredRecords(mockTopUpRecords, searchOrderId).map((record, i) => (
                    <tr key={record.id} className={cn(i % 2 === 1 && 'bg-muted/40', 'hover:bg-muted/50')}>
                      <td className="px-4 py-2.5 font-mono text-xs">{record.orderId}</td>
                      <td className="px-4 py-2.5">{record.method === 'alipay' ? '支付宝' : '微信'}</td>
                      <td className="px-4 py-2.5 text-right tabular-nums">¥{record.amount.toFixed(2)}</td>
                      <td className="px-4 py-2.5 text-right tabular-nums">¥{record.paidAmount.toFixed(2)}</td>
                      <td className="px-4 py-2.5 text-center"><InvoiceStatusBadge status={record.invoiceStatus} /></td>
                      <td className="px-4 py-2.5 text-xs text-muted-foreground">{record.createdAt}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
              {filteredRecords(mockTopUpRecords, searchOrderId).length === 0 && (
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
    </div>
  )
}

function StatCard({ icon: Icon, label, value }: { icon: React.ComponentType<{ className?: string; strokeWidth?: number }>; label: string; value: string }) {
  return (
    <div className="rounded-lg border border-border bg-card p-4 shadow-xs">
      <div className="flex items-center gap-2 text-xs text-muted-foreground">
        <Icon className="size-3.5" strokeWidth={1.5} />
        {label}
      </div>
      <p className="mt-2 text-xl font-semibold tabular-nums text-foreground">{value}</p>
    </div>
  )
}

function filteredRecords(records: TopUpRecord[], search: string) {
  if (!search) return records
  return records.filter((r) => r.orderId.toLowerCase().includes(search.toLowerCase()))
}
