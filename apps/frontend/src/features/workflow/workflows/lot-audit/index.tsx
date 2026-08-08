import { Button } from '@/components/ui/button'
import type { WorkflowComponentProps } from '@/features/workflow/types'
import { WorkflowPanelChrome } from '@/features/workflow/components/workflow-panel-chrome'
import { useLotAudit } from './use-lot-audit'
import { LotRow } from './lot-row'

export function LotAuditWorkflow({ entry, onClose }: WorkflowComponentProps<'lot-audit'>) {
  const { companyId, companyName, readonly, onSuccess } = entry.payload
  const {
    lots,
    walletRemainQuota,
    loading,
    refundTarget,
    refundAmount,
    setRefundAmount,
    refunding,
    openRefund,
    closeRefund,
    handleRefund,
  } = useLotAudit(companyId, readonly)

  const handleRefundConfirm = async () => {
    await handleRefund()
    onSuccess?.()
  }

  const totalGranted = lots.reduce((sum, l) => sum + l.quotaGranted / l.quotaPerUnit, 0)

  return (
    <WorkflowPanelChrome title={`Lot 审计 — ${companyName}`} onClose={onClose}>
      <div className="space-y-4">
        {/* Summary */}
        <div className="flex gap-4 text-sm rounded-lg border border-border/60 bg-muted/50 p-4">
          <div>
            <span className="text-muted-foreground">钱包剩余: </span>
            <span className="font-medium">
              {loading ? '—' : `${walletRemainQuota.toLocaleString()} quota`}
            </span>
          </div>
          <div>
            <span className="text-muted-foreground">累计配额: </span>
            <span className="font-medium">
              {loading
                ? '—'
                : `¥${totalGranted.toLocaleString('zh-CN', { minimumFractionDigits: 2 })}`}
            </span>
          </div>
          <div>
            <span className="text-muted-foreground">批次数: </span>
            <span className="font-medium">{loading ? '—' : lots.length}</span>
          </div>
        </div>

        {/* Lot list */}
        {loading ? (
          <div className="flex justify-center py-8 text-sm text-muted-foreground">加载中…</div>
        ) : lots.length === 0 ? (
          <div className="flex justify-center py-8 text-sm text-muted-foreground">暂无批次记录</div>
        ) : (
          <div className="space-y-3">
            {lots.map((lot) => (
              <LotRow key={lot.id} lot={lot} readonly={readonly} onRefund={openRefund} />
            ))}
          </div>
        )}

        {/* Refund dialog */}
        {refundTarget && (
          <div
            className="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
            onClick={closeRefund}
          >
            <div
              className="w-full max-w-sm rounded-lg bg-white p-6 shadow-xl"
              onClick={(e) => e.stopPropagation()}
            >
              <h3 className="text-base font-semibold">退费</h3>
              <p className="mt-1 text-sm text-muted-foreground">
                批次类型: {refundTarget.lotKind} | 剩余: ¥
                {(refundTarget.quotaRemaining / refundTarget.quotaPerUnit).toFixed(2)}
              </p>
              <label className="mt-4 block text-sm">
                <span className="text-muted-foreground">退费金额 (元)</span>
                <input
                  type="number"
                  className="mt-1 w-full rounded-md border px-3 py-2 text-sm"
                  placeholder="0.00"
                  value={refundAmount}
                  onChange={(e) => setRefundAmount(e.target.value)}
                  autoFocus
                />
              </label>
              <div className="mt-5 flex justify-end gap-2">
                <Button variant="outline" onClick={closeRefund}>
                  取消
                </Button>
                <Button disabled={refunding} onClick={handleRefundConfirm}>
                  {refunding ? '处理中…' : '确认退费'}
                </Button>
              </div>
            </div>
          </div>
        )}
      </div>
    </WorkflowPanelChrome>
  )
}
