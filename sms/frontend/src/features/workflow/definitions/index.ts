import type { WorkflowDefinition, WorkflowId } from '../types'

export { WORKFLOW_META } from './workflow-meta'

const definitionCache = new Map<WorkflowId, WorkflowDefinition>()

export async function getWorkflowDefinition(id: WorkflowId): Promise<WorkflowDefinition> {
  const cached = definitionCache.get(id)
  if (cached) return cached

  // ponytail: SMS 只有一个 domain，直接 lazy load 即可。
  // 升级路径：如需多 domain，参考 apps 的 DOMAIN_LOADERS 模式拆分。
  const { contractsWorkflowDefinitions } = await import('./contracts')
  const definition = (contractsWorkflowDefinitions as Record<string, WorkflowDefinition>)[id]
  if (!definition) {
    throw new Error(`Unknown workflow: ${id}`)
  }
  definitionCache.set(id, definition)
  return definition
}
