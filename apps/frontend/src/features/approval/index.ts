// === 页面入口（route page 消费）===
export { approvalKeys } from './lib/query-keys'
export { useApprovalPage } from './hooks/use-approval-page'
export { ApprovalPageShell } from './components/approval-page-shell'

// === 跨 feature/layout 共享 ===
// consumed by: components/layout/sidebar
export { useApprovalPendingCountQuery } from './hooks/use-approval-pending-count-query'
