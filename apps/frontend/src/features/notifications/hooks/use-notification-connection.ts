import { useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router'
import { useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { API_BASE_PATH } from '@/config/app'
import { getActionUrl } from '../lib/get-action-url'
import type { NotificationItem } from '@/api/types'

export interface NotificationSSEEvent {
  id: string
  eventType: string
  title: string
  body: string
  groupKey?: string
  category?: string
  payload?: Record<string, unknown>
}

// ponytail: track recent groupKeys to deduplicate rapid SSE pushes (防刷屏)
const recentGroupKeys = new Map<string, number>()
const DEDUP_WINDOW_MS = 10_000

function shouldShowToast(event: NotificationSSEEvent): boolean {
  if (!event.groupKey) return true
  const now = Date.now()
  const lastSeen = recentGroupKeys.get(event.groupKey)
  if (lastSeen && now - lastSeen < DEDUP_WINDOW_MS) return false
  recentGroupKeys.set(event.groupKey, now)
  // Cleanup old entries
  if (recentGroupKeys.size > 50) {
    for (const [key, ts] of recentGroupKeys) {
      if (now - ts > DEDUP_WINDOW_MS) recentGroupKeys.delete(key)
    }
  }
  return true
}

/**
 * Manages the SSE connection to /api/notifications/stream.
 * Pushes incoming notifications to the TanStack Query cache and shows toast.
 */
export function useNotificationConnection() {
  const [isConnected, setIsConnected] = useState(false)
  const queryClient = useQueryClient()
  const navigate = useNavigate()
  const navigateRef = useRef(navigate)
  useEffect(() => { navigateRef.current = navigate })
  const eventSourceRef = useRef<EventSource | null>(null)

  useEffect(() => {
    const url = `${API_BASE_PATH}/notifications/stream`
    const eventSource = new EventSource(url, { withCredentials: true })
    eventSourceRef.current = eventSource

    eventSource.addEventListener('connected', () => {
      setIsConnected(true)
    })

    eventSource.addEventListener('notification', (e) => {
      try {
        const notification: NotificationSSEEvent = JSON.parse(e.data)
        // Invalidate notification queries to refetch
        queryClient.invalidateQueries({ queryKey: ['notifications'] })

        // Show toast with dedup + clickable action
        if (shouldShowToast(notification)) {
          // Compute action URL from SSE event payload
          const actionUrl = getActionUrl({
            id: notification.id,
            eventType: notification.eventType,
            payload: notification.payload ?? {},
          } as NotificationItem)

          toast.info(notification.title, {
            description: notification.body || undefined,
            duration: 5000,
            ...(actionUrl && {
              action: {
                label: '查看',
                onClick: () => navigateRef.current(actionUrl),
              },
            }),
          })
        }
      } catch {
        // Ignore malformed events
      }
    })

    eventSource.onerror = () => {
      setIsConnected(false)
    }

    return () => {
      eventSource.close()
      eventSourceRef.current = null
      setIsConnected(false)
    }
  }, [queryClient])

  return { isBackendConnected: isConnected }
}
