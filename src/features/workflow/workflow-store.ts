import { createStore, type StoreApi } from 'zustand/vanilla'
import { WORKFLOW_MAX_DEPTH } from './constants'
import type { WorkflowId, WorkflowLayer, WorkflowPayload, WorkflowStackEntry } from './types'
import { getWorkflowDefinition } from './workflow-definitions'

export interface WorkflowStoreState {
  stack: WorkflowStackEntry[]
  dirty: boolean
  open: (id: WorkflowId, payload?: WorkflowPayload, title?: string) => void
  push: (id: WorkflowId, payload?: WorkflowPayload, title?: string) => void
  pop: () => void
  closeAll: () => void
  setDirty: (dirty: boolean) => void
}

function resolveLayer(id: WorkflowId, explicitLayer?: WorkflowLayer): WorkflowLayer {
  if (explicitLayer) return explicitLayer
  return getWorkflowDefinition(id).defaultLayer
}

function createEntry(
  id: WorkflowId,
  payload: WorkflowPayload = {},
  title?: string,
  layer?: WorkflowLayer,
): WorkflowStackEntry {
  const def = getWorkflowDefinition(id)
  return {
    id,
    layer: resolveLayer(id, layer),
    title: title ?? def.title,
    payload,
    dirty: false,
  }
}

export function createWorkflowStore(): StoreApi<WorkflowStoreState> {
  return createStore<WorkflowStoreState>((set, get) => ({
    stack: [],
    dirty: false,

    open: (id, payload = {}, title) => {
      set({
        stack: [createEntry(id, payload, title)],
        dirty: false,
      })
    },

    push: (id, payload = {}, title) => {
      const { stack } = get()
      if (stack.length >= WORKFLOW_MAX_DEPTH) return
      const parentLayer = stack[stack.length - 1]?.layer ?? 1
      const nextLayer = Math.min(parentLayer + 1, WORKFLOW_MAX_DEPTH) as WorkflowLayer
      set({
        stack: [...stack, createEntry(id, payload, title, nextLayer)],
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
