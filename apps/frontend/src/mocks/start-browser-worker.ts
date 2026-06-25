import { SERVICE_WORKER_SCOPE, SERVICE_WORKER_URL } from '@/config/app'

export async function startBrowserMockWorker(): Promise<void> {
  const { worker } = await import('./browser')
  await worker.start({
    onUnhandledRequest: 'bypass',
    serviceWorker: {
      url: SERVICE_WORKER_URL,
      options: {
        scope: SERVICE_WORKER_SCOPE,
      },
    },
  })
}
