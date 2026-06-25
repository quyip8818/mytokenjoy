import type { WorkflowId } from './types'
import { defineWorkflow } from './types'
import { KeyFormWorkflow } from './workflows/key-form'
import { KeyRevealWorkflow } from './workflows/key-reveal'
import { KeyRotateConfirmWorkflow } from './workflows/key-rotate-confirm'
import { ApprovalSubmitWorkflow } from './workflows/approval-submit'
import { ApprovalReviewWorkflow } from './workflows/approval-review'
import { RejectReasonWorkflow } from './workflows/reject-reason'
import { QuotaCheckWorkflow } from './workflows/quota-check'
import { ModelPickerWorkflow } from './workflows/model-picker'
import { ProviderKeyFormWorkflow } from './workflows/provider-key-form'
import { MemberFormWorkflow, MemberInviteWorkflow } from './workflows/member-form'
import { MemberImportWorkflow } from './workflows/member-import'
import { ImportPreviewWorkflow } from './workflows/import-preview'
import { CredentialFormWorkflow } from './workflows/credential-form'
import { SyncConfigWorkflow } from './workflows/sync-config'
import { OverrunPolicyWorkflow } from './workflows/overrun-policy'
import { ModelCreateWorkflow } from './workflows/model-create'
import { WhitelistConfigWorkflow } from './workflows/whitelist-config'
import { BudgetNodeEditWorkflow } from './workflows/budget-node-edit'
import { BudgetImpactPreviewWorkflow } from './workflows/budget-impact-preview'
import { BudgetGroupFormWorkflow } from './workflows/budget-group-form'
import { PickDeptWorkflow } from './workflows/pick-dept'
import { DeptFormWorkflow } from './workflows/dept-form'
import { RoleFormWorkflow } from './workflows/role-form'
import { RoleAddMemberWorkflow } from './workflows/role-add-member'
import { PermissionPickerWorkflow } from './workflows/permission-picker'
import { MemberSearchWorkflow } from './workflows/member-search'
import { PickMembersWorkflow } from './workflows/pick-members'

const REGISTRY: Record<WorkflowId, ReturnType<typeof defineWorkflow>> = {
  'credential-form': defineWorkflow(CredentialFormWorkflow, {
    defaultLayer: 1,
    title: '配置凭证',
  }),
  'sync-config': defineWorkflow(SyncConfigWorkflow, { defaultLayer: 1, title: '同步策略' }),
  'member-form': defineWorkflow(MemberFormWorkflow, { defaultLayer: 1, title: '添加/编辑成员' }),
  'member-invite': defineWorkflow(MemberInviteWorkflow, { defaultLayer: 1, title: '邀请成员' }),
  'member-import': defineWorkflow(MemberImportWorkflow, { defaultLayer: 1, title: '批量导入' }),
  'dept-form': defineWorkflow(DeptFormWorkflow, { defaultLayer: 1, title: '添加/编辑部门' }),
  'budget-node-edit': defineWorkflow(BudgetNodeEditWorkflow, {
    defaultLayer: 1,
    title: '编辑节点预算',
  }),
  'budget-group-form': defineWorkflow(BudgetGroupFormWorkflow, {
    defaultLayer: 1,
    title: '新建预算组',
  }),
  'overrun-policy': defineWorkflow(OverrunPolicyWorkflow, {
    defaultLayer: 1,
    title: '全局超限策略',
  }),
  'model-create': defineWorkflow(ModelCreateWorkflow, { defaultLayer: 1, title: '添加自定义模型' }),
  'whitelist-config': defineWorkflow(WhitelistConfigWorkflow, {
    defaultLayer: 1,
    title: '配置部门白名单',
  }),
  'key-create': defineWorkflow(KeyFormWorkflow, { defaultLayer: 1, title: '创建 Key' }),
  'key-edit': defineWorkflow(KeyFormWorkflow, { defaultLayer: 1, title: '编辑 Key' }),
  'key-rotate-confirm': defineWorkflow(KeyRotateConfirmWorkflow, {
    defaultLayer: 2,
    title: '重新生成 Key',
  }),
  'key-reveal': defineWorkflow(KeyRevealWorkflow, { defaultLayer: 3, title: 'Key 已生成' }),
  'approval-submit': defineWorkflow(ApprovalSubmitWorkflow, { defaultLayer: 1, title: '发起申请' }),
  'approval-review': defineWorkflow(ApprovalReviewWorkflow, { defaultLayer: 1, title: '审批处理' }),
  'role-form': defineWorkflow(RoleFormWorkflow, { defaultLayer: 1, title: '创建/编辑角色' }),
  'role-add-member': defineWorkflow(RoleAddMemberWorkflow, {
    defaultLayer: 1,
    title: '添加角色成员',
  }),
  'provider-key-form': defineWorkflow(ProviderKeyFormWorkflow, {
    defaultLayer: 1,
    title: '添加供应商 Key',
  }),
  'pick-dept': defineWorkflow(PickDeptWorkflow, { defaultLayer: 2, title: '选择部门' }),
  'model-picker': defineWorkflow(ModelPickerWorkflow, { defaultLayer: 2, title: '选择模型' }),
  'import-preview': defineWorkflow(ImportPreviewWorkflow, { defaultLayer: 2, title: '导入预览' }),
  'quota-check': defineWorkflow(QuotaCheckWorkflow, { defaultLayer: 3, title: '额度不足' }),
  'reject-reason': defineWorkflow(RejectReasonWorkflow, { defaultLayer: 2, title: '拒绝理由' }),
  'budget-impact-preview': defineWorkflow(BudgetImpactPreviewWorkflow, {
    defaultLayer: 2,
    title: '影响范围预览',
  }),
  'permission-picker': defineWorkflow(PermissionPickerWorkflow, {
    defaultLayer: 2,
    title: '选择权限',
  }),
  'member-search': defineWorkflow(MemberSearchWorkflow, { defaultLayer: 2, title: '搜索成员' }),
  'pick-members': defineWorkflow(PickMembersWorkflow, { defaultLayer: 2, title: '选择成员' }),
}

export function getWorkflowDefinition(id: WorkflowId) {
  const def = REGISTRY[id]
  if (!def) {
    throw new Error(`Unknown workflow: ${id}`)
  }
  return def
}
