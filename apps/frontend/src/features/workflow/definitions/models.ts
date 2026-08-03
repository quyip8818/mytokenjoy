import { defineWorkflow } from '../types'
import { ModelCreateWorkflow } from '../workflows/model-create'
import { ModelEditWorkflow } from '../workflows/model-edit'
import { WhitelistConfigWorkflow } from '../workflows/whitelist-config'
import {
  PlatformModelCreateWorkflow,
  PlatformModelEditWorkflow,
} from '../workflows/platform-model-form'

export const modelsWorkflowDefinitions = {
  'model-create': defineWorkflow(ModelCreateWorkflow, { defaultLayer: 1, title: '添加自定义模型' }),
  'model-edit': defineWorkflow(ModelEditWorkflow, { defaultLayer: 1, title: '编辑自定义模型' }),
  'whitelist-config': defineWorkflow(WhitelistConfigWorkflow, {
    defaultLayer: 1,
    title: '配置部门白名单',
  }),
  'platform-model-create': defineWorkflow(PlatformModelCreateWorkflow, {
    defaultLayer: 1,
    title: '添加平台模型',
  }),
  'platform-model-edit': defineWorkflow(PlatformModelEditWorkflow, {
    defaultLayer: 1,
    title: '编辑平台模型',
  }),
}
