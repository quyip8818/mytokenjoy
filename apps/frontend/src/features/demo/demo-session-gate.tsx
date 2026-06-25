import type { ReactNode } from 'react'
import { SessionGate } from '@/features/session'

export function DemoSessionGate({ children }: { children: ReactNode }) {
  return <SessionGate>{children}</SessionGate>
}
