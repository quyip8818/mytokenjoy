import { useContext } from 'react'
import { useStore } from 'zustand'
import { WorkflowStoreContext } from './workflow-store-context'
import type { WorkflowStoreState } from './workflow-store'
import type { StoreApi } from 'zustand/vanilla'

function useWorkflowStoreApi(): StoreApi<WorkflowStoreState> {
  const store = useContext(WorkflowStoreContext)
  if (!store) {
    throw new Error('useWorkflow must be used within WorkflowProvider')
  }
  return store
}

export function useWorkflowStore<T>(selector: (state: WorkflowStoreState) => T): T {
  const store = useWorkflowStoreApi()
  return useStore(store, selector)
}

export function useWorkflow() {
  const store = useWorkflowStoreApi()
  const stack = useStore(store, (s) => s.stack)
  const dirty = useStore(store, (s) => s.dirty)
  const open = useStore(store, (s) => s.open)
  const push = useStore(store, (s) => s.push)
  const pop = useStore(store, (s) => s.pop)
  const closeAll = useStore(store, (s) => s.closeAll)
  const setDirty = useStore(store, (s) => s.setDirty)

  return { stack, dirty, open, push, pop, closeAll, setDirty, store }
}
