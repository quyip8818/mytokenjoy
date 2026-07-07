import { defineWorkflow } from '../types'
import { ModelCreateWorkflow } from '../workflows/model-create'
import { ModelEditWorkflow } from '../workflows/model-edit'
import { WhitelistConfigWorkflow } from '../workflows/whitelist-config'

export const modelsWorkflowDefinitions = {
  'model-create': defineWorkflow(ModelCreateWorkflow, { defaultLayer: 1, title: '添加自定义模型' }),
  'model-edit': defineWorkflow(ModelEditWorkflow, { defaultLayer: 1, title: '编辑自定义模型' }),
  'whitelist-config': defineWorkflow(WhitelistConfigWorkflow, {
    defaultLayer: 1,
    title: '配置部门白名单',
  }),
}
