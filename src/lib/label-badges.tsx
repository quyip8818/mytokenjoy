import { Badge } from '@/components/ui/badge'
import { PROVIDER_BADGE_STYLES, PROVIDER_LABELS } from '@/lib/labels'

export function ProviderBadge({ provider }: { provider: string }) {
  return (
    <Badge
      variant="outline"
      className={PROVIDER_BADGE_STYLES[provider] ?? PROVIDER_BADGE_STYLES.custom}
    >
      {PROVIDER_LABELS[provider] ?? provider}
    </Badge>
  )
}

export function KeyStatusBadge({ status }: { status: string }) {
  switch (status) {
    case 'active':
      return (
        <span className="inline-flex items-center rounded-full bg-emerald-50 px-2 py-0.5 text-xs font-medium text-emerald-700">
          正常
        </span>
      )
    case 'disabled':
      return (
        <span className="inline-flex items-center rounded-full bg-slate-100 px-2 py-0.5 text-xs font-medium text-slate-600">
          已禁用
        </span>
      )
    case 'error':
    case 'expired':
      return (
        <span className="inline-flex items-center rounded-full bg-red-50 px-2 py-0.5 text-xs font-medium text-red-700">
          {status === 'error' ? '异常' : '已过期'}
        </span>
      )
    default:
      return <Badge variant="outline">{status}</Badge>
  }
}

export function ApprovalStatusBadge({ status }: { status: string }) {
  switch (status) {
    case 'pending':
      return (
        <span className="inline-flex items-center rounded-full bg-amber-50 px-2 py-0.5 text-xs font-medium text-amber-700">
          待审批
        </span>
      )
    case 'approved':
      return (
        <span className="inline-flex items-center rounded-full bg-emerald-50 px-2 py-0.5 text-xs font-medium text-emerald-700">
          已通过
        </span>
      )
    case 'rejected':
      return (
        <span className="inline-flex items-center rounded-full bg-red-50 px-2 py-0.5 text-xs font-medium text-red-700">
          已拒绝
        </span>
      )
    default:
      return <Badge variant="outline">{status}</Badge>
  }
}
