// === 页面入口（route page 消费）===
export { billingKeys } from './query-keys'
export { useBillingPage, type PaymentMethod, type TopUpRecordView } from './hooks/use-billing-page'
export { BillingPageShell } from './components/billing-page-shell'

// === 自身 components 通过 self-barrel 消费 ===
export { InvoiceStatusBadge } from './components/invoice-status-badge'

// === dev 子模块（consumed by: components/layout/header-dev-backend-chrome）===
export { SimulateConsumeDialog } from './dev/components/simulate-consume-dialog'
