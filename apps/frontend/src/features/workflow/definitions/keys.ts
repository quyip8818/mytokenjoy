import { defineWorkflow } from '../types'
import { KeyFormWorkflow } from '../workflows/key-form'
import { KeyRevealWorkflow } from '../workflows/key-reveal'
import { KeyRotateConfirmWorkflow } from '../workflows/key-rotate-confirm'
import { ProviderKeyFormWorkflow } from '../workflows/provider-key-form'

export const keysWorkflowDefinitions = {
  'key-create': defineWorkflow(KeyFormWorkflow, { title: '创建 Key' }),
  'key-edit': defineWorkflow(KeyFormWorkflow, { title: '编辑 Key' }),
  'key-rotate-confirm': defineWorkflow(KeyRotateConfirmWorkflow, {
    title: '重新生成 Key',
  }),
  'key-reveal': defineWorkflow(KeyRevealWorkflow, { title: 'Key 已生成' }),
  'provider-key-form': defineWorkflow(ProviderKeyFormWorkflow, {
    title: '添加供应商 Key',
  }),
}
