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
import { useVerifyCountdown } from '@/features/auth'

interface ChangeContactDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  type: 'phone' | 'email'
  error: string | null
  saving: boolean
  onSubmit: (value: string, code: string) => Promise<boolean>
}

export function ChangeContactDialog({
  open,
  onOpenChange,
  type,
  error,
  saving,
  onSubmit,
}: ChangeContactDialogProps) {
  const [value, setValue] = useState('')
  const [code, setCode] = useState('')
  const [codeSent, setCodeSent] = useState(false)
  const { countdown, sending, sendError, sendCode } = useVerifyCountdown()

  const isPhone = type === 'phone'
  const label = isPhone ? '新手机号' : '新邮箱'
  const placeholder = isPhone ? '请输入手机号' : '请输入邮箱地址'

  const handleOpenChange = useCallback(
    (nextOpen: boolean) => {
      onOpenChange(nextOpen)
      if (!nextOpen) {
        setValue('')
        setCode('')
        setCodeSent(false)
      }
    },
    [onOpenChange],
  )

  const handleSendCode = useCallback(async () => {
    if (!value.trim()) return
    const params = isPhone
      ? { phone: value.trim(), purpose: 'bind' }
      : { email: value.trim(), purpose: 'bind' }
    await sendCode(params)
    setCodeSent(true)
  }, [value, isPhone, sendCode])

  const handleSubmit = useCallback(
    async (e: React.FormEvent) => {
      e.preventDefault()
      if (!value.trim() || !code.trim()) return
      await onSubmit(value.trim(), code.trim())
    },
    [value, code, onSubmit],
  )

  const displayError = sendError || error

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{isPhone ? '更换手机号' : '更换邮箱'}</DialogTitle>
          <DialogDescription>更换后将使用新的{isPhone ? '手机号' : '邮箱'}登录</DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <div className="space-y-2">
            <Label htmlFor="contact-value">{label}</Label>
            <div className="flex gap-2">
              <Input
                id="contact-value"
                type={isPhone ? 'tel' : 'email'}
                placeholder={placeholder}
                value={value}
                onChange={(e) => setValue(e.target.value)}
                required
                className="flex-1"
              />
              <Button
                type="button"
                variant="outline"
                size="sm"
                disabled={!value.trim() || sending || countdown > 0}
                onClick={handleSendCode}
              >
                {countdown > 0 ? `${countdown}s` : '发送验证码'}
              </Button>
            </div>
          </div>
          {codeSent && (
            <div className="space-y-2">
              <Label htmlFor="verify-code">验证码</Label>
              <Input
                id="verify-code"
                type="text"
                inputMode="numeric"
                placeholder="请输入 6 位验证码"
                maxLength={6}
                value={code}
                onChange={(e) => setCode(e.target.value)}
                required
              />
            </div>
          )}
          {displayError && <p className="text-sm text-destructive">{displayError}</p>}
          <Button type="submit" disabled={saving || !codeSent || code.length < 6}>
            {saving ? '保存中…' : '确认更换'}
          </Button>
        </form>
      </DialogContent>
    </Dialog>
  )
}
