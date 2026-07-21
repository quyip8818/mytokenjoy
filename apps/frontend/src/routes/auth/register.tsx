import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useRegisterPage } from '@/features/auth/hooks/use-register-page'

export default function RegisterPage() {
  const {
    step,
    phone,
    setPhone,
    code,
    setCode,
    companyName,
    setCompanyName,
    password,
    setPassword,
    error,
    sending,
    verifying,
    creating,
    countdown,
    canSend,
    canSubmitInfo,
    handleSendCode,
    handleVerifyAndInit,
    handleCreateCompany,
  } = useRegisterPage()

  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-4 p-8">
      {step === 'phone' ? (
        <form
          onSubmit={handleVerifyAndInit}
          className="flex w-full max-w-md flex-col gap-4"
        >
          <div className="space-y-2 text-center">
            <h1 className="text-lg font-semibold">注册 TokenJoy</h1>
            <p className="text-sm text-muted-foreground">验证手机号后创建企业</p>
          </div>

          <div className="space-y-2">
            <Label htmlFor="reg-phone">手机号</Label>
            <div className="flex gap-2">
              <span className="flex items-center rounded-md border bg-muted px-3 text-sm text-muted-foreground">
                +86
              </span>
              <Input
                id="reg-phone"
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
            <Label htmlFor="reg-code">验证码</Label>
            <div className="flex gap-2">
              <Input
                id="reg-code"
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
            {verifying ? '验证中…' : '下一步'}
          </Button>

          <p className="text-center text-sm text-muted-foreground">
            已有账号？
            <a href="/login" className="text-primary underline-offset-4 hover:underline">
              去登录
            </a>
          </p>
        </form>
      ) : (
        <form
          onSubmit={handleCreateCompany}
          className="flex w-full max-w-md flex-col gap-4"
        >
          <div className="space-y-2 text-center">
            <h1 className="text-lg font-semibold">创建企业</h1>
            <p className="text-sm text-muted-foreground">
              设置登录密码并创建您的企业
            </p>
          </div>

          <div className="space-y-2">
            <Label htmlFor="reg-password">设置密码</Label>
            <Input
              id="reg-password"
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
            <Label htmlFor="reg-company">公司名称</Label>
            <Input
              id="reg-company"
              type="text"
              placeholder="您的企业名称"
              value={companyName}
              onChange={(e) => setCompanyName(e.target.value)}
              required
            />
          </div>

          {error ? <p className="text-sm text-destructive">{error}</p> : null}

          <Button type="submit" disabled={creating || !canSubmitInfo}>
            {creating ? '创建中…' : '创建并开始体验'}
          </Button>
        </form>
      )}
    </div>
  )
}
