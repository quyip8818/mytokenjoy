import type { WorkflowDefinition, WorkflowId } from '../types'
import { getWorkflowDomain } from './workflow-meta'

export { getWorkflowMeta, WORKFLOW_META } from './workflow-meta'

type DomainDefinitions = Partial<Record<WorkflowId, WorkflowDefinition>>

const DOMAIN_LOADERS = {
  org: () => import('./org').then((m) => m.orgWorkflowDefinitions),
  keys: () => import('./keys').then((m) => m.keysWorkflowDefinitions),
  approval: () => import('./approval').then((m) => m.approvalWorkflowDefinitions),
  models: () => import('./models').then((m) => m.modelsWorkflowDefinitions),
  shared: () => import('./shared').then((m) => m.sharedWorkflowDefinitions),
} as const

const domainCache = new Map<keyof typeof DOMAIN_LOADERS, DomainDefinitions>()
const definitionCache = new Map<WorkflowId, WorkflowDefinition>()

async function loadDomainDefinitions(
  domain: keyof typeof DOMAIN_LOADERS,
): Promise<DomainDefinitions> {
  const cached = domainCache.get(domain)
  if (cached) return cached
  const definitions = await DOMAIN_LOADERS[domain]()
  domainCache.set(domain, definitions)
  return definitions
}

export async function getWorkflowDefinition(id: WorkflowId): Promise<WorkflowDefinition> {
  const cached = definitionCache.get(id)
  if (cached) return cached

  const domain = getWorkflowDomain(id)
  const definitions = await loadDomainDefinitions(domain)
  const definition = definitions[id]
  if (!definition) {
    throw new Error(`Unknown workflow: ${id}`)
  }
  definitionCache.set(id, definition)
  return definition
}

export function getWorkflowDefinitionSync(id: WorkflowId): WorkflowDefinition | undefined {
  return definitionCache.get(id)
}
