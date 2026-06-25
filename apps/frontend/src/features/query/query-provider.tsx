import { useMemo, type ReactNode } from 'react'
import { QueryClientProvider } from '@tanstack/react-query'
import { createAppQueryClient } from './query-client'

interface QueryProviderProps {
  children: ReactNode
  client?: ReturnType<typeof createAppQueryClient>
}

export function QueryProvider({ children, client }: QueryProviderProps) {
  const queryClient = useMemo(() => client ?? createAppQueryClient(), [client])

  return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
}
