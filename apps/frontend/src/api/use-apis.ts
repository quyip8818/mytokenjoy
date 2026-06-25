import { useContext } from 'react'
import { ApiContext } from './api-context'
import type { AppApis } from './app-apis'

export function useApis(): AppApis {
  const ctx = useContext(ApiContext)
  if (!ctx) {
    throw new Error('useApis must be used within ApiProvider')
  }
  return ctx
}
