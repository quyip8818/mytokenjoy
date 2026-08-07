import type { WorkflowId } from '../types'

export type WorkflowMeta = { title: string }

export const WORKFLOW_META: Record<WorkflowId, WorkflowMeta> = {
  'contract-detail': { title: '合同详情' },
}
