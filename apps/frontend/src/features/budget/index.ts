// 11 exports: 8 external + 3 self-barrel (components must import via barrel)
// === 页面入口（route page 消费）===
export { budgetKeys } from './query-keys'
export { useBudgetPage } from './hooks/use-budget-page'
export { useBudgetAlertRulesPage } from './hooks/use-budget-alert-rules-page'
export { BudgetPageShell } from './components/budget-page-shell'
export { BudgetAlertsPageShell } from './components/budget-alerts-page-shell'

// === 跨 feature 共享 ===
// consumed by: dashboard (budget-hero-card)
export { getBudgetProgressClass, getBudgetProgressTone } from './lib/mappers'
// consumed by: dashboard (use-cost-dashboard-page)
export { findBudgetNode } from './lib/mappers'
// consumed by: keys (platform-key-table)
export { BudgetProgressCell } from './components/budget-progress-cell'

// === 自身 components 通过 self-barrel 消费 ===
export { useAsyncFetch } from './hooks/use-async-fetch'
export { groupProjectsByTeam, thresholdClass, type AlertRuleView } from './lib/alerts'
export { POLICY_LABELS, ALERT_PRESET_THRESHOLDS } from './lib/constants'
