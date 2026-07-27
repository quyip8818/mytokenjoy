import type { ReactNode } from 'react'
import { defaultApis } from '@/api/app-apis'
import { ApiProvider } from '@/api/context'
import { QueryProvider } from '@/features/query'
import { NotificationProvider } from '@/features/notifications'
import { AuthSessionProvider, AuthUnauthorizedBridge } from '@/features/session'

export function AppProviders({ children }: { children: ReactNode }) {
  return (
    <ApiProvider apis={defaultApis}>
      <QueryProvider>
        <AuthSessionProvider>
          <AuthUnauthorizedBridge />
          <NotificationProvider>{children}</NotificationProvider>
        </AuthSessionProvider>
      </QueryProvider>
    </ApiProvider>
  )
}
