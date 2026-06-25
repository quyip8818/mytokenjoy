import { useMemo, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import type { AppApis } from '@/api/app-apis'
import { defaultApis } from '@/api/app-apis'
import { queryKeys } from '@/features/query/query-keys'
import { SessionReactContext } from './context'
import { SessionGate } from './session-gate'
import type { AppSession } from './types'

interface AuthSessionProviderProps {
  children: ReactNode
  apis?: Pick<AppApis, 'sessionApi'>
}

export function AuthSessionProvider({ children, apis = defaultApis }: AuthSessionProviderProps) {
  const query = useQuery({
    queryKey: queryKeys.session.current(),
    queryFn: () => apis.sessionApi.getCurrent(),
  })

  const session = useMemo<AppSession>(() => {
    const refreshSession = async () => {
      await query.refetch()
    }

    return {
      memberId: query.data?.member.id ?? '',
      member: query.data?.member ?? null,
      permissions: query.data?.permissions ?? [],
      readOnly: query.data?.readOnly ?? false,
      loading: query.isLoading,
      sessionError: query.error,
      refreshSession,
    }
  }, [query])

  return (
    <SessionReactContext.Provider value={session}>
      <SessionGate>{children}</SessionGate>
    </SessionReactContext.Provider>
  )
}
