import type { ComponentType } from 'react'

export type WorkflowId =
  | 'credential-form'
  | 'sync-config'
  | 'member-form'
  | 'member-invite'
  | 'member-import'
  | 'dept-form'
  | 'budget-node-edit'
  | 'budget-group-form'
  | 'overrun-policy'
  | 'model-create'
  | 'whitelist-config'
  | 'key-create'
  | 'key-edit'
  | 'key-rotate-confirm'
  | 'key-reveal'
  | 'approval-submit'
  | 'approval-review'
  | 'role-form'
  | 'role-add-member'
  | 'provider-key-form'
  | 'pick-dept'
  | 'model-picker'
  | 'import-preview'
  | 'quota-check'
  | 'reject-reason'
  | 'budget-impact-preview'
  | 'permission-picker'
  | 'member-search'
  | 'pick-members'

export type WorkflowLayer = 1 | 2 | 3

export type WorkflowPayload = Record<string, unknown>

export interface WorkflowStackEntry {
  id: WorkflowId
  layer: WorkflowLayer
  title: string
  payload: WorkflowPayload
  dirty?: boolean
}

export interface WorkflowComponentProps {
  entry: WorkflowStackEntry
  onClose: () => void
  onPop: () => void
  onPush: (id: WorkflowId, payload?: WorkflowPayload, title?: string) => void
  onSetDirty: (dirty: boolean) => void
}

export interface WorkflowDefinition {
  component: ComponentType<WorkflowComponentProps>
  defaultLayer: WorkflowLayer
  title: string
}
