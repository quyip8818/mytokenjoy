import { defineWorkflow } from '../types'
import { ModelCreateWorkflow } from '../workflows/model-create'
import { ModelEditWorkflow } from '../workflows/model-edit'
import { WhitelistConfigWorkflow } from '../workflows/whitelist-config'
import {
  PlatformModelCreateWorkflow,
  PlatformModelEditWorkflow,
} from '../workflows/platform-model-form'
import { DiscountConfigWorkflow } from '../workflows/discount-config'

export const modelsWorkflowDefinitions = {
  'model-create': defineWorkflow(ModelCreateWorkflow, { title: '添加自定义模型' }),
  'model-edit': defineWorkflow(ModelEditWorkflow, { title: '编辑自定义模型' }),
  'whitelist-config': defineWorkflow(WhitelistConfigWorkflow, {
    title: '配置部门白名单',
  }),
  'platform-model-create': defineWorkflow(PlatformModelCreateWorkflow, {
    title: '添加平台模型',
  }),
  'platform-model-edit': defineWorkflow(PlatformModelEditWorkflow, {
    title: '编辑平台模型',
  }),
  'discount-config': defineWorkflow(DiscountConfigWorkflow, {
    title: '模型优惠配置',
  }),
}
