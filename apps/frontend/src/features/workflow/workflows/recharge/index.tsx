import { useState, useCallback } from 'react'
import { Gift } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { toast } from '@/lib/toast'
import { useApis } from '@/api/use-apis'
import { IS_SAAS } from '@/config/app'
import { ApiError } from '@/api/client'
import type { PaymentMethod } from '@/features/billing'
import { billingKeys } from '@/features/billing'
import type { WorkflowComponentProps } from '@/features/workflow/types'
import { WorkflowPanelChrome, WorkflowPanelFooter } from '@/features/workflow/components/workflow-panel-chrome'
import { useQueryClient } from '@tanstack/react-query'

const PRESET_AMOUNTS = [10, 20, 50, 100, 200, 500]

// ponytail: error code → user-friendly message mapping for redeem failures.
const REDEEM_ERROR_MESSAGES: Record<string, string> = {
  INVALID_REDEMPTION_CODE: '兑换码无效，请检查输入',
  CODE_ALREADY_USED: '该兑换码已被使用',
  CODE_EXPIRED: '该兑换码已过期',
  CODE_DISABLED: '该兑换码已被禁用',
  TRIAL_NO_TOPUP: '试用环境不支持兑换，升级后可使用',
}

export function RechargeWorkflow({ entry, onClose }: WorkflowComponentProps<'recharge'>) {
  const { currency, onSuccess } = entry.payload
  const apis = useApis()
  const queryClient = useQueryClient()

  const [amount, setAmount] = useState('')
  const [paymentMethod, setPaymentMethod] = useState<PaymentMethod>('alipay')
  const [redemptionCode, setRedemptionCode] = useState('')
  const [pending, setPending] = useState(false)
  const [redeeming, setRedeeming] = useState(false)

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

  const handleRedeem = useCallback(async () => {
    const code = redemptionCode.trim()
    if (!code) return
    setRedeeming(true)
    try {
      const result = await apis.billingApi.redeem(code)
      toast.success(`成功充值 ¥${result.faceValue.toFixed(2)}`)
      setRedemptionCode('')
      await queryClient.invalidateQueries({ queryKey: billingKeys.wallet() })
      onSuccess?.()
    } catch (err) {
      const errorCode = err instanceof ApiError ? err.code : undefined
      const msg = (errorCode && REDEEM_ERROR_MESSAGES[errorCode]) || '兑换失败，请重试'
      toast.error(msg)
    } finally {
      setRedeeming(false)
    }
  }, [apis, redemptionCode, queryClient, onSuccess])

  // Auto-uppercase + strip spaces on input change.
  const handleRedemptionCodeChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setRedemptionCode(e.target.value.toUpperCase().replace(/\s/g, ''))
  }

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

        {/* Redemption code — SaaS only */}
        {IS_SAAS && (
          <div className="rounded-md border border-border p-4">
            <div className="mb-3 flex items-center gap-2">
              <Gift className="size-4 text-muted-foreground" strokeWidth={1.5} />
              <span className="text-sm font-medium">兑换码充值</span>
            </div>
            <div className="flex gap-2">
              <Input
                placeholder="TJ-XXXX-XXXX-XXXX"
                value={redemptionCode}
                onChange={handleRedemptionCodeChange}
                className="max-w-sm font-mono"
                onKeyDown={(e) => {
                  if (e.key === 'Enter' && !redeeming && redemptionCode.trim()) {
                    void handleRedeem()
                  }
                }}
              />
              <Button
                onClick={handleRedeem}
                disabled={redeeming || !redemptionCode.trim()}
              >
                {redeeming ? '兑换中…' : '兑换额度'}
              </Button>
            </div>
          </div>
        )}
      </div>
    </WorkflowPanelChrome>
  )
}
