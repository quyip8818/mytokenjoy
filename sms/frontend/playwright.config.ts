import { defineConfig } from '@playwright/test'

export default defineConfig({
  testDir: './e2e',
  timeout: 30_000,
  expect: { timeout: 5_000 },
  fullyParallel: false,
  retries: 0,
  use: {
    baseURL: 'http://localhost:9100',
    headless: true,
    screenshot: 'only-on-failure',
  },
  projects: [{ name: 'chromium', use: { browserName: 'chromium' } }],
  webServer: [
    {
      command: 'make seed && make run',
      cwd: '../backend',
      url: 'http://127.0.0.1:8020/api/health',
      reuseExistingServer: !process.env.CI,
      timeout: 60_000,
      env: {
        DATABASE_URL: 'postgres://tokenjoy:tokenjoy@127.0.0.1:5510/sms?sslmode=disable',
        JWT_SECRET: 'e2e-sms-jwt-secret',
        PORT: '8020',
      },
    },
    {
      command: 'pnpm dev',
      url: 'http://localhost:9100',
      reuseExistingServer: !process.env.CI,
      timeout: 30_000,
    },
  ],
})
