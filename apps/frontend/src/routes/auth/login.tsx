import { useState } from 'react'
import { LoginForm, SmsLoginForm, useLoginPage, useSmsLoginPage } from '@/features/auth'

const supportSaas = import.meta.env.VITE_SUPPORT_SAAS === 'true'

export default function LoginPage() {
  const [mode, setMode] = useState<'sms' | 'password'>(supportSaas ? 'sms' : 'password')

  if (mode === 'sms') {
    return <SmsLoginMode onSwitchToPassword={() => setMode('password')} />
  }

  return <PasswordLoginMode onSwitchToSms={supportSaas ? () => setMode('sms') : undefined} />
}

function SmsLoginMode({ onSwitchToPassword }: { onSwitchToPassword: () => void }) {
  const props = useSmsLoginPage()
  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-4 p-8">
      <SmsLoginForm {...props} />
      <p className="text-sm text-muted-foreground">
        还没有账号？
        <a href="/register" className="text-primary underline-offset-4 hover:underline">
          立即注册
        </a>
      </p>
      <Divider />
      <button
        type="button"
        onClick={onSwitchToPassword}
        className="text-sm text-muted-foreground hover:text-foreground"
      >
        邮箱密码登录 →
      </button>
    </div>
  )
}

function PasswordLoginMode({ onSwitchToSms }: { onSwitchToSms?: () => void }) {
  const props = useLoginPage()
  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-4 p-8">
      <LoginForm {...props} />
      {onSwitchToSms ? (
        <>
          <Divider />
          <button
            type="button"
            onClick={onSwitchToSms}
            className="text-sm text-muted-foreground hover:text-foreground"
          >
            ← 短信验证码登录
          </button>
        </>
      ) : null}
    </div>
  )
}

function Divider() {
  return (
    <div className="flex w-full max-w-md items-center gap-3 text-sm text-muted-foreground">
      <div className="h-px flex-1 bg-border" />
      <span>或</span>
      <div className="h-px flex-1 bg-border" />
    </div>
  )
}
