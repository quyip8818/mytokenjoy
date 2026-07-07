import { useState } from 'react'
import { Gift, Receipt, ShoppingCart } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import type { PaymentMethod } from '@/features/wallet'

const PRESET_AMOUNTS = [10, 20, 50, 100, 200, 500]

interface RechargePanelProps {
  currency: string
  rechargePending: boolean
  onRecharge: (amount: number) => void
}

export function RechargePanel({ currency, rechargePending, onRecharge }: RechargePanelProps) {
  const [amount, setAmount] = useState('')
  const [paymentMethod, setPaymentMethod] = useState<PaymentMethod>('alipay')
  const [redemptionCode, setRedemptionCode] = useState('')
  const selectedAmount = amount ? Number(amount) : 0

  return (
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
      <div className="space-y-5 p-5">
        <div className="grid grid-cols-12 gap-4">
          <div className="col-span-5 space-y-1.5">
            <label className="text-xs font-medium text-muted-foreground">充值数量</label>
            <Input
              type="number"
              min="0"
              placeholder="充值数量，最低 ¥0"
              value={amount}
              onChange={(event) => setAmount(event.target.value)}
              className="h-9"
            />
            <p className="text-xs text-muted-foreground">
              实付金额：
              <span className="font-medium text-destructive">
                {selectedAmount} {currency}
              </span>
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

        <div className="rounded-md border border-border p-4">
          <div className="mb-3 flex items-center gap-2">
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
                    : 'border-border bg-card text-foreground hover:bg-muted',
                )}
              >
                <span className="text-sm font-semibold tabular-nums">{preset} ¥</span>
                <p className="mt-0.5 text-xs text-muted-foreground">实付 ¥{preset.toFixed(2)}</p>
              </button>
            ))}
          </div>
          <div className="mt-4 flex justify-end">
            <Button
              size="sm"
              disabled={rechargePending || selectedAmount <= 0}
              onClick={() => void onRecharge(selectedAmount)}
            >
              {rechargePending ? '充值中…' : '确认充值'}
            </Button>
          </div>
        </div>

        <div className="rounded-md border border-border p-4">
          <div className="mb-3 flex items-center gap-2">
            <Gift className="size-4 text-muted-foreground" strokeWidth={1.5} />
            <span className="text-xs font-medium">兑换码充值</span>
          </div>
          <div className="flex gap-2">
            <Input
              placeholder="请输入兑换码"
              value={redemptionCode}
              onChange={(event) => setRedemptionCode(event.target.value)}
              className="h-9 max-w-sm"
            />
            <Button size="sm" disabled>
              兑换额度
            </Button>
          </div>
          <p className="mt-2 text-xs text-muted-foreground">兑换码能力即将上线</p>
        </div>
      </div>
    </div>
  )
}
