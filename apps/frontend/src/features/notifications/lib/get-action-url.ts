import type { NotificationItem } from '@/api/types'

/**
 * Computes the navigation URL for a notification based on its eventType + payload.
 * Returns null if no specific page is associated.
 */
export function getActionUrl(notification: NotificationItem): string | null {
  const { eventType, payload } = notification

  switch (eventType) {
    case 'budget_alert_reached':
      return '/budget/alerts'
    case 'overrun_blocked':
    case 'overdraft_expanded':
      if (payload?.projectID) return `/keys/platform?projectId=${payload.projectID}`
      return '/keys/platform'
    case 'key_expiring_soon':
    case 'key_expired':
      if (payload?.keyID) return `/keys/platform?highlight=${payload.keyID}`
      return '/keys/platform'
    case 'usage_weekly_report':
      return '/dashboard/cost'
    case 'security_login_new_device':
      return '/me/settings'
    case 'system_maintenance_scheduled':
      return null // no specific target
    case 'sync_threshold_exceeded':
      return '/models/sync'
    default:
      return null
  }
}
