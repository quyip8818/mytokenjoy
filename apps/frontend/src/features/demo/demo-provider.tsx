import type { ReactNode } from 'react'
import type { StoreApi } from 'zustand/vanilla'
import { DemoSessionProvider } from '@/features/session'
import { DemoGuideProvider } from './guide/provider'
import { DemoRoleNavigationBridge } from './roles/navigation-bridge'
import { DesktopOnlyHint } from './chrome/desktop-only-hint'
import type { DemoRoleStoreState } from './roles/store'
import type { DemoGuideStoreState } from './guide/store'

interface DemoProviderProps {
  children: ReactNode
  roleStore?: StoreApi<DemoRoleStoreState>
  guideStore?: StoreApi<DemoGuideStoreState>
}

export function DemoProvider({ children, roleStore, guideStore }: DemoProviderProps) {
  return (
    <DemoSessionProvider store={roleStore}>
      <DemoGuideProvider store={guideStore}>
        <DemoRoleNavigationBridge />
        {children}
        <DesktopOnlyHint />
      </DemoGuideProvider>
    </DemoSessionProvider>
  )
}
