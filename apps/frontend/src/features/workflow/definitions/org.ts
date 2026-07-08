import { defineWorkflow } from '../types'
import { MemberSearchWorkflow } from '../workflows/member-search'

export const orgWorkflowDefinitions = {
  'member-search': defineWorkflow(MemberSearchWorkflow, { defaultLayer: 2, title: '搜索成员' }),
}
