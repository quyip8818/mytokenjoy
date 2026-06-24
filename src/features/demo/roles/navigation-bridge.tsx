import { useEffect, useRef } from 'react'
import { useNavigate } from 'react-router'
import { toast } from 'sonner'
import { DEMO_ROLE_PROFILES, getDefaultHomePath } from './constants'
import { useDemoRole } from './use-demo-role'

export function DemoRoleNavigationBridge() {
  const navigate = useNavigate()
  const { role } = useDemoRole()
  const isFirstRender = useRef(true)
  const previousRole = useRef(role)

  useEffect(() => {
    if (isFirstRender.current) {
      isFirstRender.current = false
      previousRole.current = role
      return
    }
    if (previousRole.current === role) return

    previousRole.current = role
    const profile = DEMO_ROLE_PROFILES[role]
    navigate(getDefaultHomePath(role))
    toast.info(`已切换为${profile.label}视角`)
  }, [role, navigate])

  return null
}
