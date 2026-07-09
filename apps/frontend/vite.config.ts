import path from 'path'
import type { Plugin, UserConfig } from 'vite'
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import { createApiProxyConfig, warnIfBackendUnreachable } from './vite-api-proxy'

const apiProxyTarget = process.env.VITE_API_PROXY_TARGET ?? 'http://127.0.0.1:8080'
const apiProxy = createApiProxyConfig('/', apiProxyTarget)

function apiBackendHealthPlugin(): Plugin {
  return {
    name: 'api-backend-health',
    configureServer() {
      warnIfBackendUnreachable(apiProxyTarget)
    },
  }
}

function manualChunks(id: string): string | undefined {
  if (id.includes('recharts')) return 'recharts'
  if (id.includes('@tanstack/react-table')) return 'react-table'
  if (id.includes('@tanstack/react-query')) return 'react-query'
  if (id.includes('node_modules/react-dom') || id.includes('node_modules/react/'))
    return 'react-vendor'
  return undefined
}

export const baseViteConfig: UserConfig = {
  base: '/',
  plugins: [react(), tailwindcss(), apiBackendHealthPlugin()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
      '@tests': path.resolve(__dirname, './tests'),
    },
  },
  server: {
    proxy: apiProxy,
    watch: {
      usePolling: true,
      interval: 300,
    },
  },
  preview: {
    proxy: apiProxy,
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks,
      },
    },
  },
}

export default defineConfig(baseViteConfig)
