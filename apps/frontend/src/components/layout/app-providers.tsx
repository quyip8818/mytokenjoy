import type { ReactNode } from 'react'
import { defaultApis } from '@/api/app-apis'
import { ApiProvider } from '@/api/context'
import { QueryProvider } from '@/features/query'
import { NotificationProvider } from '@/features/notifications'
import { AuthSessionProvider, AuthUnauthorizedBridge } from '@/features/session'
import { TooltipProvider } from '@/components/ui/tooltip'

export function AppProviders({ children }: { children: ReactNode }) {
  return (
    <ApiProvider apis={defaultApis}>
      <QueryProvider>
        <AuthSessionProvider>
          <AuthUnauthorizedBridge />
          <TooltipProvider>
            <NotificationProvider>{children}</NotificationProvider>
          </TooltipProvider>
        </AuthSessionProvider>
      </QueryProvider>
    </ApiProvider>
  )
}
