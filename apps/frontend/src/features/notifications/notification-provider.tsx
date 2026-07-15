import type { ReactNode } from 'react'
import { useSession } from '@/features/session/use-session'
import { useNotificationConnection } from './hooks/use-notification-connection'

/**
 * Initializes the SSE notification connection when user is authenticated.
 * Must be placed inside AuthSessionProvider and QueryProvider.
 */
function NotificationConnectionInit() {
  useNotificationConnection()
  return null
}

export function NotificationProvider({ children }: { children: ReactNode }) {
  const { memberId } = useSession()

  return (
    <>
      {memberId && <NotificationConnectionInit />}
      {children}
    </>
  )
}
