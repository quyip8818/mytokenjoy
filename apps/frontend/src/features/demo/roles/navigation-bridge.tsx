import { useCallback } from 'react'
import { useNavigate } from 'react-router'
import { toast } from 'sonner'
import { getDefaultHomePath } from '@/lib/permissions'
import { useRouteRedirect } from '@/hooks/use-route-redirect'
import { useDemoRole } from './use-demo-role'

export function DemoRoleNavigationBridge() {
  const navigate = useNavigate()
  const { memberId, displayName } = useDemoRole()

  const handleMemberIdChange = useCallback(
    ({
      permissions,
      displayName: name,
    }: {
      permissions: readonly string[]
      displayName: string
    }) => {
      const home = getDefaultHomePath(permissions)
      navigate(home ?? '/')
      toast.info(`已切换为${name}视角`)
    },
    [navigate],
  )

  const handleAccessDenied = useCallback(
    ({ permissions }: { permissions: readonly string[] }) => {
      const home = getDefaultHomePath(permissions)
      navigate(home ?? '/', { replace: true })
      toast.info('当前身份无权访问该页面')
    },
    [navigate],
  )

  useRouteRedirect({
    watchMemberId: memberId,
    memberDisplayName: displayName,
    onMemberIdChange: handleMemberIdChange,
    onAccessDenied: handleAccessDenied,
  })

  return null
}
