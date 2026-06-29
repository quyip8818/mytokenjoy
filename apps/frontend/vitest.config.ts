import { mergeConfig } from 'vite'
import { defineConfig } from 'vitest/config'
import { baseViteConfig } from './vite.config'

export default mergeConfig(
  baseViteConfig,
  defineConfig({
    test: {
      globals: true,
      environment: 'jsdom',
      include: ['tests/**/*.{test,spec}.{ts,tsx}'],
      setupFiles: ['./tests/setup.ts'],
    },
  }),
)
