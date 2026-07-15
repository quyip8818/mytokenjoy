import { useEffect, useRef, useState } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { API_BASE_PATH } from '@/config/app'

export interface NotificationSSEEvent {
  id: string
  eventType: string
  title: string
  body: string
}

/**
 * Manages the SSE connection to /api/notifications/stream.
 * Pushes incoming notifications to the TanStack Query cache and shows toast.
 */
export function useNotificationConnection() {
  const [isConnected, setIsConnected] = useState(false)
  const queryClient = useQueryClient()
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
        queryClient.invalidateQueries({ queryKey: ['notifications', 'unread-count'] })
        // Show toast for the new notification
        toast.info(notification.title, {
          description: notification.body || undefined,
        })
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
