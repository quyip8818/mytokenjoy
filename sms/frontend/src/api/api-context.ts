import { createContext } from 'react'
import type { AppApis } from './app-apis'

export const ApiContext = createContext<AppApis | null>(null)
