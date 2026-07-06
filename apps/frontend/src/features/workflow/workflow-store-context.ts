import { createContext } from 'react'
import type { WorkflowStoreState } from './workflow-store'
import type { StoreApi } from 'zustand/vanilla'

export const WorkflowStoreContext = createContext<StoreApi<WorkflowStoreState> | null>(null)
