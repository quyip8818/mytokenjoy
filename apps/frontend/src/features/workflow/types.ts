import type { ComponentType } from 'react'
import type { WorkflowPayloadMap } from './workflow-payloads'

export type WorkflowId = keyof WorkflowPayloadMap

export type WorkflowLayer = 1 | 2 | 3

export type { WorkflowPayloadMap, WorkflowPayload } from './workflow-payloads'

export interface WorkflowStackEntry<T extends WorkflowId = WorkflowId> {
  id: T
  layer: WorkflowLayer
  title: string
  payload: WorkflowPayloadMap[T]
  dirty?: boolean
}

export interface WorkflowPanelProps {
  entry: WorkflowStackEntry
  onClose: () => void
  onPop: () => void
  onPush: <U extends WorkflowId>(id: U, payload?: WorkflowPayloadMap[U], title?: string) => void
  onSetDirty: (dirty: boolean) => void
}

export interface WorkflowComponentProps<
  T extends WorkflowId = WorkflowId,
> extends WorkflowPanelProps {
  entry: WorkflowStackEntry<T>
}

export interface WorkflowDefinition {
  component: ComponentType<WorkflowPanelProps>
  defaultLayer: WorkflowLayer
  title: string
}

export function defineWorkflow<T extends WorkflowId>(
  component: ComponentType<WorkflowComponentProps<T>>,
  definition: Omit<WorkflowDefinition, 'component'>,
): WorkflowDefinition {
  return {
    ...definition,
    component: component as ComponentType<WorkflowPanelProps>,
  }
}
