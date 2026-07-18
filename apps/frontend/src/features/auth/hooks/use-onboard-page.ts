import { useCallback, useState } from 'react'
import { useLocation, useNavigate } from 'react-router'
import { authApi, type PendingInvite } from '@/api/auth'
import { ApiError } from '@/api/client'
import { ROUTES } from '@/config/routes'
import { useSession } from '@/features/session'

export function useOnboardPage() {
  const navigate = useNavigate()
  const location = useLocation()
  const { refreshSession } = useSession()
  const [error, setError] = useState<string | null>(null)
  const [accepting, setAccepting] = useState<string | null>(null)
  const [creating, setCreating] = useState(false)
  const [companyName, setCompanyName] = useState('')
  const [memberName, setMemberName] = useState('')

  const state = location.state as { invites?: PendingInvite[]; mode?: 'choose' | 'onboard' } | null
  const invites: PendingInvite[] = state?.invites ?? []
  const mode: 'choose' | 'onboard' = state?.mode ?? (invites.length > 0 ? 'choose' : 'onboard')

  const handleAcceptInvite = useCallback(
    async (inviteCode: string) => {
      const name = memberName.trim() || '新成员'
      setAccepting(inviteCode)
      setError(null)
      try {
        await authApi.registerAccept(inviteCode, name)
        await refreshSession()
        navigate(ROUTES.home, { replace: true })
      } catch (err) {
        const message = err instanceof ApiError ? err.message : '接受邀请失败'
        setError(message)
      } finally {
        setAccepting(null)
      }
    },
    [memberName, navigate, refreshSession],
  )

  const handleCreateCompany = useCallback(
    async (event: React.FormEvent) => {
      event.preventDefault()
      if (!companyName.trim()) return
      setCreating(true)
      setError(null)
      try {
        await authApi.registerCompany(companyName.trim())
        await refreshSession()
        navigate(ROUTES.home, { replace: true })
      } catch (err) {
        const message = err instanceof ApiError ? err.message : '创建失败'
        setError(message)
      } finally {
        setCreating(false)
      }
    },
    [companyName, navigate, refreshSession],
  )

  return {
    mode,
    invites,
    error,
    accepting,
    creating,
    companyName,
    setCompanyName,
    memberName,
    setMemberName,
    handleAcceptInvite,
    handleCreateCompany,
  }
}
