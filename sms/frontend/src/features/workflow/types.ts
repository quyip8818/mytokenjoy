import type { ComponentType } from 'react'
import type { WorkflowPayloadMap } from './payloads'

export type WorkflowId = keyof WorkflowPayloadMap

// ponytail: layer 纯展示属性，由 stack index 自动推导，不存储在 entry 中。
// 升级路径：如需 4+ 层，改 WORKFLOW_MAX_DEPTH 和 LAYER_STYLES 即可。
export type WorkflowLayer = 1 | 2 | 3

export type { WorkflowPayloadMap, WorkflowPayload } from './payloads'

export interface WorkflowStackEntry<T extends WorkflowId = WorkflowId> {
  id: T
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
  title: string
}

export function defineWorkflow<T extends WorkflowId>(
  component: ComponentType<WorkflowComponentProps<T>>,
  definition: { title: string },
): WorkflowDefinition {
  return {
    title: definition.title,
    component: component as ComponentType<WorkflowPanelProps>,
  }
}
