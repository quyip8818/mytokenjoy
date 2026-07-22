import { useCallback, useState } from 'react'
import { Dialog, DialogContent, DialogTitle } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { cn } from '@/lib/utils'
import { authApi, type CompanyOption, type LoginResult, type PendingInvite, type VerifyResult } from '@/api/auth'
import { ApiError } from '@/api/client'
import { useVerifyCountdown } from '../hooks/use-verify-countdown'

type AuthMode = 'login' | 'register'
type AuthStep =
  | 'login-phone-pw'
  | 'login-phone-code'
  | 'login-email-pw'
  | 'login-email-code'
  | 'reset-password'
  | 'reset-email-password'
  | 'register-phone'
  | 'register-email'
  | 'register-info'
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
  const [confirmPassword, setConfirmPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmNewPassword, setConfirmNewPassword] = useState('')
  const [email, setEmail] = useState('')
  const [companyName, setCompanyName] = useState('')
  const [industry, setIndustry] = useState('')
  const [size, setSize] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [successMessage, setSuccessMessage] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  // Multi-select state
  const [companies, setCompanies] = useState<CompanyOption[]>([])
  const [invites, setInvites] = useState<PendingInvite[]>([])
  const [memberName, setMemberName] = useState('')

  // Phone countdown
  const { sending, countdown, sendError, sendCode: sendPhoneCode } = useVerifyCountdown()
  const handleSendCode = useCallback(() => sendPhoneCode({ phone: phone.trim() }), [sendPhoneCode, phone])
  const handleSendRegisterPhoneCode = useCallback(() => sendPhoneCode({ phone: phone.trim(), purpose: 'register' }), [sendPhoneCode, phone])
  const canSend = phone.trim().length >= 11 && countdown === 0 && !sending

  // Email countdown
  const {
    sending: emailSending,
    countdown: emailCountdown,
    sendError: emailSendError,
    sendCode: sendEmailCode,
  } = useVerifyCountdown()
  const handleSendEmailCode = useCallback(() => sendEmailCode({ email: email.trim() }), [sendEmailCode, email])
  const handleSendRegisterEmailCode = useCallback(() => sendEmailCode({ email: email.trim(), purpose: 'register' }), [sendEmailCode, email])
  const canSendEmail = email.trim().length > 0 && email.includes('@') && emailCountdown === 0 && !emailSending

  const isLoginStep =
    step === 'login-phone-pw' || step === 'login-phone-code' || step === 'login-email-pw' || step === 'login-email-code'
  const showTabs = isLoginStep || step === 'register-phone' || step === 'register-email'

  // Clear sensitive fields whenever the visible step changes.
  const changeStep = useCallback((next: AuthStep) => {
    setStep(next)
    setPassword('')
    setConfirmPassword('')
    setNewPassword('')
    setConfirmNewPassword('')
    setCode('')
    setError(null)
    setSuccessMessage(null)
  }, [])

  const switchTab = useCallback((newMode: AuthMode) => {
    setMode(newMode)
    changeStep(newMode === 'login' ? 'login-phone-pw' : 'register-phone')
  }, [changeStep])

  // --- Shared login result handler (password login) ---
  const handleLoginResult = useCallback(
    (result: LoginResult) => {
      if ('action' in result && result.action === 'select_company') {
        setCompanies(result.companies)
        setStep('select-company')
      } else if ('action' in result && result.action === 'create_company') {
        setStep('register-info')
      } else {
        onSuccess?.()
      }
    },
    [onSuccess],
  )

  // --- Shared verify code result handler ---
  const handleVerifyResult = useCallback(
    (result: VerifyResult, notFoundMsg: string) => {
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
        case 'create_company':
          setStep('register-info')
          break
        case 'not_found':
          setError(notFoundMsg)
          break
      }
    },
    [onSuccess],
  )

  // --- Login: phone + password ---
  const handleLoginPhonePw = useCallback(
    async (e: React.FormEvent) => {
      e.preventDefault()
      if (!phone.trim() || !password) return
      setSubmitting(true)
      setError(null)
      setSuccessMessage(null)
      try {
        const result = await authApi.login({ email: phone.trim(), password })
        handleLoginResult(result)
      } catch (err) {
        setError(err instanceof ApiError ? err.message : '登录失败')
      } finally {
        setSubmitting(false)
      }
    },
    [phone, password, handleLoginResult],
  )

  // --- Login: email + password ---
  const handleLoginEmailPw = useCallback(
    async (e: React.FormEvent) => {
      e.preventDefault()
      if (!email.trim() || !password) return
      setSubmitting(true)
      setError(null)
      setSuccessMessage(null)
      try {
        const result = await authApi.login({ email: email.trim(), password })
        handleLoginResult(result)
      } catch (err) {
        setError(err instanceof ApiError ? err.message : '登录失败')
      } finally {
        setSubmitting(false)
      }
    },
    [email, password, handleLoginResult],
  )

  // --- Login: phone + SMS code ---
  const handleLoginPhoneCode = useCallback(
    async (e: React.FormEvent) => {
      e.preventDefault()
      if (!phone.trim() || !code.trim()) return
      setSubmitting(true)
      setError(null)
      try {
        const result = await authApi.verifyCode({ phone: phone.trim(), code: code.trim() })
        handleVerifyResult(result, '该手机号尚未注册，请切换到注册')
      } catch (err) {
        setError(err instanceof ApiError ? err.message : '验证失败')
      } finally {
        setSubmitting(false)
      }
    },
    [phone, code, handleVerifyResult],
  )

  // --- Login: email + verify code ---
  const handleLoginEmailCode = useCallback(
    async (e: React.FormEvent) => {
      e.preventDefault()
      if (!email.trim() || !code.trim()) return
      setSubmitting(true)
      setError(null)
      try {
        const result = await authApi.verifyCode({ email: email.trim(), code: code.trim() })
        handleVerifyResult(result, '该邮箱尚未注册')
      } catch (err) {
        setError(err instanceof ApiError ? err.message : '验证失败')
      } finally {
        setSubmitting(false)
      }
    },
    [email, code, handleVerifyResult],
  )

  // --- Select company ---
  const handleSelectCompany = useCallback(
    async (companyId: string) => {
      setSubmitting(true)
      setError(null)
      try {
        await authApi.selectCompany(companyId)
        onSuccess?.()
      } catch (err) {
        setError(err instanceof ApiError ? err.message : '选择失败')
      } finally {
        setSubmitting(false)
      }
    },
    [onSuccess],
  )

  // --- Accept invite ---
  const handleAcceptInvite = useCallback(
    async (inviteCode: string) => {
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
    },
    [memberName, onSuccess],
  )

  // --- Register: phone ---
  const handleRegisterVerify = useCallback(
    async (e: React.FormEvent) => {
      e.preventDefault()
      if (!phone.trim() || !code.trim() || password.length < 8) return
      if (password !== confirmPassword) { setError('两次密码输入不一致'); return }
      setSubmitting(true)
      setError(null)
      try {
        const result = await authApi.registerInit({ phone: phone.trim() }, code.trim(), password)
        if (result.action === 'login') { setError('该手机号已注册，请切换到登录'); return }
        if (result.invites && result.invites.length > 0) { setInvites(result.invites); setStep('select-invite'); return }
        setStep('register-info')
      } catch (err) {
        setError(err instanceof ApiError ? err.message : '验证失败')
      } finally {
        setSubmitting(false)
      }
    },
    [phone, code, password, confirmPassword],
  )

  // --- Register: email ---
  const handleRegisterEmailVerify = useCallback(
    async (e: React.FormEvent) => {
      e.preventDefault()
      if (!email.trim() || !code.trim() || password.length < 8) return
      if (password !== confirmPassword) { setError('两次密码输入不一致'); return }
      setSubmitting(true)
      setError(null)
      try {
        const result = await authApi.registerInit({ email: email.trim() }, code.trim(), password)
        if (result.action === 'login') { setError('该邮箱已注册，请切换到登录'); return }
        if (result.invites && result.invites.length > 0) { setInvites(result.invites); setStep('select-invite'); return }
        setStep('register-info')
      } catch (err) {
        setError(err instanceof ApiError ? err.message : '验证失败')
      } finally {
        setSubmitting(false)
      }
    },
    [email, code, password, confirmPassword],
  )

  // --- Create company ---
  const handleCreateCompany = useCallback(
    async (e: React.FormEvent) => {
      e.preventDefault()
      if (!companyName.trim()) return
      setSubmitting(true)
      setError(null)
      try {
        await authApi.registerCompany(companyName.trim(), industry, size)
        onSuccess?.()
      } catch (err) {
        setError(err instanceof ApiError ? err.message : '创建失败')
      } finally {
        setSubmitting(false)
      }
    },
    [companyName, industry, size, onSuccess],
  )

  // --- Reset password (phone) ---
  const handleResetPassword = useCallback(
    async (e: React.FormEvent) => {
      e.preventDefault()
      if (!phone.trim() || !code.trim() || newPassword.length < 8) return
      if (newPassword !== confirmNewPassword) { setError('两次密码输入不一致'); return }
      setSubmitting(true)
      setError(null)
      try {
        await authApi.resetPassword({ phone: phone.trim() }, code.trim(), newPassword)
        changeStep('login-phone-pw')
        setSuccessMessage('密码已重置，请使用新密码登录')
      } catch (err) {
        setError(err instanceof ApiError ? err.message : '重置失败')
      } finally {
        setSubmitting(false)
      }
    },
    [phone, code, newPassword, confirmNewPassword, changeStep],
  )

  // --- Reset password (email) ---
  const handleResetEmailPassword = useCallback(
    async (e: React.FormEvent) => {
      e.preventDefault()
      if (!email.trim() || !code.trim() || newPassword.length < 8) return
      if (newPassword !== confirmNewPassword) { setError('两次密码输入不一致'); return }
      setSubmitting(true)
      setError(null)
      try {
        await authApi.resetPassword({ email: email.trim() }, code.trim(), newPassword)
        changeStep('login-email-pw')
        setSuccessMessage('密码已重置，请使用新密码登录')
      } catch (err) {
        setError(err instanceof ApiError ? err.message : '重置失败')
      } finally {
        setSubmitting(false)
      }
    },
    [email, code, newPassword, confirmNewPassword, changeStep],
  )

  // --- Render ---
  const displayError = error || sendError

  return (
    <Dialog
      open={open}
      onOpenChange={closable ? (v) => { if (!v) onClose?.() } : undefined}
    >
      <DialogContent
        className="sm:max-w-[480px] gap-0 p-0 overflow-hidden border-border/50 shadow-[0_10px_50px_rgba(139,92,246,0.12)]"
        onPointerDownOutside={closable ? undefined : (e) => e.preventDefault()}
        onEscapeKeyDown={closable ? undefined : (e) => e.preventDefault()}
        showCloseButton={closable}
      >
        <DialogTitle className="sr-only">TokenJoy 认证</DialogTitle>

        {/* Header */}
        <div className="px-10 pt-10 pb-5 text-center">
          <img src="/logo.png" alt="Tokenjoy" className="mx-auto h-20 w-auto" />
          <p className="mt-2 text-sm text-muted-foreground tracking-wide">企业 AI 模型管理平台</p>
        </div>

        {/* Tabs */}
        {showTabs && (
          <div className="mx-10 flex border-b border-border">
            <button
              type="button"
              onClick={() => switchTab('login')}
              className={cn(
                'flex-1 pb-3 text-base font-medium transition-colors',
                mode === 'login' ? 'border-b-2 border-primary text-foreground' : 'text-muted-foreground hover:text-foreground',
              )}
            >
              登录
            </button>
            <button
              type="button"
              onClick={() => switchTab('register')}
              className={cn(
                'flex-1 pb-3 text-base font-medium transition-colors',
                mode === 'register' ? 'border-b-2 border-primary text-foreground' : 'text-muted-foreground hover:text-foreground',
              )}
            >
              注册
            </button>
          </div>
        )}

        {/* Content */}
        <div className="px-10 pb-10 pt-7">

          {/* === LOGIN: phone + password === */}
          {step === 'login-phone-pw' && (
            <form onSubmit={handleLoginPhonePw} className="flex flex-col gap-5">
              <FormMessage success={successMessage} />
              <PhoneField phone={phone} setPhone={setPhone} />
              <PasswordField id="lp-pw" value={password} onChange={setPassword} />
              <FormMessage error={displayError} />
              <Button type="submit" className="h-11 text-base font-medium" disabled={submitting || !phone.trim() || !password}>
                {submitting ? '登录中…' : '登录'}
              </Button>
              <LoginNav
                left={{ label: '验证码登录', onClick: () => changeStep('login-phone-code') }}
                right={{ label: '邮箱登录', onClick: () => changeStep('login-email-pw') }}
                forgot={{ onClick: () => changeStep('reset-password') }}
              />
            </form>
          )}

          {/* === LOGIN: phone + SMS code === */}
          {step === 'login-phone-code' && (
            <form onSubmit={handleLoginPhoneCode} className="flex flex-col gap-5">
              <PhoneCodeFields phone={phone} setPhone={setPhone} code={code} setCode={setCode} canSend={canSend} sending={sending} countdown={countdown} onSend={handleSendCode} />
              <FormMessage error={displayError} />
              <Button type="submit" className="h-11 text-base font-medium" disabled={submitting || !code.trim()}>
                {submitting ? '验证中…' : '登录'}
              </Button>
              <LoginNav
                left={{ label: '密码登录', onClick: () => changeStep('login-phone-pw') }}
                right={{ label: '邮箱登录', onClick: () => changeStep('login-email-pw') }}
              />
            </form>
          )}

          {/* === LOGIN: email + password === */}
          {step === 'login-email-pw' && (
            <form onSubmit={handleLoginEmailPw} className="flex flex-col gap-5">
              <FormMessage success={successMessage} />
              <EmailField id="le-email" email={email} setEmail={setEmail} />
              <PasswordField id="le-pw" value={password} onChange={setPassword} />
              <FormMessage error={displayError} />
              <Button type="submit" className="h-11 text-base font-medium" disabled={submitting || !email.trim() || !password}>
                {submitting ? '登录中…' : '登录'}
              </Button>
              <LoginNav
                left={{ label: '验证码登录', onClick: () => changeStep('login-email-code') }}
                right={{ label: '手机号登录', onClick: () => changeStep('login-phone-pw') }}
                forgot={{ onClick: () => changeStep('reset-email-password') }}
              />
            </form>
          )}

          {/* === LOGIN: email + verify code === */}
          {step === 'login-email-code' && (
            <form onSubmit={handleLoginEmailCode} className="flex flex-col gap-5">
              <EmailCodeFields id="lec" email={email} setEmail={setEmail} code={code} setCode={setCode} canSend={canSendEmail} sending={emailSending} countdown={emailCountdown} onSend={handleSendEmailCode} />
              <FormMessage error={error || emailSendError} />
              <Button type="submit" className="h-11 text-base font-medium" disabled={submitting || !code.trim()}>
                {submitting ? '验证中…' : '登录'}
              </Button>
              <LoginNav
                left={{ label: '密码登录', onClick: () => changeStep('login-email-pw') }}
                right={{ label: '手机号登录', onClick: () => changeStep('login-phone-pw') }}
              />
            </form>
          )}

          {/* === RESET PASSWORD (phone) === */}
          {step === 'reset-password' && (
            <form onSubmit={handleResetPassword} className="flex flex-col gap-5">
              <p className="text-base text-muted-foreground">通过短信验证码重置密码</p>
              <PhoneCodeFields phone={phone} setPhone={setPhone} code={code} setCode={setCode} canSend={canSend} sending={sending} countdown={countdown} onSend={handleSendCode} />
              <NewPasswordFields id="reset" password={newPassword} setPassword={setNewPassword} confirm={confirmNewPassword} setConfirm={setConfirmNewPassword} passwordLabel="新密码" confirmLabel="确认新密码" />
              <FormMessage error={displayError} />
              <Button type="submit" className="h-11 text-base font-medium" disabled={submitting || !code.trim() || newPassword.length < 8 || newPassword !== confirmNewPassword}>
                {submitting ? '重置中…' : '重置密码'}
              </Button>
              <BackLink label="返回登录" onClick={() => changeStep('login-phone-pw')} />
            </form>
          )}

          {/* === RESET PASSWORD (email) === */}
          {step === 'reset-email-password' && (
            <form onSubmit={handleResetEmailPassword} className="flex flex-col gap-5">
              <p className="text-base text-muted-foreground">通过邮箱验证码重置密码</p>
              <EmailCodeFields id="re" email={email} setEmail={setEmail} code={code} setCode={setCode} canSend={canSendEmail} sending={emailSending} countdown={emailCountdown} onSend={handleSendEmailCode} />
              <NewPasswordFields id="reset-email" password={newPassword} setPassword={setNewPassword} confirm={confirmNewPassword} setConfirm={setConfirmNewPassword} passwordLabel="新密码" confirmLabel="确认新密码" />
              <FormMessage error={error || emailSendError} />
              <Button type="submit" className="h-11 text-base font-medium" disabled={submitting || !code.trim() || newPassword.length < 8 || newPassword !== confirmNewPassword}>
                {submitting ? '重置中…' : '重置密码'}
              </Button>
              <BackLink label="返回登录" onClick={() => changeStep('login-email-pw')} />
            </form>
          )}

          {/* === REGISTER: phone === */}
          {step === 'register-phone' && (
            <form onSubmit={handleRegisterVerify} className="flex flex-col gap-5">
              <PhoneCodeFields phone={phone} setPhone={setPhone} code={code} setCode={setCode} canSend={canSend} sending={sending} countdown={countdown} onSend={handleSendRegisterPhoneCode} />
              <NewPasswordFields id="reg" password={password} setPassword={setPassword} confirm={confirmPassword} setConfirm={setConfirmPassword} />
              <FormMessage error={displayError} />
              <Button type="submit" className="h-11 text-base font-medium" disabled={submitting || !code.trim() || password.length < 8 || password !== confirmPassword}>
                {submitting ? '验证中…' : '下一步'}
              </Button>
              <SwitchLink label="使用邮箱注册" onClick={() => changeStep('register-email')} />
            </form>
          )}

          {/* === REGISTER: email === */}
          {step === 'register-email' && (
            <form onSubmit={handleRegisterEmailVerify} className="flex flex-col gap-5">
              <EmailCodeFields id="reg-email" email={email} setEmail={setEmail} code={code} setCode={setCode} canSend={canSendEmail} sending={emailSending} countdown={emailCountdown} onSend={handleSendRegisterEmailCode} />
              <NewPasswordFields id="reg-email" password={password} setPassword={setPassword} confirm={confirmPassword} setConfirm={setConfirmPassword} />
              <FormMessage error={error || emailSendError} />
              <Button type="submit" className="h-11 text-base font-medium" disabled={submitting || !code.trim() || password.length < 8 || password !== confirmPassword}>
                {submitting ? '验证中…' : '下一步'}
              </Button>
              <SwitchLink label="使用手机号注册" onClick={() => changeStep('register-phone')} />
            </form>
          )}

          {/* === REGISTER: company info === */}
          {step === 'register-info' && (
            <form onSubmit={handleCreateCompany} className="flex flex-col gap-5">
              <p className="text-base text-muted-foreground">创建您的企业</p>
              <div className="space-y-2">
                <Label htmlFor="ri-company" className="text-sm font-medium">公司名称</Label>
                <Input id="ri-company" type="text" placeholder="您的企业名称" className="h-11" value={companyName} onChange={(e) => setCompanyName(e.target.value)} required />
              </div>
              <div className="space-y-2">
                <Label className="text-sm font-medium">所属行业</Label>
                <Select value={industry} onValueChange={setIndustry}>
                  <SelectTrigger className="!h-11 w-full"><SelectValue placeholder="请选择行业" /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="互联网/科技">互联网/科技</SelectItem>
                    <SelectItem value="金融">金融</SelectItem>
                    <SelectItem value="教育">教育</SelectItem>
                    <SelectItem value="医疗健康">医疗健康</SelectItem>
                    <SelectItem value="电商/零售">电商/零售</SelectItem>
                    <SelectItem value="制造业">制造业</SelectItem>
                    <SelectItem value="游戏/娱乐">游戏/娱乐</SelectItem>
                    <SelectItem value="企业服务">企业服务</SelectItem>
                    <SelectItem value="政府/公共事业">政府/公共事业</SelectItem>
                    <SelectItem value="其他">其他</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label className="text-sm font-medium">人员规模</Label>
                <Select value={size} onValueChange={setSize}>
                  <SelectTrigger className="!h-11 w-full"><SelectValue placeholder="请选择人员规模" /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="1-10">1-10 人</SelectItem>
                    <SelectItem value="11-50">11-50 人</SelectItem>
                    <SelectItem value="51-200">51-200 人</SelectItem>
                    <SelectItem value="201-500">201-500 人</SelectItem>
                    <SelectItem value="501-1000">501-1000 人</SelectItem>
                    <SelectItem value="1000+">1000 人以上</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <FormMessage error={displayError} />
              <Button type="submit" className="h-11 text-base font-medium" disabled={submitting || !companyName.trim()}>
                {submitting ? '创建中…' : '创建并开始体验'}
              </Button>
              <BackLink label="返回" onClick={() => changeStep('register-phone')} />
            </form>
          )}

          {/* === SELECT COMPANY === */}
          {step === 'select-company' && (
            <div className="flex flex-col gap-4">
              <p className="text-base text-muted-foreground">选择企业</p>
              {companies.map((c) => (
                <button key={c.companyId} type="button" disabled={submitting} onClick={() => handleSelectCompany(c.companyId)}
                  className="flex items-center justify-between rounded-lg border px-4 py-3.5 text-left transition-colors hover:bg-muted">
                  <div>
                    <div className="font-medium text-base">{c.companyName}</div>
                    <div className="text-sm text-muted-foreground mt-0.5">{c.role}</div>
                  </div>
                </button>
              ))}
              <FormMessage error={displayError} />
            </div>
          )}

          {/* === SELECT INVITE === */}
          {step === 'select-invite' && (
            <div className="flex flex-col gap-4">
              <p className="text-base text-muted-foreground">您有待接受的邀请</p>
              <div className="space-y-2">
                <Label htmlFor="si-name" className="text-sm font-medium">您的姓名</Label>
                <Input id="si-name" placeholder="输入姓名" className="h-11" value={memberName} onChange={(e) => setMemberName(e.target.value)} />
              </div>
              {invites.map((inv) => (
                <button key={inv.inviteCode} type="button" disabled={submitting} onClick={() => handleAcceptInvite(inv.inviteCode)}
                  className="flex items-center justify-between rounded-lg border px-4 py-3.5 text-left transition-colors hover:bg-muted">
                  <div>
                    <div className="font-medium text-base">{inv.companyName}</div>
                    <div className="text-sm text-muted-foreground mt-0.5">{inv.role}</div>
                  </div>
                  <span className="text-sm text-primary font-medium">接受</span>
                </button>
              ))}
              <FormMessage error={displayError} />
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}

// ============================================================
// Shared sub-components
// ============================================================

/** Single phone input field */
function PhoneField({ phone, setPhone }: { phone: string; setPhone: (v: string) => void }) {
  return (
    <div className="space-y-2">
      <Label htmlFor="phone-field" className="text-sm font-medium">手机号</Label>
      <div className="flex gap-2">
        <span className="flex items-center rounded-md border bg-muted px-3 text-sm text-muted-foreground h-11">+86</span>
        <Input id="phone-field" type="tel" inputMode="numeric" autoComplete="tel" placeholder="请输入手机号" className="h-11" value={phone} onChange={(e) => setPhone(e.target.value)} required />
      </div>
    </div>
  )
}

/** Single email input field */
function EmailField({ id, email, setEmail }: { id: string; email: string; setEmail: (v: string) => void }) {
  return (
    <div className="space-y-2">
      <Label htmlFor={id} className="text-sm font-medium">邮箱</Label>
      <Input id={id} type="email" autoComplete="username" placeholder="name@company.com" className="h-11" value={email} onChange={(e) => setEmail(e.target.value)} required />
    </div>
  )
}

/** Password input field */
function PasswordField({ id, value, onChange }: { id: string; value: string; onChange: (v: string) => void }) {
  return (
    <div className="space-y-2">
      <Label htmlFor={id} className="text-sm font-medium">密码</Label>
      <Input id={id} type="password" autoComplete="current-password" placeholder="输入密码" className="h-11" value={value} onChange={(e) => onChange(e.target.value)} required />
    </div>
  )
}

/** Phone + verification code fields */
function PhoneCodeFields({ phone, setPhone, code, setCode, canSend, sending, countdown, onSend }: {
  phone: string; setPhone: (v: string) => void
  code: string; setCode: (v: string) => void
  canSend: boolean; sending: boolean; countdown: number; onSend: () => void
}) {
  return (
    <>
      <PhoneField phone={phone} setPhone={setPhone} />
      <CodeField id="phone-code" code={code} setCode={setCode} canSend={canSend} sending={sending} countdown={countdown} onSend={onSend} />
    </>
  )
}

/** Email + verification code fields */
function EmailCodeFields({ id, email, setEmail, code, setCode, canSend, sending, countdown, onSend }: {
  id: string; email: string; setEmail: (v: string) => void
  code: string; setCode: (v: string) => void
  canSend: boolean; sending: boolean; countdown: number; onSend: () => void
}) {
  return (
    <>
      <EmailField id={`${id}-email`} email={email} setEmail={setEmail} />
      <CodeField id={`${id}-code`} code={code} setCode={setCode} canSend={canSend} sending={sending} countdown={countdown} onSend={onSend} />
    </>
  )
}

/** Verification code input + send button (shared by phone and email) */
function CodeField({ id, code, setCode, canSend, sending, countdown, onSend }: {
  id: string; code: string; setCode: (v: string) => void
  canSend: boolean; sending: boolean; countdown: number; onSend: () => void
}) {
  return (
    <div className="space-y-2">
      <Label htmlFor={id} className="text-sm font-medium">验证码</Label>
      <div className="flex gap-2">
        <Input id={id} type="text" inputMode="numeric" autoComplete="one-time-code" placeholder="6 位验证码" className="h-11" maxLength={6} value={code} onChange={(e) => setCode(e.target.value.replace(/\D/g, ''))} required />
        <Button type="button" variant="outline" disabled={!canSend} onClick={onSend} className="shrink-0 whitespace-nowrap h-11">
          {sending ? '发送中…' : countdown > 0 ? `${countdown}s` : '获取验证码'}
        </Button>
      </div>
    </div>
  )
}

/** New password + confirm fields */
function NewPasswordFields({ id, password, setPassword, confirm, setConfirm, passwordLabel = '设置密码', confirmLabel = '确认密码' }: {
  id: string; password: string; setPassword: (v: string) => void
  confirm: string; setConfirm: (v: string) => void
  passwordLabel?: string; confirmLabel?: string
}) {
  const hint =
    password.length > 0 && password.length < 8 ? '密码至少需要 8 位'
    : confirm.length > 0 && confirm !== password ? '两次密码输入不一致'
    : null
  return (
    <>
      <div className="space-y-2">
        <Label htmlFor={`${id}-pw`} className="text-sm font-medium">{passwordLabel}</Label>
        <Input id={`${id}-pw`} type="password" autoComplete="new-password" placeholder="至少 8 位" className="h-11" value={password} onChange={(e) => setPassword(e.target.value)} required minLength={8} />
      </div>
      <div className="space-y-2">
        <Label htmlFor={`${id}-pw-confirm`} className="text-sm font-medium">{confirmLabel}</Label>
        <Input id={`${id}-pw-confirm`} type="password" autoComplete="new-password" placeholder="再次输入密码" className="h-11" value={confirm} onChange={(e) => setConfirm(e.target.value)} required minLength={8} />
        {hint && <p className="text-xs text-destructive mt-1">{hint}</p>}
      </div>
    </>
  )
}

// ============================================================
// Navigation components (unified layout for login steps)
// ============================================================

/** Login step bottom navigation: left link | right link | optional forgot password */
function LoginNav({ left, right, forgot }: {
  left: { label: string; onClick: () => void }
  right: { label: string; onClick: () => void }
  forgot?: { onClick: () => void }
}) {
  return (
    <div className="flex flex-col gap-2 pt-1">
      <div className="flex justify-between text-sm text-muted-foreground">
        <button type="button" onClick={left.onClick} className="hover:text-foreground transition-colors">{left.label}</button>
        <button type="button" onClick={right.onClick} className="hover:text-foreground transition-colors">{right.label}</button>
      </div>
      {forgot && (
        <div className="text-center">
          <button type="button" onClick={forgot.onClick} className="text-sm text-muted-foreground hover:text-foreground transition-colors">忘记密码？</button>
        </div>
      )}
    </div>
  )
}

/** Back link for sub-flows (reset password, register info) */
function BackLink({ label, onClick }: { label: string; onClick: () => void }) {
  return (
    <button type="button" onClick={onClick} className="text-sm text-muted-foreground hover:text-foreground transition-colors">
      ← {label}
    </button>
  )
}

/** Centered switch link (register phone/email toggle) */
function SwitchLink({ label, onClick }: { label: string; onClick: () => void }) {
  return (
    <div className="text-center pt-1">
      <button type="button" onClick={onClick} className="text-sm text-muted-foreground hover:text-foreground transition-colors">{label}</button>
    </div>
  )
}

/** Unified form message */
function FormMessage({ error, success, hint }: { error?: string | null; success?: string | null; hint?: string | null }) {
  const msg = error || success || hint
  if (!msg) return null
  const style = error
    ? 'bg-destructive/10 border-destructive/20 text-destructive'
    : success
      ? 'bg-emerald-50 border-emerald-200 text-emerald-700 dark:bg-emerald-950/30 dark:border-emerald-800 dark:text-emerald-400'
      : 'bg-muted border-border text-muted-foreground'
  return (
    <div className={cn('rounded-md border px-3 py-2 text-sm', style)} role={error ? 'alert' : 'status'}>{msg}</div>
  )
}
