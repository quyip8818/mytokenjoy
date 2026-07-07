import { BudgetPageShell, useBudgetPage } from '@/features/budget'

export default function BudgetPage() {
  return <BudgetPageShell {...useBudgetPage()} />
}
