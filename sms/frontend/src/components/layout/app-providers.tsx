import type { ReactNode } from 'react'
import { defaultApis } from '@/api/app-apis'
import { ApiProvider } from '@/api/context'
import { QueryProvider } from '@/features/query'
import { Toaster } from 'sonner'

export function AppProviders({ children }: { children: ReactNode }) {
  return (
    <ApiProvider apis={defaultApis}>
      <QueryProvider>
        {children}
        <Toaster position="top-center" richColors />
      </QueryProvider>
    </ApiProvider>
  )
}
