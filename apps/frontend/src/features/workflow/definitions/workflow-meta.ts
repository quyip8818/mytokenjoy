import type { WorkflowId } from '../types'

export type WorkflowMeta = { title: string }

export const WORKFLOW_META: Record<WorkflowId, WorkflowMeta> = {
  'member-search': { title: '搜索成员' },
  'key-create': { title: '创建 Key' },
  'key-edit': { title: '编辑 Key' },
  'key-rotate-confirm': { title: '重新生成 Key' },
  'key-reveal': { title: 'Key 已生成' },
  'approval-submit': { title: '发起申请' },
  'approval-review': { title: '审批处理' },
  'provider-key-form': { title: '添加供应商 Key' },
  'model-create': { title: '添加自定义模型' },
  'model-edit': { title: '编辑自定义模型' },
  'whitelist-config': { title: '配置部门白名单' },
  'model-picker': { title: '选择模型' },
  'budget-check': { title: '额度不足' },
  'reject-reason': { title: '拒绝理由' },
  'platform-model-create': { title: '添加平台模型' },
  'platform-model-edit': { title: '编辑平台模型' },
  'discount-config': { title: '模型优惠配置' },
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
  'platform-model-create': 'models',
  'platform-model-edit': 'models',
  'discount-config': 'models',
}

export function getWorkflowMeta(id: WorkflowId): WorkflowMeta {
  const meta = WORKFLOW_META[id]
  if (!meta) {
    throw new Error(`Unknown workflow: ${id}`)
  }
  return meta
}

export function getWorkflowDomain(
  id: WorkflowId,
): 'org' | 'keys' | 'models' | 'shared' | 'approval' {
  return WORKFLOW_DOMAIN[id]
}
