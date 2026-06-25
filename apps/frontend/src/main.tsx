import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { USE_MOCKS } from '@/config/app'
import './index.css'
import App from './App.tsx'

function renderMockWorkerError(error: unknown): void {
  const message = error instanceof Error ? error.message : String(error)
  createRoot(document.getElementById('root')!).render(
    <StrictMode>
      <div className="flex min-h-screen items-center justify-center p-8">
        <div className="max-w-lg space-y-3 text-center">
          <h1 className="text-lg font-semibold">Demo Mock API 启动失败</h1>
          <p className="text-sm text-muted-foreground">
            请硬刷新页面，或在浏览器设置中清除本站的 Service Worker 后重试。
          </p>
          <p className="break-all text-xs text-muted-foreground">{message}</p>
        </div>
      </div>
    </StrictMode>,
  )
}

async function bootstrap() {
  if (USE_MOCKS) {
    try {
      const { startBrowserMockWorker } = await import('./mocks/start-browser-worker')
      await startBrowserMockWorker()
    } catch (error) {
      console.error('[MSW] Failed to start mock worker', error)
      renderMockWorkerError(error)
      return
    }
  }

  createRoot(document.getElementById('root')!).render(
    <StrictMode>
      <App />
    </StrictMode>,
  )
}

bootstrap()
