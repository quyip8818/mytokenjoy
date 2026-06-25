import { useEffect, useRef } from 'react'
import { useNavigate } from 'react-router'
import { toast } from 'sonner'
import { getDefaultHomePath } from '@/lib/permissions'
import { useRouteAccess } from '@/hooks/use-route-access'
import { useDemoRole } from './use-demo-role'

export function DemoRoleNavigationBridge() {
  const navigate = useNavigate()
  const { memberId, displayName } = useDemoRole()
  const { permissions, loading, canAccess, pathname } = useRouteAccess()
  const isFirstRender = useRef(true)
  const previousMemberId = useRef(memberId)

  useEffect(() => {
    if (loading) return

    if (isFirstRender.current) {
      isFirstRender.current = false
      previousMemberId.current = memberId
      if (!canAccess) {
        const home = getDefaultHomePath(permissions)
        navigate(home ?? '/', { replace: true })
      }
      return
    }

    if (previousMemberId.current !== memberId) {
      previousMemberId.current = memberId
      const home = getDefaultHomePath(permissions)
      navigate(home ?? '/')
      toast.info(`已切换为${displayName}视角`)
      return
    }

    if (!canAccess) {
      const home = getDefaultHomePath(permissions)
      navigate(home ?? '/', { replace: true })
      toast.info('当前身份无权访问该页面')
    }
  }, [memberId, displayName, permissions, loading, canAccess, pathname, navigate])

  return null
}
