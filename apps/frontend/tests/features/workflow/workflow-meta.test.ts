import { describe, expect, it } from 'vitest'
import type { WorkflowDefinition } from '@/features/workflow/types'
import { WORKFLOW_META, getWorkflowDomain } from '@/features/workflow/definitions/workflow-meta'
import { orgWorkflowDefinitions } from '@/features/workflow/definitions/org'
import { keysWorkflowDefinitions } from '@/features/workflow/definitions/keys'
import { approvalWorkflowDefinitions } from '@/features/workflow/definitions/approval'
import { modelsWorkflowDefinitions } from '@/features/workflow/definitions/models'
import { sharedWorkflowDefinitions } from '@/features/workflow/definitions/shared'

const workflowDefinitionsByDomain: Record<string, Record<string, WorkflowDefinition>> = {
  org: orgWorkflowDefinitions,
  keys: keysWorkflowDefinitions,
  approval: approvalWorkflowDefinitions,
  models: modelsWorkflowDefinitions,
  shared: sharedWorkflowDefinitions,
}

describe('WORKFLOW_META', () => {
  it('matches domain workflow definitions', () => {
    for (const [id, meta] of Object.entries(WORKFLOW_META)) {
      const domain = getWorkflowDomain(id as keyof typeof WORKFLOW_META)
      const definition = workflowDefinitionsByDomain[domain][id]
      expect(definition, `missing definition for ${id}`).toBeDefined()
      expect(definition?.title).toBe(meta.title)
      expect(definition?.defaultLayer).toBe(meta.defaultLayer)
    }
  })
})
