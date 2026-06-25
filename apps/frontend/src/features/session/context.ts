import { createContext } from 'react'
import type { AppSession } from './types'

export const SessionReactContext = createContext<AppSession | null>(null)
