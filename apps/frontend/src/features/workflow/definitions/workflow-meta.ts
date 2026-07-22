import type { WorkflowId, WorkflowLayer } from '../types'

export type WorkflowMeta = { defaultLayer: WorkflowLayer; title: string }

export const WORKFLOW_META: Record<WorkflowId, WorkflowMeta> = {
  'member-search': { defaultLayer: 2, title: '搜索成员' },
  'key-create': { defaultLayer: 1, title: '创建 Key' },
  'key-edit': { defaultLayer: 1, title: '编辑 Key' },
  'key-rotate-confirm': { defaultLayer: 2, title: '重新生成 Key' },
  'key-reveal': { defaultLayer: 3, title: 'Key 已生成' },
  'approval-submit': { defaultLayer: 1, title: '发起申请' },
  'approval-review': { defaultLayer: 1, title: '审批处理' },
  'provider-key-form': { defaultLayer: 1, title: '添加供应商 Key' },
  'model-create': { defaultLayer: 1, title: '添加自定义模型' },
  'model-edit': { defaultLayer: 1, title: '编辑自定义模型' },
  'whitelist-config': { defaultLayer: 1, title: '配置部门白名单' },
  'model-picker': { defaultLayer: 2, title: '选择模型' },
  'budget-check': { defaultLayer: 3, title: '额度不足' },
  'reject-reason': { defaultLayer: 2, title: '拒绝理由' },
}

const WORKFLOW_DOMAIN: Record<WorkflowId, 'org' | 'keys' | 'models' | 'shared' | 'approval'> = {
  'member-search': 'org',
  'key-create': 'keys',
  'key-edit': 'keys',
  'key-rotate-confirm': 'keys',
  'key-reveal': 'keys',
  'approval-submit': 'approval',
  'approval-review': 'approval',
  'provider-key-form': 'keys',
  'model-create': 'models',
  'model-edit': 'models',
  'whitelist-config': 'models',
  'model-picker': 'shared',
  'budget-check': 'shared',
  'reject-reason': 'shared',
}

export function getWorkflowMeta(id: WorkflowId): WorkflowMeta {
  const meta = WORKFLOW_META[id]
  if (!meta) {
    throw new Error(`Unknown workflow: ${id}`)
  }
  return meta
}

export function getWorkflowDomain(id: WorkflowId): 'org' | 'keys' | 'models' | 'shared' | 'approval' {
  return WORKFLOW_DOMAIN[id]
}
