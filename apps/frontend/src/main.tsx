import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { RouterProvider } from '@tanstack/react-router'
import { initMonitoring } from '@/config/monitoring'
import { IS_SAAS } from '@/config/app'
import { router } from '@/router'
import './index.css'

initMonitoring()

// ponytail: local 模式启动时检测后端是否在 setup 阶段。
// 如果是，自动跳 /setup 让用户完成初始化。升级路径：改 service worker 或 SSR check。
async function boot() {
  if (!IS_SAAS && !window.location.pathname.startsWith('/setup')) {
    try {
      const res = await fetch('/api/setup/status')
      if (res.ok) {
        const data = await res.json()
        if (data.ready) {
          window.location.replace('/setup')
          return
        }
      }
    } catch {
      // Setup server not running (normal mode) — proceed
    }
  }

  createRoot(document.getElementById('root')!).render(
    <StrictMode>
      <RouterProvider router={router} />
    </StrictMode>,
  )
}

boot()
