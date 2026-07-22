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

interface ChangePasswordDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  hasPassword: boolean
  error: string | null
  saving: boolean
  onSubmit: (oldPassword: string | undefined, newPassword: string) => Promise<boolean>
}

export function ChangePasswordDialog({
  open,
  onOpenChange,
  hasPassword,
  error,
  saving,
  onSubmit,
}: ChangePasswordDialogProps) {
  const [oldPassword, setOldPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirm, setConfirm] = useState('')
  const [localError, setLocalError] = useState<string | null>(null)

  const handleOpenChange = useCallback(
    (nextOpen: boolean) => {
      onOpenChange(nextOpen)
      if (!nextOpen) {
        setOldPassword('')
        setNewPassword('')
        setConfirm('')
        setLocalError(null)
      }
    },
    [onOpenChange],
  )

  const handleSubmit = useCallback(
    async (e: React.FormEvent) => {
      e.preventDefault()
      if (newPassword.length < 8) {
        setLocalError('密码至少 8 位')
        return
      }
      if (newPassword !== confirm) {
        setLocalError('两次输入不一致')
        return
      }
      setLocalError(null)
      await onSubmit(hasPassword ? oldPassword : undefined, newPassword)
    },
    [oldPassword, newPassword, confirm, hasPassword, onSubmit],
  )

  const displayError = localError || error

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{hasPassword ? '修改密码' : '设置密码'}</DialogTitle>
          <DialogDescription>
            {hasPassword ? '请输入当前密码和新密码' : '设置后可用密码方式登录'}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          {hasPassword && (
            <div className="space-y-2">
              <Label htmlFor="old-password">当前密码</Label>
              <Input
                id="old-password"
                type="password"
                autoComplete="current-password"
                value={oldPassword}
                onChange={(e) => setOldPassword(e.target.value)}
                required
              />
            </div>
          )}
          <div className="space-y-2">
            <Label htmlFor="new-password">新密码</Label>
            <Input
              id="new-password"
              type="password"
              autoComplete="new-password"
              placeholder="至少 8 位"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              required
              minLength={8}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="confirm-password">确认新密码</Label>
            <Input
              id="confirm-password"
              type="password"
              autoComplete="new-password"
              value={confirm}
              onChange={(e) => setConfirm(e.target.value)}
              required
              minLength={8}
            />
          </div>
          {displayError && <p className="text-sm text-destructive">{displayError}</p>}
          <Button type="submit" disabled={saving || newPassword.length < 8}>
            {saving ? '保存中…' : '保存'}
          </Button>
        </form>
      </DialogContent>
    </Dialog>
  )
}
