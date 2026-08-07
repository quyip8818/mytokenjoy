import { useMemo, type ReactNode } from 'react'
import { defaultWorkflowStore } from './workflow-store'
import { WorkflowStoreContext } from './workflow-store-context'
import type { StoreApi } from 'zustand/vanilla'
import type { WorkflowStoreState } from './workflow-store'

interface WorkflowProviderProps {
  children: ReactNode
  store?: StoreApi<WorkflowStoreState>
}

export function WorkflowProvider({ children, store }: WorkflowProviderProps) {
  const storeInstance = useMemo(() => store ?? defaultWorkflowStore, [store])
  return (
    <WorkflowStoreContext.Provider value={storeInstance}>{children}</WorkflowStoreContext.Provider>
  )
}
