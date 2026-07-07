export const STATUS_BADGE_STYLES = {
  success: 'bg-emerald-50 text-emerald-700',
  warning: 'bg-amber-50 text-amber-700',
  danger: 'bg-red-50 text-red-700',
  neutral: 'bg-slate-100 text-slate-600',
  info: 'bg-blue-50 text-blue-700',
  violet: 'bg-blue-50 text-blue-700',
} as const

export type StatusBadgeVariant = keyof typeof STATUS_BADGE_STYLES
