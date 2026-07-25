import { useEffect } from 'react'
import { useSearchParams } from 'react-router'
import { AuthCard } from '@/features/auth'

const PARENT_ORIGIN = import.meta.env.VITE_WEB_ORIGIN || 'https://www.tokenjoy.com'

export default function AuthEmbedPage() {
  const [params] = useSearchParams()
  const mode = params.get('mode') === 'register' ? 'register' : 'login'

  // 通知父窗口 iframe 就绪
  useEffect(() => {
    if (window.parent !== window) {
      window.parent.postMessage({ type: 'auth:ready' }, PARENT_ORIGIN)
    }
  }, [])

  const handleSuccess = () => {
    if (window.parent !== window) {
      window.parent.postMessage({ type: 'auth:success' }, PARENT_ORIGIN)
    } else {
      window.location.href = '/'
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-background p-4">
      <div className="w-full max-w-[480px] rounded-xl border border-border/50 bg-background shadow-[0_10px_50px_rgba(139,92,246,0.12)] overflow-hidden">
        <AuthCard defaultMode={mode} onSuccess={handleSuccess} />
      </div>
    </div>
  )
}
