import { lazy, Suspense, useMemo, type ReactNode } from 'react'
import { QueryClientProvider } from '@tanstack/react-query'
import { createAppQueryClient } from './query-client'

const ReactQueryDevtools = import.meta.env.DEV
  ? lazy(() =>
      import('@tanstack/react-query-devtools').then((mod) => ({
        default: mod.ReactQueryDevtools,
      })),
    )
  : null

interface QueryProviderProps {
  children: ReactNode
  client?: ReturnType<typeof createAppQueryClient>
}

export function QueryProvider({ children, client }: QueryProviderProps) {
  const queryClient = useMemo(() => client ?? createAppQueryClient(), [client])

  return (
    <QueryClientProvider client={queryClient}>
      {children}
      {ReactQueryDevtools ? (
        <Suspense fallback={null}>
          <ReactQueryDevtools initialIsOpen={false} buttonPosition="bottom-left" />
        </Suspense>
      ) : null}
    </QueryClientProvider>
  )
}
