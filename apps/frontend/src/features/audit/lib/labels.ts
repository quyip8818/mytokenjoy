import type { StatusBadgeVariant } from '@/lib/labels'

export const CALL_LOG_STATUS_VARIANTS: Record<string, StatusBadgeVariant> = {
  success: 'success',
  error: 'danger',
  filtered: 'warning',
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

export function getOperationActionBadgeVariant(action: string): StatusBadgeVariant {
  if (action.startsWith('key_')) return 'warning'
  if (action.startsWith('budget_')) return 'success'
  if (action.startsWith('permission') || action.startsWith('role')) return 'info'
  return 'neutral'
}
