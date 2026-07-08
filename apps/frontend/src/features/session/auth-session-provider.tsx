import { useCallback, useEffect, useMemo, useRef, type ReactNode } from 'react'
import { useLocation } from 'react-router'
import { toast } from 'sonner'
import type { AppApis } from '@/api/app-apis'
import { defaultApis } from '@/api/app-apis'
import { setAuthzRevisionHandler, setForbiddenHandler } from '@/api/client'
import { LOGIN_PATH } from '@/config/auth'
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
  const location = useLocation()
  const isLoginPage = location.pathname === LOGIN_PATH

  const query = useInjectedQuery({
    injectedApis: apis,
    queryKey: queryKeys.session.current(),
    enabled: !isLoginPage,
    retry: false,
    queryFn: (a) => a.sessionApi.getCurrent(),
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
        authzRevisionRef.current = revision
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
      memberId: query.data?.member?.id ?? '',
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
