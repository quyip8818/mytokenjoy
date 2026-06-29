import type { ReactNode } from 'react'
import { ErrorBoundary, type FallbackProps } from 'react-error-boundary'
import { captureException } from '@/config/monitoring'
import { ErrorState } from '@/components/ui/error-state'

function ErrorFallback({ error, resetErrorBoundary }: FallbackProps) {
  const message = error instanceof Error ? error.message : 'An unexpected error occurred'

  return (
    <div className="flex min-h-[12rem] items-center justify-center p-8">
      <ErrorState
        title="页面出错"
        message={message}
        onRetry={resetErrorBoundary}
        retryLabel="重试"
      />
    </div>
  )
}

interface AppErrorBoundaryProps {
  children: ReactNode
}

export function AppErrorBoundary({ children }: AppErrorBoundaryProps) {
  return (
    <ErrorBoundary
      FallbackComponent={ErrorFallback}
      onError={(error, info) => {
        captureException(error)
        if (import.meta.env.DEV) {
          console.error('AppErrorBoundary caught:', error, info.componentStack)
        }
      }}
    >
      {children}
    </ErrorBoundary>
  )
}
