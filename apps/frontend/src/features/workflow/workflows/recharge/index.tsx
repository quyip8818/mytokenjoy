import { useState, useCallback } from 'react'
import { Gift } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { toast } from '@/lib/toast'
import { useApis } from '@/api/use-apis'
import type { PaymentMethod } from '@/features/billing'
import type { WorkflowComponentProps } from '@/features/workflow/types'
import { WorkflowPanelChrome, WorkflowPanelFooter } from '@/features/workflow/components/workflow-panel-chrome'

const PRESET_AMOUNTS = [10, 20, 50, 100, 200, 500]

export function RechargeWorkflow({ entry, onClose }: WorkflowComponentProps<'recharge'>) {
  const { currency, onSuccess } = entry.payload
  const apis = useApis()

  const [amount, setAmount] = useState('')
  const [paymentMethod, setPaymentMethod] = useState<PaymentMethod>('alipay')
  const [redemptionCode, setRedemptionCode] = useState('')
  const [pending, setPending] = useState(false)

  const selectedAmount = amount ? Number(amount) : 0

  const handleRecharge = useCallback(async () => {
    if (selectedAmount <= 0) return
    setPending(true)
    try {
      const order = await apis.billingApi.recharge({
        amount: selectedAmount,
        idempotencyKey: crypto.randomUUID(),
      })
      await apis.billingApi.confirmRecharge(order.id)
      toast.success('充值成功')
      onSuccess?.()
      onClose()
    } catch {
      toast.error('充值失败，请重试')
    } finally {
      setPending(false)
    }
  }, [apis, selectedAmount, onSuccess, onClose])

  return (
    <WorkflowPanelChrome
      title="账户充值"
      onClose={onClose}
      footer={
        <WorkflowPanelFooter
          onCancel={onClose}
          primaryLabel={pending ? '充值中…' : '确认充值'}
          onPrimary={handleRecharge}
          primaryDisabled={pending || selectedAmount <= 0}
          primaryDisabledReason={
            pending ? '充值中，请稍候' : selectedAmount <= 0 ? '请选择充值金额' : undefined
          }
        />
      }
    >
      <div className="space-y-6">
        {/* Amount + payment method */}
        <div className="space-y-4">
          <div className="space-y-1.5">
            <label className="text-sm font-medium text-muted-foreground">充值数量</label>
            <Input
              type="number"
              min="0"
              placeholder="充值数量，最低 ¥0"
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
            />
            <p className="text-xs text-muted-foreground">
              实付金额：
              <span className="font-medium text-destructive">
                {selectedAmount} {currency}
              </span>
            </p>
          </div>

          <div className="space-y-1.5">
            <label className="text-sm font-medium text-muted-foreground">选择支付方式</label>
            <div className="flex gap-2">
              <Button
                variant={paymentMethod === 'alipay' ? 'default' : 'outline'}
                onClick={() => setPaymentMethod('alipay')}
                className="gap-1.5"
              >
                支付宝
              </Button>
              <Button
                variant={paymentMethod === 'wechat' ? 'default' : 'outline'}
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
          <div className="mb-3 flex items-center gap-2">
            <span className="text-sm font-medium">选择充值额度</span>
            <span className="text-xs text-muted-foreground">如需开发票，请联系客服</span>
          </div>
          <div className="grid grid-cols-3 gap-2">
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
        </div>

        {/* Redemption code */}
        <div className="rounded-md border border-border p-4">
          <div className="mb-3 flex items-center gap-2">
            <Gift className="size-4 text-muted-foreground" strokeWidth={1.5} />
            <span className="text-sm font-medium">兑换码充值</span>
          </div>
          <div className="flex gap-2">
            <Input
              placeholder="请输入兑换码"
              value={redemptionCode}
              onChange={(e) => setRedemptionCode(e.target.value)}
              className="max-w-sm"
            />
            <Button disabled disabledReason="兑换码能力即将上线">
              兑换额度
            </Button>
          </div>
        </div>
      </div>
    </WorkflowPanelChrome>
  )
}
