import {
  AlertTriangle,
  BarChart3,
  Key,
  Settings,
  ShieldAlert,
  TrendingUp,
  type LucideIcon,
} from 'lucide-react'

export interface CategoryConfig {
  label: string
  icon: LucideIcon
  color: string // tailwind text color class
}

export const CATEGORY_MAP: Record<string, CategoryConfig> = {
  budget_alert: { label: '预算告警', icon: TrendingUp, color: 'text-orange-500' },
  key_expiration: { label: 'Key 到期', icon: Key, color: 'text-amber-500' },
  usage_report: { label: '用量报告', icon: BarChart3, color: 'text-blue-500' },
  security_event: { label: '安全事件', icon: ShieldAlert, color: 'text-red-500' },
  system_maintenance: { label: '系统维护', icon: Settings, color: 'text-slate-500' },
  overrun: { label: '超支通知', icon: AlertTriangle, color: 'text-rose-500' },
}

export const ALL_CATEGORIES = Object.entries(CATEGORY_MAP).map(([key, cfg]) => ({
  key,
  ...cfg,
}))
