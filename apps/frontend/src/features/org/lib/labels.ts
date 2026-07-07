import type { Platform, SyncLog } from '@/api/types'
import type { StatusBadgeVariant } from '@/lib/labels'

export const PLATFORM_LABELS: Record<Platform, string> = {
  feishu: '飞书',
  dingtalk: '钉钉',
  wecom: '企业微信',
}

export const SYNC_RESULT_VARIANTS: Record<string, StatusBadgeVariant> = {
  success: 'success',
  partial_failure: 'warning',
  failure: 'danger',
}

export const SYNC_TYPE_LABELS: Record<SyncLog['type'], string> = {
  scheduled: '定时',
  manual: '手动',
}

export const SYNC_RESULT_LABELS: Record<SyncLog['result'], string> = {
  success: '成功',
  partial_failure: '部分失败',
  failure: '失败',
}
