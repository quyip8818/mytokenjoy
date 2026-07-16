import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import type { useLoginPage } from '@/features/auth'

type LoginFormProps = ReturnType<typeof useLoginPage>

export function LoginForm({
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
  demoEmail,
  demoPassword,
}: LoginFormProps) {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-4 p-8">
      <form onSubmit={handleSubmit} className="flex w-full max-w-md flex-col gap-4">
        <div className="space-y-2 text-center">
          <h1 className="text-lg font-semibold">Sign in</h1>
          {import.meta.env.DEV ? (
            <p className="text-sm text-muted-foreground">
              Dev seed: {demoEmail} / {demoPassword}
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
        {supportSaas ? (
          <div className="space-y-2">
            <Label htmlFor="company-id">Company ID</Label>
            <Input
              id="company-id"
              type="text"
              inputMode="numeric"
              autoComplete="organization"
              value={companyId}
              onChange={(e) => setCompanyId(e.target.value)}
              placeholder="123"
              required
            />
          </div>
        ) : null}
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
