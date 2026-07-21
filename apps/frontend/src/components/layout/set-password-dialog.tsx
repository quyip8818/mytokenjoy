import { useCallback, useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { authApi } from '@/api/auth'
import { ApiError } from '@/api/client'

interface SetPasswordDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function SetPasswordDialog({ open, onOpenChange }: SetPasswordDialogProps) {
  const [password, setPassword] = useState('')
  const [confirm, setConfirm] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)
  const [success, setSuccess] = useState(false)

  const handleOpenChange = useCallback(
    (nextOpen: boolean) => {
      onOpenChange(nextOpen)
      if (!nextOpen) {
        setPassword('')
        setConfirm('')
        setError(null)
        setSuccess(false)
      }
    },
    [onOpenChange],
  )

  const handleSubmit = useCallback(
    async (event: React.FormEvent) => {
      event.preventDefault()
      if (password.length < 8) {
        setError('密码至少 8 位')
        return
      }
      if (password !== confirm) {
        setError('两次输入不一致')
        return
      }
      setSaving(true)
      setError(null)
      try {
        await authApi.setPassword(password)
        setSuccess(true)
        setTimeout(() => handleOpenChange(false), 1500)
      } catch (err) {
        setError(err instanceof ApiError ? err.message : '设置失败')
      } finally {
        setSaving(false)
      }
    },
    [password, confirm, handleOpenChange],
  )

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>设置密码</DialogTitle>
          <DialogDescription>设置后可用邮箱 + 密码方式登录</DialogDescription>
        </DialogHeader>
        {success ? (
          <p className="py-4 text-center text-sm text-green-600">密码设置成功</p>
        ) : (
          <form onSubmit={handleSubmit} className="flex flex-col gap-4">
            <div className="space-y-2">
              <Label htmlFor="new-password">新密码</Label>
              <Input
                id="new-password"
                type="password"
                autoComplete="new-password"
                placeholder="至少 8 位"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                minLength={8}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="confirm-password">确认密码</Label>
              <Input
                id="confirm-password"
                type="password"
                autoComplete="new-password"
                placeholder="再次输入"
                value={confirm}
                onChange={(e) => setConfirm(e.target.value)}
                required
                minLength={8}
              />
            </div>
            {error ? <p className="text-sm text-destructive">{error}</p> : null}
            <Button type="submit" disabled={saving || password.length < 8}>
              {saving ? '保存中…' : '保存'}
            </Button>
          </form>
        )}
      </DialogContent>
    </Dialog>
  )
}
