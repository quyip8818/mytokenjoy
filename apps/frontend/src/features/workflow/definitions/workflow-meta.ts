import type { WorkflowId, WorkflowLayer } from '../types'

export type WorkflowMeta = { defaultLayer: WorkflowLayer; title: string }

export const WORKFLOW_META: Record<WorkflowId, WorkflowMeta> = {
  'budget-node-edit': { defaultLayer: 1, title: '编辑节点预算' },
  'budget-group-form': { defaultLayer: 1, title: '新建预算组' },
  'overrun-policy': { defaultLayer: 1, title: '全局超限策略' },
  'budget-impact-preview': { defaultLayer: 2, title: '影响范围预览' },
  'member-quota-config': { defaultLayer: 1, title: '成员额度配置' },
  'credential-form': { defaultLayer: 1, title: '配置凭证' },
  'sync-config': { defaultLayer: 1, title: '同步策略' },
  'member-form': { defaultLayer: 1, title: '添加/编辑成员' },
  'member-invite': { defaultLayer: 1, title: '邀请成员' },
  'member-import': { defaultLayer: 1, title: '批量导入' },
  'dept-form': { defaultLayer: 1, title: '添加/编辑部门' },
  'role-form': { defaultLayer: 1, title: '创建/编辑角色' },
  'role-add-member': { defaultLayer: 1, title: '添加角色成员' },
  'import-preview': { defaultLayer: 2, title: '导入预览' },
  'pick-dept': { defaultLayer: 2, title: '选择部门' },
  'member-search': { defaultLayer: 2, title: '搜索成员' },
  'pick-members': { defaultLayer: 2, title: '选择成员' },
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
  'quota-check': { defaultLayer: 3, title: '额度不足' },
  'reject-reason': { defaultLayer: 2, title: '拒绝理由' },
  'permission-picker': { defaultLayer: 2, title: '选择权限' },
}

const WORKFLOW_DOMAIN: Record<WorkflowId, 'budget' | 'org' | 'keys' | 'models' | 'shared'> = {
  'budget-node-edit': 'budget',
  'budget-group-form': 'budget',
  'overrun-policy': 'budget',
  'budget-impact-preview': 'budget',
  'member-quota-config': 'budget',
  'credential-form': 'org',
  'sync-config': 'org',
  'member-form': 'org',
  'member-invite': 'org',
  'member-import': 'org',
  'dept-form': 'org',
  'role-form': 'org',
  'role-add-member': 'org',
  'import-preview': 'org',
  'pick-dept': 'org',
  'member-search': 'org',
  'pick-members': 'org',
  'key-create': 'keys',
  'key-edit': 'keys',
  'key-rotate-confirm': 'keys',
  'key-reveal': 'keys',
  'approval-submit': 'keys',
  'approval-review': 'keys',
  'provider-key-form': 'keys',
  'model-create': 'models',
  'model-edit': 'models',
  'whitelist-config': 'models',
  'model-picker': 'shared',
  'quota-check': 'shared',
  'reject-reason': 'shared',
  'permission-picker': 'shared',
}

export function getWorkflowMeta(id: WorkflowId): WorkflowMeta {
  const meta = WORKFLOW_META[id]
  if (!meta) {
    throw new Error(`Unknown workflow: ${id}`)
  }
  return meta
}

export function getWorkflowDomain(id: WorkflowId): 'budget' | 'org' | 'keys' | 'models' | 'shared' {
  return WORKFLOW_DOMAIN[id]
}
