import { useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router'
import { ROUTES } from '@/config/routes'
import { authApi } from '@/api/auth'
import { ApiError } from '@/api/client'
import { useSession } from '@/features/session'

const DEMO_EMAIL = 'admin@example.com'
const DEMO_PASSWORD = 'demo1234'
const supportSaas = import.meta.env.VITE_SUPPORT_SAAS === 'true'

export function useLoginPage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const { refreshSession } = useSession()
  const [email, setEmail] = useState(import.meta.env.DEV ? DEMO_EMAIL : '')
  const [password, setPassword] = useState(import.meta.env.DEV ? DEMO_PASSWORD : '')
  const [companyId, setCompanyId] = useState(
    () =>
      searchParams.get('companyid')?.trim() ||
      import.meta.env.VITE_DEFAULT_COMPANY_ID?.trim() ||
      '',
  )
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  async function handleSubmit(event: React.FormEvent) {
    event.preventDefault()
    setSubmitting(true)
    setError(null)
    try {
      const id = companyId.trim()
      await authApi.login({
        email,
        password,
        ...(id ? { companyId: id } : {}),
      })
      await refreshSession()
      navigate(ROUTES.home, { replace: true })
    } catch (err) {
      const message = err instanceof ApiError ? err.message : 'Login failed'
      setError(message)
    } finally {
      setSubmitting(false)
    }
  }

  return {
    email,
    setEmail,
    password,
    setPassword,
    companyId,
    setCompanyId,
    supportSaas,
    error,
    submitting,
    handleSubmit,
    demoEmail: DEMO_EMAIL,
    demoPassword: DEMO_PASSWORD,
  }
}
