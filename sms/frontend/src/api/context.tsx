import type { ReactNode } from 'react'
import type { AppApis } from './app-apis'
import { ApiContext } from './api-context'

export function ApiProvider({ apis, children }: { apis: AppApis; children: ReactNode }) {
  return <ApiContext.Provider value={apis}>{children}</ApiContext.Provider>
}
