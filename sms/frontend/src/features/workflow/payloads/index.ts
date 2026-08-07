import type { ContractsWorkflowPayloads } from './contracts'

export interface WorkflowPayloadMap extends ContractsWorkflowPayloads {}

export type WorkflowPayload<T extends keyof WorkflowPayloadMap = keyof WorkflowPayloadMap> =
  WorkflowPayloadMap[T]
