import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import type { useSmsLoginPage } from '../hooks/use-sms-login-page'

type SmsLoginFormProps = ReturnType<typeof useSmsLoginPage>

export function SmsLoginForm({
  phone,
  setPhone,
  code,
  setCode,
  error,
  sending,
  verifying,
  countdown,
  canSend,
  handleSendCode,
  handleVerify,
}: SmsLoginFormProps) {
  return (
    <form onSubmit={handleVerify} className="flex w-full max-w-md flex-col gap-4">
      <div className="space-y-2 text-center">
        <h1 className="text-lg font-semibold">TokenJoy</h1>
      </div>

      <div className="space-y-2">
        <Label htmlFor="phone">手机号</Label>
        <div className="flex gap-2">
          <span className="flex items-center rounded-md border bg-muted px-3 text-sm text-muted-foreground">
            +86
          </span>
          <Input
            id="phone"
            type="tel"
            inputMode="numeric"
            autoComplete="tel"
            placeholder="请输入手机号"
            value={phone}
            onChange={(e) => setPhone(e.target.value)}
            required
          />
        </div>
      </div>

      <div className="space-y-2">
        <Label htmlFor="sms-code">验证码</Label>
        <div className="flex gap-2">
          <Input
            id="sms-code"
            type="text"
            inputMode="numeric"
            autoComplete="one-time-code"
            placeholder="6位验证码"
            maxLength={6}
            value={code}
            onChange={(e) => setCode(e.target.value.replace(/\D/g, ''))}
            required
          />
          <Button
            type="button"
            variant="outline"
            disabled={!canSend}
            onClick={handleSendCode}
            className="shrink-0 whitespace-nowrap"
          >
            {sending ? '发送中…' : countdown > 0 ? `${countdown}s` : '获取验证码'}
          </Button>
        </div>
      </div>

      {error ? <p className="text-sm text-destructive">{error}</p> : null}

      <Button type="submit" disabled={verifying || !code.trim()}>
        {verifying ? '验证中…' : '进入'}
      </Button>
    </form>
  )
}
