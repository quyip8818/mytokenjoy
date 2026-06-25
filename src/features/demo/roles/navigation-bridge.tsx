import { useEffect, useRef } from 'react'
import { useLocation, useNavigate } from 'react-router'
import { toast } from 'sonner'
import { canAccessRoute, getDefaultHomePath } from '@/lib/permissions'
import { useDemoRole } from './use-demo-role'

export function DemoRoleNavigationBridge() {
  const navigate = useNavigate()
  const location = useLocation()
  const { memberId, displayName, permissions, loading } = useDemoRole()
  const isFirstRender = useRef(true)
  const previousMemberId = useRef(memberId)

  useEffect(() => {
    if (loading) return

    if (isFirstRender.current) {
      isFirstRender.current = false
      previousMemberId.current = memberId
      if (!canAccessRoute(location.pathname, permissions)) {
        navigate(getDefaultHomePath(permissions), { replace: true })
      }
      return
    }

    if (previousMemberId.current !== memberId) {
      previousMemberId.current = memberId
      const home = getDefaultHomePath(permissions)
      navigate(home)
      toast.info(`已切换为${displayName}视角`)
      return
    }

    if (!canAccessRoute(location.pathname, permissions)) {
      navigate(getDefaultHomePath(permissions), { replace: true })
      toast.info('当前身份无权访问该页面')
    }
  }, [memberId, displayName, permissions, loading, location.pathname, navigate])

  return null
}
