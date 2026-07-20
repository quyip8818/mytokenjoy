import { useState } from 'react'
import { toast } from 'sonner'
import { ApiError } from '@/api/client'
import type { BudgetApproval } from '@/api/types'
import { Sheet, SheetContent, SheetHeader, SheetTitle } from '@/components/ui/sheet'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { cn } from '@/lib/utils'
import { formatDisplayCurrency } from '@/lib/quota-display'
import { CheckCircle, XCircle, Clock } from 'lucide-react'

interface BudgetApprovalDrawerProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  approvals: BudgetApproval[]
  onResolve: (
    id: string,
    data: { status: 'approved' | 'rejected'; rejectReason?: string },
  ) => Promise<void>
  onResolved: () => void
}

interface RejectState {
  id: string
  reason: string
}

function StatusBadge({ status }: { status: BudgetApproval['status'] }) {
  if (status === 'approved') {
    return (
      <span className="inline-flex items-center gap-1 rounded-full bg-emerald-50 px-2 py-0.5 text-xs font-medium text-emerald-700">
        <CheckCircle className="h-3 w-3" />
        已通过
      </span>
    )
  }
  if (status === 'rejected') {
    return (
      <span className="inline-flex items-center gap-1 rounded-full bg-red-50 px-2 py-0.5 text-xs font-medium text-red-700">
        <XCircle className="h-3 w-3" />
        已拒绝
      </span>
    )
  }
  return (
    <span className="inline-flex items-center gap-1 rounded-full bg-amber-50 px-2 py-0.5 text-xs font-medium text-amber-700">
      <Clock className="h-3 w-3" />
      待审批
    </span>
  )
}

