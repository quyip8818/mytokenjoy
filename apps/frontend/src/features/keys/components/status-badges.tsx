import { Badge } from '@/components/ui/badge'
import { StatusBadge } from '@/components/ui/status-badge'
import { PROVIDER_BADGE_STYLES, PROVIDER_LABELS } from '@/features/models/lib/labels'

export function KeyPrefixBadge({ prefix }: { prefix: string }) {
  return (
    <span className="rounded bg-blue-50 px-1.5 py-0.5 font-mono text-xs text-muted-foreground">
      {prefix}
    </span>
  )
}

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
      return <StatusBadge variant="success">正常</StatusBadge>
    case 'disabled':
      return <StatusBadge variant="neutral">已禁用</StatusBadge>
    case 'error':
      return <StatusBadge variant="danger">异常</StatusBadge>
    case 'expired':
      return <StatusBadge variant="danger">已过期</StatusBadge>
    default:
      return <Badge variant="outline">{status}</Badge>
  }
}

export function ApprovalStatusBadge({ status }: { status: string }) {
  switch (status) {
    case 'pending':
      return <StatusBadge variant="warning">待审批</StatusBadge>
    case 'approved':
      return <StatusBadge variant="success">已通过</StatusBadge>
    case 'rejected':
      return <StatusBadge variant="danger">已拒绝</StatusBadge>
    default:
      return <Badge variant="outline">{status}</Badge>
  }
}
