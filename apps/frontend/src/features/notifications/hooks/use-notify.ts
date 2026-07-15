import { useCallback } from 'react'
import { toast } from 'sonner'
import { useNotificationCapabilities } from './use-notification-capabilities'

type ToastVariant = 'info' | 'success' | 'warning' | 'error'

interface NotifyOptions {
  title: string
  body?: string
  priority?: 'critical' | 'high' | 'normal' | 'low'
}

function mapPriorityToVariant(priority?: string): ToastVariant {
  switch (priority) {
    case 'critical':
      return 'error'
    case 'high':
      return 'warning'
    case 'low':
      return 'info'
    default:
      return 'info'
  }
}

/**
 * Unified notification entry point for frontend-triggered notifications.
 * - When backend has in_app configured: notifications arrive via SSE → Inbox (handled by NotificationProvider).
 * - When backend has no in_app channel: falls back to toast display.
 *
 * NOTE: This hook does NOT create its own SSE connection. The SSE connection
 * is managed by NotificationProvider at the app level.
 */
export function useNotify() {
  const { data: capabilities } = useNotificationCapabilities()
  const isBackendConnected = capabilities?.inAppConfigured ?? false

  const notify = useCallback(
    (options: NotifyOptions) => {
      if (isBackendConnected) {
        // Backend handles delivery — notifications come through SSE → Inbox
        return
      }

      // Fallback: no backend in_app channel, show as toast directly
      const variant = mapPriorityToVariant(options.priority)
      toast[variant](options.title, {
        description: options.body || undefined,
      })
    },
    [isBackendConnected],
  )

  return { notify, isBackendConnected }
}
