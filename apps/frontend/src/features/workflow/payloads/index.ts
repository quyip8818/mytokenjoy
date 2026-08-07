import type { OrgWorkflowPayloads } from './org'
import type { KeysWorkflowPayloads } from './keys'
import type { ModelsWorkflowPayloads } from './models'
import type { SharedWorkflowPayloads } from './shared'
import type { BillingWorkflowPayloads } from './billing'

export interface WorkflowPayloadMap
  extends
    OrgWorkflowPayloads,
    KeysWorkflowPayloads,
    ModelsWorkflowPayloads,
    SharedWorkflowPayloads,
    BillingWorkflowPayloads {}

export type WorkflowPayload<T extends keyof WorkflowPayloadMap = keyof WorkflowPayloadMap> =
  WorkflowPayloadMap[T]
