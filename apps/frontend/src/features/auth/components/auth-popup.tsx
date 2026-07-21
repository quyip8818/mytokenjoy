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
import { authApi, type CompanyOption, type PendingInvite, type SmsVerifyResult } from '@/api/auth'
import { ApiError } from '@/api/client'
import { useSmsCountdown } from '../hooks/use-sms-countdown'

type AuthMode = 'login' | 'register'
type AuthStep =
  | 'login-phone-pw' // 默认：手机号 + 密码
  | 'login-phone-code' // 手机号 + 验证码
  | 'login-email-pw' // 邮箱 + 密码
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
  const [email, setEmail] = useState('')
  const [companyName, setCompanyName] = useState('')
  const [industry, setIndustry] = useState('')
  const [size, setSize] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  // Multi-select state
  const [companies, setCompanies] = useState<CompanyOption[]>([])
  const [invites, setInvites] = useState<PendingInvite[]>([])
  const [memberName, setMemberName] = useState('')

  const { sending, countdown, sendError, sendCode } = useSmsCountdown()
  const handleSendCode = useCallback(() => sendCode(phone), [sendCode, phone])
  const canSend = phone.trim().length >= 11 && countdown === 0 && !sending

  const isLoginStep =
    step === 'login-phone-pw' || step === 'login-phone-code' || step === 'login-email-pw'
  const showTabs = isLoginStep || step === 'register-phone'

  // --- Tab switch ---
  const switchTab = useCallback((newMode: AuthMode) => {
    setMode(newMode)
    setStep(newMode === 'login' ? 'login-phone-pw' : 'register-phone')
    setError(null)
  }, [])

  // --- Login: phone + password ---
  const handleLoginPhonePw = useCallback(
    async (e: React.FormEvent) => {
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
    },
    [phone, password, onSuccess],
  )

  // --- Login: phone + SMS code ---
  const handleLoginPhoneCode = useCallback(
    async (e: React.FormEvent) => {
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
    },
    [phone, code, onSuccess],
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
        await authApi.smsSelect(companyId)
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
    },
    [phone, code, password, confirmPassword],
  )

  // --- Register: create company (password already collected in previous step) ---
  const handleCreateCompany = useCallback(
    async (e: React.FormEvent) => {
      e.preventDefault()
      if (!companyName.trim() || password.length < 8) return
      setSubmitting(true)
      setError(null)
      try {
        await authApi.registerCompany(companyName.trim(), password, industry, size)
        onSuccess?.()
      } catch (err) {
        setError(err instanceof ApiError ? err.message : '创建失败')
      } finally {
        setSubmitting(false)
      }
    },
    [companyName, password, industry, size, onSuccess],
  )

  // --- Reset password ---
  const handleResetPassword = useCallback(
    async (e: React.FormEvent) => {
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
    },
    [phone, code, newPassword],
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
              <div className="space-y-2">
                <Label htmlFor="lp-phone" className="text-sm font-medium">手机号</Label>
                <div className="flex gap-2">
                  <span className="flex items-center rounded-md border bg-muted px-3 text-sm text-muted-foreground">
                    +86
                  </span>
                  <Input
                    id="lp-phone"
                    type="tel"
                    inputMode="numeric"
                    autoComplete="tel"
                    placeholder="请输入手机号"
                    className="h-10"
                    value={phone}
                    onChange={(e) => setPhone(e.target.value)}
                    required
                  />
                </div>
              </div>
              <div className="space-y-2">
                <Label htmlFor="lp-pw" className="text-sm font-medium">密码</Label>
                <Input
                  id="lp-pw"
                  type="password"
                  autoComplete="current-password"
                  placeholder="输入密码"
                  className="h-10"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                />
              </div>
              {displayError && <p className="text-sm text-destructive">{displayError}</p>}
              <Button type="submit" className="h-10 text-sm font-medium" disabled={submitting || !phone.trim() || !password}>
                {submitting ? '登录中…' : '登录'}
              </Button>
              <div className="flex justify-between text-sm text-foreground">
                <button
                  type="button"
                  onClick={() => {
                    setStep('login-phone-code')
                    setError(null)
                  }}
                  className="hover:text-foreground transition-colors"
                >
                  验证码登录
                </button>
                <button
                  type="button"
                  onClick={() => {
                    setStep('login-email-pw')
                    setError(null)
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
                    setStep('reset-password')
                    setError(null)
                    setCode('')
                  }}
                  className="text-sm text-foreground hover:text-foreground transition-colors"
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
              {displayError && <p className="text-sm text-destructive">{displayError}</p>}
              <Button type="submit" className="h-10 text-sm font-medium" disabled={submitting || !code.trim()}>
                {submitting ? '验证中…' : '登录'}
              </Button>
              <div className="flex justify-between text-sm text-foreground">
                <button
                  type="button"
                  onClick={() => {
                    setStep('login-phone-pw')
                    setError(null)
                  }}
                  className="hover:text-foreground transition-colors"
                >
                  ← 密码登录
                </button>
                <button
                  type="button"
                  onClick={() => {
                    setStep('login-email-pw')
                    setError(null)
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
                <Label htmlFor="le-email" className="text-sm font-medium">邮箱</Label>
                <Input
                  id="le-email"
                  type="email"
                  autoComplete="username"
                  placeholder="name@company.com"
                  className="h-10"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="le-pw" className="text-sm font-medium">密码</Label>
                <Input
                  id="le-pw"
                  type="password"
                  autoComplete="current-password"
                  placeholder="输入密码"
                  className="h-10"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                />
              </div>
              {displayError && <p className="text-sm text-destructive">{displayError}</p>}
              <Button type="submit" className="h-10 text-sm font-medium" disabled={submitting || !email.trim() || !password}>
                {submitting ? '登录中…' : '登录'}
              </Button>
              <div className="flex justify-center text-sm text-foreground">
                <button
                  type="button"
                  onClick={() => {
                    setStep('login-phone-pw')
                    setError(null)
                  }}
                  className="hover:text-foreground transition-colors"
                >
                  ← 手机号登录
                </button>
              </div>
            </form>
          )}

          {/* === RESET PASSWORD === */}
          {step === 'reset-password' && (
            <form onSubmit={handleResetPassword} className="flex flex-col gap-5">
              <p className="text-sm text-muted-foreground">通过短信验证码重置密码</p>
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
              <div className="space-y-2">
                <Label htmlFor="rp-new-pw" className="text-sm font-medium">新密码</Label>
                <Input
                  id="rp-new-pw"
                  type="password"
                  autoComplete="new-password"
                  placeholder="至少 8 位"
                  className="h-10"
                  value={newPassword}
                  onChange={(e) => setNewPassword(e.target.value)}
                  required
                  minLength={8}
                />
              </div>
              {displayError && <p className="text-sm text-destructive">{displayError}</p>}
              <Button type="submit" className="h-10 text-sm font-medium" disabled={submitting || !code.trim() || newPassword.length < 8}>
                {submitting ? '重置中…' : '重置密码'}
              </Button>
              <button
                type="button"
                onClick={() => {
                  setStep('login-phone-pw')
                  setError(null)
                }}
                className="text-sm text-foreground hover:text-foreground transition-colors"
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
              <div className="space-y-2">
                <Label htmlFor="rp-pw" className="text-sm font-medium">设置密码</Label>
                <Input
                  id="rp-pw"
                  type="password"
                  autoComplete="new-password"
                  placeholder="至少 8 位"
                  className="h-10"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                  minLength={8}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="rp-pw-confirm" className="text-sm font-medium">确认密码</Label>
                <Input
                  id="rp-pw-confirm"
                  type="password"
                  autoComplete="new-password"
                  placeholder="再次输入密码"
                  className="h-10"
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  required
                  minLength={8}
                />
              </div>
              {displayError && <p className="text-sm text-destructive">{displayError}</p>}
              <Button
                type="submit"
                className="h-10 text-sm font-medium"
                disabled={submitting || !code.trim() || password.length < 8 || !confirmPassword}
              >
                {submitting ? '验证中…' : '下一步'}
              </Button>
            </form>
          )}

          {/* === REGISTER: company name + industry + size === */}
          {step === 'register-info' && (
            <form onSubmit={handleCreateCompany} className="flex flex-col gap-5">
              <p className="text-sm text-muted-foreground">创建您的企业</p>
              <div className="space-y-2">
                <Label htmlFor="ri-company" className="text-sm font-medium">公司名称</Label>
                <Input
                  id="ri-company"
                  type="text"
                  placeholder="您的企业名称"
                  className="h-10"
                  value={companyName}
                  onChange={(e) => setCompanyName(e.target.value)}
                  required
                />
              </div>
              <div className="space-y-2">
                <Label className="text-sm font-medium">所属行业</Label>
                <Select value={industry} onValueChange={setIndustry}>
                  <SelectTrigger className="h-10">
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
                  <SelectTrigger className="h-10">
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
              {displayError && <p className="text-sm text-destructive">{displayError}</p>}
              <Button type="submit" className="h-10 text-sm font-medium" disabled={submitting || !companyName.trim()}>
                {submitting ? '创建中…' : '创建并开始体验'}
              </Button>
              <button
                type="button"
                onClick={() => setStep('register-phone')}
                className="text-sm text-foreground hover:text-foreground transition-colors"
              >
                ← 返回
              </button>
            </form>
          )}

          {/* === SELECT COMPANY === */}
          {step === 'select-company' && (
            <div className="flex flex-col gap-4">
              <p className="text-sm text-muted-foreground">选择企业</p>
              {companies.map((c) => (
                <button
                  key={c.companyId}
                  type="button"
                  disabled={submitting}
                  onClick={() => handleSelectCompany(c.companyId)}
                  className="flex items-center justify-between rounded-lg border px-4 py-3.5 text-left transition-colors hover:bg-muted"
                >
                  <div>
                    <div className="font-medium text-sm">{c.companyName}</div>
                    <div className="text-xs text-muted-foreground mt-0.5">{c.role}</div>
                  </div>
                </button>
              ))}
              {displayError && <p className="text-sm text-destructive">{displayError}</p>}
            </div>
          )}

          {/* === SELECT INVITE === */}
          {step === 'select-invite' && (
            <div className="flex flex-col gap-4">
              <p className="text-sm text-muted-foreground">您有待接受的邀请</p>
              <div className="space-y-2">
                <Label htmlFor="si-name" className="text-sm font-medium">您的姓名</Label>
                <Input
                  id="si-name"
                  placeholder="输入姓名"
                  className="h-10"
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
                    <div className="font-medium text-sm">{inv.companyName}</div>
                    <div className="text-xs text-muted-foreground mt-0.5">{inv.role}</div>
                  </div>
                  <span className="text-sm text-primary font-medium">接受</span>
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
        <Label htmlFor="popup-phone" className="text-sm font-medium">手机号</Label>
        <div className="flex gap-2">
          <span className="flex items-center rounded-md border bg-muted px-3 text-sm text-muted-foreground h-10">
            +86
          </span>
          <Input
            id="popup-phone"
            type="tel"
            inputMode="numeric"
            autoComplete="tel"
            placeholder="请输入手机号"
            className="h-10"
            value={phone}
            onChange={(e) => setPhone(e.target.value)}
            required
          />
        </div>
      </div>
      <div className="space-y-2">
        <Label htmlFor="popup-code" className="text-sm font-medium">验证码</Label>
        <div className="flex gap-2">
          <Input
            id="popup-code"
            type="text"
            inputMode="numeric"
            autoComplete="one-time-code"
            placeholder="6 位验证码"
            className="h-10"
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
            className="shrink-0 whitespace-nowrap h-10"
          >
            {sending ? '发送中…' : countdown > 0 ? `${countdown}s` : '获取验证码'}
          </Button>
        </div>
      </div>
    </>
  )
}
