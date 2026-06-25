import { useMemo, type ReactNode } from 'react'
import type { AppApis } from '@/api/app-apis'
import { defaultApis } from '@/api/app-apis'
import { SessionContextSchema } from '@/api/schemas/session'
import { queryKeys, useInjectedQuery } from '@/features/query'
import { SessionReactContext } from './context'
import { SessionGate } from './session-gate'
import type { AppSession } from './types'

interface AuthSessionProviderProps {
  children: ReactNode
  apis?: AppApis
}

export function AuthSessionProvider({ children, apis = defaultApis }: AuthSessionProviderProps) {
  const query = useInjectedQuery({
    injectedApis: apis,
    queryKey: queryKeys.session.current(),
    queryFn: async (a) => {
      const data = await a.sessionApi.getCurrent()
      const parsed = SessionContextSchema.safeParse(data)
      if (!parsed.success) {
        throw new Error('Invalid session response')
      }
      return parsed.data
    },
  })

  const session = useMemo<AppSession>(() => {
    const refreshSession = async () => {
      await query.refresh()
    }

    return {
      memberId: query.data?.member.id ?? '',
      member: query.data?.member ?? null,
      permissions: query.data?.permissions ?? [],
      readOnly: query.data?.readOnly ?? false,
      loading: query.loading,
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
