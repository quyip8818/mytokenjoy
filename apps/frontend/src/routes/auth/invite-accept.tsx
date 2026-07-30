import { useCallback, useEffect, useState } from 'react'
import { useNavigate, useSearch } from '@tanstack/react-router'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { PasswordInput } from '@/components/ui/password-input'
import { Label } from '@/components/ui/label'
import { authApi } from '@/api/auth'
import { ApiError } from '@/api/client'
import { ROUTES } from '@/config/routes'
import { useSession } from '@/features/session'

export default function InviteAcceptPage() {
  const navigate = useNavigate()
  const { code: inviteCode = '' } = useSearch({ strict: false })
  const { refreshSession } = useSession()

  const [alias, setAlias] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [loading, setLoading] = useState(!!inviteCode)

  // Clear any existing session on mount (same as login page).
  // Then fetch pre-fill data (member alias).
  useEffect(() => {
    if (!inviteCode) return
    authApi
      .logout()
      .catch(() => {})
      .then(() =>
        authApi
          .getInviteInfo(inviteCode)
          .then(({ alias: prefill }) => {
            if (prefill) setAlias(prefill)
          })
          .catch(() => {
            // Non-fatal: user can still fill manually.
          })
          .finally(() => setLoading(false)),
      )
  }, [inviteCode])

  const handleSubmit = useCallback(
    async (event: React.FormEvent) => {
      event.preventDefault()
      if (!inviteCode || !alias.trim() || !password.trim()) return

      if (password !== confirmPassword) {
        setError('两次输入的密码不一致')
        return
      }

      setSubmitting(true)
      setError(null)
      try {
        await authApi.acceptInvite(inviteCode, alias.trim(), password.trim())
        await refreshSession()
        navigate({ to: ROUTES.home, replace: true })
      } catch (err) {
        const message = err instanceof ApiError ? err.message : '激活失败'
        setError(message)
      } finally {
        setSubmitting(false)
      }
    },
    [inviteCode, alias, password, confirmPassword, navigate, refreshSession],
  )

  if (!inviteCode) {
    return (
      <div className="flex min-h-screen items-center justify-center p-8">
        <p className="text-muted-foreground">邀请链接无效</p>
      </div>
    )
  }

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center p-8">
        <p className="text-muted-foreground">加载中…</p>
      </div>
    )
  }

  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-4 p-8">
      <form onSubmit={handleSubmit} className="flex w-full max-w-md flex-col gap-4">
        <h1 className="text-center text-lg font-semibold">接受邀请</h1>
        <p className="text-center text-sm text-muted-foreground">设置您的昵称和密码以加入团队</p>

        <div className="space-y-2">
          <Label htmlFor="invite-alias">昵称</Label>
          <Input
            id="invite-alias"
            type="text"
            autoComplete="nickname"
            value={alias}
            onChange={(e) => setAlias(e.target.value)}
            required
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="invite-password">密码</Label>
          <PasswordInput
            id="invite-password"
            autoComplete="new-password"
            placeholder="至少8位"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            minLength={8}
            required
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="invite-confirm-password">确认密码</Label>
          <PasswordInput
            id="invite-confirm-password"
            autoComplete="new-password"
            placeholder="再次输入密码"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            minLength={8}
            required
          />
        </div>

        {error ? <p className="text-sm text-destructive">{error}</p> : null}

        <Button type="submit" className="w-full" disabled={submitting}>
          {submitting ? '激活中…' : '加入'}
        </Button>
      </form>
    </div>
  )
}
