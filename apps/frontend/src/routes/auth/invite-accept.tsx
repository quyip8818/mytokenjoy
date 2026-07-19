import { useCallback, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { authApi } from '@/api/auth'
import { ApiError } from '@/api/client'
import { ROUTES } from '@/config/routes'
import { useSession } from '@/features/session'

export default function InviteAcceptPage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const { refreshSession } = useSession()
  const inviteCode = searchParams.get('code') ?? ''

  const [name, setName] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  const handleSubmit = useCallback(
    async (event: React.FormEvent) => {
      event.preventDefault()
      if (!inviteCode || !name.trim() || !password.trim()) return
      setSubmitting(true)
      setError(null)
      try {
        await authApi.acceptInvite(inviteCode, name.trim(), password.trim())
        await refreshSession()
        navigate(ROUTES.home, { replace: true })
      } catch (err) {
        const message = err instanceof ApiError ? err.message : '激活失败'
        setError(message)
      } finally {
        setSubmitting(false)
      }
    },
    [inviteCode, name, password, navigate, refreshSession],
  )

  if (!inviteCode) {
    return (
      <div className="flex min-h-screen items-center justify-center p-8">
        <p className="text-muted-foreground">邀请链接无效</p>
      </div>
    )
  }

  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-4 p-8">
      <form onSubmit={handleSubmit} className="flex w-full max-w-md flex-col gap-4">
        <h1 className="text-center text-lg font-semibold">接受邀请</h1>
        <p className="text-center text-sm text-muted-foreground">设置您的姓名和密码以加入团队</p>

        <div className="space-y-2">
          <Label htmlFor="invite-name">姓名</Label>
          <Input
            id="invite-name"
            type="text"
            autoComplete="name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="invite-password">密码</Label>
          <Input
            id="invite-password"
            type="password"
            autoComplete="new-password"
            placeholder="至少8位"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            minLength={8}
            required
          />
        </div>

        {error ? <p className="text-sm text-destructive">{error}</p> : null}

        <Button type="submit" disabled={submitting}>
          {submitting ? '激活中…' : '加入'}
        </Button>
      </form>
    </div>
  )
}
