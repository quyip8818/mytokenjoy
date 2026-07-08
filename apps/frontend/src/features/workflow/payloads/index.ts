import type { OrgWorkflowPayloads } from './org'
import type { KeysWorkflowPayloads } from './keys'
import type { ModelsWorkflowPayloads } from './models'
import type { SharedWorkflowPayloads } from './shared'

export interface WorkflowPayloadMap
  extends
    OrgWorkflowPayloads,
    KeysWorkflowPayloads,
    ModelsWorkflowPayloads,
    SharedWorkflowPayloads {}

export type WorkflowPayload<T extends keyof WorkflowPayloadMap = keyof WorkflowPayloadMap> =
  WorkflowPayloadMap[T]
