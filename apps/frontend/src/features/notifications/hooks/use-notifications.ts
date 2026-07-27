import { useInjectedQuery } from '@/features/query'
import type { NotificationItem, NotificationUnreadCount } from '@/api/types'
import type { AppApis } from '@/api/app-apis'

/**
 * Fetches recent notifications for the bell popover (simple flat list, latest 8).
 */
export function useNotifications(injectedApis?: AppApis) {
  return useInjectedQuery<NotificationItem[]>({
    injectedApis,
    queryKey: ['notifications', 'popover'],
    queryFn: async (apis) => {
      const res = await apis.notificationApi.list({ limit: 8, grouped: true })
      return res.items
    },
    refetchInterval: 60_000,
  })
}

export function useUnreadCount(injectedApis?: AppApis) {
  return useInjectedQuery<NotificationUnreadCount>({
    injectedApis,
    queryKey: ['notifications', 'unread-count'],
    queryFn: (apis) => apis.notificationApi.unreadCount(),
    refetchInterval: 30_000,
  })
}
