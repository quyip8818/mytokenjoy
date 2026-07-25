import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import { AuthCard } from '@/features/auth/components/auth-card'

const params = new URLSearchParams(window.location.search)
const mode = params.get('mode') === 'register' ? 'register' : 'login'

function handleSuccess() {
  const origin = import.meta.env.VITE_WEB_ORIGIN || 'https://www.tokenjoy.com'
  if (window.parent !== window) {
    window.parent.postMessage({ type: 'auth:success' }, origin)
  } else {
    window.location.href = '/'
  }
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <div className="flex min-h-screen items-center justify-center bg-white">
      <AuthCard defaultMode={mode} onSuccess={handleSuccess} />
    </div>
  </StrictMode>,
)
