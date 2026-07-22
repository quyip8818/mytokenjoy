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
import { authApi, type CompanyOption, type PendingInvite, type VerifyResult } from '@/api/auth'
import { ApiError } from '@/api/client'
import { useVerifyCountdown } from '../hooks/use-verify-countdown'

type AuthMode = 'login' | 'register'
type AuthStep =
  | 'login-phone-pw' // 默认：手机号 + 密码
  | 'login-phone-code' // 手机号 + 验证码
  | 'login-email-pw' // 邮箱 + 密码
  | 'login-email-code' // 邮箱 + 验证码
  | 'reset-password' // 忘记密码：手机号 + 验证码 + 新密码
  | 'register-phone' // 手机号 + 验证码 + 密码
  | 'register-info' // 公司名
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

  const { sending, countdown, sendError, sendCode: sendPhoneCode } = useVerifyCountdown()
  const handleSendCode = useCallback(() => sendPhoneCode({ phone: phone.trim() }), [sendPhoneCode, phone])
  const canSend = phone.trim().length >= 11 && countdown === 0 && !sending

  // Email verify code countdown (independent from phone)
  const {
    sending: emailSending,
    countdown: emailCountdown,
    sendError: emailSendError,
    sendCode: sendEmailCode,
  } = useVerifyCountdown()
  const handleSendEmailCode = useCallback(() => sendEmailCode({ email: email.trim() }), [sendEmailCode, email])
  const canSendEmail = email.trim().length > 0 && email.includes('@') && emailCountdown === 0 && !emailSending

  const isLoginStep =
    step === 'login-phone-pw' || step === 'login-phone-code' || step === 'login-email-pw' || step === 'login-email-code'
  const showTabs = isLoginStep || step === 'register-phone'

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

  // --- Tab switch ---
  const switchTab = useCallback((newMode: AuthMode) => {
    setMode(newMode)
    changeStep(newMode === 'login' ? 'login-phone-pw' : 'register-phone')
  }, [changeStep])

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
        if ('action' in result && result.action === 'select_company') {
          setCompanies(result.companies)
          setStep('select-company')
        } else if ('action' in result && result.action === 'create_company') {
          setStep('register-info')
        } else {
          onSuccess?.()
        }
      } catch (err) {
        setError(err instanceof ApiError ? err.message : '登录失败')
      } finally {
        setSubmitting(false)
      }
    },
    [phone, password, onSuccess],
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

  // --- Login: email + password ---
  const handleLoginEmailPw = useCallback(
    async (e: React.FormEvent) => {
      e.preventDefault()
      if (!email.trim() || !password) return
      setSubmitting(true)
      setError(null)
      try {
        const result = await authApi.login({ email: email.trim(), password })
        if ('action' in result && result.action === 'select_company') {
          setCompanies(result.companies)
          setStep('select-company')
        } else if ('action' in result && result.action === 'create_company') {
          setStep('register-info')
        } else {
          onSuccess?.()
        }
      } catch (err) {
        setError(err instanceof ApiError ? err.message : '登录失败')
      } finally {
        setSubmitting(false)
      }
    },
    [email, password, onSuccess],
  )

  // --- Login: select company ---
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

  // --- Login: accept invite ---
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

  // --- Register: verify phone + set password → create user ---
  const handleRegisterVerify = useCallback(
    async (e: React.FormEvent) => {
      e.preventDefault()
      if (!phone.trim() || !code.trim() || password.length < 8) return
      if (password !== confirmPassword) {
        setError('两次密码输入不一致')
        return
      }
      setSubmitting(true)
      setError(null)
      try {
        const result = await authApi.registerInit(phone.trim(), code.trim(), password)
        if (result.action === 'login') {
          setError('该手机号已注册，请切换到登录')
          return
        }
        // result.action === 'choose' — check if there are pending invites.
        if (result.invites && result.invites.length > 0) {
          setInvites(result.invites)
          setStep('select-invite')
          return
        }
        // No invites → move to company creation step.
        setStep('register-info')
      } catch (err) {
        setError(err instanceof ApiError ? err.message : '验证失败')
      } finally {
        setSubmitting(false)
      }
    },
    [phone, code, password, confirmPassword],
  )

  // --- Register: create company (password already stored in init step) ---
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

  // --- Reset password ---
  const handleResetPassword = useCallback(
    async (e: React.FormEvent) => {
      e.preventDefault()
      if (!phone.trim() || !code.trim() || newPassword.length < 8) return
      if (newPassword !== confirmNewPassword) {
        setError('两次密码输入不一致')
        return
      }
      setSubmitting(true)
      setError(null)
      try {
        await authApi.resetPassword(phone.trim(), code.trim(), newPassword)
        // Success → switch back to login with a success message.
        changeStep('login-phone-pw')
        setSuccessMessage('密码已重置，请使用新密码登录')
      } catch (err) {
        setError(err instanceof ApiError ? err.message : '重置失败')
      } finally {
        setSubmitting(false)
      }
    },
    [phone, code, newPassword, confirmNewPassword],
  )

  // --- Render ---
  const displayError = error || sendError

  return (
    <Dialog
      open={open}
      onOpenChange={
        closable
          ? (v) => {
              if (!v) onClose?.()
            }
          : undefined
      }
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
                'flex-1 pb-3 text-base font-medium transition-colors',
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
        <div className="px-10 pb-10 pt-7">
          {/* === LOGIN: phone + password (default) === */}
          {step === 'login-phone-pw' && (
            <form onSubmit={handleLoginPhonePw} className="flex flex-col gap-5">
              <FormMessage success={successMessage} />
              <div className="space-y-2">
                <Label htmlFor="lp-phone" className="text-sm font-medium">
                  手机号
                </Label>
                <div className="flex gap-2">
                  <span className="flex items-center rounded-md border bg-muted px-3 text-sm text-muted-foreground h-11">
                    +86
                  </span>
                  <Input
                    id="lp-phone"
                    type="tel"
                    inputMode="numeric"
                    autoComplete="tel"
                    placeholder="请输入手机号"
                    className="h-11"
                    value={phone}
                    onChange={(e) => setPhone(e.target.value)}
                    required
                  />
                </div>
              </div>
              <div className="space-y-2">
                <Label htmlFor="lp-pw" className="text-sm font-medium">
                  密码
                </Label>
                <Input
                  id="lp-pw"
                  type="password"
                  autoComplete="current-password"
                  placeholder="输入密码"
                  className="h-11"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                />
              </div>
              <FormMessage error={displayError} />
              <Button
                type="submit"
                className="h-11 text-base font-medium"
                disabled={submitting || !phone.trim() || !password}
              >
                {submitting ? '登录中…' : '登录'}
              </Button>
              <div className="flex justify-between text-sm text-muted-foreground">
                <button
                  type="button"
                  onClick={() => {
                    changeStep('login-phone-code')
                  }}
                  className="hover:text-foreground transition-colors"
                >
                  验证码登录
                </button>
                <button
                  type="button"
                  onClick={() => {
                    changeStep('login-email-pw')
                  }}
                  className="hover:text-foreground transition-colors"
                >
                  邮箱登录
                </button>
              </div>
              <div className="text-center">
                <button
                  type="button"
                  onClick={() => {
                    changeStep('reset-password')
                  }}
                  className="text-sm text-muted-foreground hover:text-foreground transition-colors"
                >
                  忘记密码？
                </button>
              </div>
            </form>
          )}

          {/* === LOGIN: phone + SMS code === */}
          {step === 'login-phone-code' && (
            <form onSubmit={handleLoginPhoneCode} className="flex flex-col gap-5">
              <PhoneCodeFields
                phone={phone}
                setPhone={setPhone}
                code={code}
                setCode={setCode}
                canSend={canSend}
                sending={sending}
                countdown={countdown}
                onSend={handleSendCode}
              />
              <FormMessage error={displayError} />
              <Button
                type="submit"
                className="h-11 text-base font-medium"
                disabled={submitting || !code.trim()}
              >
                {submitting ? '验证中…' : '登录'}
              </Button>
              <div className="flex justify-between text-sm text-muted-foreground">
                <button
                  type="button"
                  onClick={() => {
                    changeStep('login-phone-pw')
                  }}
                  className="hover:text-foreground transition-colors"
                >
                  ← 密码登录
                </button>
                <button
                  type="button"
                  onClick={() => {
                    changeStep('login-email-pw')
                  }}
                  className="hover:text-foreground transition-colors"
                >
                  邮箱登录
                </button>
              </div>
            </form>
          )}

          {/* === LOGIN: email + password === */}
          {step === 'login-email-pw' && (
            <form onSubmit={handleLoginEmailPw} className="flex flex-col gap-5">
              <div className="space-y-2">
                <Label htmlFor="le-email" className="text-sm font-medium">
                  邮箱
                </Label>
                <Input
                  id="le-email"
                  type="email"
                  autoComplete="username"
                  placeholder="name@company.com"
                  className="h-11"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="le-pw" className="text-sm font-medium">
                  密码
                </Label>
                <Input
                  id="le-pw"
                  type="password"
                  autoComplete="current-password"
                  placeholder="输入密码"
                  className="h-11"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                />
              </div>
              <FormMessage error={displayError} />
              <Button
                type="submit"
                className="h-11 text-base font-medium"
                disabled={submitting || !email.trim() || !password}
              >
                {submitting ? '登录中…' : '登录'}
              </Button>
              <div className="flex justify-between text-sm text-muted-foreground">
                <button
                  type="button"
                  onClick={() => {
                    changeStep('login-phone-pw')
                  }}
                  className="hover:text-foreground transition-colors"
                >
                  ← 手机号登录
                </button>
                <button
                  type="button"
                  onClick={() => {
                    changeStep('login-email-code')
                  }}
                  className="hover:text-foreground transition-colors"
                >
                  验证码登录
                </button>
              </div>
            </form>
          )}

          {/* === LOGIN: email + verify code === */}
          {step === 'login-email-code' && (
            <form onSubmit={handleLoginEmailCode} className="flex flex-col gap-5">
              <div className="space-y-2">
                <Label htmlFor="lec-email" className="text-sm font-medium">
                  邮箱
                </Label>
                <Input
                  id="lec-email"
                  type="email"
                  autoComplete="username"
                  placeholder="name@company.com"
                  className="h-11"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="lec-code" className="text-sm font-medium">
                  验证码
                </Label>
                <div className="flex gap-2">
                  <Input
                    id="lec-code"
                    type="text"
                    inputMode="numeric"
                    autoComplete="one-time-code"
                    placeholder="6 位验证码"
                    className="h-11"
                    maxLength={6}
                    value={code}
                    onChange={(e) => setCode(e.target.value.replace(/\D/g, ''))}
                    required
                  />
                  <Button
                    type="button"
                    variant="outline"
                    disabled={!canSendEmail}
                    onClick={handleSendEmailCode}
                    className="shrink-0 whitespace-nowrap h-11"
                  >
                    {emailSending ? '发送中…' : emailCountdown > 0 ? `${emailCountdown}s` : '获取验证码'}
                  </Button>
                </div>
              </div>
              <FormMessage error={error || emailSendError} />
              <Button
                type="submit"
                className="h-11 text-base font-medium"
                disabled={submitting || !code.trim()}
              >
                {submitting ? '验证中…' : '登录'}
              </Button>
              <div className="flex justify-between text-sm text-muted-foreground">
                <button
                  type="button"
                  onClick={() => {
                    changeStep('login-email-pw')
                  }}
                  className="hover:text-foreground transition-colors"
                >
                  ← 密码登录
                </button>
                <button
                  type="button"
                  onClick={() => {
                    changeStep('login-phone-pw')
                  }}
                  className="hover:text-foreground transition-colors"
                >
                  手机号登录
                </button>
              </div>
            </form>
          )}

          {/* === RESET PASSWORD === */}
          {step === 'reset-password' && (
            <form onSubmit={handleResetPassword} className="flex flex-col gap-5">
              <p className="text-base text-muted-foreground">通过短信验证码重置密码</p>
              <PhoneCodeFields
                phone={phone}
                setPhone={setPhone}
                code={code}
                setCode={setCode}
                canSend={canSend}
                sending={sending}
                countdown={countdown}
                onSend={handleSendCode}
              />
              <NewPasswordFields
                id="reset"
                password={newPassword}
                setPassword={setNewPassword}
                confirm={confirmNewPassword}
                setConfirm={setConfirmNewPassword}
                passwordLabel="新密码"
                confirmLabel="确认新密码"
              />
              <FormMessage error={displayError} />
              <Button
                type="submit"
                className="h-11 text-base font-medium"
                disabled={
                  submitting || !code.trim() || newPassword.length < 8 || newPassword !== confirmNewPassword
                }
              >
                {submitting ? '重置中…' : '重置密码'}
              </Button>
              <button
                type="button"
                onClick={() => {
                  changeStep('login-phone-pw')
                }}
                className="text-sm text-muted-foreground hover:text-foreground transition-colors"
              >
                ← 返回登录
              </button>
            </form>
          )}

          {/* === REGISTER: phone + code + password (one page) === */}
          {step === 'register-phone' && (
            <form onSubmit={handleRegisterVerify} className="flex flex-col gap-5">
              <PhoneCodeFields
                phone={phone}
                setPhone={setPhone}
                code={code}
                setCode={setCode}
                canSend={canSend}
                sending={sending}
                countdown={countdown}
                onSend={handleSendCode}
              />
              <NewPasswordFields
                id="reg"
                password={password}
                setPassword={setPassword}
                confirm={confirmPassword}
                setConfirm={setConfirmPassword}
              />
              <FormMessage error={displayError} />
              <Button
                type="submit"
                className="h-11 text-base font-medium"
                disabled={submitting || !code.trim() || password.length < 8 || password !== confirmPassword}
              >
                {submitting ? '验证中…' : '下一步'}
              </Button>
            </form>
          )}

          {/* === REGISTER: company name + industry + size === */}
          {step === 'register-info' && (
            <form onSubmit={handleCreateCompany} className="flex flex-col gap-5">
              <p className="text-base text-muted-foreground">创建您的企业</p>
              <div className="space-y-2">
                <Label htmlFor="ri-company" className="text-sm font-medium">
                  公司名称
                </Label>
                <Input
                  id="ri-company"
                  type="text"
                  placeholder="您的企业名称"
                  className="h-11"
                  value={companyName}
                  onChange={(e) => setCompanyName(e.target.value)}
                  required
                />
              </div>
              <div className="space-y-2">
                <Label className="text-sm font-medium">所属行业</Label>
                <Select value={industry} onValueChange={setIndustry}>
                  <SelectTrigger className="!h-11 w-full">
                    <SelectValue placeholder="请选择行业" />
                  </SelectTrigger>
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
                  <SelectTrigger className="!h-11 w-full">
                    <SelectValue placeholder="请选择人员规模" />
                  </SelectTrigger>
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
              <Button
                type="submit"
                className="h-11 text-base font-medium"
                disabled={submitting || !companyName.trim()}
              >
                {submitting ? '创建中…' : '创建并开始体验'}
              </Button>
              <button
                type="button"
                onClick={() => changeStep('register-phone')}
                className="text-sm text-muted-foreground hover:text-foreground transition-colors"
              >
                ← 返回
              </button>
            </form>
          )}

          {/* === SELECT COMPANY === */}
          {step === 'select-company' && (
            <div className="flex flex-col gap-4">
              <p className="text-base text-muted-foreground">选择企业</p>
              {companies.map((c) => (
                <button
                  key={c.companyId}
                  type="button"
                  disabled={submitting}
                  onClick={() => handleSelectCompany(c.companyId)}
                  className="flex items-center justify-between rounded-lg border px-4 py-3.5 text-left transition-colors hover:bg-muted"
                >
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
                <Label htmlFor="si-name" className="text-sm font-medium">
                  您的姓名
                </Label>
                <Input
                  id="si-name"
                  placeholder="输入姓名"
                  className="h-11"
                  value={memberName}
                  onChange={(e) => setMemberName(e.target.value)}
                />
              </div>
              {invites.map((inv) => (
                <button
                  key={inv.inviteCode}
                  type="button"
                  disabled={submitting}
                  onClick={() => handleAcceptInvite(inv.inviteCode)}
                  className="flex items-center justify-between rounded-lg border px-4 py-3.5 text-left transition-colors hover:bg-muted"
                >
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

// --- Shared phone + code fields ---
function PhoneCodeFields({
  phone,
  setPhone,
  code,
  setCode,
  canSend,
  sending,
  countdown,
  onSend,
}: {
  phone: string
  setPhone: (v: string) => void
  code: string
  setCode: (v: string) => void
  canSend: boolean
  sending: boolean
  countdown: number
  onSend: () => void
}) {
  return (
    <>
      <div className="space-y-2">
        <Label htmlFor="popup-phone" className="text-sm font-medium">
          手机号
        </Label>
        <div className="flex gap-2">
          <span className="flex items-center rounded-md border bg-muted px-3 text-sm text-muted-foreground h-11">
            +86
          </span>
          <Input
            id="popup-phone"
            type="tel"
            inputMode="numeric"
            autoComplete="tel"
            placeholder="请输入手机号"
            className="h-11"
            value={phone}
            onChange={(e) => setPhone(e.target.value)}
            required
          />
        </div>
      </div>
      <div className="space-y-2">
        <Label htmlFor="popup-code" className="text-sm font-medium">
          验证码
        </Label>
        <div className="flex gap-2">
          <Input
            id="popup-code"
            type="text"
            inputMode="numeric"
            autoComplete="one-time-code"
            placeholder="6 位验证码"
            className="h-11"
            maxLength={6}
            value={code}
            onChange={(e) => setCode(e.target.value.replace(/\D/g, ''))}
            required
          />
          <Button
            type="button"
            variant="outline"
            disabled={!canSend}
            onClick={onSend}
            className="shrink-0 whitespace-nowrap h-11"
          >
            {sending ? '发送中…' : countdown > 0 ? `${countdown}s` : '获取验证码'}
          </Button>
        </div>
      </div>
    </>
  )
}

// --- New password + confirm fields with real-time validation ---
function NewPasswordFields({
  id,
  password,
  setPassword,
  confirm,
  setConfirm,
  passwordLabel = '设置密码',
  confirmLabel = '确认密码',
}: {
  id: string
  password: string
  setPassword: (v: string) => void
  confirm: string
  setConfirm: (v: string) => void
  passwordLabel?: string
  confirmLabel?: string
}) {
  const hint =
    password.length > 0 && password.length < 8
      ? '密码至少需要 8 位'
      : confirm.length > 0 && confirm !== password
        ? '两次密码输入不一致'
        : null

  return (
    <>
      <div className="space-y-2">
        <Label htmlFor={`${id}-pw`} className="text-sm font-medium">
          {passwordLabel}
        </Label>
        <Input
          id={`${id}-pw`}
          type="password"
          autoComplete="new-password"
          placeholder="至少 8 位"
          className="h-11"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          required
          minLength={8}
        />
      </div>
      <div className="space-y-2">
        <Label htmlFor={`${id}-pw-confirm`} className="text-sm font-medium">
          {confirmLabel}
        </Label>
        <Input
          id={`${id}-pw-confirm`}
          type="password"
          autoComplete="new-password"
          placeholder="再次输入密码"
          className="h-11"
          value={confirm}
          onChange={(e) => setConfirm(e.target.value)}
          required
          minLength={8}
        />
        {hint && <p className="text-xs text-destructive mt-1">{hint}</p>}
      </div>
    </>
  )
}

// --- Unified form message (error / success / hint) ---
function FormMessage({ error, success, hint }: { error?: string | null; success?: string | null; hint?: string | null }) {
  const msg = error || success || hint
  if (!msg) return null
  const style = error
    ? 'bg-destructive/10 border-destructive/20 text-destructive'
    : success
      ? 'bg-emerald-50 border-emerald-200 text-emerald-700 dark:bg-emerald-950/30 dark:border-emerald-800 dark:text-emerald-400'
      : 'bg-muted border-border text-muted-foreground'
  return (
    <div className={cn('rounded-md border px-3 py-2 text-sm', style)} role={error ? 'alert' : 'status'}>
      {msg}
    </div>
  )
}
