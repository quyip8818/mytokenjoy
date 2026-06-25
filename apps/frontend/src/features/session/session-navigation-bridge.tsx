import { useEffect, useRef } from 'react'
import { useNavigate } from 'react-router'
import { getDefaultHomePath } from '@/lib/permissions'
import { useRouteAccess } from '@/hooks/use-route-access'

export function SessionNavigationBridge() {
  const navigate = useNavigate()
  const { pathname, permissions, loading, canAccess } = useRouteAccess()
  const isFirstRender = useRef(true)

  useEffect(() => {
    if (loading) return

    if (isFirstRender.current) {
      isFirstRender.current = false
      if (!canAccess) {
        const home = getDefaultHomePath(permissions)
        navigate(home ?? '/', { replace: true })
      }
      return
    }

    if (!canAccess) {
      const home = getDefaultHomePath(permissions)
      navigate(home ?? '/', { replace: true })
    }
  }, [canAccess, loading, navigate, pathname, permissions])

  return null
}
