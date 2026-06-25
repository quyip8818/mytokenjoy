import type { WorkflowDefinition, WorkflowId } from '../types'
import { orgWorkflowDefinitions } from './org'
import { budgetWorkflowDefinitions } from './budget'
import { keysWorkflowDefinitions } from './keys'
import { modelsWorkflowDefinitions } from './models'
import { sharedWorkflowDefinitions } from './shared'

export const WORKFLOW_REGISTRY: Record<WorkflowId, WorkflowDefinition> = {
  ...orgWorkflowDefinitions,
  ...budgetWorkflowDefinitions,
  ...keysWorkflowDefinitions,
  ...modelsWorkflowDefinitions,
  ...sharedWorkflowDefinitions,
}

export function getWorkflowDefinition(id: WorkflowId): WorkflowDefinition {
  const definition = WORKFLOW_REGISTRY[id]
  if (!definition) {
    throw new Error(`Unknown workflow: ${id}`)
  }
  return definition
}
