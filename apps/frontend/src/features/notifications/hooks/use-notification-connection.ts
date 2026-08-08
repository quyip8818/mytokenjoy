import { useEffect, useRef, useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
import { useQueryClient } from '@tanstack/react-query'
import { toast } from '@/lib/toast'
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

// ponytail: 模块级去重 map，防 SSE 快速推送刷屏
// 天花板：内存上限 50 条；升级路径：改用 LRU
const recentGroupKeys = new Map<string, number>()
const DEDUP_WINDOW_MS = 10_000

function shouldShowToast(event: NotificationSSEEvent): boolean {
  if (!event.groupKey) return true
  const now = Date.now()
  const lastSeen = recentGroupKeys.get(event.groupKey)
  if (lastSeen && now - lastSeen < DEDUP_WINDOW_MS) return false
  recentGroupKeys.set(event.groupKey, now)
  if (recentGroupKeys.size > 50) {
    for (const [key, ts] of recentGroupKeys) {
      if (now - ts > DEDUP_WINDOW_MS) recentGroupKeys.delete(key)
    }
  }
  return true
}

// ponytail: 指数退避重连，防后端不可达时刷屏
// 天花板：固定 5 次上限；升级路径：可配置 + jitter
const MAX_RETRIES = 5
const BASE_DELAY_MS = 2000

/**
 * Manages the SSE connection to /api/notifications/stream.
 * Pushes incoming notifications into TanStack Query cache and shows toast.
 */
export function useNotificationConnection() {
  const [isConnected, setIsConnected] = useState(false)
  const queryClient = useQueryClient()
  const navigate = useNavigate()
  const navigateRef = useRef(navigate)
  useEffect(() => { navigateRef.current = navigate })
  const eventSourceRef = useRef<EventSource | null>(null)

  useEffect(() => {
    let retryCount = 0
    let retryTimer: ReturnType<typeof setTimeout> | null = null
    let disposed = false

    function handleNotification(e: MessageEvent) {
      try {
        const n: NotificationSSEEvent = JSON.parse(e.data)
        queryClient.invalidateQueries({ queryKey: ['notifications'] })
        if (!shouldShowToast(n)) return

        const actionUrl = getActionUrl({
          id: n.id,
          eventType: n.eventType,
          payload: n.payload ?? {},
        } as NotificationItem)

        toast.info(n.title, {
          description: n.body || undefined,
          duration: 5000,
          ...(actionUrl && {
            action: { label: '查看', onClick: () => navigateRef.current({ to: actionUrl }) },
          }),
        })
      } catch { /* malformed event — skip */ }
    }

    function connect() {
      if (disposed) return
      const es = new EventSource(`${API_BASE_PATH}/notifications/stream`, { withCredentials: true })
      eventSourceRef.current = es

      es.addEventListener('connected', () => {
        retryCount = 0
        setIsConnected(true)
      })
      es.addEventListener('notification', handleNotification)

      es.onerror = () => {
        setIsConnected(false)
        es.close()
        eventSourceRef.current = null
        if (disposed || retryCount >= MAX_RETRIES) return
        const delay = BASE_DELAY_MS * 2 ** retryCount
        retryCount++
        retryTimer = setTimeout(connect, delay)
      }
    }

    function handleVisibility() {
      if (document.visibilityState !== 'visible' || disposed) return
      if (eventSourceRef.current) return // 已连接，不重复
      // 清掉可能 pending 的 retry timer 防止双重连接
      if (retryTimer) { clearTimeout(retryTimer); retryTimer = null }
      retryCount = 0
      connect()
    }

    connect()
    document.addEventListener('visibilitychange', handleVisibility)

    return () => {
      disposed = true
      if (retryTimer) clearTimeout(retryTimer)
      eventSourceRef.current?.close()
      eventSourceRef.current = null
      document.removeEventListener('visibilitychange', handleVisibility)
      setIsConnected(false)
    }
  }, [queryClient])

  return { isBackendConnected: isConnected }
}