function PendingItem({
  item,
  rejectState,
  resolving,
  onApprove,
  onRejectStart,
  onRejectCancel,
  onRejectReasonChange,
  onRejectConfirm,
}: {
  item: BudgetApproval
  rejectState: RejectState | null
  resolving: boolean
  onApprove: () => void
  onRejectStart: () => void
  onRejectCancel: () => void
  onRejectReasonChange: (reason: string) => void
  onRejectConfirm: () => void
}) {
  const isRejecting = rejectState?.id === item.id

  return (
    <div className="rounded-lg border border-border p-4">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <span className="text-sm font-medium text-foreground">{item.applicantName}</span>
            <span className="rounded-full bg-muted px-2 py-0.5 text-xs text-muted-foreground">
              {item.departmentName}
            </span>
          </div>
          <p className="mt-1.5 text-xs text-muted-foreground">{item.reason}</p>
          <p className="mt-1 text-xs text-muted-foreground">{item.createdAt}</p>
        </div>
        <span className="shrink-0 text-sm font-medium tabular-nums text-foreground">
          {formatDisplayCurrency(item.amount)}
        </span>
      </div>

      {!isRejecting ? (
        <div className="mt-3 flex items-center gap-2">
          <Button
            size="sm"
            className="h-7 bg-emerald-600 px-3 text-xs text-white hover:bg-emerald-700"
            disabled={resolving}
            onClick={onApprove}
            aria-label={`通过 ${item.applicantName} 的申请`}
          >
            通过
          </Button>
          <Button
            size="sm"
            variant="ghost"
            className="h-7 px-3 text-xs text-muted-foreground hover:text-red-600"
            disabled={resolving}
            onClick={onRejectStart}
            aria-label={`拒绝 ${item.applicantName} 的申请`}
          >
            拒绝
          </Button>
        </div>
      ) : (
        <div className="mt-3 space-y-2">
          <textarea
            placeholder="请输入拒绝原因…"
            className="w-full resize-none rounded-lg border border-input bg-background px-3 py-2 text-xs text-foreground placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-50"
            rows={3}
            value={rejectState.reason}
            onChange={(e) => onRejectReasonChange(e.target.value)}
            disabled={resolving}
            aria-label="拒绝原因"
          />
          <div className="flex items-center gap-2">
            <Button
              size="sm"
              variant="destructive"
              className="h-7 px-3 text-xs"
              disabled={resolving || !rejectState.reason.trim()}
              onClick={onRejectConfirm}
              aria-label="确认拒绝"
            >
              确认拒绝
            </Button>
            <Button
              size="sm"
              variant="ghost"
              className="h-7 px-3 text-xs text-muted-foreground"
              disabled={resolving}
              onClick={onRejectCancel}
              aria-label="取消拒绝"
            >
              取消
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}

function ResolvedItem({ item }: { item: BudgetApproval }) {
  return (
    <div className="rounded-lg border border-border p-4 opacity-80">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <span className="text-sm font-medium text-foreground">{item.applicantName}</span>
            <span className="rounded-full bg-muted px-2 py-0.5 text-xs text-muted-foreground">
              {item.departmentName}
            </span>
            <StatusBadge status={item.status} />
          </div>
          <p className="mt-1.5 text-xs text-muted-foreground">{item.reason}</p>
          {item.rejectReason && (
            <p className="mt-1 text-xs text-red-600">拒绝原因：{item.rejectReason}</p>
          )}
          <p
            className={cn(
              'mt-1 text-xs text-muted-foreground',
              item.resolvedAt && 'text-muted-foreground/70',
            )}
          >
            {item.resolvedAt ? `处理于 ${item.resolvedAt}` : item.createdAt}
          </p>
        </div>
        <span className="shrink-0 text-sm font-medium tabular-nums text-muted-foreground">
          {formatDisplayCurrency(item.amount)}
        </span>
      </div>
    </div>
  )
}

export function BudgetApprovalDrawer({
  open,
  onOpenChange,
  approvals,
  onResolve,
  onResolved,
}: BudgetApprovalDrawerProps) {
  const [rejectState, setRejectState] = useState<RejectState | null>(null)
  const [resolvingId, setResolvingId] = useState<string | null>(null)

  async function handleApprove(id: string) {
    setResolvingId(id)
    try {
      await onResolve(id, { status: 'approved' })
      toast.success('审批通过')
      onResolved()
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : '操作失败，请重试')
    } finally {
      setResolvingId(null)
    }
  }

  async function handleRejectConfirm(id: string, reason: string) {
    if (!reason.trim()) return
    setResolvingId(id)
    try {
      await onResolve(id, { status: 'rejected', rejectReason: reason.trim() })
      toast.success('已拒绝')
      setRejectState(null)
      onResolved()
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : '操作失败，请重试')
    } finally {
      setResolvingId(null)
    }
  }

  const pending = approvals.filter((a) => a.status === 'pending')
  const resolved = approvals.filter((a) => a.status !== 'pending')

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="right" className="flex w-[400px] flex-col gap-0 p-0 sm:max-w-[400px]">
        <SheetHeader className="border-b border-border px-5 py-4">
          <SheetTitle className="text-sm font-semibold">预算审批</SheetTitle>
        </SheetHeader>

        <div className="flex-1 overflow-y-auto px-5 py-4">
          <div className="space-y-6">
            <div>
              <h3 className="mb-3 text-xs font-medium uppercase text-muted-foreground">
                待审批
                {pending.length > 0 && (
                  <span className="ml-1.5 rounded-full bg-amber-100 px-1.5 py-0.5 text-amber-700">
                    {pending.length}
                  </span>
                )}
              </h3>
              {pending.length === 0 ? (
                <p className="rounded-lg border border-border py-6 text-center text-sm text-muted-foreground">
                  暂无待审批申请
                </p>
              ) : (
                <div className="space-y-3">
                  {pending.map((item) => (
                    <PendingItem
                      key={item.id}
                      item={item}
                      rejectState={rejectState?.id === item.id ? rejectState : null}
                      resolving={resolvingId === item.id}
                      onApprove={() => void handleApprove(item.id)}
                      onRejectStart={() => setRejectState({ id: item.id, reason: '' })}
                      onRejectCancel={() => setRejectState(null)}
                      onRejectReasonChange={(reason) =>
                        setRejectState((prev) => (prev ? { ...prev, reason } : null))
                      }
                      onRejectConfirm={() =>
                        void handleRejectConfirm(item.id, rejectState?.reason ?? '')
                      }
                    />
                  ))}
                </div>
              )}
            </div>

            {resolved.length > 0 && (
              <>
                <Separator />
                <div>
                  <h3 className="mb-3 text-xs font-medium uppercase text-muted-foreground">
                    已处理
                  </h3>
                  <div className="space-y-3">
                    {resolved.map((item) => (
                      <ResolvedItem key={item.id} item={item} />
                    ))}
                  </div>
                </div>
              </>
            )}
          </div>
        </div>
      </SheetContent>
    </Sheet>
  )
}
