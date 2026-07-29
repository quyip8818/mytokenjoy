import { useEffect } from 'react'
import { SetupForm, useSetupPage } from '@/features/setup'

// ponytail: /setup route — only visible when setupServer is running (first boot).
// Normal app returns 404 for /api/setup/status, so useSetupPage falls to 'not-setup'
// and we redirect to login.

export default function SetupPage() {
  const { state, error, submit } = useSetupPage()

  // Redirect to login if setup is not needed (normal app is running)
  useEffect(() => {
    if (state === 'not-setup') {
      window.location.href = '/'
    }
  }, [state])

  // Auto-reload after setup completes (server restarts on new port)
  useEffect(() => {
    if (state === 'done') {
      const timer = setTimeout(() => {
        window.location.href = '/'
      }, 5000)
      return () => clearTimeout(timer)
    }
  }, [state])

  if (state === 'loading') {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background">
        <p className="text-muted-foreground">检查系统状态…</p>
      </div>
    )
  }

  if (state === 'done') {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background">
        <div className="w-full max-w-[480px] text-center space-y-4 p-8">
          <img src="/logo.png" alt="Tokenjoy" className="mx-auto h-7 w-auto" />
          <h2 className="text-xl font-semibold">初始化完成</h2>
          <p className="text-muted-foreground">系统正在重启，请稍候…</p>
          <p className="text-sm text-muted-foreground">页面将在几秒后自动刷新</p>
        </div>
      </div>
    )
  }

  if (state === 'not-setup') {
    return null // will redirect
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-background p-4">
      <div className="w-full max-w-[480px] overflow-hidden rounded-xl border border-border/50 bg-card shadow-lg">
        {/* Header */}
        <div className="px-10 pt-10 pb-5 text-center">
          <img src="/logo.png" alt="Tokenjoy" className="mx-auto h-7 w-auto" />
          <h1 className="mt-3 text-lg font-semibold">系统初始化</h1>
          <p className="mt-1 text-sm text-muted-foreground">首次启动，请完成以下配置</p>
        </div>

        {/* Form */}
        <div className="px-10 pb-10">
          <SetupForm
            submitting={state === 'submitting'}
            error={error}
            onSubmit={submit}
          />
        </div>
      </div>
    </div>
  )
}
