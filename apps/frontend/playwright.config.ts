import { defineConfig, devices } from '@playwright/test'
import {
  createDatabase,
  E2E_BACKEND_PORT,
  E2E_BASE_URL,
  E2E_DATABASE_URL,
  E2E_HOST,
  E2E_PREVIEW_PORT,
  E2E_SMS_PORT,
} from './e2e/e2e-db'

// ponytail: 必须在 webServer 启动前建库，config 加载时即执行
createDatabase()

export default defineConfig({
  testDir: './e2e',
  globalSetup: './e2e/global-setup.ts',
  globalTeardown: './e2e/global-teardown.ts',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  workers: process.env.CI ? 2 : undefined,
  reporter: process.env.CI ? [['html', { open: 'never' }], ['list']] : 'list',
  use: {
    baseURL: E2E_BASE_URL,
    storageState: '.auth/admin.json',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },
  projects: [
    {
      name: 'read-only',
      testMatch:
        /^(?!.*(org-structure|member-delete-no-accumulate|budget-org-member-picker|org-roles)).*\.spec\.ts$/,
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'mutations',
      testMatch:
        /(org-structure|member-delete-no-accumulate|budget-org-member-picker|org-roles)\.spec\.ts$/,
      fullyParallel: false,
      use: { ...devices['Desktop Chrome'] },
    },
  ],
  webServer: [
    {
      command: 'make run',
      cwd: '../backend',
      url: `http://${E2E_HOST}:${E2E_BACKEND_PORT}/healthz`,
      reuseExistingServer: !process.env.CI,
      timeout: 120_000,
      gracefulShutdown: { signal: 'SIGTERM', timeout: 10_000 },
      env: {
        DATABASE_URL: E2E_DATABASE_URL,
        COMPANY_NAME: 'Demo Company',
        SESSION_SECRET: 'e2e-test-session-secret',
        DATA_SOURCE_CREDENTIAL_KEY: 'dGV2LWNyZWRlbnRpYWwta2V5LWZvci1sb2NhbC1kZXY=',
        BOOTSTRAP_MODE: 'demo',
        CLOCK_ANCHOR: '2026-06-19',
        DEPLOY_ENV: 'local',
        NEW_API_BASE_URL: 'http://127.0.0.1:3010',
        PORT: String(E2E_BACKEND_PORT),
      },
    },
    {
      command: 'make seed && make run',
      cwd: '../../sms/backend',
      url: `http://${E2E_HOST}:${E2E_SMS_PORT}/api/health`,
      reuseExistingServer: !process.env.CI,
      timeout: 60_000,
      gracefulShutdown: { signal: 'SIGTERM', timeout: 5_000 },
      env: {
        DATABASE_URL: 'postgres://tokenjoy:tokenjoy@127.0.0.1:5510/sms?sslmode=disable',
        JWT_SECRET: 'e2e-sms-jwt-secret',
        PORT: String(E2E_SMS_PORT),
      },
    },
    {
      command: `pnpm build && pnpm exec vite preview --port ${E2E_PREVIEW_PORT} --strictPort --host ${E2E_HOST}`,
      url: E2E_BASE_URL,
      reuseExistingServer: !process.env.CI,
      timeout: 180_000,
      env: {
        VITE_API_PROXY_TARGET: `http://${E2E_HOST}:${E2E_BACKEND_PORT}`,
      },
    },
  ],
})
