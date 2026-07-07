import { BudgetAlertsPageShell, useBudgetAlertRulesPage } from '@/features/budget'

export default function BudgetAlertsPage() {
  return <BudgetAlertsPageShell {...useBudgetAlertRulesPage()} />
}
