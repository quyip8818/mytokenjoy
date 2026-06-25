import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { USE_MOCKS } from '@/config/app'
import './index.css'
import App from './App.tsx'

async function bootstrap() {
  if (USE_MOCKS) {
    const { worker } = await import('./mocks/browser')
    await worker.start({
      onUnhandledRequest: 'bypass',
      serviceWorker: {
        url: `${import.meta.env.BASE_URL}mockServiceWorker.js`,
      },
    })
  }

  createRoot(document.getElementById('root')!).render(
    <StrictMode>
      <App />
    </StrictMode>,
  )
}

bootstrap()
