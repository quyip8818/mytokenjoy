import { CredentialForm } from '@/components/org/credential-form'
import { SyncConfigPanel } from '@/components/org/sync-config'
import { defineDelegateWorkflow } from '../define-delegate-workflow'
import { defineWorkflow } from '../types'
import { MemberFormWorkflow, MemberInviteWorkflow } from '../workflows/member-form'
import { MemberImportWorkflow } from '../workflows/member-import'
import { DeptFormWorkflow } from '../workflows/dept-form'
import { RoleFormWorkflow } from '../workflows/role-form'
import { RoleAddMemberWorkflow } from '../workflows/role-add-member'
import { ImportPreviewWorkflow } from '../workflows/import-preview'
import { PickDeptWorkflow } from '../workflows/pick-dept'
import { MemberSearchWorkflow } from '../workflows/member-search'
import { PickMembersWorkflow } from '../workflows/pick-members'

export const orgWorkflowDefinitions = {
  'credential-form': defineDelegateWorkflow({
    id: 'credential-form',
    title: '配置凭证',
    defaultLayer: 1,
    Child: CredentialForm,
    mapProps: (payload, { onSaved }) => ({
      connected: payload.connected ?? false,
      currentPlatform: payload.currentPlatform ?? null,
      onSaved,
    }),
  }),
  'sync-config': defineDelegateWorkflow({
    id: 'sync-config',
    title: '同步策略',
    defaultLayer: 1,
    Child: SyncConfigPanel,
    mapProps: (payload, { onSaved }) => ({
      onTriggerSync: payload.onTriggerSync ?? (() => {}),
      triggeringSync: payload.triggeringSync ?? false,
      onSaved,
    }),
  }),
  'member-form': defineWorkflow(MemberFormWorkflow, { defaultLayer: 1, title: '添加/编辑成员' }),
  'member-invite': defineWorkflow(MemberInviteWorkflow, { defaultLayer: 1, title: '邀请成员' }),
  'member-import': defineWorkflow(MemberImportWorkflow, { defaultLayer: 1, title: '批量导入' }),
  'dept-form': defineWorkflow(DeptFormWorkflow, { defaultLayer: 1, title: '添加/编辑部门' }),
  'role-form': defineWorkflow(RoleFormWorkflow, { defaultLayer: 1, title: '创建/编辑角色' }),
  'role-add-member': defineWorkflow(RoleAddMemberWorkflow, {
    defaultLayer: 1,
    title: '添加角色成员',
  }),
  'import-preview': defineWorkflow(ImportPreviewWorkflow, { defaultLayer: 2, title: '导入预览' }),
  'pick-dept': defineWorkflow(PickDeptWorkflow, { defaultLayer: 2, title: '选择部门' }),
  'member-search': defineWorkflow(MemberSearchWorkflow, { defaultLayer: 2, title: '搜索成员' }),
  'pick-members': defineWorkflow(PickMembersWorkflow, { defaultLayer: 2, title: '选择成员' }),
}
