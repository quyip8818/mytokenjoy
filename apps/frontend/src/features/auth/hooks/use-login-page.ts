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
  const [companySlug, setCompanySlug] = useState(
    () =>
      searchParams.get('company')?.trim() ||
      import.meta.env.VITE_DEFAULT_COMPANY_SLUG?.trim() ||
      '',
  )
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  async function handleSubmit(event: React.FormEvent) {
    event.preventDefault()
    setSubmitting(true)
    setError(null)
    try {
      const slug = companySlug.trim()
      await authApi.login({
        email,
        password,
        ...(slug ? { companySlug: slug } : {}),
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
    companySlug,
    setCompanySlug,
    supportSaas,
    error,
    submitting,
    handleSubmit,
    demoEmail: DEMO_EMAIL,
    demoPassword: DEMO_PASSWORD,
  }
}
