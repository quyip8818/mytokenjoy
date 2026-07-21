import { useCallback, useState } from 'react'
import {
  Dialog,
  DialogContent,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { cn } from '@/lib/utils'
import { authApi, type CompanyOption, type PendingInvite, type SmsVerifyResult } from '@/api/auth'
import { ApiError } from '@/api/client'
import { useSmsCountdown } from '../hooks/use-sms-countdown'

type AuthMode = 'login' | 'register'
type AuthStep =
  | 'login-phone-pw'       // 默认：手机号 + 密码
  | 'login-phone-code'     // 手机号 + 验证码
  | 'login-email-pw'       // 邮箱 + 密码
  | 'reset-password'       // 忘记密码：手机号 + 验证码 + 新密码
  | 'register-phone'       // 手机号 + 验证码 + 密码
  | 'register-info'        // 公司名
  | 'select-company'
  | 'select-invite'

interface AuthPopupProps {
  open: boolean
  defaultMode?: AuthMode
  closable?: boolean
  onSuccess?: () => void
  onClose?: () => void
}

export function AuthPopup({
  open,
  defaultMode = 'login',
  closable = false,
  onSuccess,
  onClose,
}: AuthPopupProps) {
  const [mode, setMode] = useState<AuthMode>(defaultMode)
  const [step, setStep] = useState<AuthStep>(
    defaultMode === 'register' ? 'register-phone' : 'login-phone-pw',
  )

  // Form state
  const [phone, setPhone] = useState('')
  const [code, setCode] = useState('')
  const [password, setPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [email, setEmail] = useState('')
  const [companyName, setCompanyName] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  // Multi-select state
  const [companies, setCompanies] = useState<CompanyOption[]>([])
  const [invites, setInvites] = useState<PendingInvite[]>([])
  const [memberName, setMemberName] = useState('')

  const { sending, countdown, sendError, sendCode } = useSmsCountdown()
  const handleSendCode = useCallback(() => sendCode(phone), [sendCode, phone])
  const canSend = phone.trim().length >= 11 && countdown === 0 && !sending

  const isLoginStep = step === 'login-phone-pw' || step === 'login-phone-code' || step === 'login-email-pw'
  const showTabs = isLoginStep || step === 'register-phone'

  // --- Tab switch ---
  const switchTab = useCallback((newMode: AuthMode) => {
    setMode(newMode)
    setStep(newMode === 'login' ? 'login-phone-pw' : 'register-phone')
    setError(null)
  }, [])

  // --- Login: phone + password ---
  const handleLoginPhonePw = useCallback(async (e: React.FormEvent) => {
    e.preventDefault()
    if (!phone.trim() || !password) return
    setSubmitting(true)
    setError(null)
    try {
      const result = await authApi.login({ email: phone.trim(), password })
      if ('action' in result && result.action === 'select_company') {
        setCompanies(result.companies)
        setStep('select-company')
      } else {
        onSuccess?.()
      }
    } catch (err) {
      setError(err instanceof ApiError ? err.message : '登录失败')
    } finally {
      setSubmitting(false)
    }
  }, [phone, password, onSuccess])

  // --- Login: phone + SMS code ---
  const handleLoginPhoneCode = useCallback(async (e: React.FormEvent) => {
    e.preventDefault()
    if (!phone.trim() || !code.trim()) return
    setSubmitting(true)
    setError(null)
    try {
      const result: SmsVerifyResult = await authApi.smsVerify(phone.trim(), code.trim())
      switch (result.action) {
        case 'enter':
          onSuccess?.()
          break
        case 'select_company':
          setCompanies(result.companies)
          setStep('select-company')
          break
        case 'choose':
          setInvites(result.invites)
          setStep('select-invite')
          break
        case 'not_found':
          setError('该手机号尚未注册')
          break
      }
    } catch (err) {
      setError(err instanceof ApiError ? err.message : '验证失败')
    } finally {
      setSubmitting(false)
    }
  }, [phone, code, onSuccess])

  // --- Login: email + password ---
  const handleLoginEmailPw = useCallback(async (e: React.FormEvent) => {
    e.preventDefault()
    if (!email.trim() || !password) return
    setSubmitting(true)
    setError(null)
    try {
      const result = await authApi.login({ email: email.trim(), password })
      if ('action' in result && result.action === 'select_company') {
        setCompanies(result.companies)
        setStep('select-company')
      } else {
        onSuccess?.()
      }
    } catch (err) {
      setError(err instanceof ApiError ? err.message : '登录失败')
    } finally {
      setSubmitting(false)
    }
  }, [email, password, onSuccess])

  // --- Login: select company ---
  const handleSelectCompany = useCallback(async (companyId: string) => {
    setSubmitting(true)
    setError(null)
    try {
      await authApi.smsSelect(companyId)
      onSuccess?.()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : '选择失败')
    } finally {
      setSubmitting(false)
    }
  }, [onSuccess])

  // --- Login: accept invite ---
  const handleAcceptInvite = useCallback(async (inviteCode: string) => {
    const name = memberName.trim() || '新成员'
    setSubmitting(true)
    setError(null)
    try {
      await authApi.registerAccept(inviteCode, name)
      onSuccess?.()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : '接受失败')
    } finally {
      setSubmitting(false)
    }
  }, [memberName, onSuccess])

  // --- Register: verify phone + set password → create user ---
  const handleRegisterVerify = useCallback(async (e: React.FormEvent) => {
    e.preventDefault()
    if (!phone.trim() || !code.trim() || password.length < 8) return
    setSubmitting(true)
    setError(null)
    try {
      const result = await authApi.registerInit(phone.trim(), code.trim())
      if (result.action === 'login') {
        setError('该手机号已注册，请切换到登录')
        return
      }
      // User created + register session set. Move to company name step.
      setStep('register-info')
    } catch (err) {
      setError(err instanceof ApiError ? err.message : '验证失败')
    } finally {
      setSubmitting(false)
    }
  }, [phone, code, password])

  // --- Register: create company (password already collected in previous step) ---
  const handleCreateCompany = useCallback(async (e: React.FormEvent) => {
    e.preventDefault()
    if (!companyName.trim() || password.length < 8) return
    setSubmitting(true)
    setError(null)
    try {
      await authApi.registerCompany(companyName.trim(), password)
      onSuccess?.()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : '创建失败')
    } finally {
      setSubmitting(false)
    }
  }, [companyName, password, onSuccess])

  // --- Reset password ---
  const handleResetPassword = useCallback(async (e: React.FormEvent) => {
    e.preventDefault()
    if (!phone.trim() || !code.trim() || newPassword.length < 8) return
    setSubmitting(true)
    setError(null)
    try {
      await authApi.resetPassword(phone.trim(), code.trim(), newPassword)
      // Success → switch back to login with a success message.
      setStep('login-phone-pw')
      setPassword('')
      setError(null)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : '重置失败')
    } finally {
      setSubmitting(false)
    }
  }, [phone, code, newPassword])

  // --- Render ---
  const displayError = error || sendError

  return (
    <Dialog open={open} onOpenChange={closable ? (v) => { if (!v) onClose?.() } : undefined}>
      <DialogContent
        className="sm:max-w-[420px] gap-0 p-0 overflow-hidden border-border/50 shadow-[0_10px_50px_rgba(139,92,246,0.10)]"
        onPointerDownOutside={closable ? undefined : (e) => e.preventDefault()}
        onEscapeKeyDown={closable ? undefined : (e) => e.preventDefault()}
        showCloseButton={closable}
      >
        <DialogTitle className="sr-only">TokenJoy 认证</DialogTitle>

        {/* Header */}
        <div className="px-8 pt-8 pb-4 text-center">
          <h2 className="text-xl font-bold text-foreground">TokenJoy</h2>
          <p className="mt-1 text-sm text-muted-foreground">企业 AI 管理平台</p>
        </div>

        {/* Tabs */}
        {showTabs && (
          <div className="mx-8 flex border-b border-border">
            <button
              type="button"
              onClick={() => switchTab('login')}
              className={cn(
                'flex-1 pb-2.5 text-sm font-medium transition-colors',
                mode === 'login'
                  ? 'border-b-2 border-primary text-foreground'
                  : 'text-muted-foreground hover:text-foreground',
              )}
            >
              登录
            </button>
            <button
              type="button"
              onClick={() => switchTab('register')}
              className={cn(
                'flex-1 pb-2.5 text-sm font-medium transition-colors',
                mode === 'register'
                  ? 'border-b-2 border-primary text-foreground'
                  : 'text-muted-foreground hover:text-foreground',
              )}
            >
              注册
            </button>
          </div>
        )}

        {/* Content */}
        <div className="px-8 pb-8 pt-6">

          {/* === LOGIN: phone + password (default) === */}
          {step === 'login-phone-pw' && (
            <form onSubmit={handleLoginPhonePw} className="flex flex-col gap-4">
              <div className="space-y-1.5">
                <Label htmlFor="lp-phone">手机号</Label>
                <div className="flex gap-2">
                  <span className="flex items-center rounded-md border bg-muted px-3 text-sm text-muted-foreground">+86</span>
                  <Input id="lp-phone" type="tel" inputMode="numeric" autoComplete="tel" placeholder="请输入手机号" value={phone} onChange={(e) => setPhone(e.target.value)} required />
                </div>
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="lp-pw">密码</Label>
                <Input id="lp-pw" type="password" autoComplete="current-password" placeholder="输入密码" value={password} onChange={(e) => setPassword(e.target.value)} required />
              </div>
              {displayError && <p className="text-sm text-destructive">{displayError}</p>}
              <Button type="submit" disabled={submitting || !phone.trim() || !password}>
                {submitting ? '登录中…' : '登录'}
              </Button>
              <div className="flex justify-between text-xs text-muted-foreground">
                <button type="button" onClick={() => { setStep('login-phone-code'); setError(null) }} className="hover:text-foreground">验证码登录</button>
                <button type="button" onClick={() => { setStep('login-email-pw'); setError(null) }} className="hover:text-foreground">邮箱登录</button>
              </div>
              <div className="text-center">
                <button type="button" onClick={() => { setStep('reset-password'); setError(null); setCode('') }} className="text-xs text-muted-foreground hover:text-foreground">忘记密码？</button>
              </div>
            </form>
          )}

          {/* === LOGIN: phone + SMS code === */}
          {step === 'login-phone-code' && (
            <form onSubmit={handleLoginPhoneCode} className="flex flex-col gap-4">
              <PhoneCodeFields phone={phone} setPhone={setPhone} code={code} setCode={setCode} canSend={canSend} sending={sending} countdown={countdown} onSend={handleSendCode} />
              {displayError && <p className="text-sm text-destructive">{displayError}</p>}
              <Button type="submit" disabled={submitting || !code.trim()}>
                {submitting ? '验证中…' : '登录'}
              </Button>
              <div className="flex justify-between text-xs text-muted-foreground">
                <button type="button" onClick={() => { setStep('login-phone-pw'); setError(null) }} className="hover:text-foreground">← 密码登录</button>
                <button type="button" onClick={() => { setStep('login-email-pw'); setError(null) }} className="hover:text-foreground">邮箱登录</button>
              </div>
            </form>
          )}

          {/* === LOGIN: email + password === */}
          {step === 'login-email-pw' && (
            <form onSubmit={handleLoginEmailPw} className="flex flex-col gap-4">
              <div className="space-y-1.5">
                <Label htmlFor="le-email">邮箱</Label>
                <Input id="le-email" type="email" autoComplete="username" placeholder="name@company.com" value={email} onChange={(e) => setEmail(e.target.value)} required />
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="le-pw">密码</Label>
                <Input id="le-pw" type="password" autoComplete="current-password" placeholder="输入密码" value={password} onChange={(e) => setPassword(e.target.value)} required />
              </div>
              {displayError && <p className="text-sm text-destructive">{displayError}</p>}
              <Button type="submit" disabled={submitting || !email.trim() || !password}>
                {submitting ? '登录中…' : '登录'}
              </Button>
              <div className="flex justify-center text-xs text-muted-foreground">
                <button type="button" onClick={() => { setStep('login-phone-pw'); setError(null) }} className="hover:text-foreground">← 手机号登录</button>
              </div>
            </form>
          )}

          {/* === RESET PASSWORD === */}
          {step === 'reset-password' && (
            <form onSubmit={handleResetPassword} className="flex flex-col gap-4">
              <p className="text-sm text-muted-foreground">通过短信验证码重置密码</p>
              <PhoneCodeFields phone={phone} setPhone={setPhone} code={code} setCode={setCode} canSend={canSend} sending={sending} countdown={countdown} onSend={handleSendCode} />
              <div className="space-y-1.5">
                <Label htmlFor="rp-new-pw">新密码</Label>
                <Input id="rp-new-pw" type="password" autoComplete="new-password" placeholder="至少 8 位" value={newPassword} onChange={(e) => setNewPassword(e.target.value)} required minLength={8} />
              </div>
              {displayError && <p className="text-sm text-destructive">{displayError}</p>}
              <Button type="submit" disabled={submitting || !code.trim() || newPassword.length < 8}>
                {submitting ? '重置中…' : '重置密码'}
              </Button>
              <button type="button" onClick={() => { setStep('login-phone-pw'); setError(null) }} className="text-xs text-muted-foreground hover:text-foreground">← 返回登录</button>
            </form>
          )}

          {/* === REGISTER: phone + code + password (one page) === */}
          {step === 'register-phone' && (
            <form onSubmit={handleRegisterVerify} className="flex flex-col gap-4">
              <PhoneCodeFields phone={phone} setPhone={setPhone} code={code} setCode={setCode} canSend={canSend} sending={sending} countdown={countdown} onSend={handleSendCode} />
              <div className="space-y-1.5">
                <Label htmlFor="rp-pw">设置密码</Label>
                <Input id="rp-pw" type="password" autoComplete="new-password" placeholder="至少 8 位" value={password} onChange={(e) => setPassword(e.target.value)} required minLength={8} />
              </div>
              {displayError && <p className="text-sm text-destructive">{displayError}</p>}
              <Button type="submit" disabled={submitting || !code.trim() || password.length < 8}>
                {submitting ? '验证中…' : '下一步'}
              </Button>
            </form>
          )}

          {/* === REGISTER: company name === */}
          {step === 'register-info' && (
            <form onSubmit={handleCreateCompany} className="flex flex-col gap-4">
              <p className="text-sm text-muted-foreground">创建您的企业</p>
              <div className="space-y-1.5">
                <Label htmlFor="ri-company">公司名称</Label>
                <Input id="ri-company" type="text" placeholder="您的企业名称" value={companyName} onChange={(e) => setCompanyName(e.target.value)} required />
              </div>
              {displayError && <p className="text-sm text-destructive">{displayError}</p>}
              <Button type="submit" disabled={submitting || !companyName.trim()}>
                {submitting ? '创建中…' : '创建并开始体验'}
              </Button>
              <button type="button" onClick={() => setStep('register-phone')} className="text-xs text-muted-foreground hover:text-foreground">← 返回</button>
            </form>
          )}

          {/* === SELECT COMPANY === */}
          {step === 'select-company' && (
            <div className="flex flex-col gap-3">
              <p className="text-sm text-muted-foreground">选择企业</p>
              {companies.map((c) => (
                <button key={c.companyId} type="button" disabled={submitting} onClick={() => handleSelectCompany(c.companyId)} className="flex items-center justify-between rounded-lg border px-4 py-3 text-left transition-colors hover:bg-muted">
                  <div>
                    <div className="font-medium text-sm">{c.companyName}</div>
                    <div className="text-xs text-muted-foreground">{c.role}</div>
                  </div>
                </button>
              ))}
              {displayError && <p className="text-sm text-destructive">{displayError}</p>}
            </div>
          )}

          {/* === SELECT INVITE === */}
          {step === 'select-invite' && (
            <div className="flex flex-col gap-3">
              <p className="text-sm text-muted-foreground">您有待接受的邀请</p>
              <div className="space-y-1.5">
                <Label htmlFor="si-name">您的姓名</Label>
                <Input id="si-name" placeholder="输入姓名" value={memberName} onChange={(e) => setMemberName(e.target.value)} />
              </div>
              {invites.map((inv) => (
                <button key={inv.inviteCode} type="button" disabled={submitting} onClick={() => handleAcceptInvite(inv.inviteCode)} className="flex items-center justify-between rounded-lg border px-4 py-3 text-left transition-colors hover:bg-muted">
                  <div>
                    <div className="font-medium text-sm">{inv.companyName}</div>
                    <div className="text-xs text-muted-foreground">{inv.role}</div>
                  </div>
                  <span className="text-xs text-primary">接受</span>
                </button>
              ))}
              {displayError && <p className="text-sm text-destructive">{displayError}</p>}
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}

// --- Shared phone + code fields ---
function PhoneCodeFields({ phone, setPhone, code, setCode, canSend, sending, countdown, onSend }: {
  phone: string; setPhone: (v: string) => void
  code: string; setCode: (v: string) => void
  canSend: boolean; sending: boolean; countdown: number; onSend: () => void
}) {
  return (
    <>
      <div className="space-y-1.5">
        <Label htmlFor="popup-phone">手机号</Label>
        <div className="flex gap-2">
          <span className="flex items-center rounded-md border bg-muted px-3 text-sm text-muted-foreground">+86</span>
          <Input id="popup-phone" type="tel" inputMode="numeric" autoComplete="tel" placeholder="请输入手机号" value={phone} onChange={(e) => setPhone(e.target.value)} required />
        </div>
      </div>
      <div className="space-y-1.5">
        <Label htmlFor="popup-code">验证码</Label>
        <div className="flex gap-2">
          <Input id="popup-code" type="text" inputMode="numeric" autoComplete="one-time-code" placeholder="6 位验证码" maxLength={6} value={code} onChange={(e) => setCode(e.target.value.replace(/\D/g, ''))} required />
          <Button type="button" variant="outline" disabled={!canSend} onClick={onSend} className="shrink-0 whitespace-nowrap">
            {sending ? '发送中…' : countdown > 0 ? `${countdown}s` : '获取验证码'}
          </Button>
        </div>
      </div>
    </>
  )
}
