import type { ReactNode } from 'react'
import { defaultApis } from '@/api/app-apis'
import { ApiProvider } from '@/api/context'
import { QueryProvider } from '@/features/query'
import { AuthSessionProvider, SessionNavigationBridge } from '@/features/session'
import { AuthUnauthorizedBridge } from '@/components/auth/auth-unauthorized-bridge'

export function AppProviders({ children }: { children: ReactNode }) {
  return (
    <ApiProvider apis={defaultApis}>
      <QueryProvider>
        <AuthSessionProvider>
          <AuthUnauthorizedBridge />
          <SessionNavigationBridge />
          {children}
        </AuthSessionProvider>
      </QueryProvider>
    </ApiProvider>
  )
}
