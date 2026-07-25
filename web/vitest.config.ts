import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { mergeConfig } from 'vite'
import { defineConfig } from 'vitest/config'
import { baseViteConfig } from './vite.config'

const rootDir = path.dirname(fileURLToPath(import.meta.url))

export default mergeConfig(
  baseViteConfig,
  defineConfig({
    resolve: {
      alias: {
        '@': path.resolve(rootDir, './src'),
      },
    },
    test: {
      globals: true,
      environment: 'jsdom',
      include: ['tests/**/*.{test,spec}.{ts,tsx}'],
      setupFiles: ['./tests/setup.ts'],
    },
  }),
)
