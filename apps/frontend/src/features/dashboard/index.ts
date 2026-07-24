// 12 exports: 6 external + 6 self-barrel (components must import via barrel)
// === 页面入口（route page 消费）===
export { dashboardKeys } from './query-keys'
export { useCostDashboardRoutePage } from './hooks/use-cost-dashboard-route-page'
export { useUsageDashboardRoutePage } from './hooks/use-usage-dashboard-route-page'
export { CostDashboardLayoutPageShell } from './components/cost-dashboard-layout-page-shell'
export { UsageDashboardLayoutPageShell } from './components/usage-dashboard-layout-page-shell'

// === 跨 feature 共享 ===
// consumed by: workflow/approval-submit
export { MODEL_NOT_IN_DEPT_MESSAGE } from './lib/constants'

// === 自身 components 通过 self-barrel 消费 ===
export { useCostDashboardPage } from './hooks/use-cost-dashboard-page'
export {
  formatMom,
  formatTokenCount,
  type CostStatItem,
} from './lib/dashboard'
export { departmentUsagePercent } from './lib/department-usage'
export { CostDistributionChart } from './components/cost-distribution-chart'
export { DeptComparisonTable } from './components/dept-comparison-table'
export { CostTopConsumersTable } from './components/cost-top-consumers-table'
