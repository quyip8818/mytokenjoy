import path from 'path'
import { copyFileSync } from 'node:fs'
import type { Plugin, UserConfig } from 'vite'
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import { createApiProxyConfig, warnIfBackendUnreachable } from './vite-api-proxy'

const repoName = process.env.GITHUB_REPOSITORY?.split('/')[1]
const githubPagesBase = process.env.VITE_BASE_PATH ?? (repoName ? `/${repoName}/` : undefined)
const appBase = githubPagesBase ?? '/'
const apiProxyTarget = process.env.VITE_API_PROXY_TARGET ?? 'http://127.0.0.1:8080'
const apiProxy = createApiProxyConfig(appBase, apiProxyTarget)

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
  base: appBase,
  plugins: [
    react(),
    tailwindcss(),
    apiBackendHealthPlugin(),
    {
      name: 'gh-pages-spa-fallback',
      closeBundle() {
        if (!githubPagesBase) return
        const distDir = path.resolve(__dirname, 'dist')
        copyFileSync(path.join(distDir, 'index.html'), path.join(distDir, '404.html'))
      },
    },
  ],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
      '@tests': path.resolve(__dirname, './tests'),
    },
  },
  server: {
    proxy: apiProxy,
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
