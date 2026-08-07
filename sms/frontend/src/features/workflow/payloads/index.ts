import type { ContractsWorkflowPayloads } from './contracts'

export type WorkflowPayloadMap = ContractsWorkflowPayloads

export type WorkflowPayload<T extends keyof WorkflowPayloadMap = keyof WorkflowPayloadMap> =
  WorkflowPayloadMap[T]
