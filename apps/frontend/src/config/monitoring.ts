import * as Sentry from '@sentry/react'

export function initMonitoring(): void {
  const dsn = import.meta.env.VITE_SENTRY_DSN
  if (!dsn) {
    return
  }

  Sentry.init({
    dsn,
    environment: import.meta.env.MODE,
    integrations: [Sentry.browserTracingIntegration()],
    tracesSampleRate: 0.1,
  })
}

export function captureException(error: unknown): void {
  if (!import.meta.env.VITE_SENTRY_DSN) {
    return
  }
  Sentry.captureException(error)
}
