import { useEffect, useRef } from 'react'
import { useNavigate } from 'react-router'
import { getDefaultHomePath } from '@/lib/permissions'
import { useRouteAccess } from '@/hooks/use-route-access'

export interface RouteRedirectContext {
  permissions: readonly string[]
  pathname: string
}

export interface UseRouteRedirectOptions {
  watchMemberId?: string
  memberDisplayName?: string
  onMemberIdChange?: (ctx: RouteRedirectContext & { displayName: string }) => void
  onAccessDenied?: (ctx: RouteRedirectContext) => void
}

export function useRouteRedirect(options: UseRouteRedirectOptions = {}) {
  const navigate = useNavigate()
  const { watchMemberId, memberDisplayName, onMemberIdChange, onAccessDenied } = options
  const { permissions, loading, canAccess, pathname } = useRouteAccess()
  const isFirstRender = useRef(true)
  const previousMemberId = useRef(watchMemberId)

  useEffect(() => {
    if (loading) return

    const ctx: RouteRedirectContext = { permissions, pathname }

    const redirectToHome = (replace: boolean) => {
      const home = getDefaultHomePath(permissions)
      navigate(home ?? '/', { replace })
    }

    if (isFirstRender.current) {
      isFirstRender.current = false
      if (watchMemberId !== undefined) {
        previousMemberId.current = watchMemberId
      }
      if (!canAccess) {
        redirectToHome(true)
      }
      return
    }

    if (watchMemberId !== undefined && previousMemberId.current !== watchMemberId) {
      previousMemberId.current = watchMemberId
      onMemberIdChange?.({ ...ctx, displayName: memberDisplayName ?? '' })
      return
    }

    if (!canAccess) {
      if (onAccessDenied) {
        onAccessDenied(ctx)
      } else {
        redirectToHome(true)
      }
    }
  }, [
    canAccess,
    loading,
    navigate,
    onAccessDenied,
    onMemberIdChange,
    pathname,
    permissions,
    watchMemberId,
    memberDisplayName,
  ])
}
