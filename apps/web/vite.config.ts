import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tsconfigPaths from 'vite-tsconfig-paths'

const rootDir = path.dirname(fileURLToPath(import.meta.url))

export const baseViteConfig = defineConfig({
  build: {
    sourcemap: 'hidden',
    modulePreload: false,
  },
  plugins: [react(), tsconfigPaths()],
  resolve: {
    alias: {
      '@': path.resolve(rootDir, './src'),
    },
  },
})

export default baseViteConfig
