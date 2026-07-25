import { useSearchParams } from 'react-router'
import { AuthCard } from '@/features/auth'

export default function AuthEmbedPage() {
  const [params] = useSearchParams()
  const mode = params.get('mode') === 'register' ? 'register' : 'login'

  const handleSuccess = () => {
    if (window.parent !== window) {
      const origin = import.meta.env.VITE_WEB_ORIGIN || 'https://www.tokenjoy.com'
      window.parent.postMessage({ type: 'auth:success' }, origin)
    } else {
      window.location.href = '/'
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-white">
      <AuthCard defaultMode={mode} onSuccess={handleSuccess} />
    </div>
  )
}
