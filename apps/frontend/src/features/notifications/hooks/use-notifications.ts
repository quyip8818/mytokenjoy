import { useInjectedQuery } from '@/features/query'
import type { NotificationItem, NotificationUnreadCount } from '@/api/types'
import type { AppApis } from '@/api/app-apis'

export function useNotifications(injectedApis?: AppApis) {
  return useInjectedQuery<NotificationItem[]>({
    injectedApis,
    queryKey: ['notifications'],
    queryFn: (apis) => apis.notificationApi.list({ limit: 50 }),
    refetchInterval: 60_000, // poll every 60s as backup to SSE
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
