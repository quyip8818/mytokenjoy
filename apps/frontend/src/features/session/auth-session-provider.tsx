import { useCallback, useEffect, useMemo, useRef, type ReactNode } from 'react'
import { toast } from 'sonner'
import type { AppApis } from '@/api/app-apis'
import { defaultApis } from '@/api/app-apis'
import { setAuthzRevisionHandler, setForbiddenHandler } from '@/api/client'
import { SessionContextSchema } from '@/api/schemas/session'
import { queryKeys, useInjectedQuery } from '@/features/query'
import { AUTHZ_BROADCAST_CHANNEL, SESSION_FOCUS_REFRESH_MS } from './authz-sync'
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

  const authzRevisionRef = useRef(0)
  const lastSessionFetchRef = useRef(0)
  const forbiddenRetriedRef = useRef(new Set<string>())

  const refreshSession = useCallback(async () => {
    await query.refresh()
    lastSessionFetchRef.current = Date.now()
  }, [query])

  useEffect(() => {
    authzRevisionRef.current = query.data?.authzRevision ?? 0
    lastSessionFetchRef.current = Date.now()
  }, [query.data?.authzRevision])

  useEffect(() => {
    setAuthzRevisionHandler((revision) => {
      if (revision > authzRevisionRef.current) {
        void refreshSession()
      }
    })
    return () => setAuthzRevisionHandler(null)
  }, [refreshSession])

  useEffect(() => {
    setForbiddenHandler((path) => {
      if (forbiddenRetriedRef.current.has(path)) {
        toast.error('You do not have permission to perform this action')
        return
      }
      forbiddenRetriedRef.current.add(path)
      void refreshSession()
    })
    return () => setForbiddenHandler(null)
  }, [refreshSession])

  useEffect(() => {
    const onFocus = () => {
      if (Date.now() - lastSessionFetchRef.current > SESSION_FOCUS_REFRESH_MS) {
        void refreshSession()
      }
    }
    window.addEventListener('focus', onFocus)
    return () => window.removeEventListener('focus', onFocus)
  }, [refreshSession])

  useEffect(() => {
    if (typeof BroadcastChannel === 'undefined') return
    const channel = new BroadcastChannel(AUTHZ_BROADCAST_CHANNEL)
    channel.onmessage = () => {
      void refreshSession()
    }
    return () => channel.close()
  }, [refreshSession])

  const session = useMemo<AppSession>(() => {
    return {
      companyId: query.data?.companyId ?? 0,
      authzRevision: query.data?.authzRevision ?? 0,
      memberId: query.data?.member.id ?? '',
      member: query.data?.member ?? null,
      permissions: query.data?.permissions ?? [],
      readOnly: query.data?.readOnly ?? false,
      loading: query.loading,
      sessionError: query.error,
      refreshSession,
    }
  }, [query, refreshSession])

  return (
    <SessionReactContext.Provider value={session}>
      <SessionGate>{children}</SessionGate>
    </SessionReactContext.Provider>
  )
}
