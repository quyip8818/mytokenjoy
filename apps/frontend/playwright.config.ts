import { defineConfig, devices } from '@playwright/test'

const previewPort = 4173
const previewHost = '127.0.0.1'
const previewUrl = `http://${previewHost}:${previewPort}`

export default defineConfig({
  testDir: './e2e',
  globalSetup: './e2e/global-setup.ts',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  workers: process.env.CI ? 2 : undefined,
  reporter: process.env.CI ? [['html', { open: 'never' }], ['list']] : 'list',
  use: {
    baseURL: previewUrl,
    storageState: '.auth/admin.json',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
  webServer: [
    {
      command: 'make run',
      cwd: '../backend',
      url: 'http://127.0.0.1:8080/healthz',
      reuseExistingServer: !process.env.CI,
      timeout: 120_000,
      gracefulShutdown: { signal: 'SIGTERM', timeout: 10_000 },
      env: {
        DATABASE_URL: 'postgres://tokenjoy:tokenjoy@127.0.0.1:5432/tokenjoy?sslmode=disable',
        COMPANY_NAME: 'Demo Company',
        SESSION_SECRET: 'e2e-test-session-secret',
        DATA_SOURCE_CREDENTIAL_KEY: 'dGV2LWNyZWRlbnRpYWwta2V5LWZvci1sb2NhbC1kZXY=',
        BOOTSTRAP_MODE: 'demo',
        CLOCK_ANCHOR: '2026-06-19',
        DEPLOY_ENV: 'local',
      },
    },
    {
      command: 'pnpm build && pnpm exec vite preview --port 4173 --strictPort --host 127.0.0.1',
      url: previewUrl,
      reuseExistingServer: !process.env.CI,
      timeout: 180_000,
    },
  ],
})
