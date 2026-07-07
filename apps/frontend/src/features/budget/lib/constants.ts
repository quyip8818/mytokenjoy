import type { OverrunPolicy } from '@/api/types'

export const POLICY_LABELS: Record<OverrunPolicy, { label: string; className: string }> = {
  hard_reject: { label: '硬拒绝', className: 'bg-red-50 text-red-700 border-red-200' },
  approval: { label: '审批追加', className: 'bg-primary/10 text-primary border-primary/20' },
  downgrade: { label: '降级路由', className: 'bg-amber-50 text-amber-700 border-amber-200' },
}

export const ALERT_PRESET_THRESHOLDS = [
  { label: '80%, 90%, 100%', value: [80, 90, 100] },
  { label: '90%, 100%', value: [90, 100] },
  { label: '仅 100%', value: [100] },
] as const
