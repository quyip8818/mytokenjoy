import type { CredentialFormPayload, SyncConfigPayload } from './org-delegate'
import type { OrgWorkflowPayloads, MemberFormData } from './org'
import type { BudgetWorkflowPayloads } from './budget'
import type { KeysWorkflowPayloads } from './keys'
import type { ModelsWorkflowPayloads } from './models'
import type { SharedWorkflowPayloads } from './shared'

export type { MemberFormData, CredentialFormPayload, SyncConfigPayload }

export interface WorkflowPayloadMap
  extends
    OrgWorkflowPayloads,
    BudgetWorkflowPayloads,
    KeysWorkflowPayloads,
    ModelsWorkflowPayloads,
    SharedWorkflowPayloads {
  'credential-form': CredentialFormPayload
  'sync-config': SyncConfigPayload
}

export type WorkflowPayload<T extends keyof WorkflowPayloadMap = keyof WorkflowPayloadMap> =
  WorkflowPayloadMap[T]
