import { useCallback, useEffect, useMemo, useRef, type ReactNode } from 'react'
import { useLocation } from 'react-router'
import { toast } from 'sonner'
import type { AppApis } from '@/api/app-apis'
import { defaultApis } from '@/api/app-apis'
import { setAuthzRevisionHandler, setForbiddenHandler } from '@/api/client'
import { LOGIN_PATH } from '@/config/auth'
import { queryKeys, useInjectedQuery } from '@/features/query'
import { createBillingExchange, setActiveBillingExchange } from '@/lib/points'
import { AUTHZ_BROADCAST_CHANNEL, SESSION_FOCUS_REFRESH_MS } from './authz-sync'
import { SessionReactContext } from './context'
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
  const refreshSessionRef = useRef<(() => Promise<void>) | undefined>(undefined)
  const refreshTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    refreshSessionRef.current = async () => {
      await query.refresh()
      lastSessionFetchRef.current = Date.now()
    }
  })

  const refreshSession = useCallback(async () => {
    await refreshSessionRef.current?.()
  }, [])

  const debouncedRefreshSession = useCallback(() => {
    if (refreshTimerRef.current) return
    refreshTimerRef.current = setTimeout(() => {
      refreshTimerRef.current = null
      if (Date.now() - lastSessionFetchRef.current < 2000) return
      void refreshSession()
    }, 500)
  }, [refreshSession])

  useEffect(() => {
    return () => {
      if (refreshTimerRef.current) clearTimeout(refreshTimerRef.current)
    }
  }, [])

  useEffect(() => {
    authzRevisionRef.current = query.data?.authzRevision ?? 0
    lastSessionFetchRef.current = Date.now()
  }, [query.data?.authzRevision])

  useEffect(() => {
    setAuthzRevisionHandler((revision) => {
      if (revision > authzRevisionRef.current) {
        authzRevisionRef.current = revision
        debouncedRefreshSession()
      }
    })
    return () => setAuthzRevisionHandler(null)
  }, [debouncedRefreshSession])

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
      billingCurrency: query.data?.billingCurrency ?? 'CNY',
      pointsPerUnit: query.data?.pointsPerUnit ?? 0,
      loading: query.loading,
      sessionError: query.error,
      refreshSession,
    }
  }, [query, refreshSession])

  useEffect(() => {
    setActiveBillingExchange(createBillingExchange(session.pointsPerUnit || undefined))
  }, [session.pointsPerUnit])

  return <SessionReactContext.Provider value={session}>{children}</SessionReactContext.Provider>
}
