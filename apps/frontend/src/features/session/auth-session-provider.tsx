import { useMemo, type ReactNode } from 'react'
import type { AppApis } from '@/api/app-apis'
import { defaultApis } from '@/api/app-apis'
import { useAsyncResource } from '@/hooks/use-async-resource'
import { SessionReactContext } from './context'
import { SessionGate } from './session-gate'
import type { AppSession } from './types'

interface AuthSessionProviderProps {
  children: ReactNode
  apis?: Pick<AppApis, 'sessionApi'>
}

export function AuthSessionProvider({ children, apis = defaultApis }: AuthSessionProviderProps) {
  const {
    data,
    loading,
    error: sessionError,
    refresh: refreshSession,
  } = useAsyncResource(() => apis.sessionApi.getCurrent(), [apis])

  const session = useMemo<AppSession>(
    () => ({
      memberId: data?.member.id ?? '',
      member: data?.member ?? null,
      permissions: data?.permissions ?? [],
      readOnly: data?.readOnly ?? false,
      loading,
      sessionError,
      refreshSession,
    }),
    [data, loading, sessionError, refreshSession],
  )

  return (
    <SessionReactContext.Provider value={session}>
      <SessionGate>{children}</SessionGate>
    </SessionReactContext.Provider>
  )
}
