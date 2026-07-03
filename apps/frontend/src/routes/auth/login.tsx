import { useState } from 'react'
import { useNavigate } from 'react-router'
import { ROUTES } from '@/config/routes'
import { authApi } from '@/api/auth'
import { ApiError } from '@/api/client'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

const DEMO_EMAIL = 'admin@example.com'
const DEMO_PASSWORD = 'demo1234'

export default function LoginPage() {
  const navigate = useNavigate()
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
      navigate(ROUTES.home, { replace: true })
    } catch (err) {
      const message = err instanceof ApiError ? err.message : 'Login failed'
      setError(message)
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-4 p-8">
      <form onSubmit={handleSubmit} className="flex w-full max-w-md flex-col gap-4">
        <div className="space-y-2 text-center">
          <h1 className="text-lg font-semibold">Sign in</h1>
          {import.meta.env.DEV ? (
            <p className="text-sm text-muted-foreground">
              Dev seed: {DEMO_EMAIL} / {DEMO_PASSWORD}
            </p>
          ) : null}
        </div>
        <div className="space-y-2">
          <Label htmlFor="email">Email</Label>
          <Input
            id="email"
            type="email"
            autoComplete="username"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="password">Password</Label>
          <Input
            id="password"
            type="password"
            autoComplete="current-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
        </div>
        {error ? <p className="text-sm text-destructive">{error}</p> : null}
        <Button type="submit" disabled={submitting}>
          {submitting ? 'Signing in…' : 'Sign in'}
        </Button>
      </form>
    </div>
  )
}
