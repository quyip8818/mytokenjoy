import type { Platform } from '@/api/types'

export const PLATFORM_LABELS: Record<Platform, string> = {
  feishu: '飞书',
  dingtalk: '钉钉',
  wecom: '企业微信',
}

export const PROVIDER_LABELS: Record<string, string> = {
  openai: 'OpenAI',
  anthropic: 'Anthropic',
  deepseek: 'DeepSeek',
  qwen: '通义千问',
  custom: '自定义',
}

export const PROVIDER_BADGE_STYLES: Record<string, string> = {
  openai: 'bg-emerald-50 text-emerald-700 border-emerald-100',
  anthropic: 'bg-orange-50 text-orange-700 border-orange-100',
  deepseek: 'bg-blue-50 text-blue-700 border-blue-100',
  qwen: 'bg-purple-50 text-purple-700 border-purple-100',
  custom: 'bg-slate-50 text-slate-700 border-slate-100',
}

export const PROVIDER_CHIP_STYLES: Record<string, string> = {
  openai: 'bg-emerald-50 text-emerald-700',
  anthropic: 'bg-amber-50 text-amber-700',
  deepseek: 'bg-blue-50 text-blue-700',
  qwen: 'bg-purple-50 text-purple-700',
  custom: 'bg-slate-100 text-slate-600',
}

export const STATUS_BADGE_STYLES = {
  success: 'bg-emerald-50 text-emerald-700',
  warning: 'bg-amber-50 text-amber-700',
  danger: 'bg-red-50 text-red-700',
  neutral: 'bg-slate-100 text-slate-600',
  info: 'bg-blue-50 text-blue-700',
  violet: 'bg-blue-50 text-blue-700',
} as const

export type StatusBadgeVariant = keyof typeof STATUS_BADGE_STYLES

export function getOperationActionBadgeVariant(action: string): StatusBadgeVariant {
  if (action.startsWith('key_')) return 'warning'
  if (action.startsWith('budget_')) return 'success'
  if (action.startsWith('permission') || action.startsWith('role')) return 'info'
  return 'neutral'
}

export const CALL_LOG_STATUS_VARIANTS: Record<string, StatusBadgeVariant> = {
  success: 'success',
  error: 'danger',
  filtered: 'warning',
}

export const SYNC_RESULT_VARIANTS: Record<string, StatusBadgeVariant> = {
  success: 'success',
  partial_failure: 'warning',
  failure: 'danger',
}

export const CALL_LOG_STATUS_LABELS: Record<string, string> = {
  success: '成功',
  error: '错误',
  filtered: '已过滤',
}

export const OPERATION_ACTION_LABELS: Record<string, string> = {
  key_create: 'Key 创建',
  key_disable: 'Key 禁用',
  key_rotate: 'Key 轮转',
  budget_change: '预算变更',
  budget_approve: '预算审批',
  permission_change: '权限变更',
  role_assign: '角色分配',
  model_whitelist_change: '白名单变更',
  member_add: '成员添加',
  member_remove: '成员移除',
  org_structure_change: '组织结构变更',
}
