// === 跨 feature/layout 共享 ===
// consumed by: components/layout/admin-layout
export { WorkflowProvider } from './workflow-context'
// consumed by: components/layout/admin-layout
export { WorkflowPanelStack } from './components/workflow-panel-stack'
// consumed by: budget/components (project-detail, project-members-section)
export { useWorkflow } from './hooks/use-workflow'
// consumed by: models/keys/budget hooks
export { useWorkflowRefresh } from './hooks/use-workflow-refresh'

// === 自身 workflows 通过 self-barrel 消费 ===
export { WorkflowPanelChrome, WorkflowPanelFooter } from './components/workflow-panel-chrome'
export { WorkflowFormLayout } from './components/workflow-form-layout'
