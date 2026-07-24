import { BillingPageShell, useBillingPage } from '@/features/billing'

export default function BillingPage() {
  return <BillingPageShell {...useBillingPage()} />
}
