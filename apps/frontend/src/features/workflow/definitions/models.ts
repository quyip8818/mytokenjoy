import { defineWorkflow } from '../types'
import { ModelCreateWorkflow } from '../workflows/model-create'
import { WhitelistConfigWorkflow } from '../workflows/whitelist-config'

export const modelsWorkflowDefinitions = {
  'model-create': defineWorkflow(ModelCreateWorkflow, { defaultLayer: 1, title: '添加自定义模型' }),
  'whitelist-config': defineWorkflow(WhitelistConfigWorkflow, {
    defaultLayer: 1,
    title: '配置部门白名单',
  }),
}
