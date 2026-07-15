import { useInjectedQuery } from '@/features/query'
import type { NotificationCapabilities } from '@/api/types'
import type { AppApis } from '@/api/app-apis'

/**
 * Fetches backend notification capabilities (which channels are configured).
 * Used to determine whether to show Inbox UI or Toast-only mode.
 */
export function useNotificationCapabilities(injectedApis?: AppApis) {
  return useInjectedQuery<NotificationCapabilities>({
    injectedApis,
    queryKey: ['notifications', 'capabilities'],
    queryFn: (apis) => apis.notificationApi.getCapabilities(),
    staleTime: 5 * 60_000, // capabilities rarely change, cache for 5 minutes
  })
}
