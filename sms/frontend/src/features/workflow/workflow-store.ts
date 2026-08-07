import { createStore, type StoreApi } from 'zustand/vanilla'
import { WORKFLOW_MAX_DEPTH } from './constants'
import type { WorkflowId, WorkflowPayloadMap, WorkflowStackEntry } from './types'
import { WORKFLOW_META } from './definitions/workflow-meta'

export interface WorkflowStoreState {
  stack: WorkflowStackEntry[]
  dirty: boolean
  open: <T extends WorkflowId>(id: T, payload?: WorkflowPayloadMap[T], title?: string) => void
  push: <T extends WorkflowId>(id: T, payload?: WorkflowPayloadMap[T], title?: string) => void
  pop: () => void
  closeAll: () => void
  setDirty: (dirty: boolean) => void
}

function createEntry<T extends WorkflowId>(
  id: T,
  payload: WorkflowPayloadMap[T] = {} as WorkflowPayloadMap[T],
  title?: string,
): WorkflowStackEntry<T> {
  return {
    id,
    title: title ?? WORKFLOW_META[id].title,
    payload,
    dirty: false,
  }
}

export function createWorkflowStore(): StoreApi<WorkflowStoreState> {
  return createStore<WorkflowStoreState>((set, get) => ({
    stack: [],
    dirty: false,

    open: (id, payload = {} as WorkflowPayloadMap[typeof id], title) => {
      set({
        stack: [createEntry(id, payload, title)],
        dirty: false,
      })
    },

    push: (id, payload = {} as WorkflowPayloadMap[typeof id], title) => {
      const { stack } = get()
      if (stack.length >= WORKFLOW_MAX_DEPTH) return
      set({
        stack: [...stack, createEntry(id, payload, title)],
      })
    },

    pop: () => {
      const { stack } = get()
      if (stack.length <= 1) {
        set({ stack: [], dirty: false })
      } else {
        set({ stack: stack.slice(0, -1), dirty: false })
      }
    },

    closeAll: () => {
      set({ stack: [], dirty: false })
    },

    setDirty: (dirty) => {
      set({ dirty })
    },
  }))
}

export const defaultWorkflowStore = createWorkflowStore()
