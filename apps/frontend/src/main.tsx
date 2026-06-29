import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { initMonitoring } from '@/config/monitoring'
import './index.css'
import App from './App.tsx'

initMonitoring()

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
