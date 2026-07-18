import { useCallback, useState } from 'react'
import { useLocation, useNavigate } from 'react-router'
import { authApi, type CompanyOption } from '@/api/auth'
import { ApiError } from '@/api/client'
import { ROUTES } from '@/config/routes'
import { useSession } from '@/features/session'

export function useSelectPage() {
  const navigate = useNavigate()
  const location = useLocation()
  const { refreshSession } = useSession()
  const [selecting, setSelecting] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)

  const companies: CompanyOption[] = (location.state as { companies?: CompanyOption[] })?.companies ?? []

  const handleSelect = useCallback(
    async (companyId: string) => {
      setSelecting(companyId)
      setError(null)
      try {
        await authApi.smsSelect(companyId)
        await refreshSession()
        navigate(ROUTES.home, { replace: true })
      } catch (err) {
        const message = err instanceof ApiError ? err.message : '选择失败'
        setError(message)
      } finally {
        setSelecting(null)
      }
    },
    [navigate, refreshSession],
  )

  return {
    companies,
    selecting,
    error,
    handleSelect,
  }
}
