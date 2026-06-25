import type { ReactNode } from 'react'
import type { StoreApi } from 'zustand/vanilla'
import { DemoProvider } from './demo-provider'
import { DemoCtaHighlightProvider } from './guide/cta-highlight-provider'
import { DemoApprovalPendingCountProvider } from './nav/approval-pending-count-provider'
import type { DemoRoleStoreState } from './roles/store'
import type { DemoGuideStoreState } from './guide/store'

interface DemoShellProps {
  children: ReactNode
  roleStore?: StoreApi<DemoRoleStoreState>
  guideStore?: StoreApi<DemoGuideStoreState>
}

export function DemoShell({ children, roleStore, guideStore }: DemoShellProps) {
  return (
    <DemoProvider roleStore={roleStore} guideStore={guideStore}>
      <DemoCtaHighlightProvider>
        <DemoApprovalPendingCountProvider>{children}</DemoApprovalPendingCountProvider>
      </DemoCtaHighlightProvider>
    </DemoProvider>
  )
}
