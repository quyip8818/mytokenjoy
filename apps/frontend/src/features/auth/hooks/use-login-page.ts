import { useState } from 'react'
import { useNavigate } from 'react-router'
import { ROUTES } from '@/config/routes'
import { authApi } from '@/api/auth'
import { ApiError } from '@/api/client'
import { useSession } from '@/features/session'

const DEMO_EMAIL = 'admin@example.com'
const DEMO_PASSWORD = 'demo1234'

export function useLoginPage() {
  const navigate = useNavigate()
  const { refreshSession } = useSession()
  const [email, setEmail] = useState(import.meta.env.DEV ? DEMO_EMAIL : '')
  const [password, setPassword] = useState(import.meta.env.DEV ? DEMO_PASSWORD : '')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  async function handleSubmit(event: React.FormEvent) {
    event.preventDefault()
    setSubmitting(true)
    setError(null)
    try {
      await authApi.login({ email, password })
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
    error,
    submitting,
    handleSubmit,
    demoEmail: DEMO_EMAIL,
    demoPassword: DEMO_PASSWORD,
  }
}
