import path from 'path'
import { copyFileSync } from 'node:fs'
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

const repoName = process.env.GITHUB_REPOSITORY?.split('/')[1]
const githubPagesBase = process.env.VITE_BASE_PATH ?? (repoName ? `/${repoName}/` : undefined)

// https://vite.dev/config/
export default defineConfig({
  base: githubPagesBase ?? '/',
  plugins: [
    react(),
    tailwindcss(),
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
    },
  },
})
